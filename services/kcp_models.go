package services

import "github.com/ButterHost69/PKr-Base/logger"

type KCPHandler struct {
	WorkspaceLogger   *logger.WorkspaceLogger
	UserConfingLogger *logger.UserLogger
}

type NotifyToPunchRequest struct {
	SendersUsername string
	SendersIP       string
	SendersPort     string
}

type NotifyToPunchResponse struct{
	RecieversPublicIP	string
	RecieversPublicPort	int

	Response int
}

