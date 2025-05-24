package chat

import "go.minekube.com/gate/pkg/util/uuid"

// NetworkChatMessagePayload defines the structure of a chat message sent over NATS.
type NetworkChatMessagePayload struct {
	PlayerID uuid.UUID `json:"playerId"`
	Server   string    `json:"server"`
	Username string    `json:"username"`
	Message  string    `json:"message"`
}
