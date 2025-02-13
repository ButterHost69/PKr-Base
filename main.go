// Listening Server
// Listen For other Connections
// Responsible to Create the Server that will Send Data

package main

import (
	"flag"
	"fmt"
	"net"
	"os"
	"time"

	"github.com/ButterHost69/PKr-Base/dialer"
	"github.com/ButterHost69/PKr-Base/logger"
	"github.com/ButterHost69/PKr-Base/models"
	"github.com/ButterHost69/PKr-Base/pb"
	"github.com/ButterHost69/PKr-Base/services"
	"google.golang.org/grpc"
)

const (
	ROOT_DIR     = "..\\tmp"
	MY_KEYS_PATH = ROOT_DIR + "\\mykeys"
	CONFIG_FILE  = ROOT_DIR + "\\userConfig.json"
	LOG_FILE = ROOT_DIR + "\\logs.txt"
	SERVER_LOG_FILE = ROOT_DIR + "\\serverlogs.txt"
)

var (
	IP_ADDR 			string
	LOG_IN_TERMINAL		bool
	LOG_LEVEL			int
)

// Loggers
var (
	workspace_logger	*logger.WorkspaceLogger
	userconfing_logger	*logger.UserLogger
)

func Init() {
	flag.StringVar(&IP_ADDR, "ip", "", "Use Application in TUI Mode.")
	flag.BoolVar(&LOG_IN_TERMINAL, "lt", false, "Log Events in Terminal.")
	flag.IntVar(&LOG_LEVEL, "ll", 4, "Set Log Levels.") // 4 -> No Logs
	flag.Parse()
	
	// Create and Initialize Loggers
	workspace_logger = logger.InitWorkspaceLogger()
	userconfing_logger = logger.InitUserLogger(LOG_FILE)

	workspace_logger.SetLogLevel(logger.IntToLog(LOG_LEVEL))
	userconfing_logger.SetLogLevel(logger.IntToLog(LOG_LEVEL))

	workspace_logger.SetPrintToTerminal(LOG_IN_TERMINAL)
	userconfing_logger.SetPrintToTerminal(LOG_IN_TERMINAL)

	workspaces, err := models.GetAllGetWorkspaces()
	if err != nil {
		userconfing_logger.Critical(fmt.Sprintf("could not get all get workspaces.\nError: %v", err))
		return
	}

	workspace_to_path := make(map[string]string)
	for _, fp := range workspaces {
		workspace_to_path[fp.WorkspaceName] = fp.WorkspacePath
	}

	workspace_logger.SetWorkspacePaths(workspace_to_path)

	// If ip is Not Provided during execution as flags check ENV
	if IP_ADDR == "" {
		IP_ADDR = os.Getenv("PKR-IP")
		if IP_ADDR == "" {
			IP_ADDR = ":9000"
		}
	}

	models.UpdateBasePort(IP_ADDR)
}

func main() {
	Init()

	lis, err := net.Listen("tcp", IP_ADDR)
	if err != nil {
		userconfing_logger.Critical(fmt.Sprintf("Error: %v\n", err))
	}

	grpcServer := grpc.NewServer()
	backgroundService := services.BackgroundServer{
		WorkspaceLogger: workspace_logger,
		UserConfingLogger: userconfing_logger,
	}

	pb.RegisterBackgroundServiceServer(grpcServer, &backgroundService)
	userconfing_logger.Info(fmt.Sprintf("Base Service Running on Port: %s" , IP_ADDR))

	// TODO: [ ] Test this code, neither human test nor code test done....
	// All The functions written with it are not tested
	go func() {
		userconfing_logger.Info("Update me Service Started")
		for {
			// Read Each Time... So can automatically detect changes without manual anything....
			serverList, err := models.GetAllServers()
			if err != nil {
				userconfing_logger.Debug(fmt.Sprintf("Could Get Server List.\nError: %v", err))
			}

			// Quit For Loop if no Server list
			if len(serverList) == 0 {
				break
			}

			for _, server := range serverList {
				// [ ] Log this
				services.SendUpdateMeRequest(server.ServerIP, server.Username, server.Password)
			}
			time.Sleep(5 * time.Minute)
		}
	}()

	// [ ] Look for a better way to call this function instead of using go-routines
	if err = dialer.ScanForUpdatesOnStart(userconfing_logger); err != nil {
		userconfing_logger.Critical(fmt.Sprintf("Error in Scan For Updates on Start.\nError: %v", err))
	}

	if err := grpcServer.Serve(lis); err != nil {
		userconfing_logger.Critical(fmt.Sprintf("Error Listening gRPC Server.\nError: %v", err))
		userconfing_logger.Critical("Closing Server...")
		os.Exit(1)
	}
}
