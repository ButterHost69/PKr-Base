package servicestemp

import (
	"context"
	"encoding/base64"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/ButterHost69/PKr-Base/dialer"
	"github.com/ButterHost69/PKr-Base/encrypt"
	// "github.com/ButterHost69/PKr-Base/config"
	"github.com/ButterHost69/PKr-Base/pb"
	"github.com/ButterHost69/PKr-Base/logger"

	"google.golang.org/grpc/peer"
	emptypb "google.golang.org/protobuf/types/known/emptypb"
)

type BackgroundServer struct {
	pb.UnimplementedBackgroundServiceServer
	WorkspaceLogger		*logger.WorkspaceLogger
	UserConfingLogger	*logger.UserLogger
}

func (s *BackgroundServer) GetPublicKey(ctx context.Context, request *emptypb.Empty) (*pb.PublicKey, error) {
	keyData, err := ReadPublicKey()
	p, _ := peer.FromContext(ctx)
	ip := p.Addr.String()
	if err != nil {
		logentry := "Could Not Provide Public Key To IP: " + ip
		s.UserConfingLogger.Debug(logentry)
		s.UserConfingLogger.Debug(err)
		
		return &pb.PublicKey{
			Key: nil,
		}, err
	}
	logentry := "Successfully Provided Public Key To IP: " + ip
	// models.AddUsersLogEntry(logentry)
	s.UserConfingLogger.Info(logentry)

	return &pb.PublicKey{
		Key: []byte(keyData),
	}, nil

}

// FIXME : IP is not stored when Connection is formed ... Look into it
// FIXME: Public Key Is checked from somewhere else...in the root dir ../..
func (s *BackgroundServer) InitNewWorkSpaceConnection(ctx context.Context, request *pb.InitRequest) (*pb.InitResponse, error) {
	// 1. Decrypt password [X]
	// 2. Authenticate Request [X]
	// 3. Add the New Connection to the .PKr Config File [X]
	// 4. Store the Public Key [X]
	// 5. Send the Response with port [X]
	// 6. Open a Data Transfer Port and shit [Will be a separate Function not here] [X]

	// [ ] Find a Better Alternative
	p, _ := peer.FromContext(ctx)
	ip := p.Addr.String()

	// [ ] Check Could be Causing Errors
	// Could Have Regex
	re := regexp.MustCompile(`^\[::1\]`)
	ip = re.ReplaceAllString(ip, "192.168.29.182")
	ip = strings.Split(ip, ":")[0]
	ip = ip + ":9000"

	encrypted_password := request.Password

	// This is Comment temporary
	password, err := encrypt.DecryptData(encrypted_password)
	// password := encrypted_password

	// UNCOMMENT THIS SHIT --------
	if err != nil {
		// AddUserLogEntry("Failed to Init Workspace Connection for User IP: " + ip)
		// AddUserLogEntry(err)
		s.UserConfingLogger.Debug("Failed to Init Workspace Connection for User IP: " + ip)
		s.UserConfingLogger.Debug(err)

		return &pb.InitResponse{
			Response: 4000,
			Port:     0000,
		}, nil
	}

	// Authenticates Workspace Name and Password and Get the Workspace File Path
	file_path := models.AuthenticateWorkspaceInfo(request.WorkspaceName, password)
	if file_path == "" {
		// models.AddLogEntry(request.WorkspaceName, "Failed to Init Workspace Connection for User IP: "+ip)
		s.WorkspaceLogger.Debug(request.Username, "Failed to Init Workspace Connection for User IP: "+ip)
		return &pb.InitResponse{
			Response: 4000,
			Port:     0000,
		}, nil
	}

	var connection models.Connection
	connection.CurrentIP = ip
	connection.CurrentPort = request.Port

	// Save Public Key
	publicKey, err := base64.StdEncoding.DecodeString(string(request.PublicKey))
	if err != nil {
		s.WorkspaceLogger.Debug(request.Username, "Failed to convert key to Base64 for User IP: "+ip)
		s.WorkspaceLogger.Debug(request.Username, err)
		// models.AddLogEntry(request.WorkspaceName, "Failed to convert key to Base64 for User IP: "+ip)
		// models.AddLogEntry(request.WorkspaceName, err)
		return &pb.InitResponse{
			Response: 4000,
			Port:     0000,
		}, err
	}

	keysPath, err := models.StorePublicKeys(file_path+"\\.PKr\\keys\\", string(publicKey))
	if err != nil {
		// models.AddLogEntry(request.WorkspaceName, "Failed to Init Workspace Connection for User IP: "+ip)
		// models.AddLogEntry(request.WorkspaceName, err)
		s.WorkspaceLogger.Debug(request.Username, "Failed to Init Workspace Connection for User IP: "+ip)
		s.WorkspaceLogger.Debug(request.Username, err)
		return &pb.InitResponse{
			Response: 4000,
			Port:     0000,
		}, err
	}

	// Store the New Connection in the .PKr Config file
	connection.PublicKeyPath = keysPath
	if err := models.AddConnectionToPKRConfigFile(file_path+"\\.PKr\\workspaceConfig.json", connection); err != nil {
		// models.AddLogEntry(request.WorkspaceName, "Failed to Init Workspace Connection for User IP: "+ip)
		// models.AddLogEntry(request.WorkspaceName, err)
		s.WorkspaceLogger.Debug(request.Username, "Failed to Init Workspace Connection for User IP: "+ip)
		s.WorkspaceLogger.Debug(request.Username, err)
		return &pb.InitResponse{
			Response: 4000,
			Port:     0000,
		}, err
	}
	// models.AddLogEntry(request.WorkspaceName, fmt.Sprintf("Added User with IP: %v to the Connection List", ip))
	s.WorkspaceLogger.Info(request.Username, fmt.Sprintf("Added User with IP: %v to the Connection List", ip))

	// Start New Data grpc Server and Transfer Data
	portchan := make(chan int)
	errorchan := make(chan error)

	go StartDataServer(120*time.Minute, request.WorkspaceName, file_path, portchan, errorchan, s.WorkspaceLogger, s.UserConfingLogger)
	select {
	case port_num := <-portchan:
		return &pb.InitResponse{
			Response: 200,
			Port:     int32(port_num),
		}, nil
	case err := <-errorchan:
		return &pb.InitResponse{
			Response: 4000,
			Port:     0000,
		}, err
	}
}

