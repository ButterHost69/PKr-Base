package handler

import (
	"encoding/base64"
	"errors"
	"log"

	"os"
	"strings"

	"github.com/ButterHost69/PKr-Base/config"
	"github.com/ButterHost69/PKr-Base/encrypt"
	"github.com/ButterHost69/PKr-Base/filetracker"
	"github.com/ButterHost69/PKr-Base/models"
)

var (
	ErrIncorrectPassword             = errors.New("incorrect password")
	ErrServerNotFound                = errors.New("server not found in config")
	ErrInternalSeverError            = errors.New("internal server error")
	ErrUserAlreadyHasLatestWorkspace = errors.New("you already've latest version of workspace")
)

type ClientHandler struct{}

func (h *ClientHandler) GetPublicKey(req models.PublicKeyRequest, res *models.PublicKeyResponse) error {
	log.Println("Get Public Key Called ...")

	keyData, err := config.ReadPublicKey()
	if err != nil {
		log.Println("Error while reading My Public Key from config:", err)
		log.Println("Source: GetPublicKey()")
		return ErrInternalSeverError
	}

	res.PublicKey = []byte(keyData)
	return nil
}

func (h *ClientHandler) InitNewWorkSpaceConnection(req models.InitWorkspaceConnectionRequest, res *models.InitWorkspaceConnectionResponse) error {
	// 1. Decrypt password [X]
	// 2. Authenticate Request [X]
	// 3. Add the New Connection to the .PKr Config File [X]
	// 4. Store the Public Key [X]

	log.Println("Init New Work Space Connection Called ...")

	password, err := encrypt.DecryptData(req.WorkspacePassword)
	if err != nil {
		log.Println("Failed to Decrypt the Workspace Pass Received from Listener:", err)
		log.Println("Source: InitNewWorkSpaceConnection()")
		return ErrInternalSeverError
	}
	log.Println("Decrypted Data ...\nAuthenticating ...")

	// Authenticates Workspace Name and Password and Get the Workspace File Path
	file_path, err := config.AuthenticateWorkspaceInfo(req.WorkspaceName, password)
	if err != nil {
		if errors.Is(err, ErrIncorrectPassword) {
			log.Println("Error: Incorrect Credentials for Workspace")
			log.Println("Source: InitNewWorkSpaceConnection()")
			return ErrIncorrectPassword
		}
		log.Println("Failed to Authenticate Password of Listener:", err)
		log.Println("Source: InitNewWorkSpaceConnection()")
		return ErrIncorrectPassword
	}

	log.Println("Auth Successfull ...\nGetting Server Details Using Server IP")
	server, err := config.GetServerDetailsUsingServerIP(req.ServerIP)
	if err != nil {
		if errors.Is(err, ErrServerNotFound) {
			log.Println("Error: Server Not Found")
			log.Println("Source: InitNewWorkSpaceConnection()")
			return ErrServerNotFound
		}
		log.Println("Failed to Get Server Details Using Server IP:", err)
		log.Println("Source: InitNewWorkSpaceConnection()")
		return ErrInternalSeverError
	}
	log.Println("Getting Server Details Using Server IP Successfull ...\n Decoding Public Keys ...")

	var connection config.Connection
	connection.Username = req.MyUsername
	connection.ServerAlias = server.ServerAlias

	publicKey, err := base64.StdEncoding.DecodeString(string(req.MyPublicKey))
	if err != nil {
		log.Println("Failed to Decode Public Key from base64:", err)
		log.Println("Source: InitNewWorkSpaceConnection()")
		return ErrInternalSeverError
	}

	// Save Public Key
	log.Println("Storing Public Keys ...")
	keysPath, err := config.StorePublicKeys(file_path+"\\.PKr\\keys\\", string(publicKey))
	if err != nil {
		log.Println("Failed to Store Public Keys at '.PKr\\keys':", err)
		log.Println("Source: InitNewWorkSpaceConnection()")
		return ErrInternalSeverError
	}
	log.Println("Adding New Conn in Pkr Config File")

	// Store the New Connection in the .PKr Config file
	connection.PublicKeyPath = keysPath
	if err := config.AddConnectionToPKRConfigFile(file_path+"\\.PKr\\workspaceConfig.json", connection); err != nil {
		log.Println("Failed to Add Connection to .PKr Config File:", err)
		log.Println("Source: InitNewWorkSpaceConnection()")
		return ErrInternalSeverError
	}

	log.Println("Added New Conn in Pkr Config file")
	log.Println("Init New Workspace Conn Successful")
	return nil
}

