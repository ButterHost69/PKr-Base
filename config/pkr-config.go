package config

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
)

const (
	WORKSPACE_PKR_DIR = ".PKr"
)

var (
	LOGS_PKR_FILE_PATH         = filepath.Join(WORKSPACE_PKR_DIR, "logs.txt")
	WORKSPACE_CONFIG_FILE_PATH = filepath.Join(WORKSPACE_PKR_DIR, "workspaceConfig.json")
)

func CreatePKRConfigIfNotExits(workspace_name string, workspace_file_path string) error {
	pkr_config_file_path := filepath.Join(workspace_file_path, WORKSPACE_CONFIG_FILE_PATH)
	if _, err := os.Stat(pkr_config_file_path); os.IsExist(err) {
		fmt.Println("~ workspaceConfig.jso already Exists")
		return err
	}

	pkrconf := PKRConfig{
		WorkspaceName: workspace_name,
	}

	jsonBytes, err := json.Marshal(pkrconf)
	if err != nil {
		fmt.Println("~ Unable to Parse PKrConfig to JSON")
		return err
	}

	// Creating Workspace Config File
	err = os.WriteFile(pkr_config_file_path, jsonBytes, 0777)
	if err != nil {
		fmt.Println("~ Unable to Write PKrConfig to File")
		return err
	}

	return nil
}

func AddConnectionToPKRConfigFile(workspace_config_path string, connection Connection) error {
	pkrConfig, err := ReadFromPKRConfigFile(workspace_config_path)
	if err != nil {
		return err
	}

	pkrConfig.AllConnections = append(pkrConfig.AllConnections, connection)
	if err := writeToPKRConfigFile(workspace_config_path, pkrConfig); err != nil {
		return err
	}

	return nil
}

func GetConnectionsPublicKeyUsingUsername(workspace_path, username string) (string, error) {
	pkrconfig, err := ReadFromPKRConfigFile(filepath.Join(workspace_path, WORKSPACE_CONFIG_FILE_PATH))
	if err != nil {
		return "", err
	}

	for _, connection := range pkrconfig.AllConnections {
		if connection.Username == username {
			return connection.PublicKeyPath, nil
		}
	}

	return "", fmt.Errorf("no such ip exists : %v", username)
}

func StorePublicKeys(username, workspace_keys_path string, key string) (string, error) {
	keyPath := filepath.Join(workspace_keys_path, username+".pem")
	file, err := os.OpenFile(keyPath, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0600)
	if err != nil {
		return "", err
	}
	defer file.Close()

	_, err = file.Write([]byte(key))
	if err != nil {
		return "", err
	}

	fullpath, err := filepath.Abs(keyPath)
	if err != nil {
		return keyPath, nil
	}

	return fullpath, nil
}

func ReadFromPKRConfigFile(workspace_config_path string) (PKRConfig, error) {
	file, err := os.Open(workspace_config_path)
	if err != nil {
		log.Println("error in opening PKR config file.... pls check if .PKr/workspaceConfig.json available ")
		// AddUsersLogEntry("error in opening PKR config file.... pls check if .PKr/workspaceConfig.json available ")
		return PKRConfig{}, err
	}
	defer file.Close()

	var pkrConfig PKRConfig
	decoder := json.NewDecoder(file)
	err = decoder.Decode(&pkrConfig)
	if err != nil {
		log.Println("error in decoding json data")
		// AddUsersLogEntry("error in decoding json data")
		return PKRConfig{}, err
	}

	// fmt.Println(pkrConfig)
	return pkrConfig, nil
}

func writeToPKRConfigFile(workspace_config_path string, newPKRConfing PKRConfig) error {
	jsonData, err := json.MarshalIndent(newPKRConfing, "", "	")
	// fmt.Println(jsonData)
	if err != nil {
		fmt.Println("error occured in Marshalling the data to JSON")
		fmt.Println(err)
		return err
	}

	// fmt.Println(string(jsonData))
	err = os.WriteFile(workspace_config_path, jsonData, 0777)
	if err != nil {
		fmt.Println("error occured in storing data in userconfig file")
		fmt.Println(err)
		return err
	}

	return nil
}