func (s *BackgroundServer) NotifyPush(ctx context.Context, request *pb.NotifyPushRequest) (*pb.NotifyPushResponse, error) {
	workspace_name := request.WorkspaceName

	log_entry := "NEW UPDATE IN FILES OF " + workspace_name
	// models.AddLogEntry(workspace_name, log_entry) // [ ] Idk why this line isn't working maybe cuz log.txt isn't generated
	s.WorkspaceLogger.Debug(workspace_name, log_entry)

	// [ ] Fetch the new data

	// [ ] Compare Hashes

	return &pb.NotifyPushResponse{Response: 200}, dialer.PullData(s.UserConfingLogger, workspace_name)
}

func (s *BackgroundServer) ScanForUpdatesOnStart(ctx context.Context, request *pb.ScanForUpdatesRequest) (*pb.ScanForUpdatesResponse, error) {
	// [ ] Check whether Workspace name is valid, log properly
	// Check whether Receiver's Hash is latest or not
	// Return true if there're new updates else false

	workspaceName := request.WorkspaceName
	receiverHash := request.LastHash

	workspacePath, err := models.GetWorkspaceFilePath(workspaceName)
	if err != nil {
		log_entry := "cannot get path of workspace\nError: " + err.Error() + "\nSource: ScanForUpdatesOnStart() Handler" + err.Error()
		// log.Println(log_entry)
		// models.AddLogEntry(workspaceName, log_entry)

		s.WorkspaceLogger.Debug(workspaceName, log_entry)
		return nil, err
	}
	workspacePath = workspacePath + "\\" + models.WORKSPACE_CONFIG_FILE_PATH

	workspace_config, err := models.ReadFromPKRConfigFile(workspacePath)
	if err != nil {
		log_entry := "cannot read from workspace config file\nError: " + err.Error() + "\nSource: ScanForUpdatesOnStart() Handler" + err.Error()
		// log.Println(log_entry)
		// models.AddLogEntry(workspaceName, log_entry)

		s.WorkspaceLogger.Debug(workspaceName, log_entry)
		return nil, err
	}

	// [ ] Debugging
	// fmt.Println("Last Hash: ", workspace_config.LastHash)
	// fmt.Println("Receiver Hash: ", receiverHash)

	s.WorkspaceLogger.Info(workspaceName, fmt.Sprintf("Last Hash: %v", workspace_config.LastHash))
	s.WorkspaceLogger.Info(workspaceName, fmt.Sprintf("Receiver Hash: %v", receiverHash))

	if workspace_config.LastHash == receiverHash {
		return &pb.ScanForUpdatesResponse{NewUpdates: false}, nil
	}
	return &pb.ScanForUpdatesResponse{NewUpdates: true}, nil
}

func (s *BackgroundServer) PullData(ctx context.Context, request *pb.PullDataRequest) (*pb.PullDataResponse, error) {
	// Start New Data grpc Server and Transfer Data
	portchan := make(chan int)
	errorchan := make(chan error)

	workspacePath, err := models.GetWorkspaceFilePath(request.WorkspaceName)
	if err != nil {
		log_entry := fmt.Sprintf("cannot get workspace's file path\nError: %s\nSource: PullData() Handler", err.Error())
		// models.AddLogEntry(request.WorkspaceName, log_entry)
		// fmt.Println(log_entry) // [ ]: Debug
		s.WorkspaceLogger.Debug(request.WorkspaceName, log_entry)
		return nil, err
	}

	go StartDataServer(120*time.Minute, request.WorkspaceName, workspacePath, portchan, errorchan, s.WorkspaceLogger, s.UserConfingLogger)
	select {
	case port_num := <-portchan:
		return &pb.PullDataResponse{
			Port: int32(port_num),
		}, nil
	case err := <-errorchan:
		log_entry := fmt.Sprintf("cannot get workspace's file path\nError: %s\nSource: PullData() Handler", err.Error())
		// models.AddLogEntry(request.WorkspaceName, log_entry)
		// fmt.Println(log_entry) // [ ]: Debug
		s.WorkspaceLogger.Debug(request.WorkspaceName, log_entry)
		return &pb.PullDataResponse{
			Port: 0000,
		}, err
	}
}
