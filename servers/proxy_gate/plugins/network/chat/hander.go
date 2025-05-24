package chat

import (
	"encoding/json"

	"github.com/bafbi/minecraft-network/servers/proxy_gate/plugins/network/constants"
	"github.com/go-logr/logr"
	"github.com/nats-io/nats.go"
	"go.minekube.com/common/minecraft/color"
	c "go.minekube.com/common/minecraft/component"
	"go.minekube.com/gate/pkg/edition/java/proxy"
)

// CreatePlayerChatEventHandler creates an event handler for player chat messages.
// This handler publishes the chat message to NATS.
func CreatePlayerChatEventHandler(nc *nats.Conn, log logr.Logger) func(*proxy.PlayerChatEvent) {
	return func(e *proxy.PlayerChatEvent) {
		if !e.Allowed() {
			return
		}
		// Cancel the event so that the message is not sent to the backend server by Gate.
		e.SetAllowed(false)

		player := e.Player()
		currentServer := player.CurrentServer()
		if currentServer == nil {
			log.Error(nil, "Player has no current server, cannot send chat message", "player", player.Username())
			_ = player.SendMessage(&c.Text{Content: "Error: Not connected to a server.", S: c.Style{Color: color.Red}})
			return
		}
		serverName := currentServer.Server().ServerInfo().Name()

		payload := NetworkChatMessagePayload{
			PlayerID: player.ID(),
			Server:   serverName,
			Username: player.Username(),
			Message:  e.Message(),
		}

		data, err := json.Marshal(payload)
		if err != nil {
			log.Error(err, "Failed to marshal chat message payload", "player", player.Username())
			_ = player.SendMessage(&c.Text{Content: "Error sending message (marshal failed).", S: c.Style{Color: color.Red}})
			return
		}

		if err := nc.Publish(constants.ChatChannelSubject, data); err != nil {
			log.Error(err, "Failed to publish chat message to NATS", "player", player.Username())
			_ = player.SendMessage(&c.Text{Content: "Error sending message (publish failed).", S: c.Style{Color: color.Red}})
			return
		}
		log.V(1).Info("Published chat message to NATS", "player", player.Username(), "message", e.Message())
	}
}
