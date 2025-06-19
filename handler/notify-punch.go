package handler

import (
	"fmt"
	"log"
	"math/rand"
	"net"
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
		defer udp_conn.Close()
		time.Sleep(5 * time.Second)
		log.Println("Initializing UDP NAT Hole Punching")
		err = dialer.WorkspaceOwnerUdpNatPunching(udp_conn, peer_addr, clientHandlerName)
		if err != nil {
			log.Println("Error while Performing UDP NAT Hole Punching:", err)
			log.Println("Source: HandleNotifyToPunch()")
			return
		}

		log.Println("Starting New New Server `Connection` server on local port:", local_port)
		StartNewNewServer(udp_conn, clientHandlerName)
	}()

	// Sending Response to Server
	ip_port_split := strings.Split(myPublicIP, ":")
	myPublicIPOnly := ip_port_split[0]
	myPublicPortOnly := ip_port_split[1]
	return myPublicIPOnly, myPublicPortOnly, nil
}

func StartNewNewServer(udp_conn *net.UDPConn, clientHandlerName string) {
	log.Println("ClientHandler"+clientHandlerName, "Started")
	err := RegisterName("ClientHandler"+clientHandlerName, &ClientHandler{})
	if err != nil {
		log.Println("Error while Register ClientHandler:", err)
		log.Println("Source: StartNewNewServer()")
		return
	}

	kcp_lis, err := kcp.ListenWithOptionsAndConn(udp_conn, nil, 0, 0)
	if err != nil {
		log.Println("Error while Listening KCP With Options & Conn:", err)
		log.Println("Source: StartNewNewServer()")
		return
	}
	log.Println("Started New KCP Server Started ...")

	err = kcp_lis.SetReadDeadline(time.Now().Add(5 * time.Minute))
	if err != nil {
		log.Println("Error while Setting Deadline for KCP Listener:", err)
		log.Println("Source: StartNewNewServer()")
		return
	}

	for {
		kcp_session, err := kcp_lis.AcceptKCP()
		if err != nil {
			log.Println("Error while Accepting KCP from KCP Listener:", err)
			log.Println("Source: StartNewNewServer()")
			kcp_lis.Close()
			log.Println("Closing NewNewServer ...")
			return
		}
		log.Println("New Incoming Connection in NewNewServer from:", kcp_session.RemoteAddr())

		// KCP Params for Congestion Control
		kcp_session.SetWindowSize(128, 1024)
		kcp_session.SetNoDelay(1, 10, 2, 1)
		kcp_session.SetACKNoDelay(true)
		kcp_session.SetDSCP(46)

		go func() {
			defer kcp_session.Close()
			log.Println("Deciding the Type of Session ...")

			var buff [3]byte
			_, err = kcp_session.Read(buff[:])
			if err != nil {
				log.Println("Error while Reading the type of Session(KCP-RPC or KCP-Plain):", err)
				log.Println("Source: StartNewNewServer()")
				return
			}
			log.Println("Type of Session Received from Listener ...")

			kcp_buff := [3]byte{'K', 'C', 'P'}
			rpc_buff := [3]byte{'R', 'P', 'C'}

			if buff == kcp_buff {
				log.Println("KCP-Plain:", kcp_session.RemoteAddr().String())
				GetDataHandler(kcp_session)
			} else if buff == rpc_buff {
				log.Println("KCP-RPC:", kcp_session.RemoteAddr().String())
				ServeConn(kcp_session)
			} else {
				log.Println("Unknown Type of Session Sent:", string(buff[:]))
			}
		}()
	}
}
