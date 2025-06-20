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

	LastPushNum int // -1 => Cloning for First Time
}

type GetMetaDataResponse struct {
	LenData  int
	KeyBytes []byte
	IVBytes  []byte

	Updates          map[string]string // {"fileA": "Updated", "fileB": "Removed"}, Updated & Created're treated equally
	RequestPushRange string            // "Request Push Range" is to be sent during GetData, i.e., "2-5", Push2 to Push5
	LastPushNum      int               // Latest Push Num of the Entire Workspace
	LastPushDesc     string            // Latest Push Desc of the Entire Workspace
}
