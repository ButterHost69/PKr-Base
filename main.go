// Listening Server
// Listen For other Connections
// Responsible to Create the Server that will Send Data

package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/ButterHost69/PKr-Base/config"
	"github.com/ButterHost69/PKr-Base/dialer"
	"github.com/ButterHost69/PKr-Base/logger"
	"github.com/ButterHost69/PKr-Base/services"
	// "github.com/ButterHost69/PKr-Base/servicestemp"
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
	PORT				int
	LOG_IN_TERMINAL		bool
	LOG_LEVEL			int
)

// Loggers
var (
	workspace_logger	*logger.WorkspaceLogger
	userconfing_logger	*logger.UserLogger
	serverRpcHandler 	*dialer.CallHandler
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

	workspaces, err := config.GetAllGetWorkspaces()
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
			IP_ADDR = ":9069"
			PORT = 9069
		}
	} else {
		
		PORT, err = strconv.Atoi(strings.Split(IP_ADDR, ":")[1])
		if err != nil {
			log.Panic("Error: ", err)
		}
	}

	serverRpcHandler = &dialer.CallHandler{
		Lipaddr: "0.0.0.0:9091",
		WorkspaceLogger: workspace_logger,
		UserConfingLogger: userconfing_logger,
	}
	config.UpdateBasePort(IP_ADDR)

}

func main() {
	Init()

	// TODO: [ ] Test this code, neither human test nor code test done....
	// All The functions written with it are not tested
	go func() {
		userconfing_logger.Info("Update me Service Started")
		for {
			// Read Each Time... So can automatically detect changes without manual anything....
			serverList, err := config.GetAllServers()
			if err != nil {
				userconfing_logger.Debug(fmt.Sprintf("Could Get Server List.\nError: %v", err))
			}

			// Quit For Loop if no Server list
			if len(serverList) == 0 {
				break
			}

			myPublicIPPort, err := dialer.GetMyPublicIP(PORT)
			myPublicIp := strings.Split(myPublicIPPort, ":")[0]
			myPublicPort := strings.Split(myPublicIPPort, ":")[1]

			if err != nil {
				log.Panic("Error in getting Public IP: ", err)
			}
			for _, server := range serverList {
				if err := serverRpcHandler.CallPing(server.ServerIP, server.Username, server.Password, myPublicIp, myPublicPort); err != nil {
					userconfing_logger.Critical(err)
				}
			}
			time.Sleep(5 * time.Minute)
		}
	}()

	// [ ] Look for a better way to call this function instead of using go-routines
	if err := dialer.ScanForUpdatesOnStart(userconfing_logger); err != nil {
		userconfing_logger.Critical(fmt.Sprintf("Error in Scan For Updates on Start.\nError: %v", err))
	}

	err := services.InitKCPServer(IP_ADDR, workspace_logger, userconfing_logger)
	if err != nil {
		userconfing_logger.Critical(fmt.Sprintf("Error: %v\n", err))
	}

	userconfing_logger.Info(fmt.Sprintf("Base Service Running on Port: %s" , IP_ADDR))
}
