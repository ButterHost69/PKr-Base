package models

type PublicKeyRequest struct{}

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

type InitWorkspaceConnectionResponse struct{}

type GetDataRequest struct {
	WorkspaceName     string
	WorkspacePassword string

	Username string
	ServerIP string

	LastHash string
}

type GetDataResponse struct {
	NewHash string
	Data    []byte

	KeyBytes []byte
	IVBytes  []byte
}
