package chatbot

type Status string

const (
	StatusConnecting   Status = "connecting"
	StatusConnected    Status = "connected"
	StatusDisconnected Status = "disconnected"
)
