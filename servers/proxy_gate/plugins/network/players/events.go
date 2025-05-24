package players

import (
	"fmt"

	"github.com/bafbi/minecraft-network/servers/proxy_gate/util/metadata"
	"github.com/go-logr/logr"
	"go.minekube.com/gate/pkg/edition/java/proxy"
)

var eventsLog logr.Logger

// InitEvents initializes the events module, primarily for logging.
func InitEvents(log logr.Logger) {
	eventsLog = log.WithName("PlayerEvents")
}

// TrackServerConnection returns an event handler for when a player connects to a server.
func TrackServerConnection(
	fnGetServerMeta func(string) (metadata.Metadata, bool),
	chatZoneLabel string,
	chatListenZonesAnnotation string,
) func(e *proxy.ServerConnectedEvent) {
	return func(e *proxy.ServerConnectedEvent) {
		player := e.Player()
		server := e.Server()
		serverInfo := server.ServerInfo()
		playerID := player.ID()

		serverMeta, exists := fnGetServerMeta(serverInfo.Name())
		if !exists {
			eventsLog.Error(fmt.Errorf("server not found"), "Server not found for player tracking", "serverName", serverInfo.Name(), "player", player.Username())
			return
		}
		serverChatZone, serverChatZoneExists := serverMeta.GetLabel(chatZoneLabel)

		var previousServerChatZone string
		var previousServerChatZoneExists bool
		if previousServer := e.PreviousServer(); previousServer != nil {
			if prevServerMeta, ok := fnGetServerMeta(previousServer.ServerInfo().Name()); ok {
				previousServerChatZone, previousServerChatZoneExists = prevServerMeta.GetLabel(chatZoneLabel)
			}
		}

		err := UpdateMetadataByUUID(playerID, func(meta *metadata.Metadata) {
			meta.SetAnnotation("network/location", serverInfo.Name())
			meta.SetAnnotation("player/name", player.GameProfile().Name)
			if previousServerChatZoneExists {
				meta.RemoveAnnotationStringValue(chatListenZonesAnnotation, previousServerChatZone)
			}
			if serverChatZoneExists {
				meta.AddAnnotationStringValue(chatListenZonesAnnotation, serverChatZone)
			}
		})
		if err != nil {
			eventsLog.Error(err, "Failed to update player metadata on server connect", "player", player.Username())
			return
		}
		eventsLog.Info("Tracked player server connection", "player", player.Username(), "server", serverInfo.Name())
	}
}

// TrackDisconnect returns an event handler for when a player disconnects.
func TrackDisconnect() func(e *proxy.DisconnectEvent) {
	return func(e *proxy.DisconnectEvent) {
		player := e.Player()
		err := RemovePlayer(player.ID()) // Uses API to remove from KV and cache
		if err != nil {
			eventsLog.Error(err, "Failed to remove player metadata on disconnect", "player", player.Username())
			return
		}
		eventsLog.Info("Player disconnected, tracking removed", "player", player.Username())
	}
}
