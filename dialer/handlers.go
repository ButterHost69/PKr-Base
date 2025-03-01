package dialer

import (
	"context"
	"fmt"
	"net"
	"net/rpc"
	"time"

	"github.com/ButterHost69/kcp-go"
	// "github.com/MohitSilwal16/kcp-go"
)

const (
	HANDLER_NAME = "Handler"
)

func callWithContextAndConn(ctx context.Context, rpcname string, args interface{}, reply interface{}, ripaddr string, udpconn *net.UDPConn) error {
	// Dial the remote address
	conn, err := kcp.DialWithConnAndOptions(ripaddr, nil, 0, 0, udpconn)
	if err != nil {
		return err
	}

	// Find a Way to close the kcp conn without closing UDP Connection
	// defer conn.Close()

	c := rpc.NewClient(conn)
	// defer c.Close()

	// Create a channel to handle the RPC call with context
	done := make(chan error, 1)
	go func() {
		done <- c.Call(rpcname, args, reply)
	}()

	select {
	case <-ctx.Done():
		// if err := c.Close(); err != nil {
		// 	return fmt.Errorf("RPC call timed out - %s\nAlso Error in Closing RPC %v", ripaddr, err)
		// }
		return fmt.Errorf("RPC call timed out - %s", ripaddr)
	case err := <-done:
		// if cerr := c.Close(); err != nil {
		// 	return fmt.Errorf("%v, Also Error in Closing RPC %v", err, cerr)
		// }
		return err
	}
}

func call(rpcname string, args interface{}, reply interface{}, ripaddr, lipaddr string) error {

	conn, err := kcp.Dial(ripaddr, lipaddr)
	if err != nil {
		return err
	}
	defer conn.Close()

	c := rpc.NewClient(conn)
	defer c.Close()

	err = c.Call(rpcname, args, reply)
	if err != nil {
		return err
	}

	return nil
}

func (h *CallHandler) CallPing(server_ip, username, password, public_ip, public_port string) error {

	var req PingRequest
	var res PingResponse

	req.Username = username
	req.Password = password
	req.PublicIP = public_ip
	req.PublicPort = public_port

	ctx, cancel := context.WithTimeout(context.Background(), 25*time.Second)
	defer cancel()

	if err := callWithContextAndConn(ctx, HANDLER_NAME+".Ping", req, &res, server_ip, h.Conn); err != nil {
		return fmt.Errorf("Error while Calling %s.Ping RPC...\nSource: CallPing\nError: %v", HANDLER_NAME, err)
	}

	if res.Response != 200 {
		return fmt.Errorf("calling Ping Method was not Successful.\nReturn Code - %d", res.Response)
	}

	return nil
}
