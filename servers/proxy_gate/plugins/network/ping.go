package network

import (
	"fmt"
	"os"

	"github.com/bafbi/minecraft-network/servers/proxy_gate/plugins/network/players"
	. "github.com/bafbi/minecraft-network/servers/proxy_gate/util"
	"github.com/bafbi/minecraft-network/servers/proxy_gate/util/mini"
	"github.com/go-logr/logr"
	"go.minekube.com/common/minecraft/color"
	c "go.minekube.com/common/minecraft/component"
	"go.minekube.com/gate/pkg/edition/java/proto/version"
	"go.minekube.com/gate/pkg/edition/java/proxy"
)

func onPing(log logr.Logger) func(*proxy.PingEvent) {
	podName, exists := os.LookupEnv("POD_NAME")
	if !exists {
		podName = "unknow"
	}
	message := fmt.Sprintf("Join (%s)", podName)
	// log.Info("Ping message", "message", message)
	line2 := mini.Gradient(
		message,
		c.Style{Bold: c.True},
		*color.Yellow.RGB, *color.Gold.RGB, *color.Red.RGB,
	)

	return func(e *proxy.PingEvent) {
		clientVersion := version.Protocol(e.Connection().Protocol())
		line1 := mini.Gradient(
			fmt.Sprintf("Hey %s user!\n", clientVersion),
			c.Style{},
			*color.White.RGB, *color.LightPurple.RGB,
		)

		p := e.Ping()
		p.Description = Join(line1, line2)

		playerCount := players.GetPlayerCount()
		p.Players.Online = playerCount
		p.Players.Max = p.Players.Online + 1

	}
}
