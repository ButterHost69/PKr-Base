package ws

import (
	"bufio"
	"errors"
	"fmt"
	"log"
	"math/rand"
	"net"
	"net/rpc"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/ButterHost69/PKr-Base/config"
	"github.com/ButterHost69/PKr-Base/dialer"
	"github.com/ButterHost69/PKr-Base/encrypt"
	"github.com/ButterHost69/PKr-Base/filetracker"
	"github.com/ButterHost69/PKr-Base/handler"
	"github.com/ButterHost69/PKr-Base/models"
	"github.com/ButterHost69/PKr-Base/utils"
	"github.com/ButterHost69/kcp-go"
	"github.com/gorilla/websocket"
)

const DATA_CHUNK = handler.DATA_CHUNK

var MY_USERNAME string
var MY_SERVER_IP string
var MY_SERVER_ALIAS string

func connectToAnotherUser(workspace_owner_username string, conn *websocket.Conn) (string, string, *net.UDPConn, *kcp.UDPSession, error) {
	local_port := rand.Intn(16384) + 16384
	log.Println("My Local Port:", local_port)

	// Get My Public IP
	myPublicIP, err := dialer.GetMyPublicIP(local_port)
	if err != nil {
		log.Println("Error while Getting my Public IP:", err)
		log.Println("Source: connectToAnotherUser()")
		return "", "", nil, nil, err
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
		return "", "", nil, nil, err

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
		return "", "", nil, nil, errors.New("workspace owner isn't responding")
	}

	if req_punch_from_receiver_response.Error != "" {
		log.Println("Error Received from Server's WS:", err)
		log.Println("Description: Could Not Request Punch From Receiver")
		log.Println("Source: connectToAnotherUser()")
		return "", "", nil, nil, err
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
		return "", "", nil, nil, err
	}
	log.Println("Starting UDP NAT Hole Punching ...")

	workspace_owner_public_ip := req_punch_from_receiver_response.WorkspaceOwnerPublicIP + ":" + req_punch_from_receiver_response.WorkspaceOwnerPublicPort
	client_handler_name, err := dialer.WorkspaceListenerUdpNatHolePunching(udp_conn, workspace_owner_public_ip)
	if err != nil {
		log.Println("Error while Punching to Remote Addr:", err)
		log.Println("Source: connectToAnotherUser()")
		return "", "", nil, nil, err

	}
	log.Println("UDP NAT Hole Punching Completed Successfully")

	// Creating KCP-Conn, KCP = Reliable UDP
	kcp_conn, err := kcp.DialWithConnAndOptions(workspace_owner_public_ip, nil, 0, 0, udp_conn)
	if err != nil {
		log.Println("Error while Dialing KCP Connection to Remote Addr:", err)
		log.Println("Source: connectToAnotherUser()")
		return "", "", nil, nil, err
	}

	// KCP Params for Congestion Control
	kcp_conn.SetWindowSize(128, 1024)
	kcp_conn.SetNoDelay(1, 10, 2, 1)
	kcp_conn.SetACKNoDelay(true)
	kcp_conn.SetDSCP(46)

	return client_handler_name, workspace_owner_public_ip, udp_conn, kcp_conn, nil
}

