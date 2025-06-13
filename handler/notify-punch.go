package handler

import (
	"fmt"
	"log"
	"math/rand"
	"net"
	"net/rpc"
	"slices"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/ButterHost69/PKr-Base/dialer"
	"github.com/ButterHost69/PKr-Base/utils"

	"github.com/ButterHost69/kcp-go"
)

type ClientHandlerNameManager struct {
	sync.Mutex
	RandomStringList []string
}

var clientHandlerNameManager = ClientHandlerNameManager{
	RandomStringList: []string{},
}

func HandleNotifyToPunch(peer_addr string) (string, string, error) {
	local_port := rand.Intn(16384) + 16384
	fmt.Println("My Local Port:", local_port)

	// Get My Public IP
	myPublicIP, err := dialer.GetMyPublicIP(local_port)
	if err != nil {
		fmt.Println("Error while Getting my Public IP:", err)
		fmt.Println("Source: HandleNotifyToPunch()")
		return "", "", err
	}
	fmt.Println("My Public IP Addr:", myPublicIP)

	udp_local_addr, err := net.ResolveUDPAddr("udp", ":"+strconv.Itoa(local_port))
	if err != nil {
		fmt.Println("Error while Resolving UDP Addr for Random Local Port:", err)
		fmt.Println("Source: HandleNotifyToPunch()")
		return "", "", err
	}

	// Creating UDP Conn to Perform UDP NAT Hole Punching
	udp_conn, err := net.ListenUDP("udp", udp_local_addr)
	if err != nil {
		fmt.Printf("Error while Listening to %d: %v\n", local_port, err)
		fmt.Println("Source: HandleNotifyToPunch()")
		return "", "", err
	}

	// Creating Unique ClientHandlerName
	clientHandlerName := utils.RandomString(4)
	for slices.Contains(clientHandlerNameManager.RandomStringList, clientHandlerName) {
		clientHandlerName = utils.RandomString(4)
	}

	go func() {
		time.Sleep(5 * time.Second)
		log.Println("Initializing UDP NAT Hole Punching")
		err = dialer.WorkspaceOwnerUdpNatPunching(udp_conn, peer_addr, clientHandlerName)
		if err != nil {
			log.Println("Error while Performing UDP NAT Hole Punching:", err)
			return
		}

		log.Println("Starting New New Server `Connection` server on local port:", local_port)
		// TODO Start Reciever on private ip
		// TODO Pass context to close server in 5min
		StartNewNewServer(udp_conn, clientHandlerName)
		udp_conn.Close()
	}()

	// Sending Response to Server
	ip_port_split := strings.Split(myPublicIP, ":")
	myPublicIPOnly := ip_port_split[0]
	myPublicPortOnly := ip_port_split[1]
	return myPublicIPOnly, myPublicPortOnly, nil
}

func StartNewNewServer(conn *net.UDPConn, clientHandlerName string) {
	log.Println("ClientHandler"+clientHandlerName, "Started")
	err := rpc.RegisterName("ClientHandler"+clientHandlerName, &ClientHandler{})
	if err != nil {
		log.Println("Error while Register ClientHandler:", err)
		log.Println("Source: StartNewNewServer()")
		return
	}

	lis, err := kcp.ListenWithOptionsAndConn(conn, nil, 0, 0)
	if err != nil {
		log.Println("Error while Listening KCP With Options & Conn:", err)
		log.Println("Source: StartNewNewServer()")
		return
	}
	log.Println("Started New KCP Server Started ...")

	err = lis.SetReadDeadline(time.Now().Add(10 * time.Minute))
	if err != nil {
		log.Println("Error while Setting Deadline for KCP Listener:", err)
		log.Println("Source: StartNewNewServer()")
		return
	}

	for {
		session, err := lis.AcceptKCP()
		if err != nil {
			log.Println("Error while Accepting KCP from KCP Listener:", err)
			log.Println("Source: StartNewNewServer()")
			conn.Close()
			lis.Close()
			log.Println("Closing NewNewServer ...")
			return
		}
		session.SetWindowSize(128, 512)
		session.SetNoDelay(1, 20, 0, 1)
		session.SetACKNoDelay(false)

		go rpc.ServeConn(session)
	}
}
