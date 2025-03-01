package config

import (
	// "ButterHost69/PKr-client/encrypt"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"

	"github.com/ButterHost69/PKr-cli/encrypt"
	// "github.com/go-delve/delve/cmd/dlv/cmds"
)

const (
	ROOT_DIR     = "tmp"
	MY_KEYS_PATH = ROOT_DIR + "\\mykeys"
	CONFIG_FILE  = ROOT_DIR + "\\userConfig.json"
	LOG_FILE     = ROOT_DIR + "\\logs.txt"
)

var (
	MY_USERNAME string
)

// FIXME: NOT IMPORTANT : Remove Prints - return stuff

func CreateUserIfNotExists() {
	if _, err := os.Stat(ROOT_DIR + "/userConfig.json"); os.IsNotExist(err) {
		fmt.Println("!! 'tmp' No such DIR exists ")

		usconf := UsersConfig{
			User: "temporary",
		}

		jsonbytes, err := json.Marshal(usconf)
		if err != nil {
			fmt.Println("~ Unable to Parse Username to Json")
		}

		if err = os.Mkdir(ROOT_DIR, 0777); err != nil {
			fmt.Println("~ Folder tmp exists")
		}
		err = os.WriteFile(ROOT_DIR+"/userConfig.json", jsonbytes, 0777)
		if err != nil {
			log.Fatal(err.Error())
		}

		if err = os.Mkdir(MY_KEYS_PATH, 0777); err != nil {
			fmt.Println("~ Folder tmp exists")
		}

		private_key, public_key := encrypt.GenerateRSAKeys()
		if private_key == nil && public_key == nil {
			panic("Could Not Generate Keys")
		}

		if err = encrypt.StorePrivateKeyInFile(MY_KEYS_PATH+"/privatekey.pem", private_key); err != nil {
			panic(err.Error())
		}

		if err = encrypt.StorePublicKeyInFile(MY_KEYS_PATH+"/publickey.pem", public_key); err != nil {
			panic(err.Error())
		}

		return
	}

	fmt.Println("It Seems PKr is Already Installed...")
}

// Send Workspaces are workspaces you create
// This workspaces will be broadcasted to other users
func RegisterNewSendWorkspace(server_alias, workspace_name, workspace_path, workspace_password string) error {
	userConfig, err := ReadFromUserConfigFile()
	if err != nil {
		fmt.Println("Error in reading From the UserConfig File...")
		return err
	}

	workspaceFolder := SendWorkspaceFolder{
		WorkspaceName:     workspace_name,
		WorkspacePath:     workspace_path,
		WorkSpacePassword: workspace_password,
	}

	for idx, server := range userConfig.ServerLists {
		if server.ServerAlias == server_alias {
			userConfig.ServerLists[idx].SendWorkspaces = append(userConfig.ServerLists[idx].SendWorkspaces, workspaceFolder)
			if err := writeToUserConfigFile(userConfig); err != nil {
				fmt.Println("Error Occured in Writing To the UserConfig File")
				return err
			}
			return nil
		}
	}

	fmt.Println("No Such Server Alias Exists...")
	return nil
}

func GetWorkspaceFilePath(workspace_name string) (string, error) {
	userConfig, err := ReadFromUserConfigFile()
	if err != nil {
		return "", err
	}

	servers := userConfig.ServerLists
	for _, server := range servers {
		for _, workspace := range server.GetWorkspaces {
			if workspace.WorkspaceName == workspace_name {
				return workspace.WorkspacePath, nil
			}
		}
	}

	return "", errors.New("no such workspace found")
}

// Returns Workspace Path if Username and Password Correct
func AuthenticateWorkspaceInfo(workspace_name string, workspace_password string) (string, error) {
	userConfig, err := ReadFromUserConfigFile()
	if err != nil {
		return "", err
	}

	fmt.Println("User Config Fetched ...")
	fmt.Println(userConfig)
	fmt.Println(userConfig.ServerLists)
	fmt.Println(userConfig.User)

	servers := userConfig.ServerLists
	for _, server := range servers {
		for _, workspace := range server.SendWorkspaces {
			if workspace.WorkspaceName == workspace_name {
				if workspace.WorkSpacePassword == workspace_password {
					return workspace.WorkspacePath, nil
				}
				return "", errors.New("incorrect password")
			}
		}
	}

	return "", errors.New("could not find workspace")
}

func ReadFromUserConfigFile() (UsersConfig, error) {
	file, err := os.Open(CONFIG_FILE)
	if err != nil {
		fmt.Println("error in opening config file.... pls check if tmp/userConfig.json available ")
		return UsersConfig{}, err
	}
	defer file.Close()

	var userConfig UsersConfig
	decoder := json.NewDecoder(file)
	err = decoder.Decode(&userConfig)
	if err != nil {
		fmt.Println("error in decoding json data")
		return UsersConfig{}, err
	}

	// fmt.Println(userConfig)
	return userConfig, nil
}

