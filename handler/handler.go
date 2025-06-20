package handler

import (
	"encoding/base64"
	"errors"
	"log"
	"path/filepath"
	"strconv"
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
	ErrInvalidLastPushNum            = errors.New("invalid last push number")
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
	keysPath, err := config.StorePublicKeys(req.MyUsername, filepath.Join(file_path, ".PKr", "keys"), string(publicKey))
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

	// Reading Last Hash from Config
	workspace_conf, err := config.ReadFromPKRConfigFile(filepath.Join(workspace_path, ".PKr", "workspaceConfig.json"))
	if err != nil {
		log.Println("Error while Reading from PKr Config File:", err)
		log.Println("Source: GetMetaData()")
		return ErrInternalSeverError
	}

	log.Println("Conf Last Push Num:", workspace_conf.LastPushNum)
	log.Println("Req Last Push Num:", req.LastPushNum)
	if workspace_conf.LastPushNum == req.LastPushNum {
		log.Println("User has the Latest Workspace, according to Last Push Num")
		log.Println("No need to transfer data")
		return ErrUserAlreadyHasLatestWorkspace
	}

	if req.LastPushNum > workspace_conf.LastPushNum {
		log.Println("User has Requested Invalid Last Push Num")
		return ErrInvalidLastPushNum
	}

	zip_destination_path := filepath.Join(workspace_path, ".PKr", "Files", "Current") + string(filepath.Separator)
	zip_enc_filepath := zip_destination_path + strconv.Itoa(workspace_conf.LastPushNum) + ".enc"

	res.RequestPushRange = strconv.Itoa(workspace_conf.LastPushNum)
	res.Updates = nil

	if req.LastPushNum != -1 {
		res.Updates = map[string]string{}
		log.Println("Pull")

		log.Println("Merging Required Updates between the Hashes")
		merged_changes, err := config.MergeUpdates(workspace_path, req.LastPushNum, workspace_conf.LastPushNum)
		if err != nil {
			log.Println("Unable to Merge Updates:", err)
			log.Println("Source: GetMetaData()")
			return ErrInternalSeverError
		}
		log.Println("Merged Changes:", merged_changes)

		log.Println("Generating Changes Hash Name ...")
		files_hash_list := []string{}
		for _, changes := range merged_changes {
			res.Updates[changes.FilePath] = changes.Type
			if changes.Type == "Updated" {
				log.Println(changes.FilePath)
				files_hash_list = append(files_hash_list, changes.FilePath)
				files_hash_list = append(files_hash_list, changes.FileHash)
				res.Updates[changes.FilePath] = changes.Type
			}
		}
		log.Println("Files Hash List:", files_hash_list)
		log.Println("Res.Updates:", res.Updates)
		res.RequestPushRange = strconv.Itoa(req.LastPushNum) + "-" + strconv.Itoa(workspace_conf.LastPushNum)
		log.Println("Request Push Range:", res.RequestPushRange)

		is_updates_cache_present, err := filetracker.AreUpdatesCached(workspace_path, res.RequestPushRange)
		if err != nil {
			log.Println("Error while Checking Whether Updates're Already Cached or Not")
			log.Println("Source: GetMetaData()")
			return ErrInternalSeverError
		}
		log.Println("Is Update Cache Present:", is_updates_cache_present)

		if is_updates_cache_present {
			zip_destination_path = filepath.Join(workspace_path, ".PKr", "Files", "Changes", res.RequestPushRange) + string(filepath.Separator)
			zip_enc_filepath = zip_destination_path + res.RequestPushRange + ".enc"
		} else {
			log.Println("Generating Changes Zip")
			last_push_num_str := strconv.Itoa(workspace_conf.LastPushNum)
			src_path := filepath.Join(workspace_path, ".PKr", "Current", last_push_num_str, last_push_num_str+".zip")
			dst_path := filepath.Join(workspace_path, ".PKr", "Changes", res.RequestPushRange, res.RequestPushRange+".zip")

			err = filetracker.ZipUpdates(merged_changes, src_path, dst_path)
			if err != nil {
				log.Println("Error while Creating Zip for Changes:", err)
				log.Println("Source: GetMetaData()")
				return ErrInternalSeverError
			}
			changes_path := filepath.Join(workspace_path, ".PKr", "Files", "Changes", res.RequestPushRange)
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

			changes_zipped_filepath := filepath.Join(changes_path, res.RequestPushRange+".zip")
			changes_enc_zip_filepath := strings.Replace(changes_zipped_filepath, ".zip", ".enc", 1)

			err = encrypt.EncryptZipFileAndStore(changes_zipped_filepath, changes_enc_zip_filepath, changes_key, changes_iv)
			if err != nil {
				log.Println("Error while Encrypting Zip File of Entire Workspace, Storing it & Deleting Zip File:", err)
				log.Println("Source: GetMetaData()")
				return ErrInternalSeverError
			}
			zip_destination_path = changes_path + string(filepath.Separator)
			zip_enc_filepath = zip_destination_path + res.RequestPushRange + ".enc"
		}
	}

	key, err := os.ReadFile(zip_destination_path + "AES_KEY")
	if err != nil {
		log.Println("Failed to Fetch AES Keys:", err)
		log.Println("Source: GetMetaData()")
		return ErrInternalSeverError
	}

	iv, err := os.ReadFile(zip_destination_path + "AES_IV")
	if err != nil {
		log.Println("Failed to Fetch IV Keys:", err)
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

	file_info, err := os.Stat(zip_enc_filepath)
	if err != nil {
		log.Println("Failed to Get FileInfo of Encrypted Zip File:", err)
		log.Println("Source: GetMetaData()")
		return ErrInternalSeverError
	}

	res.KeyBytes = []byte(encrypt_key)
	res.IVBytes = []byte(encrypt_iv)
	res.LenData = int(file_info.Size())

	res.LastPushNum = workspace_conf.LastPushNum
	res.LastPushDesc = workspace_conf.AllUpdates[workspace_conf.LastPushNum].PushDesc

	log.Println(res.LenData)
	log.Println(res.RequestPushRange)
	log.Println(res.LastPushNum)
	log.Println(res.LastPushDesc)
	log.Println(res.Updates)

	log.Println("Done with Everything now returning Response")
	return nil
}
