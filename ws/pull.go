package ws

import (
	"errors"
	"log"
	"math/rand"
	"net"
	"net/rpc"
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
	kcp_conn.SetWindowSize(128, 512)
	kcp_conn.SetNoDelay(1, 10, 1, 1)

	return client_handler_name, workspace_owner_public_ip, udp_conn, kcp_conn, nil
}

// TODO: Instead of writing whole data_bytes into file at once,
// Write received encrypted data in chunks, after the transfer is completed, read from encrpyted file
// & decrypt it
// We can use Cipher Block Methods to decrypt & encrpyt with AES
func storeDataIntoWorkspace(workspace_name string, res *models.GetMetaDataResponse, data_bytes []byte) error {
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

	log.Println("Workspace Name:", workspace_name)
	workspace_path, err := config.GetGetWorkspaceFilePath(workspace_name)
	if err != nil {
		log.Println("Error while Fetching Workspace Path from Config:", err)
		log.Println("Source: storeDataIntoWorkspace()")
		return err
	}
	log.Println("Workspace Path: ", workspace_path)

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

func fetchData(workspace_owner_public_ip, workspace_name, workspace_hash string, udp_conn *net.UDPConn, len_data_bytes int) ([]byte, error) {
	// Now Transfer Data using KCP ONLY, No RPC in chunks
	log.Println("Connecting Again to Workspace Owner")
	kcp_conn, err := kcp.DialWithConnAndOptions(workspace_owner_public_ip, nil, 0, 0, udp_conn)
	if err != nil {
		log.Println("Error while Dialing Workspace Owner to Get Data:", err)
		log.Println("Source: fetchData()")
		return nil, err
	}
	log.Println("Connected Successfully to Workspace Owner")

	// KCP Params for Congestion Control
	kcp_conn.SetWindowSize(128, 512)
	kcp_conn.SetNoDelay(1, 10, 1, 1)

	// Sending the Type of Session
	kpc_buff := [3]byte{'K', 'C', 'P'}
	_, err = kcp_conn.Write(kpc_buff[:])
	if err != nil {
		log.Println("Error while Writing the type of Session(KCP-RPC or KCP-Plain):", err)
		log.Println("Source: fetchData()")
		return nil, err
	}

	log.Println("Sending Workspace Name & Hash to Workspace Owner")
	// Sending Workspace Name & Hash
	_, err = kcp_conn.Write([]byte(workspace_name))
	if err != nil {
		log.Println("Error while Sending Workspace Name to Workspace Owner:", err)
		log.Println("Source: fetchData()")
		return nil, err
	}

	_, err = kcp_conn.Write([]byte(workspace_hash))
	if err != nil {
		log.Println("Error while Sending Workspace Name to Workspace Owner:", err)
		log.Println("Source: fetchData()")
		return nil, err
	}
	log.Println("Workspace Name & Hash Sent to Workspace Owner")

	CHUNK_SIZE := min(DATA_CHUNK, len_data_bytes)

	log.Println("Len Data Bytes:", len_data_bytes)
	log.Println("Len Buffer:", len_data_bytes+CHUNK_SIZE)
	data_bytes := make([]byte, len_data_bytes+CHUNK_SIZE)
	offset := 0

	log.Println("Now Reading Data from Workspace Owner ...")
	for offset < len_data_bytes {

		n, err := kcp_conn.Read(data_bytes[offset : offset+CHUNK_SIZE])
		// Check for Errors on Workspace Owner's Side
		if n < 30 {
			msg := string(data_bytes[offset : offset+n])
			if msg == "Incorrect Workspace Name/Hash" || msg == "Internal Server Error" {
				log.Println("\nError while Reading from Workspace on his/her side:", msg)
				log.Println("Source: fetchData()")
				return nil, errors.New(msg)
			}
		}

		if err != nil {
			log.Println("\nError while Reading from Workspace Owner:", err)
			log.Println("Source: fetchData()")
			return nil, err
		}
		offset += n
		utils.PrintProgressBar(offset, len_data_bytes, 100)
	}
	log.Println("\nData Transfer Completed:", offset)

	_, err = kcp_conn.Write([]byte("Data Received"))
	if err != nil {
		log.Println("Error while Sending Data Received Message:", err)
		log.Println("Source: fetchData()")
		// Not Returning Error because, we got data, we don't care if workspace owner now is offline
	}
	return data_bytes[:offset], nil
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
	defer kcp_conn.Close()

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
	res, err := rpcClientHandler.CallGetMetaData(MY_USERNAME, MY_SERVER_IP, workspace_name, encrypted_password, last_hash, client_handler_name, rpc_client)
	if err != nil {
		log.Println("Error while Calling GetMetaData:", err)
		log.Println("Source: pullWorkspace()")
		return
	}
	log.Println("Get Data Responded, now storing files into workspace")

	data_bytes, err := fetchData(workspace_owner_public_ip, workspace_name, res.NewHash, udp_conn, res.LenData)
	if err != nil {
		log.Println("Error while Fetching Data:", err)
		log.Println("Source: pullWorkspace()")
		return
	}

	// Store Data into workspace
	err = storeDataIntoWorkspace(workspace_name, res, data_bytes)
	if err != nil {
		log.Println("Error while Storing Requested Data into Workspace:", err)
		log.Println("Source: pullWorkspace()")
		return
	}

	// Update tmp/userConfig.json
	err = config.UpdateLastHashInGetWorkspaceFolderToUserConfig(workspace_name, res.NewHash)
	if err != nil {
		log.Println("Error while Registering New GetWorkspace:", err)
		log.Println("Source: pullWorkspace()")
		return
	}
	log.Println("Pull Done")
}
