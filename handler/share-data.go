package handler

import (
	"bufio"
	"io"
	"log"
	"os"

	"github.com/ButterHost69/PKr-Base/config"
	"github.com/ButterHost69/kcp-go"
)

const DATA_CHUNK = 1024 // 1KB

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

	destination_filepath := workspace_path + "\\.PKr\\" + workspace_hash + ".enc"
	log.Println("Destination FilePath to share:", destination_filepath)

	if _, err := os.Stat(destination_filepath); err == nil {
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

	log.Println("Preparing to Transfer Data")
	for {
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
			log.Println("Data Transfer Completed")
			return
		}
		if err != nil {
			log.Println("Error while Sending Workspace Chunk:", err)
			log.Println("Source: GetDataHandler()")
			sendErrorMessage(kcp_session, "Internal Server Error")
			return
		}
	}
}
