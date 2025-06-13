package dialer

import (
	"context"
	"fmt"
	"net/rpc"
	"time"

	"github.com/ccding/go-stun/stun"
)

const (
	CLIENT_BASE_HANDLER_NAME = "ClientHandler"
	CONTEXT_TIMEOUT          = 25 * time.Second
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

func CallKCP_RPC_WithContext(args, reply any, rpc_name string, rpc_client *rpc.Client) error {
	ctx, cancel := context.WithTimeout(context.Background(), CONTEXT_TIMEOUT)
	defer cancel()

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
