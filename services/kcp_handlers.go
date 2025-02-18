package services

import (
	"fmt"
	"math/rand"
	"strconv"
	"strings"

	"github.com/ButterHost69/PKr-Base/dialer"
)


func (h *KCPHandler) NotifyToPunch(req NotifyToPunchRequest, res *NotifyToPunchResponse) error {
	dialPort := rand.Intn(16384) + 16384 

	publicIPAddr, err := dialer.GetMyPublicIP(dialPort)
	if err != nil {
		res.Response = 500
		return err
	}

	privateIP := fmt.Sprintf(":%d", dialPort)
	sendersIPAddr := fmt.Sprintf("%s:%s", req.SendersIP, req.SendersPort)
	
	if err = dialer.PuchToIP(sendersIPAddr, privateIP); err != nil {
		res.Response = 500
		return err
	}

	publicIP := strings.Split(publicIPAddr, ":")[0]
	publicPort := strings.Split(publicIPAddr, ":")[1]
	
	i_publicPort, err := strconv.Atoi(publicPort)
	if err != nil {
		res.Response = 500
		return err
	}

	res.Response = 200
	res.RecieversPublicIP = publicIP
	res.RecieversPublicPort = i_publicPort

	return nil
}
