package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/ButterHost69/PKr-Base/encrypt"
)

var (
	USER_CONFIG_ROOT_DIR = filepath.Join(os.Getenv("LOCALAPPDATA"), "PKr")
	MY_KEYS_PATH         = filepath.Join(USER_CONFIG_ROOT_DIR, "Config", "Keys", "My")
	OTHERS_KEYS_PATH     = filepath.Join(USER_CONFIG_ROOT_DIR, "Config", "Keys", "Others")
	USER_CONFIG_FILE     = filepath.Join(USER_CONFIG_ROOT_DIR, "Config", "user-config.json")
)

func CreateUserConfigIfNotExists(username, password, server_ip string) error {
	_, err := os.Stat(USER_CONFIG_FILE)
	if err == nil {
		fmt.Println("It Seems PKr is Already Installed...")
		return nil
	} else if os.IsNotExist(err) {
		fmt.Println("Creating user-config.json ...")
	} else {
		fmt.Println("Error while checking Existence of user-config file:", err)
		fmt.Println("Source: CreateUserConfigIfNotExists()")
		return err
	}

	// Creating %LOCALAPPDATA%/.PKr/Config/Keys/My
	err = os.MkdirAll(filepath.Join(USER_CONFIG_ROOT_DIR, "Config", "Keys", "My"), 0600)
	if err != nil {
		fmt.Println("Error while Creating %LOCALAPPDATA%/.PKr/Config/Keys/My Dir:", err)
		fmt.Println("Source: CreateUserConfigIfNotExists()")
		return err
	}

	// Creating %LOCALAPPDATA%/.PKr/Config/Keys/Others
	err = os.MkdirAll(filepath.Join(USER_CONFIG_ROOT_DIR, "Config", "Keys", "Others"), 0600)
	if err != nil {
		fmt.Println("Error while Creating %LOCALAPPDATA%/.PKr/Config/Keys/Others Dir:", err)
		fmt.Println("Source: CreateUserConfigIfNotExists()")
		return err
	}

	user_conf := UserConfig{
		Username: username,
		Password: password,
		ServerIP: server_ip,
	}

	conf_bytes, err := json.Marshal(user_conf)
	if err != nil {
		fmt.Println("Error while Parsing user-config:", err)
		fmt.Println("Source: CreateUserConfigIfNotExists()")
		return err
	}

	err = os.WriteFile(USER_CONFIG_FILE, conf_bytes, 0600)
	if err != nil {
		fmt.Println("Error while Writing in user-config:", err)
		fmt.Println("Source: CreateUserConfigIfNotExists()")
		return err
	}

	private_key, public_key := encrypt.GenerateRSAKeys()
	if private_key == nil && public_key == nil {
		panic("Could Not Generate Keys")
	}

	if err = encrypt.StorePrivateKeyInFile(filepath.Join(MY_KEYS_PATH, "private.pem"), private_key); err != nil {
		fmt.Println("Error while Storing My Private Key:", err)
		fmt.Println("Source: CreateUserConfigIfNotExists()")
		return err
	}

	if err = encrypt.StorePublicKeyInFile(filepath.Join(MY_KEYS_PATH, "public.pem"), public_key); err != nil {
		fmt.Println("Error while Storing My Public Key:", err)
		fmt.Println("Source: CreateUserConfigIfNotExists()")
		return err
	}
	return nil
}

func ReadFromUserConfigFile() (UserConfig, error) {
	file, err := os.Open(USER_CONFIG_FILE)
	if err != nil {
		fmt.Println("Error while opening user-config file:", err)
		fmt.Println("Source: ReadFromUserConfigFile()")
		return UserConfig{}, err
	}
	defer file.Close()

	var user_conf UserConfig
	decoder := json.NewDecoder(file)
	err = decoder.Decode(&user_conf)
	if err != nil {
		fmt.Println("Error while Decoding JSON Data from user-config file:", err)
		fmt.Println("Source: ReadFromUserConfigFile()")
		return UserConfig{}, err
	}

	return user_conf, nil
}

func writeToUserConfigFile(new_user_conf UserConfig) error {
	jsonData, err := json.MarshalIndent(new_user_conf, "", "	")
	if err != nil {
		fmt.Println("Error while Marshalling the user-conf to JSON:", err)
		fmt.Println("Source: writeToUserConfigFile()")
		return err
	}

	err = os.WriteFile(USER_CONFIG_FILE, jsonData, 0600)
	if err != nil {
		fmt.Println("Error while writing data in user-config file", err)
		fmt.Println("Source: writeToUserConfigFile()")
		return err
	}
	return nil
}

