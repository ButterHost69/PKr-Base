package services

import "github.com/ButterHost69/PKr-Base/logger"

// Handles Requests from other Clients
type Handler struct {
	WorkspaceLogger   *logger.WorkspaceLogger
	UserConfingLogger *logger.UserLogger
}

type PublicKeyRequest struct {
}

type PublicKeyResponse struct {
	PublicKey []byte
}

type InitWorkspaceConnectionRequest struct {
	WorkspaceName     string
	MyUsername        string
	MyPublicKey       []byte

	ServerAlias       string
	WorkspacePassword string
}

type InitWorkspaceConnectionResponse struct {
	Response int32 // 200 [Valid / ACK / OK] ||| 4000 [InValid / You Fucked Up Somewhere]
}
