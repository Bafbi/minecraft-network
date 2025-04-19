package network

import (
	"math/rand"
	"strings"

	"github.com/go-logr/logr"
	"go.minekube.com/common/minecraft/color"
	c "go.minekube.com/common/minecraft/component"
	"go.minekube.com/gate/pkg/edition/java/proxy"
)

func handleKickedFromServerEvent(p *proxy.Proxy, log logr.Logger) func(e *proxy.KickedFromServerEvent) {
	return func(e *proxy.KickedFromServerEvent) {
		// Find any server with type=lobby label as fallback
		lobbyServers := findServersByLabels(map[string]string{"type": "lobby"})
		// log the lobby servers find
		log.Info("Lobby servers found", "servers", lobbyServers)

		if len(lobbyServers) > 0 {
			// Choose random lobby
			targetServer := lobbyServers[rand.Intn(len(lobbyServers))]

			e.SetResult(&proxy.RedirectPlayerKickResult{
				Server: targetServer,
				Message: &c.Text{
					Extra: []c.Component{
						&c.Text{Content: "You have been redirected to another server",
							S: c.Style{Color: color.Green}},
						&c.Text{Content: "\n"},
						&c.Text{Content: "Reason: ", S: c.Style{Color: color.Red}},
						e.OriginalReason(),
					},
				},
			})
			log.Info("Redirected player to lobby server", "player", e.Player().Username(), "server", targetServer.ServerInfo().Addr())
		} else {
			e.SetResult(&proxy.DisconnectPlayerKickResult{
				Reason: &c.Text{Content: "No available server to redirect to."},
			})
			log.Info("No available server to redirect player", "player", e.Player().Username())
		}
	}
}

func isServerShutdownReason(reason c.Component) bool {
	// Check common shutdown messages
	if text, ok := reason.(*c.Text); ok {
		content := strings.ToLower(text.Content)
		shutdownPhrases := []string{"server closed", "shutting down", "restart", "maintenance", "Internal server connection error"}
		for _, phrase := range shutdownPhrases {
			if strings.Contains(content, phrase) {
				return true
			}
		}
	}
	return false
}
