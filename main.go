package main

import (
	"log"
	"net/url"
	"os"
	"os/signal"
	"strings"

	"github.com/ButterHost69/PKr-Base/config"
	"github.com/ButterHost69/PKr-Base/ws"

	"github.com/gorilla/websocket"
)

var WEBSOCKET_SERVER_ADDR_WITH_QUERY url.URL

func init() {
	servers, err := config.GetAllServers()
	if err != nil {
		log.Println("Failed to Get all Servers from Config:", err)
		log.Println("Source: init()")
		os.Exit(1)
	}

	// TODO: Handle multiple server urls
	for _, server := range servers {
		escaped_username := url.QueryEscape(server.Username)
		escaped_password := url.QueryEscape(server.Password)
		ws.MY_USERNAME = server.Username
		ws.MY_SERVER_IP = server.ServerIP
		ws.MY_SERVER_ALIAS = server.ServerAlias

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

	conn, _, err := websocket.DefaultDialer.Dial(WEBSOCKET_SERVER_ADDR_WITH_QUERY.String(), nil)
	if err != nil {
		log.Println("Error while Dialing Websocket Connection to Server:", err)
		log.Println("Source: main()")
		return
	}
	defer conn.Close()

	log.Println("Connected to Server")
	done := make(chan struct{})

	go ws.ReadJSONMessage(done, conn)
	go ws.PingPongWriter(done, conn)

	select {
	case <-done:
	case <-interrupt:
		log.Println("Interrupt Received, Closing Connection ...")

		err := conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, "Bye"))
		if err != nil {
			log.Println("Error:", err)
			return
		}
		conn.Close()
	}
}
