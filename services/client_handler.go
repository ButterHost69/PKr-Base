package services

import (
	"encoding/base64"
	"errors"
	"fmt"

	"github.com/ButterHost69/PKr-Base/config"
	"github.com/ButterHost69/PKr-Base/encrypt"
)

var (
	ErrIncorrectPassword = errors.New("incorrect password")
)

func (h *Handler) GetPublicKey(req PublicKeyRequest, res *PublicKeyResponse) error {
	keyData, err := config.ReadPublicKey()
	if err != nil {
		logentry := fmt.Sprintf("Could Not Provide Public Key To IP. Error: %v", keyData)
		h.UserConfingLogger.Debug(logentry)
	}
	
	logentry := "Successfully Provided Public Key to a Client "
	h.UserConfingLogger.Info(logentry)

	res.PublicKey = []byte(keyData)
	return nil
}

func (h *Handler) InitNewWorkSpaceConnection(req InitWorkspaceConnectionRequest, res *InitWorkspaceConnectionResponse) (error) {
	// 1. Decrypt password [X]
	// 2. Authenticate Request [X]
	// 3. Add the New Connection to the .PKr Config File [X]
	// 4. Store the Public Key [X]
	// 5. Send the Response with port [X]
	// 6. Open a Data Transfer Port and shit [Will be a separate Function not here] [X]

	password, err := encrypt.DecryptData(req.WorkspacePassword)
	if err != nil {
		h.UserConfingLogger.Debug(fmt.Sprintf("Failed to Init Workspace Connection for User - %s from Server %s: ", req.MyUsername, req.ServerAlias))
		h.UserConfingLogger.Debug(err)

		res.Response = 4000 
		return nil
	}

	// Authenticates Workspace Name and Password and Get the Workspace File Path
	file_path, err := config.AuthenticateWorkspaceInfo(req.WorkspaceName, password)
	if err != nil {
		if errors.Is(err, ErrIncorrectPassword) {
			h.WorkspaceLogger.Debug(req.WorkspaceName, fmt.Sprintf("Incorrect Credentials for Workspace - %s, By User - %s, Server: %s", ))
		} else{
			h.UserConfingLogger.Debug(fmt.Sprintf("could not init workspace for user %s, server %s ", req.MyUsername, req.ServerAlias))
		}

		res.Response = 4000
		return nil
	}

	var connection config.Connection
	connection.Username = req.MyUsername
	connection.ServerAlias = req.ServerAlias

	// Save Public Key
	publicKey, err := base64.StdEncoding.DecodeString(string(req.MyPublicKey))
	if err != nil {
		h.WorkspaceLogger.Debug(req.WorkspaceName, "Failed to convert key to Base64 for User: "+ req.MyUsername)
		h.WorkspaceLogger.Debug(req.WorkspaceName, err)
		
		res.Response = 4000
		return err
	}

	keysPath, err := config.StorePublicKeys(file_path+"\\.PKr\\keys\\", string(publicKey))
	if err != nil {
		h.WorkspaceLogger.Debug(req.WorkspaceName, "Failed to Init Workspace Connection for User: " + req.MyUsername)
		h.WorkspaceLogger.Debug(req.WorkspaceName, err)
		
		res.Response = 4000
		return err
	}

	// Store the New Connection in the .PKr Config file
	connection.PublicKeyPath = keysPath
	if err := config.AddConnectionToPKRConfigFile(file_path+"\\.PKr\\workspaceConfig.json", connection); err != nil {
		h.WorkspaceLogger.Debug(req.WorkspaceName, "Failed to Init Workspace Connection for User IP: " + req.MyUsername)
		h.WorkspaceLogger.Debug(req.WorkspaceName, err)
		
		res.Response = 4000
		return nil
	}
	// models.AddLogEntry(request.WorkspaceName, fmt.Sprintf("Added User with IP: %v to the Connection List", ip))
	h.WorkspaceLogger.Info(req.WorkspaceName, fmt.Sprintf("Added User IP: %v of Server %s to the Connection List", req.MyUsername, req.ServerAlias))

	// TODO The Client Will make another new Request Entirely through the entire server process to retrieve the workspace data
	// TODO Create a RPC Reciever for GetData -> if Last Hash "" than send entire zip 
	res.Response = 200
	return nil
}