func writeToUserConfigFile(newUserConfig UsersConfig) error {
	jsonData, err := json.MarshalIndent(newUserConfig, "", "	")
	// fmt.Println(jsonData)
	if err != nil {
		fmt.Println("error occured in Marshalling the data to JSON")
		fmt.Println(err)
		return err
	}

	// fmt.Println(string(jsonData))
	err = os.WriteFile(CONFIG_FILE, jsonData, 0777)
	if err != nil {
		fmt.Println("error occured in storing data in userconfig file")
		fmt.Println(err)
		return err
	}

	return nil
}

// FIXME: This is an Old Function, the caller passes connection_slug
// func AddConnectionInUserConfig(connection_slug string, password string, connectionIP string, cmdPort int) error {
// 	userConfig, err := ReadFromUserConfigFile()
// 	if err != nil {
// 		return err
// 	}

// 	connection := Connections{
// 		ConnectionSlug: connection_slug,
// 		// Password:       password,
// 		CurrentIP:   connectionIP,
// 		CurrentPort: strconv.Itoa(cmdPort),
// 	}

// 	userConfig.AllConnections = append(userConfig.AllConnections, connection)
// 	newUserConfig := UsersConfig{
// 		User:           userConfig.User,
// 		AllConnections: userConfig.AllConnections,
// 		Sendworkspaces: userConfig.Sendworkspaces,
// 		GetWorkspaces:  userConfig.GetWorkspaces,
// 	}

// 	if err := writeToUserConfigFile(newUserConfig); err != nil {
// 		return err
// 	}
// 	return nil
// }

// This old function doesnt do anyting ??
// func AddNewConnectionToTheWorkspace(wName string, connectionSlug string) error {
// 	userConfig, err := ReadFromUserConfigFile()
// 	if err != nil {
// 		return err
// 	}

// 	wFound := false
// 	for _, server := range userConfig.ServerLists {
// 		for _, newSWork := range server.SendWorkspaces {
// 			if wName == newSWork.WorkspaceName {
// 				wFound = true
// 				break
// 			}
// 		}
// 	}

// 	if !wFound {
// 		fmt.Println(" No Such Workspace Exists !!")
// 		return nil
// 	}

// 	if err := writeToUserConfigFile(userConfig); err != nil {
// 		fmt.Println("error in writting to the user config file ...")
// 		return err
// 	}

// 	fmt.Printf(" New Connection Added To %s Workspace \n", wName)
// 	return nil
// }

// This CODE Might Be Useless.
// This Function Doesnt Seem to be Used Anywhere
// Please Delete this Future ME
func CreateNewWorkspace(serverAlias, wName, wPassword, wPath string) error {
	wfolder := SendWorkspaceFolder{
		WorkspaceName:     wName,
		WorkSpacePassword: wPassword,
		WorkspacePath:     wPath,
	}

	userConfig, err := ReadFromUserConfigFile()
	if err != nil {
		return err
	}

	for _, server := range userConfig.ServerLists {
		if server.ServerAlias == serverAlias {
			server.SendWorkspaces = append(server.SendWorkspaces, wfolder)
			if err := writeToUserConfigFile(userConfig); err != nil {
				fmt.Println("error: could not write to userconfig file")
				return err
			}

			return nil
		}
	}

	return fmt.Errorf("server with the alias - %s Not Found", serverAlias)
}

// TODO: See what this func is used for and rewrite a better one
// func GetAllConnections() []Connections {
// 	userConfigFile, err := ReadFromUserConfigFile()
// 	if err != nil {
// 		fmt.Println("error in reading from the userConfig File")
// 	}

// 	return userConfigFile.AllConnections
// }

// func ValidateConnection(connSlug string, connPassword string) bool {
// 	userConfigFile, err := readFromUserConfigFile()
// 	if err != nil {
// 		fmt.Println("error in reading from the userConfig File")
// 		return false
// 	}

// 	for _, conn := range userConfigFile.AllConnections {
// 		if conn.ConnectionSlug == connPassword && conn.Password == connPassword{
// 			return true
// 		}
// 	}

// 	return false
// }

// Creates Log Entry in the Main tmp file
func AddUsersLogEntry(log_entry any) error {
	// Adds the "root_dir/logs.txt"
	workspace_path := LOG_FILE

	// Opens or Creates the Log File
	file, err := os.OpenFile(workspace_path, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		return err
	}

	defer file.Close()
	log.SetOutput(file)
	log.Println(log_entry)

	return nil
}

// FIXME: This code is not needed... Fix the Caller and remove it
// func UpdateBasePort(port string) error {
// 	file, err := ReadFromUserConfigFile()
// 	if err != nil {
// 		return err
// 	}
//
// 	file.BasePort = port
// 	err = writeToUserConfigFile(file)
//
// 	return err
// }

