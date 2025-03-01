package services

import "github.com/ButterHost69/PKr-Base/logger"

// Handles Requests from other Clients
type ClientHandler struct {
	WorkspaceLogger   *logger.WorkspaceLogger
	UserConfingLogger *logger.UserLogger
}

type PublicKeyRequest struct {
}

type PublicKeyResponse struct {
	PublicKey []byte
}

type InitWorkspaceConnectionRequest struct {
	WorkspaceName string
	MyUsername    string
	MyPublicKey   []byte

	ServerIP          string
	WorkspacePassword string
}

type InitWorkspaceConnectionResponse struct {
	Response int32 // 200 [Valid / ACK / OK] ||| 4000 [InValid / You Fucked Up Somewhere]
}

type GetDataRequest struct {
	WorkspaceName     string
	WorkspacePassword string

	Username string
	ServerIP string

	LastHash string
}

type GetDataResponse struct {
	Response int

	NewHash string
	Data    []byte

	KeyBytes []byte
	IVBytes  []byte
}
