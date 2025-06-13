package models

type WSMessage struct {
	MessageType string // Error, NotifyToPunchResponse
	Message     any
}
