package dialer

import "github.com/ButterHost69/PKr-Base/logger"

type CallHandler struct {
	Lipaddr string

	WorkspaceLogger   *logger.WorkspaceLogger
	UserConfingLogger *logger.UserLogger
}

type PingRequest struct {
	PublicIP   string
	PublicPort string

	Username string
	Password string
}

type PingResponse struct {
	Response int
}

// type RegisterUserRequest struct {
// 	PublicIP	string
// 	PublicPort	string

// 	Username	string
// 	Password	string
// }

// type RegisterUserResponse struct {
// 	UniqueUsername	string
// 	Response		int
// }
