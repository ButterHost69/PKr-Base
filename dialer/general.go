package dialer

import "github.com/ccding/go-stun/stun"

func GetMyPublicIP (port int) (string, error){
	stunClient := stun.NewClient()
  	stunClient.SetServerAddr("stun.l.google.com:19302")
  	stunClient.SetLocalPort(port)

	  _, myExtAddr, err := stunClient.Discover()
  	if err != nil {
		return "", err
  	}

	return myExtAddr.String(), err
}