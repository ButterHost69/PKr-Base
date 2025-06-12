package dialer

import (
	"context"
	"fmt"
	"log"
	"net"
	rpc "github.com/ButterHost69/PKr-Base/prpc"
	"time"

	"github.com/ButterHost69/PKr-Base/logger"
	"github.com/ButterHost69/kcp-go"
	// "github.com/MohitSilwal16/kcp-go"
)

const (
	HANDLER_NAME = "Handler"
)


func callWithContextAndConn(ctx context.Context, rpcname string, args interface{}, reply interface{}, ripaddr string, udpconn *net.UDPConn, userConfingLogger *logger.UserLogger) error {
	log.Println("Calling - callWithContextAndConn")
	log.Println("Arg - ", args)
	// Dial the remote address
	conn, err := kcp.DialWithConnAndOptions(ripaddr, nil, 0, 0, udpconn)
	if err != nil {
		return err
	}
	// defer conn.Close()
	conn.SetWindowSize(2, 32)                               // Only 2 unacked packets maximum
	conn.SetWriteDeadline(time.Now().Add(10 * time.Second)) // Limits total retry time
	// conn.SetNoDelay(0, 15000, 0, 0)
	conn.SetNoDelay(1, 10, 0, 0) // No need to batch packets ; Faster ACKS ; Retransmit is only timeout expire
	conn.SetDeadline(time.Now().Add(30 * time.Second)) // Overall timeout
	// conn.SetACKNoDelay(false)                          // Batch ACKs to reduce traffic
	conn.SetACKNoDelay(true)                          // Doesnt Send ACK in batch to close conns faster so less resending

	// Find a Way to close the kcp conn without closing UDP Connection
	// defer conn.Close()

	c := rpc.NewClient(conn)
	// defer c.Close()

	// Create a channel to handle the RPC call with context
	done := make(chan error, 1)

	go func() {
		log.Println("Calling Call RPC Method")
		done <- c.Call(rpcname, args, reply)
	}()

	log.Println("RPC CALL Go Func Running...")

	select {
	case <-ctx.Done():
		log.Println("Closing KPC Connection Because of timeout")
		if err := c.Close(); err != nil {
			return fmt.Errorf("RPC call timed out - %s\nAlso Error in Closing RPC %v", ripaddr, err)
		}
		return fmt.Errorf("RPC call timed out - %s", ripaddr)
	case err := <-done:
		log.Println("Closing KPC Connection")
		if cerr := c.Close(); err != nil {
			return fmt.Errorf("%v, Also Error in Closing RPC %v", err, cerr)
		}
		log.Println(reply)
		return nil
	}
}

// Go func this ... it is a blocking function
// Does Not Handle Errors - Only Writes and Send Request...
// Access the errors if any using call.Error
func initiateCallWithContextAndConn(ctx context.Context, rpcname string, args interface{}, reply interface{}, ripaddr string, udpconn *net.UDPConn, userConfingLogger *logger.UserLogger) (*rpc.Call) {
	log.Println("Calling - callWithContextAndConn")
	log.Println("Arg - ", args)
	// Dial the remote address
	conn, err := kcp.DialWithConnAndOptions(ripaddr, nil, 0, 0, udpconn)
	if err != nil {
		return nil
	}

	c := rpc.NewClient(conn)
	// done <- c.Call(rpcname, args, reply)
	call := c.Go(rpcname, args, reply, nil)
	log.Printf("Calling Call RPC Method (%s) | Seq - %d | args - %v\n", rpcname, call)

	return call
}

func (h *CallHandler) CallPing(server_ip, username, password, public_ip, public_port string, ping_num int) error {

	var req PingRequest
	var res PingResponse

	req.Username = username
	req.Password = password
	req.PublicIP = public_ip
	req.PublicPort = public_port
	req.PingNum = ping_num

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	log.Println("Calling callWithContextAndConn()")
	if err := callWithContextAndConn(ctx, HANDLER_NAME+".Ping", req, &res, server_ip, h.Conn, h.UserConfingLogger); err != nil {
		cancel()
		return fmt.Errorf("Error while Calling %s.Ping RPC...\nSource: CallPing\nError: %v", HANDLER_NAME, err)
	}

	if res.Response != 200 {
		return fmt.Errorf("calling Ping Method was not Successful.\nReturn Code - %d", res.Response)
	}
	
	log.Println("Ping Num:", res.PingNum)
	cancel()
	log.Println("Called Cancel for - ", res.PingNum)

	return nil
}
