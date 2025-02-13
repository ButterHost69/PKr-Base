package models

import (
	// "ButterHost69/PKr-client/encrypt"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/ButterHost69/PKr-cli/encrypt"
	// "github.com/go-delve/delve/cmd/dlv/cmds"
)

type Connections struct {
	ConnectionSlug string `json:"connection_slug"`
	// Password       string `json:"password"`
	CurrentIP   string `json:"current_ip"`
	CurrentPort string `json:"current_port"`
}

type ConnectionInfo struct {
	// ConnectionSlug string `json:"connection_slug"`
	Username    string `json:"username"`
	CurrentIP   string `json:"current_ip"`
	CurrentPort string `json:"current_port"`
}

type WorkspaceFolder struct {
	WorkspaceName     string `json:"workspace_name"`
	WorkspacePath     string `json:"workspace_path"`
	WorkSpacePassword string `json:"workspace_password"`
	// ConnectionSlugs []string `json:"connection_slug"`
}

type Files struct {
	FileName string `json:"file_name"`
	FileLoc  string `json:"file_loc"`
	FileSize string `json:"file_size"`
}

type GetWorkspaceFolder struct {
	WorkspaceName string `json:"workspace_name"`
	WorkspacePath string `json:"workspace_path"`
	WorkspcaceIP  string `json:"workspace_ip"`
	LastHash      string `json:"last_hash"`
}

type UsersConfig struct {
	User           string        `json:"user"`
	BasePort       string        `json:"base_port"`
	AllConnections []Connections `json:"all_connections"`

	Sendworkspaces []WorkspaceFolder    `json:"send_workspace"`
	GetWorkspaces  []GetWorkspaceFolder `json:"get_workspace"`
}

const (
	ROOT_DIR     = "tmp"
	MY_KEYS_PATH = ROOT_DIR + "\\mykeys"
	CONFIG_FILE  = ROOT_DIR + "\\userConfig.json"
	LOG_FILE     = ROOT_DIR + "\\logs.txt"
)

var (
	MY_USERNAME string
)

