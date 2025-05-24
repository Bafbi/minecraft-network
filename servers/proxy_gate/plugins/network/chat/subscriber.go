package chat

import (
	"encoding/json"
	"fmt"

	"github.com/bafbi/minecraft-network/servers/proxy_gate/plugins/network/constants"
	"github.com/bafbi/minecraft-network/servers/proxy_gate/plugins/network/players"
	"github.com/bafbi/minecraft-network/servers/proxy_gate/plugins/network/servers"
	"github.com/go-logr/logr"
	"github.com/nats-io/nats.go"
	"go.minekube.com/gate/pkg/edition/java/proxy"
)

// StartNatsSubscriber subscribes to NATS chat messages and broadcasts them to relevant local players.
func StartNatsSubscriber(
	p *proxy.Proxy,
	nc *nats.Conn,
	log logr.Logger,
) (*nats.Subscription, error) {
	if nc == nil {
		return nil, fmt.Errorf("NATS connection (nc) is not initialized for chat subscriber")
	}

	subscription, err := nc.Subscribe(constants.ChatChannelSubject, func(msg *nats.Msg) {
		var payload NetworkChatMessagePayload
		if err := json.Unmarshal(msg.Data, &payload); err != nil {
			log.Error(err, "Failed to unmarshal chat message from NATS", "data", string(msg.Data))
			return
		}

		log.V(1).Info("Received chat message from NATS", "payload", payload)

		// Determine sender's published layers
		playerMeta, playerMetaExists := players.GetMetadataByUUID(payload.PlayerID)
		serverMeta, serverMetaExists := servers.GetMetadataByName(payload.Server)

		var pubLayers []string
		var layersFound bool
		var err error

		if playerMetaExists {
			pubLayers, layersFound, err = (&playerMeta).GetAnnotationStringSlice(constants.ChatPubLayersAnnotation)
			if err != nil {
				log.Error(err, "Failed to get pub-layers from player metadata", "player", payload.Username)
			}
		}
		// Fallback to server's pub-layers if player has none
		if !layersFound && serverMetaExists {
			pubLayers, layersFound, err = (&serverMeta).GetAnnotationStringSlice(constants.ChatPubLayersAnnotation)
			if err != nil {
				log.Error(err, "Failed to get pub-layers from server metadata", "server", payload.Server)
			}
		}

		if !layersFound || len(pubLayers) == 0 {
			log.V(1).Info("No chat pub-layers found for player or server, skipping message",
				"sender", payload.Username, "server", payload.Server)
			return
		}

		// Find listeners
		listeners := make([]proxy.MessageSink, 0)
		for _, onlinePlayer := range p.Players() {
			// Get sub-layers for the listener
			listenerPlayerMeta, listenerPlayerMetaExists := players.GetMetadataByUUID(onlinePlayer.ID())

			var subLayers []string
			var subLayersFound bool
			var err error

			if listenerPlayerMetaExists {
				subLayers, subLayersFound, err = listenerPlayerMeta.GetAnnotationStringSlice(constants.ChatSubLayersAnnotation)
			}
			if err != nil {
				log.Error(err, "Failed to get sub-layers from player metadata", "player", onlinePlayer.Username())
			}
			// Fallback to server's sub-layers if player has none
			if !subLayersFound {
				playerCurrentSrv := onlinePlayer.CurrentServer()
				if playerCurrentSrv != nil {
					serverInfo := playerCurrentSrv.Server().ServerInfo()
					if serverInfo != nil {
						listenerServerName := serverInfo.Name()
						listenerServerMeta, listenerServerMetaExists := servers.GetMetadataByName(listenerServerName)
						if listenerServerMetaExists {
							subLayers, subLayersFound, err = listenerServerMeta.GetAnnotationStringSlice(constants.ChatSubLayersAnnotation)
							if err != nil {
								log.Error(err, "Failed to get sub-layers from server metadata", "server", listenerServerName)
							}
						}
					}
				}
			}

			// Check if any pubLayer is in subLayers
			shouldReceive := false
			if subLayersFound && len(subLayers) > 0 {
				for _, pubLayer := range pubLayers {
					for _, subLayer := range subLayers {
						if pubLayer == subLayer {
							shouldReceive = true
							break
						}
					}
					if shouldReceive {
						break
					}
				}
			}

			if shouldReceive {
				listeners = append(listeners, onlinePlayer)
			}
		}

		if len(listeners) > 0 {
			chatMsgComponent := formatChatMessage(payload.Username, payload.Message)
			proxy.BroadcastMessage(listeners, chatMsgComponent)
			log.V(1).Info("Broadcasted NATS chat message to local players",
				"sender", payload.Username, "listeners", len(listeners), "pubLayers", pubLayers)
		}

	})

	if err != nil {
		log.Error(err, "Failed to subscribe to NATS chat messages")
		return nil, err
	}
	log.Info("Subscribed to NATS chat messages", "subject", constants.ChatChannelSubject)
	return subscription, nil
}
