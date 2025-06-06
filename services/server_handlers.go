package services

import (
	"fmt"
	"log"
	"math/rand"
	"net"
	"slices"
	"strconv"
	"strings"

	"github.com/ButterHost69/PKr-Base/dialer"
	"github.com/ButterHost69/PKr-Base/utils"
)

func (h *ServerHandler) NotifyToPunch(req NotifyToPunchRequest, res *NotifyToPunchResponse) error {
	h.UserConfingLogger.Info("Notify To Punch Called ...")
	log.Println("Notify to Punch Called")

	dialPort := rand.Intn(16384) + 16384
	myPublicIPAddr, err := dialer.GetMyPublicIP(dialPort)
	if err != nil {
		res.Response = 500
		h.UserConfingLogger.Critical("Unable to Get Public IP\nSource: NotifyToPunch\nError: " + err.Error())
		return fmt.Errorf("Unable to Get Public IP\nSource: NotifyToPunch\nError: %v", err)
	}
	log.Println("My New Public IP Addr:", myPublicIPAddr)
	h.UserConfingLogger.Info("My New Public IP Addr: " + myPublicIPAddr)

	privateIPStr := ":" + strconv.Itoa(dialPort)
	privateIP, err := net.ResolveUDPAddr("udp", privateIPStr)
	if err != nil {
		h.UserConfingLogger.Critical("Error Occured while Resolving Private UDP Addr\nSource: NotifyToPunch\nError:" + err.Error())
		return fmt.Errorf("Error Occured while Resolving Private UDP Addr\nSource: NotifyToPunch\nError:%v", err)
	}
	log.Println("Resolving UDP Addr")

	udpConn, err := net.ListenUDP("udp", privateIP)
	if err != nil {
		h.UserConfingLogger.Critical("Error Occured while Listening to UDP\nSource: NotifyToPunch\nError:" + err.Error())
		return fmt.Errorf("Error Occured while Listening to UDP\nSource: NotifyToPunch\nError:%v", err)
	}

	sendersIPAddr := fmt.Sprintf("%s:%s", req.SendersIP, req.SendersPort)

	myPublicIPOnly := strings.Split(myPublicIPAddr, ":")[0]
	myPublicPortOnlyStr := strings.Split(myPublicIPAddr, ":")[1]

	myPublicPortOnlyInt, err := strconv.Atoi(myPublicPortOnlyStr)
	if err != nil {
		res.Response = 500
		h.UserConfingLogger.Critical(fmt.Sprintf("Unable to Convert myPublicPortOnlyStr to Integer\nSource: NotifyToPunch\nError:%v", myPublicPortOnlyStr))
		return fmt.Errorf("Unable to Convert myPublicPortOnlyStr to Integer\nSource: NotifyToPunch\nError:%v", myPublicPortOnlyStr)
	}

	clientHandlerName := utils.RandomString(4)
	for slices.Contains(h.RandomStringList, clientHandlerName) {
		clientHandlerName = utils.RandomString(4)
	}

	log.Println("Setting My Public IP & Port into Response")
	res.Response = 200
	res.RecieversPublicIP = myPublicIPOnly
	res.RecieversPublicPort = myPublicPortOnlyInt

	go func() {
		log.Println("Initializing UDP NAT Hole Punching")
		err = dialer.UdpNatPunching(udpConn, sendersIPAddr, clientHandlerName)
		if err != nil {
			h.UserConfingLogger.Critical("Unable to Perform NAT Hole Punching\nSource: NotifyToPunch\nError:" + err.Error())
			return
		}

		h.UserConfingLogger.Info("Starting New New Server `Connection` server on local port: " + strconv.Itoa(dialPort))
		// TODO Start Reciever on private ip
		// TODO Pass context to close server in 5min
		StartNewNewServer(udpConn, h.WorkspaceLogger, h.UserConfingLogger, clientHandlerName)
		udpConn.Close()
	}()
	log.Println("Sending Response to Server After")

	return nil
}
