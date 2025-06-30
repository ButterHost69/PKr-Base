package models

import "sync"

type NotifyToPunchRequest struct {
	ListenerUsername    string `json:"listener_username"`
	ListenerPublicIp    string `json:"listener_public_ip"`
	ListenerPublicPort  string `json:"listener_public_port"`
	ListenerPrivateIp   string `json:"listener_private_ip"`
	ListenerPrivatePort string `json:"listener_private_port"`
}

type NotifyToPunchResponse struct {
	WorkspaceOwnerPublicIp    string `json:"workspace_owner_public_ip"`
	WorkspaceOwnerPublicPort  string `json:"workspace_owner_public_port"`
	WorkspaceOwnerPrivateIp   string `json:"workspace_owner_private_ip"`
	WorkspaceOwnerPrivatePort string `json:"workspace_owner_private_port"`
	ListenerUsername          string `json:"listener_username"`
}

type NotifyNewPushToListeners struct {
	WorkspaceOwnerUsername string `json:"workspace_owner_username"`
	WorkspaceName          string `json:"workspace_name"`
	NewWorkspacePushNum    int    `json:"workspace_new_push_num"`
}

type RequestPunchFromReceiverRequest struct {
	ListenerUsername       string `json:"listener_username"`
	ListenerPublicIp       string `json:"listener_public_ip"`
	ListenerPublicPort     string `json:"listener_public_port"`
	ListenerPrivateIp      string `json:"listener_private_ip"`
	ListenerPrivatePort    string `json:"listener_private_port"`
	WorkspaceOwnerUsername string `json:"workspace_owner_username"`
}

type RequestPunchFromReceiverResponse struct {
	Error                     string `json:"error"`
	WorkspaceOwnerUsername    string `json:"workspace_owner_username"`
	WorkspaceOwnerPublicIp    string `json:"workspace_owner_public_ip"`
	WorkspaceOwnerPublicPort  string `json:"workspace_owner_public_port"`
	WorkspaceOwnerPrivateIp   string `json:"workspace_owner_private_ip"`
	WorkspaceOwnerPrivatePort string `json:"workspace_owner_private_port"`
}

type WorkspaceOwnerIsOnline struct {
	Error              string `json:"error"`
	WorkspaceOwnerName string `json:"workspace_owner_username"`
}

type RequestPunchFromReceiverResponseMap struct {
	sync.RWMutex
	Map map[string]RequestPunchFromReceiverResponse
}
