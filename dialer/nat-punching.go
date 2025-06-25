package dialer

import (
	"fmt"
	"net"
	"strings"
	"time"
)

const (
	STUN_SERVER_ADDR = "stun.l.google.com:19302"
	PUNCH_ATTEMPTS   = 5 // Number of Punch Packets to Send
)

func WorkspaceOwnerUdpNatPunching(conn *net.UDPConn, peerAddr, clientHandlerName string) error {
	fmt.Println("Attempting to Dial Peer ...")
	peerUDPAddr, err := net.ResolveUDPAddr("udp", peerAddr)
	if err != nil {
		fmt.Println("Error while resolving UDP Addr:", err)
		fmt.Println("Source: WorkspaceOwnerUdpNatPunching()")
		return err
	}

	fmt.Println("Punching ", peerAddr)
	for range PUNCH_ATTEMPTS {
		conn.WriteToUDP([]byte("Punch"+";"+clientHandlerName), peerUDPAddr)
	}

	err = conn.SetReadDeadline(time.Now().Add(30 * time.Second))
	if err != nil {
		fmt.Println("Error while Setting Deadline during UDP NAT Hole Punching:", err)
		fmt.Println("Source: WorkspaceOwnerUdpNatPunching()")
		return err
	}

	// Reset Read Deadline, else it'll close conn during data transfers
	defer func() {
		err = conn.SetReadDeadline(time.Time{})
		if err != nil {
			fmt.Println("Error while Setting Deadline after UDP NAT Hole Punching:", err)
			fmt.Println("Source: WorkspaceOwnerUdpNatPunching()")
			return
		}
	}()

	var buff [512]byte
	for {
		n, addr, err := conn.ReadFromUDP(buff[0:])
		if err != nil {
			fmt.Println("Error while reading from Udp:", err)
			fmt.Println("Source: WorkspaceOwnerUdpNatPunching()")
			return err
		}
		msg := string(buff[:n])
		fmt.Printf("Received message: %s from %v\n", msg, addr)

		if addr.String() == peerAddr {
			fmt.Println("Expected User Messaged:", addr.String())
			if msg == "Punch" {
				_, err = conn.WriteToUDP([]byte("Punch ACK"+";"+clientHandlerName), peerUDPAddr)
				if err != nil {
					fmt.Println("Error while Writing 'Punch ACK;clientHandlerName':", err)
					fmt.Println("Source: WorkspaceOwnerUdpNatPunching()")
					continue
				}
				fmt.Println("Connection Established with", addr.String())
				return nil
			} else if msg == "Punch ACK" {
				fmt.Println("Connection Established with", addr.String())
				return nil
			} else {
				fmt.Println("Something Else is in Message:", msg)
			}
		} else {
			fmt.Println("Unexpected User Messaged:", addr.String())
			fmt.Println(msg)
		}
	}
}

func WorkspaceListenerUdpNatHolePunching(conn *net.UDPConn, peerAddr string) (string, error) {
	fmt.Println("Attempting to Dial Peer ...")
	peerUDPAddr, err := net.ResolveUDPAddr("udp", peerAddr)
	if err != nil {
		fmt.Println("Error while resolving UDP Addr:", err)
		fmt.Println("Source: WorkspaceListenerUdpNatHolePunching()")
		return "", err
	}

	fmt.Println("Punching ", peerAddr)
	for range PUNCH_ATTEMPTS {
		conn.WriteToUDP([]byte("Punch"), peerUDPAddr)
	}

	err = conn.SetReadDeadline(time.Now().Add(30 * time.Second))
	if err != nil {
		fmt.Println("Error while Setting Deadline during UDP NAT Hole Punching:", err)
		fmt.Println("Source: WorkspaceListenerUdpNatHolePunching()")
		return "", err
	}

	// Reset Read Deadline, else it'll close conn during data transfers
	defer func() {
		err = conn.SetReadDeadline(time.Time{})
		if err != nil {
			fmt.Println("Error while Setting Deadline after UDP NAT Hole Punching:", err)
			fmt.Println("Source: WorkspaceListenerUdpNatHolePunching()")
			return
		}
	}()

	var buff [512]byte
	for {
		n, addr, err := conn.ReadFromUDP(buff[0:])
		if err != nil {
			fmt.Println("Error while reading from Udp:", err)
			fmt.Println("Source: WorkspaceListenerUdpNatHolePunching()")
			return "", err
		}
		msg := string(buff[:n])
		fmt.Printf("Received message: %s from %v\n", msg, addr)

		if addr.String() == peerAddr {
			fmt.Println("Expected User Messaged:", addr.String())
			if strings.HasPrefix(msg, "Punch") {
				clientHandlerName := strings.Split(msg, ";")[1]
				_, err = conn.WriteToUDP([]byte("Punch ACK"), peerUDPAddr)
				if err != nil {
					fmt.Println("Error while Writing Punch ACK:", err)
					fmt.Println("Source: WorkspaceListenerUdpNatHolePunching()")
					continue
				}
				fmt.Println("Connection Established with", addr.String())
				return clientHandlerName, nil
			} else if strings.HasPrefix(msg, "Punch ACK") {
				fmt.Println("Connection Established with", addr.String())
				clientHandlerName := strings.Split(msg, ";")[1]
				return clientHandlerName, nil
			} else {
				fmt.Println("Something Else is in Message:", msg)
			}
		} else {
			fmt.Println("Unexpected User Messaged:", addr.String())
		}
	}
}
