package handler

import (
	"encoding/base64"
	"errors"
	"path/filepath"
	"strconv"
	"strings"

	"os"

	"github.com/ButterHost69/PKr-Base/config"
	"github.com/ButterHost69/PKr-Base/encrypt"
	"github.com/ButterHost69/PKr-Base/filetracker"
	"github.com/ButterHost69/PKr-Base/logger"
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
	logger.USER_LOGGER.Println("Get Public Key Called ...")

	keyData, err := config.ReadPublicKey()
	if err != nil {
		logger.USER_LOGGER.Println("Error while reading My Public Key from config:", err)
		logger.USER_LOGGER.Println("Source: GetPublicKey()")
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

	logger.USER_LOGGER.Println("Init New Work Space Connection Called ...")

	password, err := encrypt.DecryptData(req.WorkspacePassword)
	if err != nil {
		logger.USER_LOGGER.Println("Failed to Decrypt the Workspace Pass Received from Listener:", err)
		logger.USER_LOGGER.Println("Source: InitNewWorkSpaceConnection()")
		return ErrInternalSeverError
	}
	logger.USER_LOGGER.Println("Decrypted Data ...\nAuthenticating ...")

	// Authenticates Workspace Name and Password and Get the Workspace File Path
	file_path, err := config.AuthenticateWorkspaceInfo(req.WorkspaceName, password)
	if err != nil {
		if errors.Is(err, ErrIncorrectPassword) {
			logger.USER_LOGGER.Println("Error: Incorrect Credentials for Workspace")
			logger.USER_LOGGER.Println("Source: InitNewWorkSpaceConnection()")
			return ErrIncorrectPassword
		}
		logger.USER_LOGGER.Println("Failed to Authenticate Password of Listener:", err)
		logger.USER_LOGGER.Println("Source: InitNewWorkSpaceConnection()")
		return ErrIncorrectPassword
	}

	logger.USER_LOGGER.Println("Auth Successfull ...\nGetting Server Details Using Server IP")
	server, err := config.GetServerDetailsUsingServerIP(req.ServerIP)
	if err != nil {
		if errors.Is(err, ErrServerNotFound) {
			logger.USER_LOGGER.Println("Error: Server Not Found")
			logger.USER_LOGGER.Println("Source: InitNewWorkSpaceConnection()")
			return ErrServerNotFound
		}
		logger.USER_LOGGER.Println("Failed to Get Server Details Using Server IP:", err)
		logger.USER_LOGGER.Println("Source: InitNewWorkSpaceConnection()")
		return ErrInternalSeverError
	}
	logger.USER_LOGGER.Println("Getting Server Details Using Server IP Successfull ...\n Decoding Public Keys ...")

	var connection config.Connection
	connection.Username = req.MyUsername
	connection.ServerAlias = server.ServerAlias

	publicKey, err := base64.StdEncoding.DecodeString(string(req.MyPublicKey))
	if err != nil {
		logger.USER_LOGGER.Println("Failed to Decode Public Key from base64:", err)
		logger.USER_LOGGER.Println("Source: InitNewWorkSpaceConnection()")
		return ErrInternalSeverError
	}

	// Save Public Key
	logger.USER_LOGGER.Println("Storing Public Keys ...")
	keysPath, err := config.StorePublicKeys(req.MyUsername, filepath.Join(file_path, ".PKr", "keys"), string(publicKey))
	if err != nil {
		logger.USER_LOGGER.Println("Failed to Store Public Keys at '.PKr\\keys':", err)
		logger.USER_LOGGER.Println("Source: InitNewWorkSpaceConnection()")
		return ErrInternalSeverError
	}
	logger.USER_LOGGER.Println("Adding New Conn in Pkr Config File")

	// Store the New Connection in the .PKr Config file
	connection.PublicKeyPath = keysPath
	if err := config.AddConnectionToPKRConfigFile(filepath.Join(file_path, ".PKr", "workspaceConfig.json"), connection); err != nil {
		logger.USER_LOGGER.Println("Failed to Add Connection to .PKr Config File:", err)
		logger.USER_LOGGER.Println("Source: InitNewWorkSpaceConnection()")
		return ErrInternalSeverError
	}

	logger.USER_LOGGER.Println("Added New Conn in Pkr Config file")
	logger.USER_LOGGER.Println("Init New Workspace Conn Successful")
	return nil
}

func (h *ClientHandler) GetMetaData(req models.GetMetaDataRequest, res *models.GetMetaDataResponse) error {
	password, err := encrypt.DecryptData(req.WorkspacePassword)
	if err != nil {
		logger.USER_LOGGER.Println("Failed to Decrypt the Workspace Pass Received from Listener:", err)
		logger.USER_LOGGER.Println("Source: GetMetaData()")
		return ErrInternalSeverError
	}
	logger.USER_LOGGER.Println("Decrypted Data ...\nAuthenticating ...")

	// Authenticates Workspace Name and Password and Get the Workspace File Path
	_, err = config.AuthenticateWorkspaceInfo(req.WorkspaceName, password)
	if err != nil {
		if errors.Is(err, ErrIncorrectPassword) {
			logger.USER_LOGGER.Println("Error: Incorrect Credentials for Workspace")
			logger.USER_LOGGER.Println("Source: GetMetaData()")
			return ErrIncorrectPassword
		}
		logger.USER_LOGGER.Println("Failed to Authenticate Password of Listener:", err)
		logger.USER_LOGGER.Println("Source: GetMetaData()")
		return ErrIncorrectPassword
	}
	logger.USER_LOGGER.Printf("Data Requested For Workspace: %s\n", req.WorkspaceName)

	workspace_path, err := config.GetSendWorkspaceFilePath(req.WorkspaceName)
	if err != nil {
		logger.USER_LOGGER.Println("Failed to Get Workspace Path from Config:", err)
		logger.USER_LOGGER.Println("Source: GetMetaData()")
		return ErrInternalSeverError
	}

	// Reading Last Push Num from Config
	workspace_conf, err := config.ReadFromPKRConfigFile(filepath.Join(workspace_path, ".PKr", "workspaceConfig.json"))
	if err != nil {
		logger.USER_LOGGER.Println("Error while Reading from PKr Config File:", err)
		logger.USER_LOGGER.Println("Source: GetMetaData()")
		return ErrInternalSeverError
	}

	logger.USER_LOGGER.Println("Conf Last Push Num:", workspace_conf.LastPushNum)
	logger.USER_LOGGER.Println("Req Last Push Num:", req.LastPushNum)
	if workspace_conf.LastPushNum == req.LastPushNum {
		logger.USER_LOGGER.Println("User has the Latest Workspace, according to Last Push Num")
		logger.USER_LOGGER.Println("No need to transfer data")
		return ErrUserAlreadyHasLatestWorkspace
	}

	if req.LastPushNum > workspace_conf.LastPushNum {
		logger.USER_LOGGER.Println("User has Requested Invalid Last Push Num")
		return ErrInvalidLastPushNum
	}

	zip_destination_path := filepath.Join(workspace_path, ".PKr", "Files", "Current") + string(filepath.Separator)

	res.RequestPushRange = strconv.Itoa(workspace_conf.LastPushNum)
	res.Updates = nil

	// LastPushNum = -1 => Requesting for first time,i.e, Clone
	if req.LastPushNum == -1 {
		logger.USER_LOGGER.Println("Clone")
		file_info, err := os.Stat(zip_destination_path + strconv.Itoa(workspace_conf.LastPushNum) + ".zip")
		if err != nil {
			logger.USER_LOGGER.Println("Failed to Get FileInfo of Zip File:", err)
			logger.USER_LOGGER.Println("Source: GetMetaData()")
			return ErrInternalSeverError
		}
		res.LenData = int(file_info.Size())
	} else {
		var zip_enc_filepath string
		res.Updates = map[string]string{}
		logger.USER_LOGGER.Println("Pull")

		logger.USER_LOGGER.Println("Merging Required Updates between the Pushes")
		merged_changes, err := config.MergeUpdates(workspace_path, req.LastPushNum, workspace_conf.LastPushNum)
		if err != nil {
			logger.USER_LOGGER.Println("Unable to Merge Updates:", err)
			logger.USER_LOGGER.Println("Source: GetMetaData()")
			return ErrInternalSeverError
		}
		// logger.USER_LOGGER.Println("Merged Changes:", merged_changes)

		logger.USER_LOGGER.Println("Generating Changes Push Name ...")
		for _, changes := range merged_changes {
			res.Updates[changes.FilePath] = changes.Type
		}
		// logger.USER_LOGGER.Println("Res.Updates:", res.Updates)
		res.RequestPushRange = strconv.Itoa(req.LastPushNum) + "-" + strconv.Itoa(workspace_conf.LastPushNum)
		logger.USER_LOGGER.Println("Request Push Range:", res.RequestPushRange)

		is_updates_cache_present, err := filetracker.AreUpdatesCached(workspace_path, res.RequestPushRange)
		if err != nil {
			logger.USER_LOGGER.Println("Error while Checking Whether Updates're Already Cached or Not")
			logger.USER_LOGGER.Println("Source: GetMetaData()")
			return ErrInternalSeverError
		}
		logger.USER_LOGGER.Println("Is Update Cache Present:", is_updates_cache_present)

		if is_updates_cache_present {
			zip_destination_path = filepath.Join(workspace_path, ".PKr", "Files", "Changes", res.RequestPushRange) + string(filepath.Separator)
			zip_enc_filepath = zip_destination_path + res.RequestPushRange + ".enc"
		} else {
			logger.USER_LOGGER.Println("Generating Changes Zip")
			last_push_num_str := strconv.Itoa(workspace_conf.LastPushNum)
			src_path := filepath.Join(workspace_path, ".PKr", "Files", "Current", last_push_num_str+".zip")
			dst_path := filepath.Join(workspace_path, ".PKr", "Files", "Changes", res.RequestPushRange, res.RequestPushRange+".zip")

			err = filetracker.ZipUpdates(merged_changes, src_path, dst_path)
			if err != nil {
				logger.USER_LOGGER.Println("Error while Creating Zip for Changes:", err)
				logger.USER_LOGGER.Println("Source: GetMetaData()")
				return ErrInternalSeverError
			}
			changes_path := filepath.Join(workspace_path, ".PKr", "Files", "Changes", res.RequestPushRange)
			logger.USER_LOGGER.Println("Generating Keys for Changes File ...")

			changes_key, err := encrypt.AESGenerakeKey(16)
			if err != nil {
				logger.USER_LOGGER.Println("Failed to Generate AES Keys:", err)
				logger.USER_LOGGER.Println("Source: GetMetaData()")
				return ErrInternalSeverError
			}

			err = os.WriteFile(filepath.Join(changes_path, "AES_KEY"), changes_key, 0644)
			if err != nil {
				logger.USER_LOGGER.Println("Failed to Write AES Key to File:", err)
				logger.USER_LOGGER.Println("Source: GetMetaData()")
				return ErrInternalSeverError
			}

			logger.USER_LOGGER.Println("Generating IV for Changes File ...")
			changes_iv, err := encrypt.AESGenerateIV()
			if err != nil {
				logger.USER_LOGGER.Println("Failed to Generate IV Keys:", err)
				logger.USER_LOGGER.Println("Source: GetMetaData()")
				return ErrInternalSeverError
			}

			err = os.WriteFile(filepath.Join(changes_path, "AES_IV"), changes_iv, 0644)
			if err != nil {
				logger.USER_LOGGER.Println("Failed to Write AES IV to File:", err)
				logger.USER_LOGGER.Println("Source: GetMetaData()")
				return ErrInternalSeverError
			}

			changes_zipped_filepath := filepath.Join(changes_path, res.RequestPushRange+".zip")
			changes_enc_zip_filepath := strings.Replace(changes_zipped_filepath, ".zip", ".enc", 1)

			err = encrypt.EncryptZipFileAndStore(changes_zipped_filepath, changes_enc_zip_filepath, changes_key, changes_iv)
			if err != nil {
				logger.USER_LOGGER.Println("Error while Encrypting Zip File of Entire Workspace, Storing it & Deleting Zip File:", err)
				logger.USER_LOGGER.Println("Source: GetMetaData()")
				return ErrInternalSeverError
			}
			zip_destination_path = changes_path + string(filepath.Separator)
			zip_enc_filepath = zip_destination_path + res.RequestPushRange + ".enc"
		}
		file_info, err := os.Stat(zip_enc_filepath)
		if err != nil {
			logger.USER_LOGGER.Println("Failed to Get FileInfo of Encrypted Zip File:", err)
			logger.USER_LOGGER.Println("Source: GetMetaData()")
			return ErrInternalSeverError
		}
		res.LenData = int(file_info.Size())
	}

	key, err := os.ReadFile(zip_destination_path + "AES_KEY")
	if err != nil {
		logger.USER_LOGGER.Println("Failed to Fetch AES Keys:", err)
		logger.USER_LOGGER.Println("Source: GetMetaData()")
		return ErrInternalSeverError
	}

	iv, err := os.ReadFile(zip_destination_path + "AES_IV")
	if err != nil {
		logger.USER_LOGGER.Println("Failed to Fetch IV Keys:", err)
		logger.USER_LOGGER.Println("Source: GetMetaData()")
		return ErrInternalSeverError
	}

	logger.USER_LOGGER.Println("Fetching Public Key of Listener")
	public_key_path, err := config.GetConnectionsPublicKeyUsingUsername(workspace_path, req.Username)
	if err != nil {
		logger.USER_LOGGER.Println("Failed to Get Public Key's Path of Listener Using Username:", err)
		logger.USER_LOGGER.Println("Source: GetMetaData()")
		return ErrInternalSeverError
	}

	public_key, err := os.ReadFile(public_key_path)
	if err != nil {
		logger.USER_LOGGER.Println("Failed to Read Public Key of Listener:", err)
		logger.USER_LOGGER.Println("Source: GetMetaData()")
		return ErrInternalSeverError
	}

	logger.USER_LOGGER.Println("Encrypting Key")
	encrypt_key, err := encrypt.EncryptData(string(key), string(public_key))
	if err != nil {
		logger.USER_LOGGER.Println("Failed to Encrypt AES Keys using Listener's Public Key:", err)
		logger.USER_LOGGER.Println("Source: GetMetaData()")
		return ErrInternalSeverError
	}

	encrypt_iv, err := encrypt.EncryptData(string(iv), string(public_key))
	if err != nil {
		logger.USER_LOGGER.Println("Failed to Encrypt IV Keys using Listener's Public Key:", err)
		logger.USER_LOGGER.Println("Source: GetMetaData()")
		return ErrInternalSeverError
	}

	res.KeyBytes = []byte(encrypt_key)
	res.IVBytes = []byte(encrypt_iv)

	res.LastPushNum = workspace_conf.LastPushNum
	res.LastPushDesc = workspace_conf.AllUpdates[workspace_conf.LastPushNum].PushDesc

	logger.USER_LOGGER.Println(res.LenData)
	logger.USER_LOGGER.Println(res.RequestPushRange)
	logger.USER_LOGGER.Println(res.LastPushNum)
	logger.USER_LOGGER.Println(res.LastPushDesc)
	// logger.USER_LOGGER.Println(res.Updates)

	logger.USER_LOGGER.Println("Done with Everything now returning Response")
	return nil
}
