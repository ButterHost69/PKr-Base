package services

import (
	"fmt"
	"log"
	"net"
	"net/rpc"
	"time"

	"github.com/ButterHost69/PKr-Base/logger"
	"github.com/ButterHost69/kcp-go"
)

func InitKCPServer(conn *net.UDPConn, workspace_logger *logger.WorkspaceLogger, userconfing_logger *logger.UserLogger) error {
	handler := ServerHandler{
		WorkspaceLogger:   workspace_logger,
		UserConfingLogger: userconfing_logger,
	}

	err := rpc.Register(&handler)
	if err != nil {
		userconfing_logger.Critical(fmt.Sprintf("Could Not Register KCP RPC to Handler...Error: %v", err))
		return err
	}

	lis, err := kcp.ServeConn(nil, 0, 0, conn)
	if err != nil {
		userconfing_logger.Critical(fmt.Sprintf("Could Not Start the KCP Server...\nError: %v", err))
		return err
	}

	userconfing_logger.Info("Started KCP Server at Port: ")
	for {
		session, err := lis.AcceptKCP()
		if err != nil {
			userconfing_logger.Critical(fmt.Sprint("Error accepting KCP connection: ", err))
			continue
		}

		remoteAddr := session.RemoteAddr().String()
		userconfing_logger.Info("New incoming connection from " + remoteAddr)
		session.SetNoDelay(0, 15000, 0, 0)
		session.SetDeadline(time.Now().Add(30 * time.Second)) // Overall timeout
		session.SetACKNoDelay(false)                          // Batch ACKs to reduce traffic

		// userconfing_logger.Info(session.Read())
		// Wrap connection and pass it to RPC
		go rpc.ServeConn(session)

	}

	// return nil
}

// TODO Close Server if no Connections in 5 Min... IDK How ??
func StartNewNewServer(conn *net.UDPConn, workspace_logger *logger.WorkspaceLogger, userconfing_logger *logger.UserLogger) {
	err := rpc.Register(&ClientHandler{
		WorkspaceLogger:   workspace_logger,
		UserConfingLogger: userconfing_logger,
	})
	if err != nil {
		userconfing_logger.Critical(fmt.Sprintf("Could Not Register Client RPC to Handler...Error: %v", err))
		return
	}

	lis, err := kcp.ListenWithOptionsAndConn(conn, nil, 0, 0)
	if err != nil {
		userconfing_logger.Critical(fmt.Sprintf("Could Not Start the Client Server...\nError: %v", err))
		return
	}

	userconfing_logger.Info("Started New KCP Server START ...")
	log.Println("Started New KCP Server Started ...")
	rpc.Accept(lis)
	userconfing_logger.Info("Started New KCP Server END ...")
}
