package handler

import (
	"encoding/base64"
	"errors"
	"log"
	"path/filepath"
	"strings"

	"os"

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
	keysPath, err := config.StorePublicKeys(filepath.Join(file_path, ".PKr", "Keys"), string(publicKey))
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
	ifHashPresent, err := config.IfValidHash(req.LastHash, filepath.Join(workspace_path, ".PKr", "workspaceConfig.json"))
	if err != nil {
		log.Println("Failed to verify Hash:", err)
		log.Println("Source: GetMetaData()")
		return ErrInternalSeverError
	}

	snapshot_folder_path := ""
	enc_zip_path := ""
	iv_path := ""
	aeskey_path := ""
	if_changes := false
	requestedHash := ""
	resUpdates := map[string]string{}

	// If Hash is Valid -> Create Changes File zip if not Present, Send Changes only
	if ifHashPresent {
		if_changes = true

		log.Println("Provided Hash is Valid: ", req.LastHash)

		log.Println("Merging Required Updates between the Hashes")
		mupdates, err := config.MergeUpdates(filepath.Join(workspace_path, ".PKr", "workspaceConfig.json"), req.LastHash, workspace_config.LastHash)
		if err != nil {
			log.Println("Unable to Merge Updates:", err)
			log.Println("Source: GetMetaData()")
			return ErrInternalSeverError
		}

		log.Println("Zipping Changed Files")
		log.Println("Generating Changes Hash Name ...")
		files_hash_list := []string{}
		for _, changes := range mupdates.Changes {
			resUpdates[changes.FilePath] = changes.Type
			if changes.Type == "Updated" {
				files_hash_list = append(files_hash_list, mupdates.Hash)
			}
		}
		changes_hash_name := encrypt.GeneratHashFromFileNames(files_hash_list)
		log.Println("Changes Hash:", changes_hash_name)
		requestedHash = changes_hash_name

		log.Println("Checking if Required Hash File Name already Generated")
		ifPresent, err := filetracker.IfUpdateHashCached(workspace_path, changes_hash_name)
		if err != nil {
			log.Println("Unable to Fetch Update Hashes:", err)
			log.Println("Source: GetMetaData()")
			return ErrInternalSeverError
		}

		// This Updates/Changes are already generated in Changes Directory
		if ifPresent {
			log.Println("Required Changes File is Already Cached...")
			log.Println("Sending and Preparing: ", changes_hash_name)

			snapshot_folder_path = filepath.Join(workspace_path, ".PKr", "Files", "Changes", changes_hash_name)
			enc_zip_path = filepath.Join(snapshot_folder_path, changes_hash_name+".enc")
			iv_path = filepath.Join(snapshot_folder_path, "AES_IV")
			aeskey_path = filepath.Join(snapshot_folder_path, "AES_KEY")
		} else { // Generating Hash Zip -> Enc Zip -> Store AES Key and IV
			log.Println("Required Changes File is Not Cached...")
			log.Println("Generating Enc Zip , Keys and IV")

			log.Println("Generating Changes Zip")
			err = filetracker.ZipUpdates(mupdates, workspace_path, changes_hash_name)
			if err != nil {
				log.Println("Error while Creating Zip for Changes:", err)
				log.Println("Source: GetMetaData()")
				return ErrInternalSeverError
			}

			// Encrypt Zip and Store Keys
			log.Println("Encrypting Changes Zip File...")
			changes_path := filepath.Join(workspace_path, ".PKr", "Files", "Changes", changes_hash_name)

			log.Println("Generating Keys for Changes File ...")
			changes_key, err := encrypt.AESGenerakeKey(16)
			if err != nil {
				log.Println("Failed to Generate AES Keys:", err)
				log.Println("Source: GetMetaData()")
				return ErrInternalSeverError
			}

			err = os.WriteFile(filepath.Join(changes_path, "AES_KEY"), changes_key, 0644)
			if err != nil {
				log.Println("Failed to Write AES Key to File:", err)
				log.Println("Source: GetMetaData()")
				return ErrInternalSeverError
			}

			log.Println("Generating IV for Changes File ...")
			changes_iv, err := encrypt.AESGenerateIV()
			if err != nil {
				log.Println("Failed to Generate IV Keys:", err)
				log.Println("Source: GetMetaData()")
				return ErrInternalSeverError
			}

			err = os.WriteFile(filepath.Join(changes_path, "AES_IV"), changes_iv, 0644)
			if err != nil {
				log.Println("Failed to Write AES IV to File:", err)
				log.Println("Source: GetMetaData()")
				return ErrInternalSeverError
			}

			// Encrypting Zip File
			log.Println("Encrypting Zip and Storing for Workspace ...")
			changes_zipped_filepath := filepath.Join(changes_path, changes_hash_name+".zip")
			changes_destination_filepath := strings.Replace(changes_zipped_filepath, ".zip", ".enc", 1)
			if err := encrypt.AESEncrypt(changes_zipped_filepath, changes_destination_filepath, changes_key, changes_iv); err != nil {
				log.Println("Failed to Encrypt Data using AES:", err)
				log.Println("Source: GetMetaData()")
				return ErrInternalSeverError
			}

			err = os.Remove(changes_zipped_filepath)
			if err != nil {
				log.Println("Error deleting zip file:", err)
				log.Println("Source: Push()")
				return ErrInternalSeverError
			}
			log.Println("Removed Changes Zip File - ", changes_zipped_filepath)

			snapshot_folder_path = filepath.Join(workspace_path, ".PKr", "Files", "Changes", changes_hash_name)
			enc_zip_path = filepath.Join(snapshot_folder_path, changes_hash_name+".enc")
			iv_path = filepath.Join(snapshot_folder_path, "AES_IV")
			aeskey_path = filepath.Join(snapshot_folder_path, "AES_KEY")
		}

	} else {
		// Send Last Snapshot if last hash is garbage or "" it means clone or some funny business
		// Send Last Hash
		// Encrypt Given AES Key and IV
		// Send that shiii....
		log.Println("User is Cloning For the First Time")
		log.Println("Provide Latest Snapshot from .Pkr/Files/Current/")

		snapshot_folder_path = filepath.Join(workspace_path, ".PKr", "Files", "Current")
		enc_zip_path = filepath.Join(snapshot_folder_path, workspace_config.LastHash+".enc")
		iv_path = filepath.Join(snapshot_folder_path, "AES_IV")
		aeskey_path = filepath.Join(snapshot_folder_path, "AES_KEY")
		requestedHash = workspace_config.LastHash
	}

	log.Println("~ Files Loc:")
	log.Println("Snapshot Folder Path: ", snapshot_folder_path)
	log.Println("Encrypted Zip Path: ", enc_zip_path)
	log.Println("AES IV Path: ", iv_path)
	log.Println("AES Key Path: ", aeskey_path)

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
	res.RequestHash = requestedHash
	res.UpdatedHash = workspace_config.LastHash
	res.KeyBytes = []byte(encrypt_key)
	res.IVBytes = []byte(encrypt_iv)
	res.IsChanges = if_changes
	res.Updates = resUpdates

	log.Println("[Sending MetaDataResponse] New Hash:", res.RequestHash)
	log.Println("[Sending MetaDataResponse] Key Len:", len(string(res.KeyBytes)))
	log.Println("[Sending MetaDataResponse] IV Len", len(string(res.IVBytes)))
	log.Println("[Sending MetaDataResponse] Length of Data:", res.LenData)
	log.Println("[Sending MetaDataResponse] Is Changes:", res.IsChanges)
	log.Println("[Sending MetaDataResponse] Updates:", res.Updates)

	log.Println("Get Meta Data Done")
	return nil

}
