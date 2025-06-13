package models

import "sync"

type NotifyToPunchRequest struct {
	ListenerUsername   string `json:"listener_username"`
	ListenerPublicIP   string `json:"listener_public_ip"`
	ListenerPublicPort string `json:"listener_public_port"`
}

type NotifyToPunchResponse struct {
	WorkspaceOwnerPublicIP   string `json:"workspace_owner_public_ip"`
	WorkspaceOwnerPublicPort string `json:"workspace_owner_public_port"`
	ListenerUsername         string `json:"listener_username"`
}

type NotifyNewPushToListeners struct {
	WorkspaceOwnerUsername string `json:"workspace_owner_username"`
	WorkspaceName          string `json:"workspace_name"`
	NewWorkspaceHash       string `json:"workspace_new_hash"`
}

type RequestPunchFromReceiverRequest struct {
	ListenerUsername       string `json:"listener_username"`
	ListenerPublicIP       string `json:"listener_public_ip"`
	ListenerPublicPort     string `json:"listener_public_port"`
	WorkspaceOwnerUsername string `json:"workspace_owner_username"`
}

type RequestPunchFromReceiverResponse struct {
	Error                    string `json:"error"`
	WorkspaceOwnerUsername   string `json:"workspace_owner_username"`
	WorkspaceOwnerPublicIP   string `json:"workspace_owner_public_ip"`
	WorkspaceOwnerPublicPort string `json:"workspace_owner_public_port"`
}

type RequestPunchFromReceiverResponseMap struct {
	sync.RWMutex
	Map map[string]RequestPunchFromReceiverResponse
}
