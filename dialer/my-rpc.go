package dialer

import (
	"context"
	"log"
	"net/rpc"

	"github.com/ButterHost69/PKr-Base/models"
)

type ClientCallHandler struct{}

func (h *ClientCallHandler) CallGetPublicKey(clientHandlerName string, rpc_client *rpc.Client) ([]byte, error) {
	var req models.PublicKeyRequest
	var res models.PublicKeyResponse

	ctx, cancel := context.WithTimeout(context.Background(), CONTEXT_TIMEOUT)
	defer cancel()

	rpc_name := CLIENT_BASE_HANDLER_NAME + clientHandlerName + ".GetPublicKey"
	if err := CallKCP_RPC_WithContext(ctx, req, &res, rpc_name, rpc_client); err != nil {
		log.Println("Error while Calling GetPublicKey:", err)
		log.Println("Source: CallGetPublicKey()")
		return nil, err
	}
	return res.PublicKey, nil
}

func (h *ClientCallHandler) CallInitNewWorkSpaceConnection(workspace_name, my_username, server_ip, workspace_password string, my_public_key []byte, clientHandlerName string, rpc_client *rpc.Client) error {
	var req models.InitWorkspaceConnectionRequest
	var res models.InitWorkspaceConnectionResponse

	req.WorkspaceName = workspace_name
	req.MyUsername = my_username
	req.MyPublicKey = my_public_key

	req.ServerIP = server_ip
	req.WorkspacePassword = workspace_password

	ctx, cancel := context.WithTimeout(context.Background(), CONTEXT_TIMEOUT)
	defer cancel()

	rpc_name := CLIENT_BASE_HANDLER_NAME + clientHandlerName + ".InitNewWorkSpaceConnection"
	if err := CallKCP_RPC_WithContext(ctx, req, &res, rpc_name, rpc_client); err != nil {
		log.Println("Error while Calling Init New Workspace Connection:", err)
		log.Println("Source: CallInitNewWorkSpaceConnection()")
		return err
	}
	return nil
}

func (h *ClientCallHandler) CallGetData(my_username, server_ip, workspace_name, workspace_password, last_hash, clientHandlerName string, rpc_client *rpc.Client) (*models.GetDataResponse, error) {
	var req models.GetDataRequest
	var res models.GetDataResponse

	req.Username = my_username
	req.WorkspaceName = workspace_name
	req.WorkspacePassword = workspace_password
	req.LastHash = last_hash
	req.ServerIP = server_ip

	ctx, cancel := context.WithTimeout(context.Background(), LONG_CONTEXT_TIMEOUT)
	defer cancel()

	rpc_name := CLIENT_BASE_HANDLER_NAME + clientHandlerName + ".GetData"
	if err := CallKCP_RPC_WithContext(ctx, req, &res, rpc_name, rpc_client); err != nil {
		log.Println("Error while Calling Get Data:", err)
		log.Println("Source: CallGetData()")
		return nil, err
	}
	return &res, nil
}
