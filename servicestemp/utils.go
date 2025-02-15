package servicestemp

import (
	"log"
	"os"
)

const (
	ROOT_DIR        = "tmp"
	MY_KEYS_PATH    = ROOT_DIR + "\\mykeys"
	CONFIG_FILE     = ROOT_DIR + "\\userConfig.json"
	LOG_FILE        = ROOT_DIR + "\\logs.txt"
	SERVER_LOG_FILE = ROOT_DIR + "\\serverlogs.txt"
)

func ReadPublicKey() (string, error) {
	keyData, err := os.ReadFile(MY_KEYS_PATH + "\\publickey.pem")
	if err != nil {
		return "", err
	}

	return string(keyData), nil
}

func AddUserLogEntry(log_text any) {
	file, err := os.OpenFile(SERVER_LOG_FILE, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		return
	}
	defer file.Close()

	log.SetOutput(file)
	log.Println(log_text)
}
