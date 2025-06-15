package config

type PKRConfig struct {
	WorkspaceName  	string       	`json:"workspace_name"`
	AllConnections 	[]Connection 	`json:"all_connections"`
	LastHash       	string       	`json:"last_hash"`
	AllUpdates		[]Updates		`json:"all_updates"`
}

type FileChange struct {
	FilePath	string	`json:"file_path"`
	Type		string	`json:"type"` //Created ; Updated ; Removed
}
type Updates struct {
	Hash	string		`json:"hash"`
	Changes	[]FileChange	`json:"file_change"`
}

type Connection struct {
	ServerAlias   string `json:"server_alias"`
	Username      string `json:"username"`
	PublicKeyPath string `json:"public_key_path"`
}
