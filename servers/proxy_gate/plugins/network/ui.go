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

		// Get server metadata safely
		serversMu.RLock()
		meta, ok := serversMetadata[serverName]
		serversMu.RUnlock()

		footerContent := fmt.Sprintf("\n# Server: %s\n", serverName)
		if ok {
			if len(meta.Labels) > 0 {
				footerContent += "Labels:\n"
				for k, v := range meta.Labels {
					footerContent += fmt.Sprintf("  %s: %s\n", k, v)
				}
			}
			if len(meta.Annotations) > 0 {
				footerContent += "Annotations:\n"
				for k, v := range meta.Annotations {
					footerContent += fmt.Sprintf("  %s: %s\n", k, v)
				}
			}
		}

		footer := &c.Text{
			Content: footerContent,
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
