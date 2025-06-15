package dialer

import (
	"context"
	"fmt"
	"log"

	pb "github.com/ButterHost69/PKr-Base/pb"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func NewGRPCClients(address string) (pb.CliServiceClient, error) {
	conn, err := grpc.NewClient(address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}
	return pb.NewCliServiceClient(conn), nil
}

func CheckForNewChanges(grpc_client pb.CliServiceClient, workspace_name, workspace_owner_name, listener_username, listener_password, last_hash string) (bool, error) {
	log.Println("Preparing gRPC Request ...")
	// Prepare req
	req := &pb.GetLastHashOfWorkspaceRequest{
		WorkspaceOwner:   workspace_owner_name,
		WorkspaceName:    workspace_name,
		ListenerUsername: listener_username,
		ListenerPassword: listener_password,
	}

	// Request Timeout
	ctx, cancelFunc := context.WithTimeout(context.Background(), CONTEXT_TIMEOUT)
	defer cancelFunc()

	log.Println("Sending gRPC Request ...")
	// Sending Request ...
	res, err := grpc_client.GetLastHashOfWorkspace(ctx, req)
	if err != nil {
		fmt.Println("Error:", err)
		fmt.Println("Description: Cannot Register User")
		fmt.Println("Source: Install()")
		return false, err
	}
	log.Println("Latest Hash Received from Server:", res.LastHash)
	log.Println("My Latest Hash:", last_hash)
	return res.LastHash != last_hash, nil
}
