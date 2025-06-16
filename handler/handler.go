package handler

import (
	"encoding/base64"
	"errors"
	"log"
	"path/filepath"

	"os"

	"github.com/ButterHost69/PKr-Base/config"
	"github.com/ButterHost69/PKr-Base/encrypt"
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
	keysPath, err := config.StorePublicKeys(filepath.Join(file_path, ".PKr", "keys"), string(publicKey))
	if err != nil {
		log.Println("Failed to Store Public Keys at '.PKr\\keys':", err)
		log.Println("Source: InitNewWorkSpaceConnection()")
		return ErrInternalSeverError
	}
	log.Println("Adding New Conn in Pkr Config File")

	// Store the New Connection in the .PKr Config file
	connection.PublicKeyPath = keysPath
	if err := config.AddConnectionToPKRConfigFile(filepath.Join(file_path, ".PKr", "workspaceConfig.json"), connection); err != nil {
		log.Println("Failed to Add Connection to .PKr Config File:", err)
		log.Println("Source: InitNewWorkSpaceConnection()")
		return ErrInternalSeverError
	}

	log.Println("Added New Conn in Pkr Config file")
	log.Println("Init New Workspace Conn Successful")
	return nil
}

// TODO: Send Updated File Structure too...
// Compare Provided Hash File Structure and The Current File Structure Hash
// Compres and Hash those files - Store them
// Send that Hash
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

	workspace_config, err := config.ReadFromPKRConfigFile(filepath.Join(workspace_path, ".PKr", "workspaceConfig.json"))
	if err != nil {
		log.Println("Failed to Fetch Config Struct Using Workspace Path:", err)
		log.Println("Source: GetMetaData()")
		return ErrInternalSeverError
	}

	// If Provided Last Hash == Last Snapshot then Do Nothing
	// If Provided Last Hash == "" or garbage then User Cloning For the First time ; Provide Latest Snapshot
	if req.LastHash == workspace_config.LastHash {
		log.Println("User has the Latest Workspace, according to Last Hash")
		log.Println("No need to transfer data")
		return ErrUserAlreadyHasLatestWorkspace
	}

	log.Println("Check if Hash Provided is Valid and Present in Updates Hash List")
	ifHashPresent, err := config.IfHashContains(req.LastHash, filepath.Join(workspace_path, ".PKr", "workspaceConfig.json"))
	if err != nil {
		log.Println("Failed to verify Hash:", err)
		log.Println("Source: GetMetaData()")
		return ErrInternalSeverError
	}

	if ifHashPresent {

	} else {
		// Send Last Snapshot if last hash is garbage or "" it means clone or some funny business
		// Send Last Hash
		// Encrypt Given AES Key and IV
		// Send that shiii....
		log.Println("User is Cloning For the First Time")
		log.Println("Provide Latest Snapshot from .Pkr/Files/Current/")

		snapshot_folder_path := filepath.Join(workspace_path, ".PKr", "Files", "Current")
		enc_zip_path := filepath.Join(snapshot_folder_path, workspace_config.LastHash+".enc")
		iv_path := filepath.Join(snapshot_folder_path, "AES_IV")
		aeskey_path := filepath.Join(snapshot_folder_path, "AES_KEY")

		log.Println("Reading Latest Snapshot Hash IV")
		log.Println("IV Loc: ", iv_path)
		iv_data, err := os.ReadFile(iv_path)
		if err != nil {
			log.Println("Error reading IV:", err)
			log.Println("Source: GetMetaData()")
			return err
		}

		log.Println("Reading Latest Snapshot Hash AES Key")
		log.Println("AES Key Loc: ", aeskey_path)
		aeskey_data, err := os.ReadFile(aeskey_path)
		if err != nil {
			log.Println("Error reading AES Key:", err)
			log.Println("Source: GetMetaData()")
			return err
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

		log.Println("Encrypting AES Keys")
		encrypt_key, err := encrypt.EncryptData(string(aeskey_data), string(public_key))
		if err != nil {
			log.Println("Failed to Encrypt AES Keys using Listener's Public Key:", err)
			log.Println("Source: GetMetaData()")
			return ErrInternalSeverError
		}

		encrypt_iv, err := encrypt.EncryptData(string(iv_data), string(public_key))
		if err != nil {
			log.Println("Failed to Encrypt IV Keys using Listener's Public Key:", err)
			log.Println("Source: GetMetaData()")
			return ErrInternalSeverError
		}

		file_info, err := os.Stat(enc_zip_path)
		if err != nil {
			log.Println("Failed to Get FileInfo of Encrypted Zip File:", err)
			log.Println("Source: GetMetaData()")
			return ErrInternalSeverError
		}

		res.LenData = int(file_info.Size())
		res.NewHash = workspace_config.LastHash
		res.KeyBytes = []byte(encrypt_key)
		res.IVBytes = []byte(encrypt_iv)

		log.Println(res.NewHash)
		log.Println(len(string(res.KeyBytes)))
		log.Println(len(string(res.IVBytes)))
		log.Println("Length of Data:", res.LenData)

		log.Println("Get Meta Data Done")
		return nil

	}

	return nil
}
