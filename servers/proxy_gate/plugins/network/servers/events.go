package servers

import (
	"github.com/go-logr/logr"
	"github.com/robinbraemer/event"
	"go.minekube.com/common/minecraft/color"
	c "go.minekube.com/common/minecraft/component"
	"go.minekube.com/gate/pkg/edition/java/proxy"
)

var eventsLog logr.Logger

// InitEvents initializes the server events module.
func InitEvents(p *proxy.Proxy, log logr.Logger) {
	eventsLog = log.WithName("ServerEvents")

	event.Subscribe(p.Event(), 0, playerChooseInitialServer(p, log.WithName("PlayerChooseInitialServer")))
	event.Subscribe(p.Event(), 0, handleKickedFromServerEvent(p, log.WithName("KickedFromServer")))
}

// HandlePlayerChooseInitialServer provides an event handler for initial server selection.
func playerChooseInitialServer(p *proxy.Proxy, log logr.Logger) func(*proxy.PlayerChooseInitialServerEvent) {
	return func(e *proxy.PlayerChooseInitialServerEvent) {
		player := e.Player()
		chosen, found := GetRandomDefaultServer() // Uses API

		if !found {
			log.Info("No default servers available for initial connection", "player", player.Username())
			player.Disconnect(&c.Text{
				Content: "No default servers available.",
				S:       c.Style{Color: color.Red},
			})
			return
		}
		e.SetInitialServer(chosen)
		log.Info("Routed player to default server", "player", player.Username(), "server", chosen.ServerInfo().Name())
	}
}

func handleKickedFromServerEvent(p *proxy.Proxy, log logr.Logger) func(e *proxy.KickedFromServerEvent) {
	return func(e *proxy.KickedFromServerEvent) {
		// Find any server with type=lobby label as fallback
		fallbackServers, some := GetRandomDefaultServer()
		if !some {
			e.SetResult(&proxy.DisconnectPlayerKickResult{
				Reason: &c.Text{Content: "No available server to redirect to."},
			})
			log.Info("No available server to redirect player", "player", e.Player().Username())
			return
		}

		e.SetResult(&proxy.RedirectPlayerKickResult{
			Server: fallbackServers,
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
		log.Info("Redirected player to lobby server", "player", e.Player().Username(), "server", fallbackServers.ServerInfo().Addr())

	}
}
