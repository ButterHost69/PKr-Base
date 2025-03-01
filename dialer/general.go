package dialer

import (
	"fmt"
	"net"
	"time"

	"github.com/ButterHost69/kcp-go"
	"github.com/ccding/go-stun/stun"
)

/*
	README
	NAT Hole Punching Flow:

	Step 0: Send Request to STUN Server
	Step 1: Send Punch Packet to Peer
	Step 2: Identify the type of user by listening to dialed connection
	Step 3: ONLY FOR User-A, start new listener

	User-A & User-B send request to STUN Server Discovering their Public IP & PORT
	User-A sends req to User-B, but it's dropped since B's Router doesn't have entry for A
	User-B sends req to User-A, it's accepted since A tried to send request earlier
*/

const (
	LOCAL_ADDRESS    = "0.0.0.0"
	STUN_SERVER_ADDR = "stun.l.google.com:19302"
	MAX_BUFFER_SIZE  = 512
	PUNCH_ATTEMPTS   = 5 // Number of Punch Packets to Send
	READ_TIMEOUT     = 10 * time.Second
)

func GetMyPublicIP(port int) (string, error) {
	stunClient := stun.NewClient()
	stunClient.SetServerAddr("stun.l.google.com:19302")
	stunClient.SetLocalPort(port)

	_, myExtAddr, err := stunClient.Discover()
	if err != nil && err.Error() != "Server error: no changed address" {
		return "", err
	}
	return myExtAddr.String(), nil
}

// TODO: Add Timeout & Maybe there's gonna be Timeout Prob

// Returns if NAT Hole Punching is Done
func handleConnection(conn net.Conn) {
	buff := make([]byte, 1024)
	for {
		n, err := conn.Read(buff)
		if err != nil {
			fmt.Println("Error While Reading from Conn during NAT Hole Punching\nSource: handleConnection\nError:", err)
			continue
		}
		msg := string(buff[:n])
		fmt.Printf("Received: %s from %v\n", msg, conn.RemoteAddr())

		if msg == "Punch" {
			conn.Write([]byte("Punch ACK"))
			fmt.Println("Connection Established with", conn.RemoteAddr().String())
			return
		} else if msg == "Punch ACK" {
			fmt.Println("Connection Established with", conn.RemoteAddr().String())
			return
		}
	}
}

func RudpNatPunching(udpConn *net.UDPConn, peerAddr string) error {
	// Step 1: Send Punch Packets
	fmt.Println("Attempting to Dial Peer ...")
	conn, err := kcp.DialWithConnAndOptions(peerAddr, nil, 0, 0, udpConn)
	if err != nil {
		fmt.Println("Error while Dialing With Conn & Options\nSource: RudpNatPunching\nError:", err)
		return err
	}

	fmt.Println("Punching ", peerAddr)
	for range PUNCH_ATTEMPTS {
		conn.Write([]byte("Punch"))
	}

	// Step 2: Identify whether it's User-A or User-B
	conn.SetReadDeadline(time.Now().Add(READ_TIMEOUT))

	buff := make([]byte, 1024)
	for {
		n, err := conn.Read(buff)
		if err != nil {
			if err.Error() == "timeout" {
				// No message received within the timeout; act as User-A
				fmt.Println("No message received. Acting as User-A (listener).")
				conn.Close()
				break
			} else {
				fmt.Println("Error Reading from Dialer's Conn\nSource: RudpNatPunching\nError:", err)
				return err
			}
		} else {
			// Message received; act as User-B
			fmt.Printf("Received message: %s. Acting as User-B (dialer).\n", string(buff[:n]))
			conn.SetReadDeadline(time.Time{})
			handleConnection(conn)
			return nil
		}
	}

	// Step 3: Only for User-A
	// If dialing failed or no message was received, start listening
	listener, err := kcp.ListenWithOptionsAndConn(udpConn, nil, 0, 0)
	if err != nil {
		fmt.Println("Error Occured while Listening With Options & Conn\nSource: RudpNatPunching\nError:", err)
		return err
	}

	fmt.Println("Waiting for incoming connection...")
	for {
		listenConn, err := listener.Accept()
		if err != nil {
			fmt.Println("Error Accepting Connection from New Listener\nSource: RudpNatPunching\nError:", err)
			return err
		}

		// Only Receive Connection from peerAddr
		if listenConn.RemoteAddr().String() == peerAddr {
			fmt.Println("Received connection from", listenConn.RemoteAddr().String())
			handleConnection(listenConn)
			return nil
		} else {
			fmt.Println("Ye Kutta Kyu Bhok rha h", listenConn.RemoteAddr())
		}
	}
}
