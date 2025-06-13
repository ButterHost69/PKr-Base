package handler

import (
	"encoding/base64"
	"errors"
	"io/ioutil"
	"log"

	"os"
	"strings"

	"github.com/ButterHost69/PKr-Base/config"
	"github.com/ButterHost69/PKr-Base/encrypt"
	"github.com/ButterHost69/PKr-Base/filetracker"
	"github.com/ButterHost69/PKr-Base/models"
)

var (
	ErrIncorrectPassword  = errors.New("incorrect password")
	ErrServerNotFound     = errors.New("server not found in config")
	ErrInternalSeverError = errors.New("internal server error")
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
	// 5. Send the Response with port [X]
	// 6. Open a Data Transfer Port and shit [Will be a separate Function not here] [X]

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

func (h *ClientHandler) GetData(req models.GetDataRequest, res *models.GetDataResponse) error {
	// FIXME AUTH req.workspace_name, req.workspace_password
	// FIXME Store Keys when called GetPublicKey in cli and reuse it

	// TODO Compare Hash ...
	// TODO - If Hash == "" : Send Entire File
	// TODO - If Hash == Last_Hash : Do Nothing

	// TODO Maintain 3 Last Hash object files, and Send Files Accordingly

	log.Printf("Data Requested For Workspace: %s\n", req.WorkspaceName)
	workspacePath, err := config.GetSendWorkspaceFilePath(req.WorkspaceName)
	if err != nil {
		log.Println("Failed to Get Workspace Path from Config:", err)
		log.Println("Source: GetData()")
		return ErrInternalSeverError
	}

	zipped_file_name, err := filetracker.ZipData(workspacePath)
	if err != nil {
		log.Println("Failed to Hash & Zip Data:", err)
		log.Println("Source: GetData()")
		return ErrInternalSeverError
	}

	zipped_hash := strings.Split(zipped_file_name, ".")[0]
	log.Println("Data Service Hash File Name: " + zipped_file_name)

	err = config.AddNewPushToConfig(req.WorkspaceName, zipped_hash)
	if err != nil {
		log.Println("Failed to Add New Push to Config:", err)
		log.Println("Source: GetData()")
		return ErrInternalSeverError
	}
	log.Println("Added New Push to Config")
	// config.AddLogEntry(req.WorkspaceName, true, "Workspace Zipped")

	log.Println("Generating Keys ...")
	key, err := encrypt.AESGenerakeKey(16)
	if err != nil {
		log.Println("Failed to Generate AES Keys:", err)
		log.Println("Source: GetData()")
		return ErrInternalSeverError
	}

	iv, err := encrypt.AESGenerateIV()
	if err != nil {
		log.Println("Failed to Generate IV Keys:", err)
		log.Println("Source: GetData()")
		return ErrInternalSeverError
	}

	log.Println("Zipping Workspace ...")

	zipped_filepath := workspacePath + "\\.PKr\\" + zipped_file_name
	destination_filepath := strings.Replace(zipped_filepath, ".zip", ".enc", 1)
	if err := encrypt.AESEncrypt(zipped_filepath, destination_filepath, key, iv); err != nil {
		log.Println("Failed to Encrypt Data using AES:", err)
		log.Println("Source: GetData()")
		return ErrInternalSeverError
	}

	log.Println("Fetching Public Key of Listener")
	publicKeyPath, err := config.GetConnectionsPublicKeyUsingUsername(workspacePath, req.Username)
	if err != nil {
		log.Println("Failed to Get Public Key's Path of Listener Using Username:", err)
		log.Println("Source: GetData()")
		return ErrInternalSeverError
	}

	publicKey, err := os.ReadFile(publicKeyPath)
	if err != nil {
		log.Println("Failed to Read Public Key of Listener:", err)
		log.Println("Source: GetData()")
		return ErrInternalSeverError
	}

	log.Println("Encrypting Key")
	encrypt_key, err := encrypt.EncryptData(string(key), string(publicKey))
	if err != nil {
		log.Println("Failed to Encrypt AES Keys using Listener's Public Key:", err)
		log.Println("Source: GetData()")
		return ErrInternalSeverError
	}

	encrypt_iv, err := encrypt.EncryptData(string(iv), string(publicKey))
	if err != nil {
		log.Println("Failed to Encrypt IV Keys using Listener's Public Key:", err)
		log.Println("Source: GetData()")
		return ErrInternalSeverError
	}

	log.Println("Opening Destination File")
	encrypt_file, err := os.Open(destination_filepath)
	if err != nil {
		log.Println("Failed to Open Destination Filepath:", err)
		log.Println("Source: GetData()")
		return ErrInternalSeverError
	}
	defer encrypt_file.Close()

	log.Println("Reading Destination File")
	fileData, err := ioutil.ReadAll(encrypt_file)
	if err != nil {
		log.Println("Failed to Read Content of Destination FilePath:", err)
		log.Println("Source: GetData()")
		return ErrInternalSeverError
	}

	log.Println("Done with Everything now returning Response")

	res.NewHash = zipped_hash
	res.KeyBytes = []byte(encrypt_key)
	res.IVBytes = []byte(encrypt_iv)
	res.Data = fileData

	log.Println(res.NewHash)
	log.Println(len(string(res.KeyBytes)))
	log.Println(len(string(res.IVBytes)))
	log.Println(len(string(res.Data)))

	log.Println("Get Data Done")
	return nil
}
