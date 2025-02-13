package services

import (
	
	"github.com/ButterHost69/PKr-Base/encrypt"
	"github.com/ButterHost69/PKr-Base/models"
	"github.com/ButterHost69/PKr-Base/pb"
	"github.com/ButterHost69/PKr-Base/logger"
	"github.com/ButterHost69/PKr-Base/filetracker"
	

	"strings"

	"fmt"
	"io"
	"math/rand"
	"net"
	"os"
	"strconv"
	"sync"
	"time"

	// "github.com/brianvoe/gofakeit/v7/source"
	"google.golang.org/grpc"
)

type DataServer struct {
	pb.UnimplementedDataServiceServer
	wg             *sync.WaitGroup
	workspace_path string

	Workspace_Logger  *logger.WorkspaceLogger
	UserConfig_Logger *logger.UserLogger
}

func (server *DataServer) GetData(request *pb.DataRequest, stream pb.DataService_GetDataServer) error {

	// 1. Zip Data [x]
	//		|-> Check if already zipped ???? Later Not Now
	// 2. Encrypt Data [x] -> Check Ip or ID than find that public key
	// 		Encrypt using AES ; Encrypt AES using RSA ; SEND AES KEY ; SEND DATA
	// 3. Encrypt AES key and iv [X]
	// 4. Send / Stream Data [X]

	// workspace_path + "\\.PKr\\" + zipFileName
	zipped_file_name, err := filetracker.ZipData(server.workspace_path)
	zipped_hash := strings.Split(zipped_file_name, ".")[0]

	fmt.Println("Data Service Hash File Name: " + zipped_file_name)
	if err != nil {
		logdata := fmt.Sprintf("Could Not Zip The File\nError: %v", err)
		// models.AddLogEntry(request.WorkspaceName, logdata)
		server.Workspace_Logger.Critical(request.WorkspaceName, logdata)
		return err
	}

	err = models.AddNewPushToConfig(request.WorkspaceName, zipped_hash)
	if err != nil {
		logdata := fmt.Sprintf("could add entry to PKR config file.\nError: %v", err)
		// models.AddLogEntry(request.WorkspaceName, logdata)
		server.Workspace_Logger.Critical(request.WorkspaceName, logdata)
		return err
	}

	models.AddLogEntry(request.WorkspaceName, "Workspace Zipped")

	key, err := encrypt.AESGenerakeKey(16)
	if err != nil {
		logdata := fmt.Sprintf("Could Not Generate AES Key\nError: %v", err)
		// models.AddLogEntry(request.WorkspaceName, logdata)
		server.Workspace_Logger.Critical(request.WorkspaceName, logdata)
		return err
	}

	iv, err := encrypt.AESGenerateIV()
	if err != nil {
		logdata := fmt.Sprintf("Could Not Generate AES IV\nError: %v", err)
		// models.AddLogEntry(request.WorkspaceName, logdata)
		server.Workspace_Logger.Critical(request.WorkspaceName, logdata)
		return err
	}

	// models.AddLogEntry(request.WorkspaceName, "AES Keys Generated")
	server.Workspace_Logger.Info(request.WorkspaceName, "AES Keys Generated")

	zipped_filepath := server.workspace_path + "\\.PKr\\" + zipped_file_name
	destination_filepath := strings.Replace(zipped_filepath, ".zip", ".enc", 1)
	if err := encrypt.AESEncrypt(zipped_filepath, destination_filepath, key, iv); err != nil {
		logdata := fmt.Sprintf("Could Not Encrypt File\nError: %v\nFilePath: %v", err, zipped_filepath)
		// models.AddLogEntry(request.WorkspaceName, logdata)
		server.Workspace_Logger.Critical(request.WorkspaceName, logdata)
		return err
	}

	// models.AddLogEntry(request.WorkspaceName, "Zip AES is Encrypted")
	server.Workspace_Logger.Info(request.WorkspaceName, "Zip AES is Encrypted")

	publicKeyPath, err := models.GetConnectionsPublicKeyUsingIP(server.workspace_path, request.ConnectionIp)
	if err != nil {
		logdata := fmt.Sprintf("Could Not Find Users Public Key\nError: %v", err)
		// models.AddLogEntry(request.WorkspaceName, logdata)
		server.Workspace_Logger.Critical(request.WorkspaceName, logdata)
		return err
	}

	publicKey, err := os.ReadFile(publicKeyPath)
	if err != nil {
		logdata := fmt.Sprintf("Could Not Read Users Public Key\nError: %v", err)
		// models.AddLogEntry(request.WorkspaceName, logdata)
		server.Workspace_Logger.Critical(request.WorkspaceName, logdata)
		return err
	}

	encrypt_key, err := encrypt.EncryptData(string(key), string(publicKey))
	if err != nil {
		logdata := fmt.Sprintf("Could Not Encrypt Key\nError: %v", err)
		// models.AddLogEntry(request.WorkspaceName, logdata)
		server.Workspace_Logger.Critical(request.WorkspaceName, logdata)
		return err
	}

	encrypt_iv, err := encrypt.EncryptData(string(iv), string(publicKey))
	if err != nil {
		logdata := fmt.Sprintf("Could Not Encrypt IV\nError: %v", err)
		// models.AddLogEntry(request.WorkspaceName, logdata)
		server.Workspace_Logger.Critical(request.WorkspaceName, logdata)
		return err
	}

	encrypt_file, err := os.Open(destination_filepath)
	if err != nil {
		logdata := fmt.Sprintf("Could Not Open The Encrypted File\nError: %v", err)
		// models.AddLogEntry(request.WorkspaceName, logdata)
		server.Workspace_Logger.Critical(request.WorkspaceName, logdata)
		return err
	}
	defer encrypt_file.Close()

	// models.AddLogEntry(request.WorkspaceName, "AES Keys Encrypted")
	server.Workspace_Logger.Info(request.WorkspaceName, "AES Keys Encrypted")

	buff := make([]byte, 2048)
	chunknumber := 1

	// logdata = fmt.Sprintf()
	// models.AddLogEntry(request.WorkspaceName, "Sending Chunks...")
	server.Workspace_Logger.Info(request.WorkspaceName, "Sending Chunks...")

	// Sending Last Hash with File Type = 3
	chunk := []byte(zipped_hash)
	if err := stream.Send(&pb.Data{Filetype: 3, Chunk: chunk}); err != nil {
		logdata := fmt.Sprintf("Could Not Send Last Hash\nError: %v", err)
		models.AddLogEntry(request.WorkspaceName, logdata)
		return err
	}

	// Send the File
	for {
		num, err := encrypt_file.Read(buff)
		// ## Uncomment Later ##
		if err == io.EOF {
			// If file Send Over then Send the Key and IV
			// models.AddLogEntry(request.WorkspaceName, "Sending Encrypted Keys")
			server.Workspace_Logger.Info(request.WorkspaceName, "Sending Encrypted Keys")
			if err := stream.Send(&pb.Data{Filetype: 1, Chunk: []byte(encrypt_key)}); err != nil {
				logdata := fmt.Sprintf("Could Not Send Encrypted Key\nError: %v", err)
				// models.AddLogEntry(request.WorkspaceName, logdata)
				server.Workspace_Logger.Critical(request.WorkspaceName, logdata)
				return err
			}

			models.AddLogEntry(request.WorkspaceName, "Sending Encrypted IV")
			if err := stream.Send(&pb.Data{Filetype: 2, Chunk: []byte(encrypt_iv)}); err != nil {
				logdata := fmt.Sprintf("Could Not Send Encrypted IV\nError: %v", err)
				// models.AddLogEntry(request.WorkspaceName, logdata)
				server.Workspace_Logger.Critical(request.WorkspaceName, logdata)
				return err
			}

			break
		}

		if err != nil {
			logdata := fmt.Sprintf("Could Not Read Encrypted File\nError: %v", err)
			// models.AddLogEntry(request.WorkspaceName, logdata)
			server.Workspace_Logger.Critical(request.WorkspaceName, logdata)
			return err
		}

		chunk := buff[:num]
		if err := stream.Send(&pb.Data{Filetype: 0, Chunk: chunk}); err != nil {
			logdata := fmt.Sprintf("Could Not Send Encrypted File\nError: %v", err)
			// models.AddLogEntry(request.WorkspaceName, logdata)
			server.Workspace_Logger.Critical(request.WorkspaceName, logdata)
			return err
		}

		chunknumber += 1
	}
	logdata := fmt.Sprintf("File Sent SuccessFully to User IP: %v\nTotal Size: %v", request.ConnectionIp, chunknumber*256)
	// models.AddLogEntry(request.WorkspaceName, logdata)
	server.Workspace_Logger.Info(request.WorkspaceName, logdata)

	server.wg.Done()
	return nil
}