func fetchAndStoreDataIntoWorkspace(workspace_owner_public_ip, workspace_name string, udp_conn *net.UDPConn, res models.GetMetaDataResponse) error {
	// Decrypting AES Key
	key, err := encrypt.DecryptData(string(res.KeyBytes))
	if err != nil {
		log.Println("Error while Decrypting Key:", err)
		log.Println("Source: fetchAndStoreDataIntoWorkspace()")
		return err
	}

	// Decrypting AES IV
	iv, err := encrypt.DecryptData(string(res.IVBytes))
	if err != nil {
		log.Println("Error while Decrypting 'IV':", err)
		log.Println("Source: fetchAndStoreDataIntoWorkspace()")
		return err
	}

	workspace_path, err := config.GetGetWorkspaceFilePath(workspace_name)
	if err != nil {
		log.Println("Error while Fetching Workspace Path from Config:", err)
		log.Println("Source: fetchAndStoreDataIntoWorkspace()")
		return err
	}
	log.Println("Workspace Path: ", workspace_path)

	zip_file_path := filepath.Join(workspace_path, ".PKr", res.RequestPushRange+".zip")
	// Create Zip File
	zip_file_obj, err := os.Create(zip_file_path)
	if err != nil {
		log.Println("Failed to Open & Create Zipped File:", err)
		log.Println("Source: fetchAndStoreDataIntoWorkspace()")
		return err
	}

	// To Write Decrypted Data in Chunks
	writer := bufio.NewWriter(zip_file_obj)

	// Now Transfer Data using KCP ONLY, No RPC in chunks
	log.Println("Connecting Again to Workspace Owner")
	kcp_conn, err := kcp.DialWithConnAndOptions(workspace_owner_public_ip, nil, 0, 0, udp_conn)
	if err != nil {
		log.Println("Error while Dialing Workspace Owner to Get Data:", err)
		log.Println("Source: fetchAndStoreDataIntoWorkspace()")
		return err
	}
	defer kcp_conn.Close()
	log.Println("Connected Successfully to Workspace Owner")

	// KCP Params for Congestion Control
	kcp_conn.SetWindowSize(128, 1024)
	kcp_conn.SetNoDelay(1, 10, 2, 1)
	kcp_conn.SetACKNoDelay(true)
	kcp_conn.SetDSCP(46)

	// Sending the Type of Session
	kpc_buff := [3]byte{'K', 'C', 'P'}
	_, err = kcp_conn.Write(kpc_buff[:])
	if err != nil {
		log.Println("Error while Writing the type of Session(KCP-RPC or KCP-Plain):", err)
		log.Println("Source: fetchAndStoreDataIntoWorkspace()")
		return err
	}

	log.Println("Sending Workspace Name & Hash to Workspace Owner")
	// Sending Workspace Name & Hash
	_, err = kcp_conn.Write([]byte(workspace_name))
	if err != nil {
		log.Println("Error while Sending Workspace Name to Workspace Owner:", err)
		log.Println("Source: fetchAndStoreDataIntoWorkspace()")
		return err
	}

	_, err = kcp_conn.Write([]byte(res.RequestPushRange))
	if err != nil {
		log.Println("Error while Sending Workspace Name to Workspace Owner:", err)
		log.Println("Source: fetchAndStoreDataIntoWorkspace()")
		return err
	}
	log.Println("Workspace Name & Push Num Sent to Workspace Owner")

	_, err = kcp_conn.Write([]byte("Pull"))
	if err != nil {
		log.Println("Error while Sending 'Pull' to Workspace Owner:", err)
		log.Println("Source: fetchAndStoreDataIntoWorkspace()")
		return err
	}

	buffer := make([]byte, DATA_CHUNK)

	log.Println("Len Data Bytes:", res.LenData)
	offset := 0

	log.Println("Now Reading Data from Workspace Owner ...")
	for offset < res.LenData {
		n, err := kcp_conn.Read(buffer)
		if err != nil {
			log.Println("\nError while Reading from Workspace Owner:", err)
			log.Println("Source: fetchAndStoreDataIntoWorkspace()")
			return err
		}

		// Check for Errors on Workspace Owner's Side
		if n < 30 {
			msg := string(buffer[:n])
			if msg == "Incorrect Workspace Name/Hash" || msg == "Internal Server Error" {
				log.Println("\nError while Reading from Workspace on his/her side:", msg)
				log.Println("Source: fetchAndStoreDataIntoWorkspace()")
				return errors.New(msg)
			}
		}

		// Decrypt Data
		decrypted_data, err := encrypt.EncryptDecryptChunk(buffer[:n], []byte(key), []byte(iv))
		if err != nil {
			log.Println("Error while Decrypting Chunk:", err)
			log.Println("Source: fetchAndStoreDataIntoWorkspace()")
			return err
		}

		// Store data in chunks using 'writer'
		_, err = writer.Write(decrypted_data)
		if err != nil {
			log.Println("Error while Writing Decrypted Data in Chunks:", err)
			log.Println("Source: fetchAndStoreDataIntoWorkspace()")
			return err
		}

		// Flush buffer to disk after 'FLUSH_AFTER_EVERY_X_MB'
		if offset%handler.FLUSH_AFTER_EVERY_X_MB == 0 {
			err = writer.Flush()
			if err != nil {
				fmt.Println("Error flushing 'writer' buffer:", err)
				fmt.Println("Soure: fetchAndStoreDataIntoWorkspace()")
				return err
			}
		}

		offset += n
		utils.PrintProgressBar(offset, res.LenData, 100)
	}
	log.Println("\nData Transfer Completed ...")

	// Flush buffer to disk at end
	err = writer.Flush()
	if err != nil {
		log.Println("Error flushing 'writer' buffer:", err)
		log.Println("Soure: fetchAndStoreDataIntoWorkspace()")
		return err
	}
	zip_file_obj.Close()

	_, err = kcp_conn.Write([]byte("Data Received"))
	if err != nil {
		log.Println("Error while Sending Data Received Message:", err)
		log.Println("Source: fetchAndStoreDataIntoWorkspace()")
		// Not Returning Error because, we got data, we don't care if workspace owner now is offline or not responding
	}

	unzip_dest := filepath.Join(workspace_path, ".PKr", "Contents", res.RequestPushRange)
	err = os.MkdirAll(unzip_dest, 0666)
	if err != nil {
		log.Println("Error while Creating .PKr/Hash Directory:", err)
		log.Println("Source: fetchAndStoreDataIntoWorkspace()")
		return err
	}

	// Unzip Content
	if err = filetracker.UnzipData(zip_file_path, unzip_dest); err != nil {
		log.Println("Error while Unzipping Data into Workspace:", err)
		log.Println("Source: fetchAndStoreDataIntoWorkspace()")
		return err
	}

	err = filetracker.UpdateFilesFromWorkspace(workspace_path, unzip_dest, res.Updates)
	if err != nil {
		log.Println("Error while Updating Files From Workspace:", err)
		log.Println("Source: fetchAndStoreDataIntoWorkspace()")
		return err
	}

	// Remove Zip File After Unzipping it
	err = os.Remove(zip_file_path)
	if err != nil {
		log.Println("Error while Removing the Zip File After Use:", err)
		log.Println("Source: fetchAndStoreDataIntoWorkspace()")
		// No need to return err, else it won't register in configs
	}

	// Remove files from the place where changes were temporarily un-zipped
	err = os.RemoveAll(unzip_dest)
	if err != nil {
		log.Println("Error while Removing the Files from '.PKr/Hash/':", err)
		log.Println("Source: fetchAndStoreDataIntoWorkspace()")
		return err
	}
	return nil
}

