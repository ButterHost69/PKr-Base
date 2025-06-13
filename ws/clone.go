package ws

import (
	"encoding/base64"
	"errors"
	"fmt"
	"log"
	"math/rand"
	"net"
	"net/rpc"
	"os"
	"strings"
	"time"

	"github.com/ButterHost69/PKr-Base/config"
	"github.com/ButterHost69/PKr-Base/dialer"
	"github.com/ButterHost69/PKr-Base/encrypt"
	"github.com/ButterHost69/PKr-Base/filetracker"
	"github.com/ButterHost69/PKr-Base/models"
	"github.com/ButterHost69/kcp-go"
	"github.com/gorilla/websocket"
)

var MY_USERNAME string
var MY_SERVER_IP string
var MY_SERVER_ALIAS string

func connectToAnotherUser(workspace_owner_username string, conn *websocket.Conn) (string, *kcp.UDPSession, error) {
	local_port := rand.Intn(16384) + 16384
	log.Println("My Local Port:", local_port)

	// Get My Public IP
	myPublicIP, err := dialer.GetMyPublicIP(local_port)
	if err != nil {
		log.Println("Error while Getting my Public IP:", err)
		log.Println("Source: connectToAnotherUser()")
		return "", nil, err
	}
	log.Println("My Public IP Addr:", myPublicIP)

	myPublicIPSplit := strings.Split(myPublicIP, ":")
	myPublicIPOnly := myPublicIPSplit[0]
	myPublicPortOnly := myPublicIPSplit[1]

	var req_punch_from_receiver_request models.RequestPunchFromReceiverRequest
	req_punch_from_receiver_request.WorkspaceOwnerUsername = workspace_owner_username
	req_punch_from_receiver_request.ListenerUsername = MY_USERNAME
	req_punch_from_receiver_request.ListenerPublicIP = myPublicIPOnly
	req_punch_from_receiver_request.ListenerPublicPort = myPublicPortOnly

	log.Println("Calling RequestPunchFromReceiverRequest")
	err = conn.WriteJSON(models.WSMessage{
		MessageType: "RequestPunchFromReceiverRequest",
		Message:     req_punch_from_receiver_request,
	})
	if err != nil {
		log.Println("Error while Sending RequestPunchFromReceiverRequest to WS Server:", err)
		log.Println("Source: connectToAnotherUser()")
		return "", nil, err

	}

	var req_punch_from_receiver_response models.RequestPunchFromReceiverResponse
	var ok, invalid_flag bool
	count := 0

	for {
		time.Sleep(5 * time.Second)
		RequestPunchFromReceiverResponseMap.Lock()
		req_punch_from_receiver_response, ok = RequestPunchFromReceiverResponseMap.Map[workspace_owner_username]
		RequestPunchFromReceiverResponseMap.Unlock()
		if ok {
			RequestPunchFromReceiverResponseMap.Lock()
			delete(RequestPunchFromReceiverResponseMap.Map, workspace_owner_username)
			RequestPunchFromReceiverResponseMap.Unlock()
			break
		}

		if count == 6 {
			invalid_flag = true
			break
		}
		count += 1
	}

	if invalid_flag {
		log.Println("Error: Workspace Owner isn't Responding\nSource: connectToAnotherUser()")
		return "", nil, errors.New("workspace owner isn't responding")
	}

	if req_punch_from_receiver_response.Error != "" {
		log.Println("Error Received from Server's WS:", err)
		log.Println("Description: Could Not Request Punch From Receiver")
		log.Println("Source: connectToAnotherUser()")
		return "", nil, err
	}
	log.Println("Called RequestPunchFromReceiverRequest ...")

	// Creating UDP Conn to Perform UDP NAT Hole Punching
	udp_conn, err := net.ListenUDP("udp", &net.UDPAddr{
		Port: local_port,
		IP:   net.IPv4zero, // or nil
	})
	if err != nil {
		log.Printf("Error while Listening to %d: %v\n", local_port, err)
		log.Println("Source: connectToAnotherUser()")
		return "", nil, err
	}
	log.Println("Starting UDP NAT Hole Punching ...")

	workspace_owner_ip := req_punch_from_receiver_response.WorkspaceOwnerPublicIP + ":" + req_punch_from_receiver_response.WorkspaceOwnerPublicPort
	client_handler_name, err := dialer.WorkspaceListenerUdpNatHolePunching(udp_conn, workspace_owner_ip)
	if err != nil {
		log.Println("Error while Punching to Remote Addr:", err)
		log.Println("Source: connectToAnotherUser()")
		return "", nil, err

	}
	log.Println("UDP NAT Hole Punching Completed Successfully")

	// Creating KCP-Conn, KCP = Reliable UDP
	kcp_conn, err := kcp.DialWithConnAndOptions(workspace_owner_ip, nil, 0, 0, udp_conn)
	if err != nil {
		log.Println("Error while Dialing KCP Connection to Remote Addr:", err)
		log.Println("Source: connectToAnotherUser()")
		return "", nil, err
	}

	// KCP Params for Congestion Control
	kcp_conn.SetWindowSize(128, 512)
	kcp_conn.SetNoDelay(1, 20, 0, 1)
	kcp_conn.SetACKNoDelay(false)

	return client_handler_name, kcp_conn, nil
}