func isPortInUse(port int) bool {
	conn, err := net.DialTimeout("tcp", ":"+strconv.Itoa(port), time.Second)
	if err != nil {
		return false
	}
	conn.Close()
	return true
}

func StartDataServer(time_till_wait time.Duration, workspace_name string, workspace_path string, portchan chan int, errorchan chan error, workspace_logger *logger.WorkspaceLogger, userconfig_logger *logger.UserLogger) {
	// Pass the port using channels
	// Look into it later, i mean soon, after the DataServer soon[]

	// Generate a Random not Taken Port Number
	rand.Seed(time.Now().UnixNano())
	var port int
	for {
		port = rand.Intn(65535-1024) + 1024
		if !isPortInUse(port) {
			break
		}
	}

	// Register and Start gRPC Server
	str_port := strconv.Itoa(port)

	logdata := fmt.Sprintf("Port Number: %v Has Been Selected For Data Transfer", str_port)
	// models.AddLogEntry(workspace_name, logdata)
	workspace_logger.Info(workspace_name, logdata)

	lis, err := net.Listen("tcp", ":"+str_port)
	if err != nil {
		// models.AddLogEntry(workspace_name, err)
		workspace_logger.Debug(workspace_name, err)
		errorchan <- err
		os.Exit(1)
	}

	logdata = fmt.Sprintf("Data Server Started on %d", port)
	workspace_logger.Info(workspace_name, logdata)

	logdata = fmt.Sprintf("Started Listening to the Port: %v", str_port)
	// models.AddLogEntry(workspace_name, logdata)
	workspace_logger.Info(workspace_name, logdata)

	var wg sync.WaitGroup
	wg.Add(1)

	grpcServer := grpc.NewServer()
	dataServer := DataServer{
		wg:                &wg,
		workspace_path:    workspace_path,
		Workspace_Logger:  workspace_logger,
		UserConfig_Logger: userconfig_logger,
	}

	pb.RegisterDataServiceServer(grpcServer, &dataServer)

	go func() {
		// Add Serve With Time Out Later
		if err := grpcServer.Serve(lis); err != nil {
			// models.AddLogEntry(workspace_name, err)
			// models.AddLogEntry(workspace_name, "Closing Data Server")
			workspace_logger.Critical(workspace_name, err)
			workspace_logger.Critical(workspace_name, "Closing Data Server")
			errorchan <- err
			os.Exit(1)
		}
	}()

	logdata = fmt.Sprintf("gRPC Server Data Started on Port: %v", str_port)
	// models.AddLogEntry(workspace_name, logdata)
	workspace_logger.Info(workspace_name, logdata)

	portchan <- port

	wg.Wait()
	// fmt.Println("Data Transfer Done ... Closing Data Server") // [ ] Debug
	// models.AddLogEntry(workspace_name, "Data Transfer Done... Closing Data Server")
	workspace_logger.Info(workspace_name, logdata)
	grpcServer.GracefulStop()
}
