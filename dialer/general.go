package dialer

import (
	"github.com/ButterHost69/kcp-go"
	"github.com/ccding/go-stun/stun"
)

const (
	PUNCH_ATTEMPTS = 5
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

func PuchToIP(senderIP, privateIP string) error {
	conn, err := kcp.Dial(senderIP, privateIP)
	if err != nil {
		return err
	}
	defer conn.Close()

	for i := 0; i < PUNCH_ATTEMPTS; i++ {
		conn.Write([]byte("Punch"))
	}

	return nil
}
