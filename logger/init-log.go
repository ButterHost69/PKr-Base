package logger

import (
	"errors"
	"log"
	"os"
	"path/filepath"
)

var LOGGER *log.Logger

func InitLogger() error {
	local_app_data_path := os.Getenv("LOCALAPPDATA")
	if local_app_data_path == "" {
		return errors.New("localappdata path not set in env")
	}

	dir_name := filepath.Join(local_app_data_path, "PKr")
	err := os.MkdirAll(dir_name, 0600)
	if err != nil {
		log.Printf("Error while Creating '%s' dir: %v\n", dir_name, err)
		log.Println("Source: InitUserLogger()")
		return err
	}

	log_file_path := filepath.Join(local_app_data_path, "PKr", "PKr-Base.log")
	log_file, err := os.OpenFile(log_file_path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
	if err != nil {
		log.Printf("Error while Opening '%s' file: %v\n", log_file_path, err)
		log.Println("Source: InitUserLogger()")
		return err
	}

	LOGGER = log.New(log_file, "", log.Ldate|log.Ltime|log.Lshortfile)
	return nil
}
