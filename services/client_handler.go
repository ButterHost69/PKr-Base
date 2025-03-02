package services

import (
	"encoding/base64"
	"errors"
	"fmt"
	"io/ioutil"
	"log"

	"os"
	"strings"

	"github.com/ButterHost69/PKr-Base/config"
	"github.com/ButterHost69/PKr-Base/encrypt"
	"github.com/ButterHost69/PKr-Base/filetracker"
)

var (
	ErrIncorrectPassword = errors.New("incorrect password")
	ErrServerNotFound    = errors.New("server not found in config")
)

func (h *ClientHandler) GetPublicKey(req PublicKeyRequest, res *PublicKeyResponse) error {
	h.UserConfingLogger.Info("Get Public Key Called ...")
	log.Println("Get Public Key Called ...")

	keyData, err := config.ReadPublicKey()
	if err != nil {
		fmt.Println("Error while reading Public Key from config\nSource: GetPublicKey\nError:", err)
		logentry := fmt.Sprintf("Could Not Provide Public Key To IP. Error: %v", keyData)
		h.UserConfingLogger.Debug(logentry)
	}

	logentry := "Successfully Provided Public Key to a Client "
	h.UserConfingLogger.Info(logentry)

	res.PublicKey = []byte(keyData)
	return nil
}

func (h *ClientHandler) InitNewWorkSpaceConnection(req InitWorkspaceConnectionRequest, res *InitWorkspaceConnectionResponse) error {
	// 1. Decrypt password [X]
	// 2. Authenticate Request [X]
	// 3. Add the New Connection to the .PKr Config File [X]
	// 4. Store the Public Key [X]
	// 5. Send the Response with port [X]
	// 6. Open a Data Transfer Port and shit [Will be a separate Function not here] [X]

	h.UserConfingLogger.Info("Init New Work Space Connection Called ...")
	log.Println("Init New Work Space Connection Called ...")

	password, err := encrypt.DecryptData(req.WorkspacePassword)
	if err != nil {
		h.UserConfingLogger.Debug(fmt.Sprintf("Failed to Init Workspace Connection for User - %s from Server %s: ", req.MyUsername, req.ServerIP))
		h.UserConfingLogger.Debug(err)

		res.Response = 4000
		return nil
	}

	fmt.Println("Decrypted Data ...\nAuthenticating ...")

	// Authenticates Workspace Name and Password and Get the Workspace File Path
	file_path, err := config.AuthenticateWorkspaceInfo(req.WorkspaceName, password)
	if err != nil {
		if errors.Is(err, ErrIncorrectPassword) {
			h.WorkspaceLogger.Debug(req.WorkspaceName, fmt.Sprintf("Incorrect Credentials for Workspace - %s, By User - %s, Server: %s", req.MyUsername, req.MyUsername, req.ServerIP))
		} else {
			h.UserConfingLogger.Debug(fmt.Sprintf("could not init workspace for user %s, server %s\nError:%v", req.MyUsername, req.ServerIP, err.Error()))
		}

		res.Response = 4000
		return nil
	}

	fmt.Println("Auth Successfull ...\nGetting Server Details Using Server IP")

	server, err := config.GetServerDetailsUsingServerIP(req.ServerIP)
	if err != nil {
		if errors.Is(err, ErrServerNotFound) {
			h.WorkspaceLogger.Debug(req.WorkspaceName, fmt.Sprintf("Unable to find Server with such IP: %s", req.ServerIP))
		} else {
			h.UserConfingLogger.Debug(fmt.Sprintf("could not init workspace for user %s, server %s\nError: %v", req.MyUsername, req.ServerIP, err))
		}

		res.Response = 4000
		return nil
	}

	fmt.Println("Getting Server Details Using Server IP Successfull ...\n Decoding Public Keys ...")

	var connection config.Connection
	connection.Username = req.MyUsername
	connection.ServerAlias = server.ServerAlias

	// Save Public Key
	publicKey, err := base64.StdEncoding.DecodeString(string(req.MyPublicKey))
	if err != nil {
		h.WorkspaceLogger.Debug(req.WorkspaceName, "Failed to convert key to Base64 for User: "+req.MyUsername)
		h.WorkspaceLogger.Debug(req.WorkspaceName, err)

		res.Response = 4000
		return err
	}

	fmt.Println("Storing Public Keys ...")

	keysPath, err := config.StorePublicKeys(file_path+"\\.PKr\\keys\\", string(publicKey))
	if err != nil {
		h.WorkspaceLogger.Debug(req.WorkspaceName, "Failed to Init Workspace Connection for User: "+req.MyUsername)
		h.WorkspaceLogger.Debug(req.WorkspaceName, err)

		res.Response = 4000
		return err
	}

	fmt.Println("Adding New Conn in Pkr Config File")

	// Store the New Connection in the .PKr Config file
	connection.PublicKeyPath = keysPath
	if err := config.AddConnectionToPKRConfigFile(file_path+"\\.PKr\\workspaceConfig.json", connection); err != nil {
		h.WorkspaceLogger.Debug(req.WorkspaceName, "Failed to Init Workspace Connection for User IP: "+req.MyUsername)
		h.WorkspaceLogger.Debug(req.WorkspaceName, err)

		res.Response = 4000
		return nil
	}

	fmt.Println("Added New Conn in Pkr Config file")
	fmt.Println("Init New Workspace Conn Successful")
	// models.AddLogEntry(request.WorkspaceName, fmt.Sprintf("Added User with IP: %v to the Connection List", ip))
	h.WorkspaceLogger.Info(req.WorkspaceName, fmt.Sprintf("Added User IP: %v of Server %s to the Connection List", req.MyUsername, server.ServerAlias))

	// TODO The Client Will make another new Request Entirely through the entire server process to retrieve the workspace data
	// TODO Create a RPC Reciever for GetData -> if Last Hash "" than send entire zip
	res.Response = 200
	return nil
}

