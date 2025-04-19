package network

import (
	"fmt"
	"strings"

	"github.com/go-logr/logr"
	"go.minekube.com/brigodier"
	"go.minekube.com/common/minecraft/color"
	c "go.minekube.com/common/minecraft/component"
	"go.minekube.com/gate/pkg/command"
	"go.minekube.com/gate/pkg/edition/java/proxy"
)

// Register all commands
func registerCommands(p *proxy.Proxy, log logr.Logger) {
	p.Command().Register(registerJoinCommand(p, log))
	// p.Command().Register(registerServerSelectCommand(p, log))
	// p.Command().Register(registerListServersCommand(p, log))
}

// Join command implementation - unchanged
func registerJoinCommand(p *proxy.Proxy, log logr.Logger) brigodier.LiteralNodeBuilder {
	executeJoin := command.Command(func(ctx *command.Context) error {
		sender, ok := ctx.Source.(proxy.Player)
		if !ok {
			return ctx.Source.SendMessage(&c.Text{
				Content: "Only players can use this command",
				S:       c.Style{Color: color.Red},
			})
		}

		targetName := ctx.String("player")

		meta := playersMetadata[targetName]
		serverName, exists := GetAnnotation(&meta, "network/location")
		if !exists {
			return sender.SendMessage(&c.Text{
				Content: fmt.Sprintf("Player %s is not registered in the network", targetName),
				S:       c.Style{Color: color.Red},
			})
		}

		serversMu.RLock()
		defer serversMu.RUnlock()
		// Check if the server is registered
		var targetServer proxy.RegisteredServer
		for _, server := range servers {
			if server.ServerInfo().Name() == serverName {
				targetServer = server
				break // Found the server, no need to continue
			}
		}

		if targetServer == nil {
			return sender.SendMessage(&c.Text{
				Content: fmt.Sprintf("Server %s is not registered in the network", serverName),
				S:       c.Style{Color: color.Red},
			})
		}

		// Notify the user
		sender.SendMessage(&c.Text{
			Content: fmt.Sprintf("Connecting you to %s's server (%s)...", targetName, serverName),
			S:       c.Style{Color: color.Green},
		})

		// Connect to the server
		_, err := sender.CreateConnectionRequest(targetServer).Connect(ctx)
		if err != nil {
			log.Error(err, "Failed to connect player", "player", sender.Username(), "target", targetName)
			return sender.SendMessage(&c.Text{
				Content: "Failed to connect to server",
				S:       c.Style{Color: color.Red},
			})
		}

		return nil
	})

	return brigodier.Literal("join").
		Then(brigodier.Argument("player", brigodier.StringWord).Suggests(suggestNetworkPlayers()).
			Executes(executeJoin))
}

func suggestNetworkPlayers() brigodier.SuggestionProvider {
	return command.SuggestFunc(func(context *command.Context, builder *brigodier.SuggestionsBuilder) *brigodier.Suggestions {
		remaining := builder.RemainingLowerCase

		for name, _ := range playersMetadata {
			if strings.HasPrefix(strings.ToLower(name), remaining) {
				builder.Suggest(name)
			}
		}

		return builder.Build()
	})
}