// Logs Entry of all the events occurred related to the workspace
// Also Creates the Log File by default
func AddLogEntry(workspace_name string, isSendWorkspace bool, log_entry any) error {
	var workspace_path string
	var err error
	if isSendWorkspace {
		workspace_path, err = GetSendWorkspaceFilePath(workspace_name)
		if err != nil {
			return err
		}
	} else {
		workspace_path, err = GetGetWorkspaceFilePath(workspace_name)
		if err != nil {
			return err
		}
	}

	// Adds the ".Pkr/logs.txt"
	workspace_path += "\\" + LOGS_PKR_FILE_PATH

	// Opens or Creates the Log File
	file, err := os.OpenFile(workspace_path, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0600)
	if err != nil {
		return err
	}

	defer file.Close()
	log.SetOutput(file)
	log.Println(log_entry)
	// log.Println(log_entry, log.LstdFlags)

	return nil
}

func UpdateLastPushNum(workspace_name string, last_push_num int) error {
	workspace_path, err := GetSendWorkspaceFilePath(workspace_name)
	if err != nil {
		return err
	}

	workspace_path = filepath.Join(workspace_path, WORKSPACE_CONFIG_FILE_PATH)
	workspace_json, err := ReadFromPKRConfigFile(workspace_path)
	if err != nil {
		return fmt.Errorf("could not read from config file.\nError: %v", err)
	}

	workspace_json.LastPushNum = last_push_num
	if err := writeToPKRConfigFile(workspace_path, workspace_json); err != nil {
		return fmt.Errorf("error in writing the update push num to file: %s.\nError: %v", workspace_path, err)
	}
	return nil
}

func ReadPublicKey() (string, error) {
	keyData, err := os.ReadFile(filepath.Join(MY_KEYS_PATH, "publickey.pem"))
	if err != nil {
		return "", err
	}

	return string(keyData), nil
}

func GetWorkspaceConnectionsUsingPath(workspace_path string) ([]Connection, error) {
	workspace_json, err := ReadFromPKRConfigFile(workspace_path)
	if err != nil {
		return []Connection{}, fmt.Errorf("could not read from config file.\nError: %v", err)
	}

	return workspace_json.AllConnections, nil
}

func AppendWorkspaceUpdates(updates Updates, workspace_path string) error {
	workspace_config_path := filepath.Join(workspace_path, WORKSPACE_CONFIG_FILE_PATH)
	workspace_json, err := ReadFromPKRConfigFile(workspace_config_path)
	if err != nil {
		return fmt.Errorf("could not read from config file.\nError: %v", err)
	}

	workspace_json.AllUpdates = append(workspace_json.AllUpdates, updates)
	if err := writeToPKRConfigFile(workspace_config_path, workspace_json); err != nil {
		return fmt.Errorf("error in writing new update to workspace_conf\nError: %v", err)
	}
	return nil
}

func MergeUpdates(workspace_path string, start_push_num, end_push_num int) ([]FileChange, error) {
	workspace_conf, err := ReadFromPKRConfigFile(filepath.Join(workspace_path, WORKSPACE_CONFIG_FILE_PATH))
	if err != nil {
		return nil, fmt.Errorf("could not read from config file.\nError: %v", err)
	}

	updates_list := make(map[string]FileChange)
	for i := start_push_num + 1; i <= end_push_num; i++ {
		for _, change := range workspace_conf.AllUpdates[i].Changes {
			update, exists := updates_list[change.FilePath]
			if exists {
				if update.Type == "Updated" && change.Type == "Removed" {
					delete(updates_list, change.FilePath)
				}
			} else {
				updates_list[change.FilePath] = FileChange{
					FilePath: change.FilePath,
					FileHash: change.FileHash,
					Type:     change.Type,
				}
			}
		}
	}

	merged_changes := []FileChange{}
	for _, hash_type := range updates_list {
		merged_changes = append(merged_changes, hash_type)
	}
	return merged_changes, nil
}
