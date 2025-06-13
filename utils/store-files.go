package utils

import (
	"log"
	"os"
)

const (
	CONNECTION_KEYS_PATH = "tmp/connections/"
	ROOT_DIR             = "tmp"
	CONFIG_FILE          = "tmp/userConfig.json"
)

func StoreInitPublicKeys(connection_slug string, key string) error {
	if err := os.Mkdir(ROOT_DIR, 0766); err != nil {
		log.Println("~ Folder tmp Exists !!")
	}

	if err := os.Mkdir(CONNECTION_KEYS_PATH, 0766); err != nil {
		log.Println("~ Folder Connections Exists !!")
	}

	if err := os.Mkdir(CONNECTION_KEYS_PATH+connection_slug+"/", 0766); err != nil {
		log.Printf("~ Folder %s Exists !!\n", connection_slug)
	}
	connectionFilePath := CONNECTION_KEYS_PATH + connection_slug + "/publickey.pem"

	return os.WriteFile(connectionFilePath, []byte(key), 0666)
}