func storeDataIntoWorkspace(res *models.GetDataResponse, workspace_name string) error {
	data_bytes := res.Data
	key_bytes := res.KeyBytes
	iv_bytes := res.IVBytes

	decrypted_key, err := encrypt.DecryptData(string(key_bytes))
	if err != nil {
		log.Println("Error while Decrypting Key:", err)
		log.Println("Source: storeDataIntoWorkspace()")
		return err
	}

	decrypted_iv, err := encrypt.DecryptData(string(iv_bytes))
	if err != nil {
		log.Println("Error while Decrypting 'IV':", err)
		log.Println("Source: storeDataIntoWorkspace()")
		return err
	}

	data, err := encrypt.AESDecrypt(data_bytes, decrypted_key, decrypted_iv)
	if err != nil {
		log.Println("Error while Decrypting Data:", err)
		log.Println("Source: storeDataIntoWorkspace()")
		return err
	}

	user_config, err := config.ReadFromUserConfigFile()
	if err != nil {
		log.Println("Error while Reading User Config File:", err)
		log.Println("Source: storeDataIntoWorkspace()")
		return err
	}

	var workspace_path string
	workspace_path = "."
	for _, server := range user_config.ServerLists {
		for _, workspace := range server.GetWorkspaces {
			if workspace.WorkspaceName == workspace_name {
				workspace_path = workspace.WorkspacePath
			}
		}
	}
	log.Println("Workspace Path:", workspace_path)

	zip_file_path := workspace_path + "\\.PKr\\" + res.NewHash + ".zip"
	if err = filetracker.SaveDataToFile(data, zip_file_path); err != nil {
		log.Println("Error while Saving Data into '.PKr/abc.zip':", err)
		log.Println("Source: storeDataIntoWorkspace()")
		return err
	}

	if err = filetracker.CleanFilesFromWorkspace(workspace_path); err != nil {
		log.Println("Error while Cleaning Workspace :", err)
		log.Println("Source: storeDataIntoWorkspace()")
		return err
	}

	// Unzip Content
	if err = filetracker.UnzipData(zip_file_path, workspace_path+"\\"); err != nil {
		log.Println("Error while Unzipping Data into Workspace:", err)
		log.Println("Source: storeDataIntoWorkspace()")
		return err
	}
	return nil
}

func cloneWorkspace(workspace_owner_username, workspace_name string, conn *websocket.Conn) {
	client_handler_name, kcp_conn, err := connectToAnotherUser(workspace_owner_username, conn)
	if err != nil {
		log.Println("Error while Connecting to Another User:", err)
		log.Println("Source: cloneWorkspace()")
		return
	}
	defer kcp_conn.Close()

	user_config, err := config.ReadFromUserConfigFile()
	if err != nil {
		log.Println("Error while Reading User Config File:", err)
		log.Println("Source: cloneWorkspace()")
		return
	}

	var workspace_password, last_hash string
	for _, server := range user_config.ServerLists {
		for _, workspace := range server.GetWorkspaces {
			if workspace.WorkspaceName == workspace_name {
				workspace_password = workspace.WorkspacePassword
				last_hash = workspace.LastHash
			}
		}
	}

	// Creating RPC Client
	rpc_client := rpc.NewClient(kcp_conn)
	defer rpc_client.Close()

	rpcClientHandler := dialer.ClientCallHandler{}

	log.Println("Calling Get Public Key")
	// Get Public Key of Workspace Owner
	public_key, err := rpcClientHandler.CallGetPublicKey(client_handler_name, rpc_client)
	if err != nil {
		log.Println("Error while Calling GetPublicKey:", err)
		log.Println("Source: cloneWorkspace()")
		return
	}

	// Encrypting Workspace Password with Public Key
	encrypted_password, err := encrypt.EncryptData(workspace_password, string(public_key))
	if err != nil {
		log.Println("Error while Encrypting Workspace Password via Public Key:", err)
		log.Println("Source: cloneWorkspace()")
		return
	}

	// Reading my Public Key
	my_public_key, err := os.ReadFile("./tmp/mykeys/publickey.pem")
	if err != nil {
		log.Println("Error while Reading Public Key:", err)
		log.Println("Source: cloneWorkspace()")
		return
	}
	base64_public_key := []byte(base64.StdEncoding.EncodeToString(my_public_key))

	log.Println("Calling InitWorkspaceConnection")
	// Requesting InitWorkspaceConnection
	err = rpcClientHandler.CallInitNewWorkSpaceConnection(workspace_name, MY_USERNAME, MY_SERVER_IP, encrypted_password, base64_public_key, client_handler_name, rpc_client)
	if err != nil {
		log.Println("Error while Calling Init New Workspace Connection:", err)
		log.Println("Source: cloneWorkspace()")
		return
	}

	// Create .PKr folder to store zipped data
	currDir, err := os.Getwd()
	if err != nil {
		log.Println("Error while Getting Current Directory:", err)
		log.Println("Source: cloneWorkspace()")
		return
	}
	err = os.MkdirAll(currDir+"\\.PKr\\", 0777)
	if err != nil {
		log.Println("Error while using MkdirAll for '.PKr' folder:", err)
		log.Println("Source: cloneWorkspace()")
		return
	}

	log.Println("Calling GetData ...")
	// Calling GetData
	res, err := rpcClientHandler.CallGetData(MY_USERNAME, MY_SERVER_IP, workspace_name, workspace_password, last_hash, client_handler_name, rpc_client)
	if err != nil {
		log.Println("Error while Calling GetData:", err)
		log.Println("Source: cloneWorkspace()")
		return
	}

	log.Println("Get Data Responded, now storing files into workspace")
	// Store Data into workspace
	err = storeDataIntoWorkspace(res, workspace_name)
	if err != nil {
		log.Println("Error while Storing Requested Data into Workspace:", err)
		log.Println("Source: cloneWorkspace()")
		return
	}

	// Update tmp/userConfig.json
	err = config.UpdateLastHashInGetWorkspaceFolderToUserConfig(workspace_name, res.NewHash)
	if err != nil {
		fmt.Println("Error while Registering New GetWorkspace:", err)
		fmt.Println("Source: cloneWorkspace()")
		return
	}
	fmt.Println("Clone Done")
}
