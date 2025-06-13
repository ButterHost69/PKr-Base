package models

type User struct {
	Username string
	Password string
}

type WSMessage struct {
	MessageType string // Error, NotifyToPunchResponse
	Message     any
}
