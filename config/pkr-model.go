package config

type PKRConfig struct {
	WorkspaceName string    `json:"workspace_name"`
	LastPushNum   int       `json:"last_push_num"`
	AllUpdates    []Updates `json:"all_updates"`
}

type FileChange struct {
	FilePath string `json:"file_path"`
	FileHash string `json:"file_hash"`
	Type     string `json:"type"` //Updated ; Removed [Created and Updated are the same -> When File Created -> Tag as Updated]
	// Because They will be replaced Anyway
}

type Updates struct {
	PushNum  int          `json:"push_num"`
	PushDesc string       `json:"push_desc"`
	Changes  []FileChange `json:"file_change"`
}