// Update Last Hash (Used during Pulls)
func UpdateGetWorkspaceFolderToUserConfig(workspace_name, workspace_path, workspace_ip string, last_hash string) error {
	// WorkspaceName		string		`json:"workspace_name"`
	// WorkspacePath    	string		`json:"workspace_path"`
	// WorkspcaceIP			string		`json:"workspace_ip"`
	// LastHash				string		`json:"last_hash"`

	userConfig, err := ReadFromUserConfigFile()
	if err != nil {
		return err
	}

	for idx, server := range userConfig.ServerLists {
		for widx, workspace := range server.GetWorkspaces {
			if workspace.WorkspaceName == workspace_name {
				userConfig.ServerLists[idx].GetWorkspaces[widx].LastHash = last_hash
				break
			}
		}
	}

	if err := writeToUserConfigFile(userConfig); err != nil {
		return err
	}

	return nil
}

func GetAllGetWorkspaces() ([]GetWorkspaceFolder, error) {
	userConfig, err := ReadFromUserConfigFile()
	if err != nil {
		return []GetWorkspaceFolder{}, err
	}

	allGetWorkspaces := make([]GetWorkspaceFolder, 0)

	for _, server := range userConfig.ServerLists {
		allGetWorkspaces = append(allGetWorkspaces, server.GetWorkspaces...)
	}

	return allGetWorkspaces, nil
}

func GetAllSendWorkspaces() ([]GetWorkspaceFolder, error) {
	userConfig, err := ReadFromUserConfigFile()
	if err != nil {
		return []GetWorkspaceFolder{}, err
	}

	allGetWorkspaces := make([]GetWorkspaceFolder, 0)

	for _, server := range userConfig.ServerLists {
		allGetWorkspaces = append(allGetWorkspaces, server.GetWorkspaces...)
	}

	return allGetWorkspaces, nil
}

func AddNewServerToConfig(server_alias, server_ip, username, password string) error {
	// serverConfig, err := readFromServerConfigFile()

	userConfig, err := ReadFromUserConfigFile()
	if err != nil {
		fmt.Println("Error in reading From the UserConfig File...")
		return err
	}

	sconf := ServerConfig{
		Username:    username,
		Password:    password,
		ServerAlias: server_alias,
		ServerIP:    server_ip,
	}

	userConfig.ServerLists = append(userConfig.ServerLists, sconf)
	if err := writeToUserConfigFile(userConfig); err != nil {
		fmt.Println("Error Occured in Writing To the UserConfigr File")
		return err
	}

	return nil
}

func GetAllServers() ([]ServerConfig, error) {
	serverConfig, err := ReadFromUserConfigFile()
	if err != nil {
		return serverConfig.ServerLists, fmt.Errorf("error in reading From the ServerConfig File...\nError: %v", err)
	}

	return serverConfig.ServerLists, nil
}

// Returns - ServerIp, ServerUsername, ServerPassword
func GetServerDetails(server_alias string) (string, string, string, error) {
	// var username, password string
	serverConfig, err := ReadFromUserConfigFile()
	if err != nil {
		return "", "", "", fmt.Errorf("error in reading From the ServerConfig File...\nError: %v", err)
	}

	for _, server := range serverConfig.ServerLists {
		if server.ServerAlias == server_alias {
			return server.ServerIP, server.Username, server.Password, nil
		}
	}

	return "", "", "", fmt.Errorf("server with the server alias - %s not found", server_alias)
}

func GetServerIPThroughAlias(server_alias string) (string, error) {
	serverConfig, err := ReadFromUserConfigFile()
	if err != nil {
		return "", fmt.Errorf("error in reading From the ServerConfig File...\nError: %v", err)
	}

	for _, server := range serverConfig.ServerLists {
		if server.ServerAlias == server_alias {
			return server.ServerIP, nil
		}
	}

	return "", errors.New("no such server alias found")
}

func GetServerDetailsUsingServerAlias(server_alias string) (ServerConfig, error) {
	serverConfig, err := ReadFromUserConfigFile()
	if err != nil {
		return ServerConfig{}, fmt.Errorf("error in reading From the ServerConfig File...\nError: %v", err)
	}

	for _, server := range serverConfig.ServerLists {
		if server.ServerAlias == server_alias {
			return server, nil
		}
	}

	return ServerConfig{}, errors.New("no such server alias found")
}

func GetServerDetailsUsingServerIP(server_ip string) (ServerConfig, error) {
	serverConfig, err := ReadFromUserConfigFile()
	if err != nil {
		return ServerConfig{}, fmt.Errorf("error in reading From the ServerConfig File...\nError: %v", err)
	}

	for _, server := range serverConfig.ServerLists {
		if server.ServerIP == server_ip {
			return server, nil
		}
	}

	return ServerConfig{}, errors.New("server not found in config")
}