func PullWorkspace(workspace_owner_username, workspace_name string, conn *websocket.Conn) {
	log.Println("Pulling Workspace:", workspace_name)
	log.Println("Workspace Owner:", workspace_owner_username)

	client_handler_name, workspace_owner_public_ip, udp_conn, kcp_conn, err := connectToAnotherUser(workspace_owner_username, conn)
	if err != nil {
		log.Println("Error while Connecting to Another User:", err)
		log.Println("Source: pullWorkspace()")
		return
	}

	rpc_buff := [3]byte{'R', 'P', 'C'}
	_, err = kcp_conn.Write(rpc_buff[:])
	if err != nil {
		log.Println("Error while Writing the type of Session(KCP-RPC or KCP-Plain):", err)
		log.Println("Source: pullWorkspace()")
		return
	}

	user_config, err := config.ReadFromUserConfigFile()
	if err != nil {
		log.Println("Error while Reading User Config File:", err)
		log.Println("Source: pullWorkspace()")
		return
	}

	var workspace_password string
	var last_push_num int
	for _, server := range user_config.ServerLists {
		for _, workspace := range server.GetWorkspaces {
			if workspace.WorkspaceName == workspace_name && workspace.WorkspaceOwnerName == workspace_owner_username {
				workspace_password = workspace.WorkspacePassword
				last_push_num = workspace.LastPushNum
			}
		}
	}

	// Creating RPC Client
	rpc_client := rpc.NewClient(kcp_conn)
	rpcClientHandler := dialer.ClientCallHandler{}

	workspace_path, err := config.GetGetWorkspaceFilePath(workspace_name)
	if err != nil {
		log.Println("Error while Fetching Workspace Path from Config:", err)
		log.Println("Source: Pull()")
		return
	}

	// Get Public Key of Workspace Owner
	log.Println("Fetching Public Key of Workspace Owner .PKr/Keys")
	public_key, err := os.ReadFile(filepath.Join(workspace_path, ".PKr", "Keys", workspace_owner_username+".pem"))
	if err != nil {
		log.Println("Error while Getting Public Key of Workspace Owner from .PKr/Keys:", err)
		log.Println("Source: pullWorkspace()")
		return
	}

	// Encrypting Workspace Password with Public Key
	encrypted_password, err := encrypt.EncryptData(workspace_password, string(public_key))
	if err != nil {
		log.Println("Error while Encrypting Workspace Password via Public Key:", err)
		log.Println("Source: pullWorkspace()")
		return
	}

	log.Println("Calling GetMetaData ...")
	// Calling GetMetaData
	res, err := rpcClientHandler.CallGetMetaData(MY_USERNAME, MY_SERVER_IP, workspace_name, encrypted_password, client_handler_name, last_push_num, rpc_client)
	if err != nil {
		log.Println("Error while Calling GetMetaData:", err)
		log.Println("Source: pullWorkspace()")
		return
	}
	log.Println("Get Data Responded, now storing files into workspace")
	log.Println(res.LastPushNum)
	log.Println(res.LastPushDesc)
	log.Println(res.LenData)
	log.Println(res.RequestPushRange)
	log.Println(res.Updates)

	kcp_conn.Close()
	rpc_client.Close()

	err = fetchAndStoreDataIntoWorkspace(workspace_owner_public_ip, workspace_name, udp_conn, *res)
	if err != nil {
		log.Println("Error while Fetching Data & Storing it in Workspace:", err)
		log.Println("Source: pullWorkspace()")
		return
	}

	// Update tmp/userConfig.json
	err = config.UpdateLastPushNumInGetWorkspaceFolderToUserConfig(workspace_name, res.LastPushNum)
	if err != nil {
		log.Println("Error while Registering New GetWorkspace:", err)
		log.Println("Source: pullWorkspace()")
		return
	}
	log.Println("Pull Done")
}
