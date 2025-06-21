package main

import (
	"log"
	"net"
	"net/url"
	"os"
	"os/signal"
	"strings"
	"time"

	"github.com/ButterHost69/PKr-Base/config"
	"github.com/ButterHost69/PKr-Base/dialer"
	"github.com/ButterHost69/PKr-Base/ws"

	"github.com/gorilla/websocket"
)

var WEBSOCKET_SERVER_ADDR_WITH_QUERY url.URL
var SERVER config.ServerConfig

func init() {
	servers, err := config.GetAllServers()
	if err != nil {
		log.Println("Failed to Get all Servers from Config:", err)
		log.Println("Source: init()")
		os.Exit(1)
	}

	if len(servers) == 0 {
		log.Println("No Server're found in Config\nExiting Base ...")
		os.Exit(1)
	}

	// Will Pick the last server from user config & maitain ws connection with that
	for _, server := range servers {
		escaped_username := url.QueryEscape(server.Username)
		escaped_password := url.QueryEscape(server.Password)
		ws.MY_USERNAME = server.Username
		ws.MY_SERVER_IP = server.ServerIP
		ws.MY_SERVER_ALIAS = server.ServerAlias
		SERVER = server

		raw_query := "username=" + escaped_username + "&password=" + escaped_password
		websock_server_ip := strings.Split(server.ServerIP, ":")[0]

		WEBSOCKET_SERVER_ADDR_WITH_QUERY = url.URL{
			Scheme:   "ws",
			Host:     websock_server_ip + ":8080",
			Path:     "/ws",
			RawQuery: raw_query,
		}
	}
}

func main() {
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	var server_err error
	ws_conn, _, server_err := websocket.DefaultDialer.Dial(WEBSOCKET_SERVER_ADDR_WITH_QUERY.String(), nil)
	for server_err != nil {
		// Check if error is because server is offline
		if opErr, ok := server_err.(*net.OpError); ok {
			// Check if it's a syscall error underneath
			if sysErr, ok := opErr.Err.(*os.SyscallError); ok {
				if strings.Contains(sysErr.Error(), "actively refused") {
					log.Println("Server seems offline")
					log.Println("Failed to Connect to PKr-Server")
					log.Println("Will be retrying in 5 Mins")

					ws_conn, _, server_err = websocket.DefaultDialer.Dial(WEBSOCKET_SERVER_ADDR_WITH_QUERY.String(), nil)
					select {
					case <-time.After(15 * time.Minute):
						ws_conn, _, server_err = websocket.DefaultDialer.Dial(WEBSOCKET_SERVER_ADDR_WITH_QUERY.String(), nil)

					case <-interrupt:
						log.Println("Interrupt received. Exiting Program...")
						return
					}
					continue
				}
			} else {
				log.Println("Error while Dialing Websocket Connection to Server: ", server_err)
				log.Println("Source: main()")
				return
			}
		} else {
			log.Println("Error while Dialing Websocket Connection to Server: ", server_err)
			log.Println("Source: main()")
			return
		}
	}

	defer ws_conn.Close()
	log.Println("Connected to Server")

	done := make(chan struct{})

	go ws.ReadJSONMessage(done, ws_conn)
	go ws.PingPongWriter(done, ws_conn)

	log.Println("Preparing gRPC Client ...")
	// New GRPC Client
	gRPC_cli_service_client, err := dialer.NewGRPCClients(SERVER.ServerIP)
	if err != nil {
		log.Println("Error:", err)
		log.Println("Description: Cannot Create New GRPC Client")
		log.Println("Source: Install()")
		return
	}

	log.Println("Checking for New Changes")
	// Checking for New Changes
	for _, get_workspace := range SERVER.GetWorkspaces {
		log.Println("GET Workspace: ")
		log.Println(get_workspace)
		are_there_new_changes, err := dialer.CheckForNewChanges(gRPC_cli_service_client, get_workspace.WorkspaceName, get_workspace.WorkspaceOwnerName, SERVER.Username, SERVER.Password, get_workspace.LastPushNum)
		if err != nil {
			log.Println("Error while Checking For New Changes:", err)
			log.Println("Source: main()")
			continue
		}
		log.Println("Are there new changes:", are_there_new_changes)

		if are_there_new_changes {
			ws.PullWorkspace(get_workspace.WorkspaceOwnerName, get_workspace.WorkspaceName, ws_conn)
		}
	}
	log.Println("Done with Checking for New Changes ...")

	select {
	case <-done:
	case <-interrupt:
		log.Println("Interrupt Received, Closing Connection ...")

		err := ws_conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, "Bye"))
		if err != nil {
			log.Println("Error while Writing Close Message to Server via WS:", err)
			log.Println("Source: main()")
			return
		}
	}
}