func (h *ClientHandler) GetData(req GetDataRequest, res *GetDataResponse) error {
	// FIXME AUTH req.workspace_name, req.workspace_password
	// FIXME Store Keys when called GetPublicKey in cli and reuse it

	// TODO Compare Hash ...
	// TODO - If Hash == "" : Send Entire File
	// TODO - If Hash == Last_Hash : Do Nothing

	// TODO Maintain 3 Last Hash object files, and Send Files Accordingly

	fmt.Printf("Data Requested For Workspace: %s\n", req.WorkspaceName)
	h.WorkspaceLogger.Info(req.WorkspaceName, fmt.Sprintf("Data Requested For Workspace: %s", req.WorkspaceName))
	workspacePath, err := config.GetSendWorkspaceFilePath(req.WorkspaceName)
	if err != nil {
		log_entry := fmt.Sprintf("cannot get workspace's file path\nError: %s\nSource: PullData() Handler", err.Error())
		h.WorkspaceLogger.Debug(req.WorkspaceName, log_entry)
		return err
	}

	zipped_file_name, err := filetracker.ZipData(workspacePath)
	zipped_hash := strings.Split(zipped_file_name, ".")[0]
	fmt.Println("Data Service Hash File Name: " + zipped_file_name)

	if err != nil {
		logdata := fmt.Sprintf("Could Not Zip The File\nError: %v", err)
		h.WorkspaceLogger.Critical(req.WorkspaceName, logdata)
		return err
	}

	err = config.AddNewPushToConfig(req.WorkspaceName, zipped_hash)
	if err != nil {
		logdata := fmt.Sprintf("could add entry to PKR config file.\nError: %v", err)
		h.WorkspaceLogger.Critical(req.WorkspaceName, logdata)
		return err
	}
	config.AddLogEntry(req.WorkspaceName, true, "Workspace Zipped")

	key, err := encrypt.AESGenerakeKey(16)
	if err != nil {
		logdata := fmt.Sprintf("Could Not Generate AES Key\nError: %v", err)
		h.WorkspaceLogger.Critical(req.WorkspaceName, logdata)
		return err
	}

	iv, err := encrypt.AESGenerateIV()
	if err != nil {
		logdata := fmt.Sprintf("Could Not Generate AES IV\nError: %v", err)
		h.WorkspaceLogger.Critical(req.WorkspaceName, logdata)
		return err
	}

	h.WorkspaceLogger.Info(req.WorkspaceName, "AES Keys Generated")

	zipped_filepath := workspacePath + "\\.PKr\\" + zipped_file_name
	destination_filepath := strings.Replace(zipped_filepath, ".zip", ".enc", 1)
	if err := encrypt.AESEncrypt(zipped_filepath, destination_filepath, key, iv); err != nil {
		logdata := fmt.Sprintf("Could Not Encrypt File\nError: %v\nFilePath: %v", err, zipped_filepath)
		h.WorkspaceLogger.Critical(req.WorkspaceName, logdata)
		return err
	}

	h.WorkspaceLogger.Info(req.WorkspaceName, "Zip AES is Encrypted")

	publicKeyPath, err := config.GetConnectionsPublicKeyUsingUsername(workspacePath, req.Username)
	if err != nil {
		logdata := fmt.Sprintf("Could Not Find Users Public Key\nError: %v", err)
		h.WorkspaceLogger.Critical(req.WorkspaceName, logdata)
		return err
	}

	publicKey, err := os.ReadFile(publicKeyPath)
	if err != nil {
		logdata := fmt.Sprintf("Could Not Read Users Public Key\nError: %v", err)
		h.WorkspaceLogger.Critical(req.WorkspaceName, logdata)
		return err
	}

	encrypt_key, err := encrypt.EncryptData(string(key), string(publicKey))
	if err != nil {
		logdata := fmt.Sprintf("Could Not Encrypt Key\nError: %v", err)
		h.WorkspaceLogger.Critical(req.WorkspaceName, logdata)
		return err
	}

	encrypt_iv, err := encrypt.EncryptData(string(iv), string(publicKey))
	if err != nil {
		logdata := fmt.Sprintf("Could Not Encrypt IV\nError: %v", err)
		h.WorkspaceLogger.Critical(req.WorkspaceName, logdata)
		return err
	}

	encrypt_file, err := os.Open(destination_filepath)
	if err != nil {
		logdata := fmt.Sprintf("Could Not Open The Encrypted File\nError: %v", err)
		h.WorkspaceLogger.Critical(req.WorkspaceName, logdata)
		return err
	}
	defer encrypt_file.Close()
	h.WorkspaceLogger.Info(req.WorkspaceName, "AES Keys Encrypted")

	h.WorkspaceLogger.Info(req.WorkspaceName, "Sending Data...")
	fileData, err := ioutil.ReadAll(encrypt_file)
	if err != nil {
		// Client Error
		logdata := fmt.Sprintf("Could Not Read File.\nError: %v", err)
		h.WorkspaceLogger.Critical(req.WorkspaceName, logdata)
		res.Response = 500
		return err
	}

	res.Response = 200
	res.NewHash = zipped_hash
	res.KeyBytes = []byte(encrypt_key)
	res.IVBytes = []byte(encrypt_iv)
	res.Data = fileData

	fmt.Println("Get Data Done")
	return nil
}
