package dialer

import (
	"context"
	"fmt"
	"net"
	"net/rpc"
	"time"

	"github.com/ButterHost69/PKr-Base/logger"
	"github.com/ccding/go-stun/stun"
)

const (
	CLIENT_BASE_HANDLER_NAME = "ClientHandler"
	CONTEXT_TIMEOUT          = 45 * time.Second
	LONG_CONTEXT_TIMEOUT     = 10 * time.Minute
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

func CallKCP_RPC_WithContext(ctx context.Context, args, reply any, rpc_name string, rpc_client *rpc.Client) error {
	// Create a channel to handle the RPC call with context
	done := make(chan error, 1)
	go func() {
		done <- rpc_client.Call(rpc_name, args, reply)
	}()

	select {
	case <-ctx.Done():
		return fmt.Errorf("RPC call timed out")
	case err := <-done:
		return err
	}
}

// 192.168.65.1 & 192.168.137.1 are used as VM-Bridge, so ignoring those IP's
func isPrivateIPv4(ip net.IP) bool {
	return ip[0] == 10 ||
		(ip[0] == 172 && ip[1] >= 16 && ip[1] <= 31) ||
		(ip[0] == 192 && ip[1] == 168 && ip[2] != 65 && ip[3] != 1) ||
		(ip[0] == 192 && ip[1] == 168 && ip[2] != 137 && ip[3] != 1)
}

func ReturnListOfPrivateIPs() ([]string, error) {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		logger.LOGGER.Println("Error while Retrieving Address of all Network Interfaces:", err)
		logger.LOGGER.Println("Source: ReturnListOfPrivateIPs()")
		return nil, err
	}

	private_ips := []string{}
	for _, addr := range addrs {
		if ipNet, ok := addr.(*net.IPNet); ok && !ipNet.IP.IsLoopback() {
			if ip4 := ipNet.IP.To4(); ip4 != nil {
				// Check if it's a private IP
				if isPrivateIPv4(ip4) {
					private_ips = append(private_ips, ip4.String())
				}
			}
		}
	}
	return private_ips, nil
}
