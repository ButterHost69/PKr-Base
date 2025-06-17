package handler

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/ButterHost69/PKr-Base/config"
	"github.com/ButterHost69/PKr-Base/utils"
	"github.com/ButterHost69/kcp-go"
)

const DATA_CHUNK = 1024                        // 1KB
const FLUSH_AFTER_EVERY_X_MB = 5 * 1024 * 1024 // 5 MB

func sendErrorMessage(kcp_session *kcp.UDPSession, error_msg string) {
	_, err := kcp_session.Write([]byte(error_msg))
	if err != nil {
		log.Println("Error while Sending Error Message:", err)
		log.Println("Source: sendMessage()")
	}
}

func GetDataHandler(kcp_session *kcp.UDPSession) {
	log.Println("Get Data Handler Called")
	log.Println("Reading Workspace Name ...")

	var buff [512]byte
	n, err := kcp_session.Read(buff[:])
	if err != nil {
		log.Println("Error while Reading Workspace Name:", err)
		log.Println("Source: GetDataHandler()")
		return
	}
	workspace_name := string(buff[:n])
	log.Println("Workspace Name:", workspace_name)
	log.Println("Reading Workspace Hash ...")

	n, err = kcp_session.Read(buff[:])
	if err != nil {
		log.Println("Error while Reading Workspace Hash:", err)
		log.Println("Source: GetDataHandler()")
		return
	}
	workspace_hash := string(buff[:n])
	log.Println("Workspace Hash:", workspace_hash)

	workspace_path, err := config.GetSendWorkspaceFilePath(workspace_name)
	if err != nil {
		log.Println("Failed to Get Workspace Path from Config:", err)
		log.Println("Source: GetDataHandler()")
		sendErrorMessage(kcp_session, "Internal Server Error")
		return
	}
	log.Println("Workspace Path:", workspace_path)

	// Check if hash is last hash -> Send The .Pkr/Files/Current dir
	// TODO: Else Check in Changes Hash, Send it (should be there;to be created during metadata)
	config_file, err := config.ReadFromPKRConfigFile(filepath.Join(workspace_path, ".PKr", "workspaceConfig.json"))
	if err != nil {
		log.Println("Failed to Get Workspace  Config:", err)
		log.Println("Source: GetDataHandler()")
		sendErrorMessage(kcp_session, "Internal Server Error")
		return
	}

	config_file.LastHash = strings.TrimSpace(config_file.LastHash)
	workspace_hash = strings.TrimSpace(workspace_hash)

	destination_filepath := ""
	if workspace_hash == config_file.LastHash {
		log.Println("Requested Hash is of the Snapshot")
		destination_filepath = filepath.Join(workspace_path, ".PKr", "Files", "Current", workspace_hash+".enc")
		log.Println("Updated Destination File Path: ", destination_filepath)
	} else {
		log.Println("Checking if Requested Hash is of the Changes ...")

		changes_cache_path := filepath.Join(workspace_path, ".PKr", "Files", "Changes")
		entries, err := os.ReadDir(changes_cache_path)
		if err != nil {
			log.Println("Failed to Read Dir From Workspace Changes Cache:", err)
			log.Println("Source: GetDataHandler()")
			sendErrorMessage(kcp_session, "Internal Server Error")
			return
		}

		if_present := false
		for _, entry := range entries {
			if entry.IsDir() && entry.Name() == workspace_hash {
				if_present = true
			}
		}

		if if_present {
			log.Println("Provided Hash is Present in the Workspace Changes Cache")
			destination_filepath = filepath.Join(workspace_path, ".PKr", "Files", "Changes", workspace_hash, workspace_hash+".enc")
			log.Println("Destination File Path is set: ", destination_filepath)
		} else {
			log.Println("Garbage Hash")
			log.Println("Provided Hash: ", workspace_hash)
			log.Println("Last Config Hash: ", config_file.LastHash)
			sendErrorMessage(kcp_session, "Internal Server Error")
			return
		}
	}

	if destination_filepath == ""  {
		log.Println("Garbage Hash")
		log.Println("Provided Hash: ", workspace_hash)
		log.Println("Last Config Hash: ", config_file.LastHash)
		sendErrorMessage(kcp_session, "Internal Server Error")
		return
	}

	log.Println("Destination FilePath to share:", destination_filepath)

	fileInfo, err := os.Stat(destination_filepath)
	if err == nil {
		log.Println("Destination File exists")
	} else if os.IsNotExist(err) {
		log.Println("Destination File does not exist")
		sendErrorMessage(kcp_session, "Incorrect Workspace Name/Hash")
		return
	} else {
		log.Println("Error while checking Existence of Destination file:", err)
		log.Println("Source: GetDataHandler()")
		sendErrorMessage(kcp_session, "Internal Server Error")
		return
	}

	log.Println("Opening Destination File")
	file, err := os.Open(destination_filepath)
	if err != nil {
		log.Println("Error while Opening Destination File:", err)
		log.Println("Source: GetDataHandler()")
		sendErrorMessage(kcp_session, "Internal Server Error")
		return
	}
	defer file.Close()

	buffer := make([]byte, DATA_CHUNK)
	reader := bufio.NewReader(file)

	len_data_bytes := int(fileInfo.Size())
	log.Println("Length of File:", len_data_bytes)

	log.Println("Preparing to Transfer Data")
	offset := 0
	for {
		utils.PrintProgressBar(offset, len_data_bytes, 100)

		n, err := reader.Read(buffer)
		if n > 0 {
			_, err := kcp_session.Write([]byte(buffer[:n]))
			if err != nil {
				log.Println("Error while Sending Data:", err)
				log.Println("Source: GetDataHandler()")
				sendErrorMessage(kcp_session, "Internal Server Error")
				return
			}
		}
		if err == io.EOF {
			fmt.Println()
			log.Println("Done Sent, now waiting for ack from listener ...")
			n, err = kcp_session.Read(buff[:])
			if err != nil {
				log.Println("Error while Reading 'Data Received' Message from Listener:", err)
				log.Println("Source: GetDataHandler()")
			}
			//Data Received
			msg := string(buff[:n])
			if msg == "Data Received" {
				log.Println("Data Transfer Completed:", offset)
				return
			}
			log.Println("Received Unexpected Message:", msg)
			return
		}
		if err != nil {
			log.Println("Error while Sending Workspace Chunk:", err)
			log.Println("Source: GetDataHandler()")
			sendErrorMessage(kcp_session, "Internal Server Error")
			return
		}
		offset += n
	}
}
