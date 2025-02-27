package services

import (
	"fmt"
	"math/rand"
	"strconv"
	"strings"

	"github.com/ButterHost69/PKr-Base/dialer"
)

func (h *ServerHandler) NotifyToPunch(req NotifyToPunchRequest, res *NotifyToPunchResponse) error {
	h.UserConfingLogger.Info("Notify To Punch Called ...")

	dialPort := rand.Intn(16384) + 16384

	publicIPAddr, err := dialer.GetMyPublicIP(dialPort)
	if err != nil {
		res.Response = 500
		h.UserConfingLogger.Critical("Unable to Get Public IP")
		return err
	}

	privateIP := fmt.Sprintf(":%d", dialPort)
	sendersIPAddr := fmt.Sprintf("%s:%s", req.SendersIP, req.SendersPort)

	if err = dialer.PuchToIP(sendersIPAddr, privateIP); err != nil {
		res.Response = 500
		h.UserConfingLogger.Debug(fmt.Sprintf("Unable to Punch to ip - %s", sendersIPAddr))
		return err
	}
	
	h.UserConfingLogger.Info(fmt.Sprintf("Punched Successfully to ip - %s", sendersIPAddr))
	
	publicIP := strings.Split(publicIPAddr, ":")[0]
	publicPort := strings.Split(publicIPAddr, ":")[1]

	i_publicPort, err := strconv.Atoi(publicPort)
	if err != nil {
		res.Response = 500
		h.UserConfingLogger.Critical(fmt.Sprint("unable to convert port to integer", publicPort))
		return err
	}

	res.Response = 200
	res.RecieversPublicIP = publicIP
	res.RecieversPublicPort = i_publicPort
	
	h.UserConfingLogger.Info("Starting New New Server `Connection` server")
	// TODO Start Reciever on private ip
	// TODO Pass context to close server in 5min
	go StartNewNewServer(strconv.Itoa(dialPort), h.WorkspaceLogger, h.UserConfingLogger)
	return nil
}
