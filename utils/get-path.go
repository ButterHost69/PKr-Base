package utils

import (
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"runtime"
)

var USER_CONFIF_FILE_DIR string

func SetUserConfigDir(file_path string) error {
	if file_path == "" {
		var err error
		USER_CONFIF_FILE_DIR, err = GetUserConfigRootDir()
		if err != nil {
			fmt.Println("Error while Getting Path of Local App Data:", err)
			fmt.Println("Source: GetUserConfigFilePath()")
			return err
		}
	}
	USER_CONFIF_FILE_DIR = file_path
	return nil
}

func GetUserConfigRootDir() (string, error) {
	if USER_CONFIF_FILE_DIR == "" {
		var base_dir string

		current_user, err := user.Current()
		if err != nil {
			fmt.Println("Error while Getting Current User:", err)
			fmt.Println("Source: GetUserConfigRootDir()")
			return "", err
		}

		switch runtime.GOOS {
		case "windows":
			base_dir = os.Getenv("LOCALAPPDATA") // Typically C:\Users\<User>\AppData\Local
		case "darwin":
			base_dir = filepath.Join(current_user.HomeDir, "Library", "Application Support")
		default: // Linux and other Unix-like systems
			base_dir = os.Getenv("XDG_DATA_HOME")
			if base_dir == "" {
				base_dir = filepath.Join(current_user.HomeDir, ".local", "share")
			}
		}
		return filepath.Join(base_dir, "PKr"), nil
	}
	return USER_CONFIF_FILE_DIR, nil

}

func GetMyKeysPath() (string, error) {
	user_config_root_dir, err := GetUserConfigRootDir()
	if err != nil {
		fmt.Println("Error while Getting Path of Local App Data:", err)
		fmt.Println("Source: GetMyKeysPath()")
		return "", err
	}
	return filepath.Join(user_config_root_dir, "Config", "Keys", "My"), nil
}

func GetOthersKeysPath() (string, error) {
	user_config_root_dir, err := GetUserConfigRootDir()
	if err != nil {
		fmt.Println("Error while Getting Path of Local App Data:", err)
		fmt.Println("Source: GetOthersKeysPath()")
		return "", err
	}
	return filepath.Join(user_config_root_dir, "Config", "Keys", "Others"), nil
}

func GetUserConfigFilePath() (string, error) {
	user_config_root_dir, err := GetUserConfigRootDir()
	if err != nil {
		fmt.Println("Error while Getting Path of Local App Data:", err)
		fmt.Println("Source: GetUserConfigFilePath()")
		return "", err
	}
	return filepath.Join(user_config_root_dir, "Config", "user-config.json"), nil
}
