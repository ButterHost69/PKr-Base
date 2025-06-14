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

type GetMetaDataRequest struct {
	WorkspaceName     string
	WorkspacePassword string

	Username string
	ServerIP string

	LastHash string
}

type GetMetaDataResponse struct {
	LenData  int
	KeyBytes []byte
	IVBytes  []byte
	NewHash  string
}