// Send Workspaces are workspaces you create
// This workspaces will be broadcasted to other users
func RegisterNewSendWorkspace(workspace_name, workspace_path, workspace_password string) error {
	user_conf, err := ReadFromUserConfigFile()
	if err != nil {
		fmt.Println("Error while reading from the user-config file:", err)
		fmt.Println("Source: RegisterNewSendWorkspace()")
		return err
	}

	new_send_workspace := SendWorkspaceFolder{
		WorkspaceName:     workspace_name,
		WorkspacePath:     workspace_path,
		WorkSpacePassword: workspace_password,
	}

	user_conf.SendWorkspaces = append(user_conf.SendWorkspaces, new_send_workspace)
	if err := writeToUserConfigFile(user_conf); err != nil {
		fmt.Println("Error while Writing in the user-config file:", err)
		fmt.Println("Source: RegisterNewSendWorkspace()")
		return err
	}
	return nil
}

func RegisterNewGetWorkspace(workspace_name, workspace_owner_name, workspace_path, workspace_password string, last_push_num int) error {
	user_conf, err := ReadFromUserConfigFile()
	if err != nil {
		fmt.Println("Error in reading From the UserConfig File...")
		fmt.Println("Source: RegisterNewGetWorkspace()")
		return err
	}

	new_get_workspace := GetWorkspaceFolder{
		WorkspaceOwnerName: workspace_owner_name,
		WorkspaceName:      workspace_name,
		WorkspacePath:      workspace_path,
		WorkspacePassword:  workspace_password,
		LastPushNum:        last_push_num,
	}

	user_conf.GetWorkspaces = append(user_conf.GetWorkspaces, new_get_workspace)
	if err := writeToUserConfigFile(user_conf); err != nil {
		fmt.Println("Error while Writing in the user-config file:", err)
		fmt.Println("Source: RegisterNewGetWorkspace()")
		return err
	}
	return nil
}

func GetGetWorkspaceFilePath(workspace_name string) (string, error) {
	user_conf, err := ReadFromUserConfigFile()
	if err != nil {
		fmt.Println("Error while Reading from user-config file:", err)
		fmt.Println("Source: GetGetWorkspaceFilePath()")
		return "", err
	}

	for _, workspace := range user_conf.GetWorkspaces {
		if workspace.WorkspaceName == workspace_name {
			return workspace.WorkspacePath, nil
		}
	}
	return "", errors.New("no such workspace found")
}

func GetSendWorkspaceFilePath(workspace_name string) (string, error) {
	user_conf, err := ReadFromUserConfigFile()
	if err != nil {
		fmt.Println("Error while Reading from user-config file:", err)
		fmt.Println("Source: GetSendWorkspaceFilePath()")
		return "", err
	}

	for _, workspace := range user_conf.SendWorkspaces {
		if workspace.WorkspaceName == workspace_name {
			return workspace.WorkspacePath, nil
		}
	}
	return "", errors.New("no such workspace found")
}

// Returns Workspace Path if Username and Password Correct
func AuthenticateWorkspaceInfo(workspace_name string, workspace_password string) (string, error) {
	user_conf, err := ReadFromUserConfigFile()
	if err != nil {
		fmt.Println("Error while Reading from user-config file:", err)
		fmt.Println("Source: AuthenticateWorkspaceInfo()")
		return "", err
	}

	for _, workspace := range user_conf.SendWorkspaces {
		if workspace.WorkspaceName == workspace_name {
			if workspace.WorkSpacePassword == workspace_password {
				return workspace.WorkspacePath, nil
			}
			return "", errors.New("incorrect password")
		}
	}
	return "", errors.New("no such workspace found")
}

// Update Last Push Num (Used during Pulls)
func UpdateLastPushNumInGetWorkspaceFolderToUserConfig(workspace_name string, last_push_num int) error {
	user_conf, err := ReadFromUserConfigFile()
	if err != nil {
		fmt.Println("Error while reading user-config File:", err)
		fmt.Println("Source: UpdateLastPushNumInGetWorkspaceFolderToUserConfig()")
		return err
	}

	for idx, workspace := range user_conf.GetWorkspaces {
		if workspace.WorkspaceName == workspace_name {
			user_conf.GetWorkspaces[idx].LastPushNum = last_push_num
			break
		}
	}

	if err := writeToUserConfigFile(user_conf); err != nil {
		fmt.Println("Error while writing in user-config File:", err)
		fmt.Println("Source: UpdateLastPushNumInGetWorkspaceFolderToUserConfig()")
		return err
	}
	return nil
}

func GetPublicKeyUsingUsername(username string) ([]byte, error) {
	public_key_path := filepath.Join(OTHERS_KEYS_PATH, username+".pem")
	public_key, err := os.ReadFile(public_key_path)
	if err != nil {
		fmt.Println("Error while Reading Public Key of Other User:", err)
		fmt.Println("Source: GetPublicKeyUsingUsername()")
		return nil, err
	}
	return public_key, nil
}
