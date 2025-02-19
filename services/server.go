package services

import (
	"fmt"
	"net/rpc"

	"github.com/ButterHost69/PKr-Base/logger"
	"github.com/ButterHost69/kcp-go"
)

func InitKCPServer(port string, workspace_logger *logger.WorkspaceLogger, userconfing_logger *logger.UserLogger) error {

	handler := ServerHandler{
		WorkspaceLogger:   workspace_logger,
		UserConfingLogger: userconfing_logger,
	}

	err := rpc.Register(&handler)
	if err != nil {
		userconfing_logger.Critical(fmt.Sprintf("Could Not Register KCP RPC to Handler...Error: %v", err))
		return err
	}

	lis, err := kcp.Listen(port)
	if err != nil {
		userconfing_logger.Critical(fmt.Sprintf("Could Not Start the KCP Server...\nError: %v", err))
		return err
	}

	userconfing_logger.Info("Started KCP Server...")
	rpc.Accept(lis)

	return nil
}

// TODO Close Server if no Connections in 5 Min... IDK How ??
func StartNewNewServer(port string, workspace_logger *logger.WorkspaceLogger, userconfing_logger *logger.UserLogger){
	err := rpc.Register(&Handler{
		WorkspaceLogger:  workspace_logger,
		UserConfingLogger: userconfing_logger,
	})
	if err != nil {
		userconfing_logger.Critical(fmt.Sprintf("Could Not Register Client RPC to Handler...Error: %v", err))
		return
	}

	lis, err := kcp.Listen(port)
	if err != nil {
		userconfing_logger.Critical(fmt.Sprintf("Could Not Start the Client Server...\nError: %v", err))
		return 
	}

	userconfing_logger.Info("Started KCP Server...")
	rpc.Accept(lis)
	
}
