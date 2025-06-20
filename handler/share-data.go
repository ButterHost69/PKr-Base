package handler

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"

	"github.com/ButterHost69/PKr-Base/config"
	"github.com/ButterHost69/PKr-Base/encrypt"
	"github.com/ButterHost69/PKr-Base/utils"
	"github.com/ButterHost69/kcp-go"
)

const DATA_CHUNK = encrypt.DATA_CHUNK
const FLUSH_AFTER_EVERY_X_MB = encrypt.FLUSH_AFTER_EVERY_X_MB

func sendErrorMessage(kcp_session *kcp.UDPSession, error_msg string) {
	_, err := kcp_session.Write([]byte(error_msg))
	if err != nil {
		log.Println("Error while Sending Error Message:", err)
		log.Println("Source: sendMessage()")
	}
}

func handleClone(kcp_session *kcp.UDPSession, zip_path string, len_data_bytes int, workspace_path string) {
	curr_dir := filepath.Join(workspace_path, ".PKr", "Files", "Current") + string(filepath.Separator)
	key, err := os.ReadFile(curr_dir + "AES_KEY")
	if err != nil {
		log.Println("Error while Reading AES Key:", err)
		log.Println("Source: handleClone()")
		return
	}

	iv, err := os.ReadFile(curr_dir + "AES_IV")
	if err != nil {
		log.Println("Error while Reading AES IV:", err)
		log.Println("Source: handleClone()")
		return
	}

	log.Println("Opening Destination File")
	zip_file_obj, err := os.Open(zip_path)
	if err != nil {
		log.Println("Error while Opening Destination File:", err)
		log.Println("Source: GetDataHandler()")
		sendErrorMessage(kcp_session, "Internal Server Error")
		return
	}
	defer zip_file_obj.Close()

	var buff [512]byte
	buffer := make([]byte, DATA_CHUNK)
	reader := bufio.NewReader(zip_file_obj)
	log.Println("Length of File:", len_data_bytes)

	log.Println("Preparing to Transfer Data")
	offset := 0
	for {
		utils.PrintProgressBar(offset, len_data_bytes, 100)

		n, err := reader.Read(buffer)
		if err != nil {
			if err == io.EOF {
				fmt.Println()
				log.Println("Done Sent, now waiting for ack from listener ...")
				n, err := kcp_session.Read(buff[:])
				if err != nil {
					log.Println("Error while Reading 'Data Received' Message from Listener:", err)
					log.Println("Source: GetDataHandler()")
					return
				}
				// Data Received
				msg := string(buff[:n])
				if msg == "Data Received" {
					log.Println("Data Transfer Completed:", offset)
					return
				}
				log.Println("Received Unexpected Message:", msg)
				return
			}
			log.Println("Error while Sending Workspace Chunk:", err)
			log.Println("Source: GetDataHandler()")
			sendErrorMessage(kcp_session, "Internal Server Error")
			return
		}

		if n > 0 {
			buffer, err = encrypt.EncryptDecryptChunk(buffer[:n], key, iv)
			if err != nil {
				log.Println("Error while Encrypting Data Chunk ...:", err)
				log.Println("Source: handleClone()")
				return
			}

			_, err := kcp_session.Write([]byte(buffer[:n]))
			if err != nil {
				log.Println("Error while Sending Data:", err)
				log.Println("Source: GetDataHandler()")
				sendErrorMessage(kcp_session, "Internal Server Error")
				return
			}
		}
		offset += n
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
	log.Println("Reading Workspace Push Num ...")

	n, err = kcp_session.Read(buff[:])
	if err != nil {
		log.Println("Error while Reading Workspace Push Num:", err)
		log.Println("Source: GetDataHandler()")
		return
	}
	workspace_push_num := string(buff[:n])
	log.Println("Workspace Push Num:", workspace_push_num)

	// Read Data Request Type (Pull/Clone)
	n, err = kcp_session.Read(buff[:])
	if err != nil {
		log.Println("Error while Reading Type of Data Request Type:", err)
		log.Println("Source: GetDataHandler()")
		return
	}
	data_req_type := string(buff[:n])
	log.Println("Data Request Type(Clone/Pull):", data_req_type)

	workspace_path, err := config.GetSendWorkspaceFilePath(workspace_name)
	if err != nil {
		log.Println("Failed to Get Workspace Path from Config:", err)
		log.Println("Source: GetDataHandler()")
		sendErrorMessage(kcp_session, "Internal Server Error")
		return
	}
	log.Println("Workspace Path:", workspace_path)

	if data_req_type == "Clone" {
		zip_path := filepath.Join(workspace_path, ".PKr", "Files", "Current", workspace_push_num+".zip")
		fileInfo, err := os.Stat(zip_path)
		if err == nil {
			log.Println("Destination File exists")
		} else if os.IsNotExist(err) {
			log.Println("Destination File does not exist")
			sendErrorMessage(kcp_session, "Incorrect Workspace Name/Push Num")
			return
		} else {
			log.Println("Error while checking Existence of Destination file:", err)
			log.Println("Source: GetDataHandler()")
			sendErrorMessage(kcp_session, "Internal Server Error")
			return
		}

		handleClone(kcp_session, zip_path, int(fileInfo.Size()), workspace_path)
		return
	} else if data_req_type != "Pull" {
		log.Println("Invalid Data Request Type Sent from User")
		log.Println("Source: GetDataHandler()")
		sendErrorMessage(kcp_session, "Invalid Data Request Type Sent")
		return
	}

	zip_enc_path := filepath.Join(workspace_path, ".PKr", "Files", "Changes", workspace_push_num, workspace_push_num+".enc")
	log.Println("Zip Enc FilePath to share:", zip_enc_path)

	fileInfo, err := os.Stat(zip_enc_path)
	if err == nil {
		log.Println("Destination File exists")
	} else if os.IsNotExist(err) {
		log.Println("Destination File does not exist")
		sendErrorMessage(kcp_session, "Incorrect Workspace Name/Push Num Range")
		return
	} else {
		log.Println("Error while checking Existence of Destination file:", err)
		log.Println("Source: GetDataHandler()")
		sendErrorMessage(kcp_session, "Internal Server Error")
		return
	}

	log.Println("Opening Destination File")
	zip_file_obj, err := os.Open(zip_enc_path)
	if err != nil {
		log.Println("Error while Opening Destination File:", err)
		log.Println("Source: GetDataHandler()")
		sendErrorMessage(kcp_session, "Internal Server Error")
		return
	}
	defer zip_file_obj.Close()

	buffer := make([]byte, DATA_CHUNK)
	reader := bufio.NewReader(zip_file_obj)

	len_data_bytes := int(fileInfo.Size())
	log.Println("Length of File:", len_data_bytes)

	log.Println("Preparing to Transfer Data")
	offset := 0
	for {
		utils.PrintProgressBar(offset, len_data_bytes, 100)

		n, err := reader.Read(buffer)
		if err != nil {
			if err == io.EOF {
				fmt.Println()
				log.Println("Done Sent, now waiting for ack from listener ...")
				n, err := kcp_session.Read(buff[:])
				if err != nil {
					log.Println("Error while Reading 'Data Received' Message from Listener:", err)
					log.Println("Source: GetDataHandler()")
					return
				}
				// Data Received
				msg := string(buff[:n])
				if msg == "Data Received" {
					log.Println("Data Transfer Completed:", offset)
					return
				}
				log.Println("Received Unexpected Message:", msg)
				return
			}
			log.Println("Error while Sending Workspace Chunk:", err)
			log.Println("Source: GetDataHandler()")
			sendErrorMessage(kcp_session, "Internal Server Error")
			return
		}

		if n > 0 {
			_, err := kcp_session.Write([]byte(buffer[:n]))
			if err != nil {
				log.Println("Error while Sending Data:", err)
				log.Println("Source: GetDataHandler()")
				sendErrorMessage(kcp_session, "Internal Server Error")
				return
			}
		}
		offset += n
	}
}
