package handler

import (
	"bufio"
	"encoding/base64"
	"errors"
	"io"
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

func EncryptZipFileAndStore(zipped_filepath, zip_enc_path string, key, iv []byte) error {
	zipped_filepath_obj, err := os.Open(zipped_filepath)
	if err != nil {
		log.Println("Failed to Open Zipped File:", err)
		log.Println("Source: encryptZipFileAndStore()")
		return err
	}
	defer zipped_filepath_obj.Close()

	zip_enc_file_obj, err := os.Create(zip_enc_path)
	if err != nil {
		log.Println("Failed to Create & Open Enc Zipped File:", err)
		log.Println("Source: encryptZipFileAndStore()")
		return err
	}
	defer zip_enc_file_obj.Close()

	buffer := make([]byte, DATA_CHUNK)
	reader := bufio.NewReader(zipped_filepath_obj)
	writer := bufio.NewWriter(zip_enc_file_obj)

	// Reading from Zip File, Encrypting it & Writing it to Enc Zip File
	offset := 0
	for {
		n, err := reader.Read(buffer)
		if err != nil {
			if err == io.EOF {
				log.Println("File Encryption Completed ...")
				break
			}
			log.Println("Error while Reading Zip File:", err)
			log.Println("Source: encryptZipFileAndStore()")
			return err
		}
		encrypted, err := encrypt.EncryptDecryptChunk(buffer[:n], key, iv)
		if err != nil {
			log.Println("Failed to Encrypt Chunk:", err)
			log.Println("Source: encryptZipFileAndStore()")
			return err
		}

		_, err = writer.Write(encrypted)
		if err != nil {
			log.Println("Failed to Write Chunk to File:", err)
			log.Println("Source: encryptZipFileAndStore()")
			return err
		}

		// Flush buffer to disk after 'FLUSH_AFTER_EVERY_X_CHUNK'
		if offset%FLUSH_AFTER_EVERY_X_MB == 0 {
			err = writer.Flush()
			if err != nil {
				log.Println("Error flushing 'writer' after X KB/MB buffer:", err)
				log.Println("Soure: encryptZipFileAndStore()")
				return err
			}
		}
		offset += n
	}

	// Flush buffer to disk at end
	err = writer.Flush()
	if err != nil {
		log.Println("Error flushing 'writer' buffer:", err)
		log.Println("Soure: encryptZipFileAndStore()")
		return err
	}
	zipped_filepath_obj.Close() // Close Obj now, so we can delete zip file
	zip_enc_file_obj.Close()

	// Removing Zip File
	err = os.Remove(zipped_filepath)
	if err != nil {
		log.Println("Error deleting zip file:", err)
		log.Println("Source: encryptZipFileAndStore()")
		return err
	}
	log.Println("Removed Zip File - ", zipped_filepath)
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

	// Reading Last Hash from Config
	workspace_conf, err := config.ReadFromPKRConfigFile(filepath.Join(workspace_path, ".PKr", "workspaceConfig.json"))
	if err != nil {
		log.Println("Error while Reading from PKr Config File:", err)
		log.Println("Source: GetMetaData()")
		return ErrInternalSeverError
	}

	log.Println("Conf Last Hash:", workspace_conf.LastHash)
	log.Println("Req Last Hash:", req.LastHash)
	if workspace_conf.LastHash == req.LastHash {
		log.Println("User has the Latest Workspace, according to Last Hash")
		log.Println("No need to transfer data")
		return ErrUserAlreadyHasLatestWorkspace
	}

	log.Println("Check if Hash Provided is Valid and Present in Updates Hash List")
	is_hash_present, err := config.IfValidHash(req.LastHash, workspace_path)
	if err != nil {
		log.Println("Failed to check whether Hash is valid or not:", err)
		log.Println("Source: GetMetaData()")
		return ErrInternalSeverError
	}

	zip_destination_path := filepath.Join(workspace_path, ".PKr", "Files", "Current") + string(filepath.Separator)
	zip_enc_filepath := zip_destination_path + workspace_conf.LastHash + ".enc"

	res.RequestHash = res.UpdatedHash
	res.IsChanges = false
	res.Updates = nil

	log.Println("Is Hash Present:", is_hash_present)
	if is_hash_present {
		res.Updates = map[string]string{}
		log.Println("Pull")
		res.IsChanges = true

		log.Println("Merging Required Updates between the Hashes")
		merged_updates, err := config.MergeUpdates(workspace_path, req.LastHash, workspace_conf.LastHash)
		if err != nil {
			log.Println("Unable to Merge Updates:", err)
			log.Println("Source: GetMetaData()")
			return ErrInternalSeverError
		}
		log.Println("Merged Updates:", merged_updates)

		log.Println("Generating Changes Hash Name ...")
		files_hash_list := []string{}
		for _, changes := range merged_updates.Changes {
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
		res.RequestHash = merged_updates.Hash
		log.Println("Changes Hash:", merged_updates.Hash)

		is_updates_cache_present, err := filetracker.AreUpdatesCached(workspace_path, merged_updates.Hash)
		if err != nil {
			log.Println("Error while Checking Whether Updates're Already Cached or Not")
			log.Println("Source: GetMetaData()")
			return ErrInternalSeverError
		}
		log.Println("Is Update Cache Present:", is_updates_cache_present)

		if is_updates_cache_present {
			zip_destination_path = filepath.Join(workspace_path, ".PKr", "Files", "Changes", merged_updates.Hash) + string(filepath.Separator)
			zip_enc_filepath = zip_destination_path + merged_updates.Hash + ".enc"
		} else {
			log.Println("Generating Changes Zip")
			err = filetracker.ZipUpdates(merged_updates, workspace_path, merged_updates.Hash)
			if err != nil {
				log.Println("Error while Creating Zip for Changes:", err)
				log.Println("Source: GetMetaData()")
				return ErrInternalSeverError
			}
			changes_path := filepath.Join(workspace_path, ".PKr", "Files", "Changes", merged_updates.Hash)
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

			changes_zipped_filepath := filepath.Join(changes_path, merged_updates.Hash+".zip")
			changes_enc_zip_filepath := strings.Replace(changes_zipped_filepath, ".zip", ".enc", 1)
			if err := encrypt.AESEncrypt(changes_zipped_filepath, changes_enc_zip_filepath, changes_key, changes_iv); err != nil {
				log.Println("Failed to Encrypt Data using AES:", err)
				log.Println("Source: GetMetaData()")
				return ErrInternalSeverError
			}

			err = EncryptZipFileAndStore(changes_zipped_filepath, changes_enc_zip_filepath, changes_key, changes_iv)
			if err != nil {
				log.Println("Error while Encrypting Zip File of Entire Workspace, Storing it & Deleting Zip File:", err)
				log.Println("Source: GetMetaData()")
				return ErrInternalSeverError
			}
			zip_destination_path = changes_path + string(filepath.Separator)
			zip_enc_filepath = zip_destination_path + merged_updates.Hash + ".enc"
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

	file_info, err := os.Stat(zip_enc_filepath)
	if err != nil {
		log.Println("Failed to Get FileInfo of Encrypted Zip File:", err)
		log.Println("Source: GetMetaData()")
		return ErrInternalSeverError
	}
	res.UpdatedHash = workspace_conf.LastHash
	res.KeyBytes = []byte(encrypt_key)
	res.IVBytes = []byte(encrypt_iv)
	res.LenData = int(file_info.Size())

	log.Println(res.IsChanges)
	log.Println(res.LenData)
	log.Println(res.RequestHash)
	log.Println(res.UpdatedHash)
	log.Println(res.Updates)

	log.Println("Done with Everything now returning Response")
	return nil

}
