package models

import "sync"

type NotifyToPunchRequest struct {
	ListenerUsername      string   `json:"listener_username"`
	ListenerPublicIP      string   `json:"listener_public_ip"`
	ListenerPublicPort    string   `json:"listener_public_port"`
	ListenerPrivateIPList []string `json:"listener_private_ip_list"`
	ListenerPrivatePort   string   `json:"listener_private_port"`
}

type NotifyToPunchResponse struct {
	WorkspaceOwnerPublicIP      string   `json:"workspace_owner_public_ip"`
	WorkspaceOwnerPublicPort    string   `json:"workspace_owner_public_port"`
	WorkspaceOwnerPrivateIPList []string `json:"workspace_owner_private_ip_list"`
	WorkspaceOwnerPrivatePort   string   `json:"workspace_owner_private_port"`
	ListenerUsername            string   `json:"listener_username"`
}

type NotifyNewPushToListeners struct {
	WorkspaceOwnerUsername string `json:"workspace_owner_username"`
	WorkspaceName          string `json:"workspace_name"`
	NewWorkspacePushNum    int    `json:"workspace_new_push_num"`
}

type RequestPunchFromReceiverRequest struct {
	ListenerUsername       string   `json:"listener_username"`
	ListenerPublicIP       string   `json:"listener_public_ip"`
	ListenerPublicPort     string   `json:"listener_public_port"`
	ListenerPrivateIPList  []string `json:"listener_private_ip_list"`
	ListenerPrivatePort    string   `json:"listener_private_port"`
	WorkspaceOwnerUsername string   `json:"workspace_owner_username"`
}

type RequestPunchFromReceiverResponse struct {
	Error                       string   `json:"error"`
	WorkspaceOwnerUsername      string   `json:"workspace_owner_username"`
	WorkspaceOwnerPublicIP      string   `json:"workspace_owner_public_ip"`
	WorkspaceOwnerPublicPort    string   `json:"workspace_owner_public_port"`
	WorkspaceOwnerPrivateIPList []string `json:"workspace_owner_private_ip_list"`
	WorkspaceOwnerPrivatePort   string   `json:"workspace_owner_private_port"`
}

type RequestPunchFromReceiverResponseMap struct {
	sync.RWMutex
	Map map[string]RequestPunchFromReceiverResponse
}
