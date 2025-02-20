package config

type PKRConfig struct {
	WorkspaceName  string       `json:"workspace_name"`
	AllConnections []Connection `json:"all_connections"`
	LastHash       string       `json:"last_hash"`
}

type Connection struct {
	ServerAlias   string `json:"server_alias"`
	Username      string `json:"username"`
	PublicKeyPath string `json:"public_key_path"`
}