func CreateUserIfNotExists(username string) {
	if _, err := os.Stat(ROOT_DIR + "/userConfig.json"); os.IsNotExist(err) {
		fmt.Println("!! 'tmp' No such DIR exists ")

		// var username string
		// fmt.Println(" [*] Register [*]")
		// fmt.Print(" > Username: ")
		// fmt.Scanln(&username)
		MY_USERNAME = username

		usconf := UsersConfig{
			User: username,
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

		fmt.Printf(" ~ Created User : %s\n", username)
		return
	}

	fmt.Println("It Seems PKr is Already Installed...")
}

func AddConnection(connection_slug string, password string) {

}

func RegisterNewSendWorkspace(workspace_name string, workspace_path string, workspace_password string) error {
	userConfig, err := ReadFromUserConfigFile()
	if err != nil {
		fmt.Println("Error in reading From the UserConfig File...")
		return err
	}

	workspaceFolder := WorkspaceFolder{
		WorkspaceName:     workspace_name,
		WorkspacePath:     workspace_path,
		WorkSpacePassword: workspace_password,
	}
	userConfig.Sendworkspaces = append(userConfig.Sendworkspaces, workspaceFolder)

	if err := writeToUserConfigFile(userConfig); err != nil {
		fmt.Println("Error Occured in Writing To the UserConfig File")
		return err
	}

	return nil
}

func GetWorkspaceFilePath(workspace_name string) (string, error) {
	userConfig, err := ReadFromUserConfigFile()
	if err != nil {
		return "", err
	}

	workspaces := userConfig.Sendworkspaces
	for _, workspace := range workspaces {
		if workspace.WorkspaceName == workspace_name {
			return workspace.WorkspacePath, nil
		}
	}

	return "", errors.New("no such workspace found")
}

// Returns Workspace Path if Username and Password Correct
func AuthenticateWorkspaceInfo(workspace_name string, workspace_password string) string {
	userConfig, err := ReadFromUserConfigFile()
	if err != nil {
		return ""
	}

	for _, workspace := range userConfig.Sendworkspaces {
		if workspace.WorkspaceName == workspace_name {
			if workspace.WorkSpacePassword == workspace_password {
				return workspace.WorkspacePath
			}
		}
	}

	return ""
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

func AddConnectionInUserConfig(connection_slug string, password string, connectionIP string, cmdPort int) error {
	userConfig, err := ReadFromUserConfigFile()
	if err != nil {
		return err
	}

	connection := Connections{
		ConnectionSlug: connection_slug,
		// Password:       password,
		CurrentIP:   connectionIP,
		CurrentPort: strconv.Itoa(cmdPort),
	}

	userConfig.AllConnections = append(userConfig.AllConnections, connection)
	newUserConfig := UsersConfig{
		User:           userConfig.User,
		AllConnections: userConfig.AllConnections,
		Sendworkspaces: userConfig.Sendworkspaces,
		GetWorkspaces:  userConfig.GetWorkspaces,
	}

	if err := writeToUserConfigFile(newUserConfig); err != nil {
		return err
	}
	return nil
}

func UpdateWorkSpaceFolders() {

}

// func SetWorkSpaceFolders () error {

// }

func AddNewConnectionToTheWorkspace(wName string, connectionSlug string) error {
	userConfig, err := ReadFromUserConfigFile()
	if err != nil {
		return err
	}

	wFound := false
	for _, newSWork := range userConfig.Sendworkspaces {
		if wName == newSWork.WorkspaceName {
			wFound = true
			// newSWork.ConnectionSlugs = append(newSWork.ConnectionSlugs, connectionSlug)
			break
		}
	}

	if !wFound {
		fmt.Println(" No Such Workspace Exists !!")
		return nil
	}

	if err := writeToUserConfigFile(userConfig); err != nil {
		fmt.Println("error in writting to the user config file ...")
		return err
	}

	fmt.Printf(" New Connection Added To %s Workspace \n", wName)
	return nil
}

// This CODE Might Be Useless.
// This Function Doesnt Seem to be Used Anywhere
// Please Delete this Future ME
func CreateNewWorkspace(wName string, wPath string, connectionSlug string) error {
	//connectionSlugs := make([]string, 1)
	var connectionSlugs []string
	connectionSlugs = append(connectionSlugs, connectionSlug)
	fmt.Println(connectionSlugs)
	wfolder := WorkspaceFolder{
		WorkspaceName: wName,
		WorkspacePath: wPath,
		// ConnectionSlugs: connectionSlugs,
	}

	userConfig, err := ReadFromUserConfigFile()
	if err != nil {
		return err
	}

	fmt.Println(userConfig.Sendworkspaces)

	userConfig.Sendworkspaces = append(userConfig.Sendworkspaces, wfolder)

	// fmt.Println(userConfig.Sendworkspaces)
	// jj, _ := json.MarshalIndent(userConfig, "", "	")
	// fmt.Println(string(jj))

	if err := writeToUserConfigFile(userConfig); err != nil {
		fmt.Println("error: could not write to userconfig file")
		return err
	}

	return nil
}

func GetAllConnections() []Connections {
	userConfigFile, err := ReadFromUserConfigFile()
	if err != nil {
		fmt.Println("error in reading from the userConfig File")
	}

	return userConfigFile.AllConnections
}

// func GetAllSendWorkspaceList() []string {

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

func UpdateBasePort(port string) error {
	file, err := ReadFromUserConfigFile()
	if err != nil {
		return err
	}

	file.BasePort = port
	err = writeToUserConfigFile(file)

	return err
}

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

	for idx := range userConfig.GetWorkspaces {
		if userConfig.GetWorkspaces[idx].WorkspaceName == workspace_name {
			userConfig.GetWorkspaces[idx].LastHash = last_hash
			break
		}
	}

	if err := writeToUserConfigFile(userConfig); err != nil {
		return err
	}

	return nil
}


func GetAllGetWorkspaces() ([]GetWorkspaceFolder, error) {
	file, err := ReadFromUserConfigFile()
	if err != nil {
		return []GetWorkspaceFolder{}, err
	}

	return file.GetWorkspaces, err
}