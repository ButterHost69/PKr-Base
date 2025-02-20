package dialer

import (
	"fmt"
	"net/rpc"

	"github.com/ButterHost69/kcp-go"
	// "github.com/MohitSilwal16/kcp-go"
)

const (
	HANDLER_NAME = "Handler"
)

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

	if err := call(HANDLER_NAME+".Ping", req, &res, server_ip, ""); err != nil {

		return fmt.Errorf("error in Calling RPC...\nError: %v", err)
	}

	if res.Response != 200 {
		return fmt.Errorf("calling Ping Method was not Successful.\nReturn Code - %d", res.Response)
	}

	return nil
}
