package dialer

import (
	"log"
	"net"
	"strings"
	"time"
)

const (
	STUN_SERVER_ADDR = "stun.l.google.com:19302"
	PUNCH_ATTEMPTS   = 5 // Number of Punch Packets to Send
)

func WorkspaceOwnerUdpNatPunching(conn *net.UDPConn, peerAddr, clientHandlerName string) error {
	log.Println("Attempting to Dial Peer ...")
	peerUDPAddr, err := net.ResolveUDPAddr("udp", peerAddr)
	if err != nil {
		log.Println("Error while resolving UDP Addr:", err)
		log.Println("Source: WorkspaceOwnerUdpNatPunching()")
		return err
	}

	log.Println("Punching ", peerAddr)
	for range PUNCH_ATTEMPTS {
		conn.WriteToUDP([]byte("Punch"+";"+clientHandlerName), peerUDPAddr)
	}

	err = conn.SetReadDeadline(time.Now().Add(30 * time.Second))
	if err != nil {
		log.Println("Error while Setting Deadline during UDP NAT Hole Punching:", err)
		log.Println("Source: WorkspaceOwnerUdpNatPunching()")
		return err
	}

	// Reset Read Deadline, else it'll close conn during data transfers
	defer func() {
		err = conn.SetReadDeadline(time.Time{})
		if err != nil {
			log.Println("Error while Setting Deadline after UDP NAT Hole Punching:", err)
			log.Println("Source: WorkspaceOwnerUdpNatPunching()")
			return
		}
	}()

	var buff [512]byte
	for {
		n, addr, err := conn.ReadFromUDP(buff[0:])
		if err != nil {
			log.Println("Error while reading from Udp:", err)
			log.Println("Source: WorkspaceOwnerUdpNatPunching()")
			return err
		}
		msg := string(buff[:n])
		log.Printf("Received message: %s from %v\n", msg, addr)
		log.Println(peerAddr == addr.String())

		if addr.String() == peerAddr {
			log.Println("Expected User Messaged:", addr.String())
			if msg == "Punch" {
				_, err = conn.WriteToUDP([]byte("Punch ACK"+";"+clientHandlerName), peerUDPAddr)
				if err != nil {
					log.Println("Error while Writing 'Punch ACK;clientHandlerName':", err)
					log.Println("Source: WorkspaceOwnerUdpNatPunching()")
					continue
				}
				log.Println("Connection Established with", addr.String())
				return nil
			} else if msg == "Punch ACK" {
				log.Println("Connection Established with", addr.String())
				return nil
			} else {
				log.Println("Something Else is in Message:", msg)
			}
		} else {
			log.Println("Unexpected User Messaged:", addr.String())
			log.Println(msg)
		}
	}
}

func WorkspaceListenerUdpNatHolePunching(conn *net.UDPConn, peerAddr string) (string, error) {
	log.Println("Attempting to Dial Peer ...")
	peerUDPAddr, err := net.ResolveUDPAddr("udp", peerAddr)
	if err != nil {
		log.Println("Error while resolving UDP Addr:", err)
		log.Println("Source: WorkspaceListenerUdpNatHolePunching()")
		return "", err
	}

	log.Println("Punching ", peerAddr)
	for range PUNCH_ATTEMPTS {
		conn.WriteToUDP([]byte("Punch"), peerUDPAddr)
	}

	err = conn.SetReadDeadline(time.Now().Add(30 * time.Second))
	if err != nil {
		log.Println("Error while Setting Deadline during UDP NAT Hole Punching:", err)
		log.Println("Source: WorkspaceListenerUdpNatHolePunching()")
		return "", err
	}

	// Reset Read Deadline, else it'll close conn during data transfers
	defer func() {
		err = conn.SetReadDeadline(time.Time{})
		if err != nil {
			log.Println("Error while Setting Deadline after UDP NAT Hole Punching:", err)
			log.Println("Source: WorkspaceListenerUdpNatHolePunching()")
			return
		}
	}()

	var buff [512]byte
	for {
		n, addr, err := conn.ReadFromUDP(buff[0:])
		if err != nil {
			log.Println("Error while reading from Udp:", err)
			log.Println("Source: WorkspaceListenerUdpNatHolePunching()")
			return "", err
		}
		msg := string(buff[:n])
		log.Printf("Received message: %s from %v\n", msg, addr)
		log.Println(peerAddr == addr.String())

		if addr.String() == peerAddr {
			log.Println("Expected User Messaged:", addr.String())
			if strings.HasPrefix(msg, "Punch") {
				clientHandlerName := strings.Split(msg, ";")[1]
				_, err = conn.WriteToUDP([]byte("Punch ACK"), peerUDPAddr)
				if err != nil {
					log.Println("Error while Writing Punch ACK:", err)
					log.Println("Source: WorkspaceListenerUdpNatHolePunching()")
					continue
				}
				log.Println("Connection Established with", addr.String())
				return clientHandlerName, nil
			} else if strings.HasPrefix(msg, "Punch ACK") {
				log.Println("Connection Established with", addr.String())
				clientHandlerName := strings.Split(msg, ";")[1]
				return clientHandlerName, nil
			} else {
				log.Println("Something Else is in Message:", msg)
			}
		} else {
			log.Println("Unexpected User Messaged:", addr.String())
		}
	}
}
