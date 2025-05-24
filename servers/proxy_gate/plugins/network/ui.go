package network

import (
	"fmt"

	"github.com/go-logr/logr"
	"go.minekube.com/common/minecraft/color"
	c "go.minekube.com/common/minecraft/component"
	"go.minekube.com/gate/pkg/edition/java/proxy"
)

func addServerToTabList(p *proxy.Proxy, log logr.Logger) func(e *proxy.ServerPostConnectEvent) {
	return func(e *proxy.ServerPostConnectEvent) {
		serverName := e.Player().CurrentServer().Server().ServerInfo().Name()

		header := &c.Text{
			Content: fmt.Sprintf("\nWelcome %s on my network!\n", e.Player().Username()),
			S:       c.Style{Color: color.Yellow, Bold: c.True},
		}

		footer := &c.Text{
			Content: fmt.Sprintf("You are connected to %s\n", serverName),
			S:       c.Style{Color: color.Gray},
		}

		err := e.Player().TabList().SetHeaderFooter(header, footer)
		if err != nil {
			log.Error(err, "Failed to set tab list header and footer", "player", e.Player().Username())
		} else {
			log.Info("Set tab list header and footer", "player", e.Player().Username(), "server", serverName)
		}
	}
}