func (h *ClientHandler) GetMetaData(req models.GetMetaDataRequest, res *models.GetMetaDataResponse) error {
	password, err := encrypt.DecryptData(req.WorkspacePassword)
	if err != nil {
		log.Println("Failed to Decrypt the Workspace Pass Received from Listener:", err)
		log.Println("Source: GetMetaData()")
		return ErrInternalSeverError
	}
	log.Println("Decrypted Data ...\nAuthenticating ...")

	// Authenticates Workspace Name and Password and Get the Workspace File Path
	_, err = config.AuthenticateWorkspaceInfo(req.WorkspaceName, password)
	if err != nil {
		if errors.Is(err, ErrIncorrectPassword) {
			log.Println("Error: Incorrect Credentials for Workspace")
			log.Println("Source: GetMetaData()")
			return ErrIncorrectPassword
		}
		log.Println("Failed to Authenticate Password of Listener:", err)
		log.Println("Source: GetMetaData()")
		return ErrIncorrectPassword
	}

	log.Printf("Data Requested For Workspace: %s\n", req.WorkspaceName)
	workspace_path, err := config.GetSendWorkspaceFilePath(req.WorkspaceName)
	if err != nil {
		log.Println("Failed to Get Workspace Path from Config:", err)
		log.Println("Source: GetMetaData()")
		return ErrInternalSeverError
	}

	zipped_file_name, err := filetracker.ZipData(workspace_path)
	if err != nil {
		log.Println("Failed to Hash & Zip Data:", err)
		log.Println("Source: GetMetaData()")
		return ErrInternalSeverError
	}

	zipped_hash := strings.Split(zipped_file_name, ".")[0]
	log.Println("Data Service Hash File Name: " + zipped_file_name)

	if zipped_hash == req.LastHash {
		log.Println("User has the Latest Workspace, according to Last Hash")
		log.Println("No need to transfer data")
		return ErrUserAlreadyHasLatestWorkspace
	}

	log.Println("Generating Keys ...")
	key, err := encrypt.AESGenerakeKey(16)
	if err != nil {
		log.Println("Failed to Generate AES Keys:", err)
		log.Println("Source: GetMetaData()")
		return ErrInternalSeverError
	}

	iv, err := encrypt.AESGenerateIV()
	if err != nil {
		log.Println("Failed to Generate IV Keys:", err)
		log.Println("Source: GetMetaData()")
		return ErrInternalSeverError
	}

	log.Println("Zipping Workspace ...")

	zipped_filepath := workspace_path + "\\.PKr\\" + zipped_file_name
	destination_filepath := strings.Replace(zipped_filepath, ".zip", ".enc", 1)
	if err := encrypt.AESEncrypt(zipped_filepath, destination_filepath, key, iv); err != nil {
		log.Println("Failed to Encrypt Data using AES:", err)
		log.Println("Source: GetMetaData()")
		return ErrInternalSeverError
	}

	log.Println("Fetching Public Key of Listener")
	public_key_path, err := config.GetConnectionsPublicKeyUsingUsername(workspace_path, req.Username)
	if err != nil {
		log.Println("Failed to Get Public Key's Path of Listener Using Username:", err)
		log.Println("Source: GetMetaData()")
		return ErrInternalSeverError
	}

	public_key, err := os.ReadFile(public_key_path)
	if err != nil {
		log.Println("Failed to Read Public Key of Listener:", err)
		log.Println("Source: GetMetaData()")
		return ErrInternalSeverError
	}

	log.Println("Encrypting Key")
	encrypt_key, err := encrypt.EncryptData(string(key), string(public_key))
	if err != nil {
		log.Println("Failed to Encrypt AES Keys using Listener's Public Key:", err)
		log.Println("Source: GetMetaData()")
		return ErrInternalSeverError
	}

	encrypt_iv, err := encrypt.EncryptData(string(iv), string(public_key))
	if err != nil {
		log.Println("Failed to Encrypt IV Keys using Listener's Public Key:", err)
		log.Println("Source: GetMetaData()")
		return ErrInternalSeverError
	}

	file_info, err := os.Stat(destination_filepath)
	if err != nil {
		log.Println("Failed to Get FileInfo of Encrypted Zip File:", err)
		log.Println("Source: GetMetaData()")
		return ErrInternalSeverError
	}

	err = config.AddNewPushToConfig(req.WorkspaceName, zipped_hash)
	if err != nil {
		log.Println("Failed to Add New Push to Config:", err)
		log.Println("Source: GetMetaData()")
		return ErrInternalSeverError
	}
	log.Println("Added New Push to Config")
	log.Println("Done with Everything now returning Response")

	res.LenData = int(file_info.Size())
	res.NewHash = zipped_hash
	res.KeyBytes = []byte(encrypt_key)
	res.IVBytes = []byte(encrypt_iv)

	log.Println(res.NewHash)
	log.Println(len(string(res.KeyBytes)))
	log.Println(len(string(res.IVBytes)))
	log.Println("Length of Data:", res.LenData)

	log.Println("Get Meta Data Done")
	return nil
}
