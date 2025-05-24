package network

import (
	"fmt"
	"strings"

	"github.com/bafbi/minecraft-network/servers/proxy_gate/plugins/network/players"
	"github.com/bafbi/minecraft-network/servers/proxy_gate/plugins/network/servers"
	"github.com/bafbi/minecraft-network/servers/proxy_gate/util/metadata"
	"github.com/bafbi/minecraft-network/servers/proxy_gate/util/mini"
	"github.com/go-logr/logr"
	"go.minekube.com/brigodier"
	"go.minekube.com/gate/pkg/command"
	"go.minekube.com/gate/pkg/edition/java/proxy"
)

// Register all commands
func registerCommands(p *proxy.Proxy, log logr.Logger) {
	p.Command().Register(registerJoinCommand(p, log))
	p.Command().Register(metadataCommand(p, log))
	// p.Command().Register(registerServerSelectCommand(p, log))
	// p.Command().Register(registerListServersCommand(p, log))
}

// Join command implementation - unchanged
func registerJoinCommand(p *proxy.Proxy, log logr.Logger) brigodier.LiteralNodeBuilder {
	executeJoin := command.Command(func(ctx *command.Context) error {
		sender, ok := ctx.Source.(proxy.Player)
		if !ok {
			return ctx.Source.SendMessage(mini.Parse("<red>Only players can use this command</red>"))
		}

		targetName := ctx.String("player")
		meta, exists := players.GetMetadataByName(targetName)
		if !exists {
			return sender.SendMessage(mini.Parse(fmt.Sprintf("<red>Player %s is not registered in the network</red>", targetName)))
		}
		serverName, exists := (&meta).GetAnnotation("network/location")
		if !exists {
			return sender.SendMessage(mini.Parse(fmt.Sprintf("<red>Player %s does not have a server assigned</red>", targetName)))
		}

		targetServer, found := servers.GetRegisteredServerByName(serverName)
		if !found {
			return sender.SendMessage(mini.Parse(fmt.Sprintf("<red>Server %s is not registered in the network</red>", serverName)))
		}

		// Notify the user
		sender.SendMessage(mini.Parse(fmt.Sprintf("<green>Connecting you to %s's server (<yellow>%s</yellow>)...</green>", targetName, serverName)))

		// Connect to the server
		_, err := sender.CreateConnectionRequest(targetServer).Connect(ctx)
		if err != nil {
			log.Error(err, "Failed to connect player", "player", sender.Username(), "target", targetName)
			return sender.SendMessage(mini.Parse("<red>Failed to connect to server</red>"))
		}

		return nil
	})

	return brigodier.Literal("join").
		Then(brigodier.Argument("player", brigodier.StringWord).Suggests(suggestNetworkPlayers()).
			Executes(executeJoin))
}

func metadataCommand(p *proxy.Proxy, log logr.Logger) brigodier.LiteralNodeBuilder {
	executeJoin := command.Command(func(ctx *command.Context) error {
		sender, ok := ctx.Source.(proxy.Player)
		if !ok {
			return ctx.Source.SendMessage(mini.Parse("<red>Only players can use this command</red>"))
		}

		targetName := ctx.String("target_player")
		meta, exists := players.GetMetadataByName(targetName)
		if !exists {
			return sender.SendMessage(mini.Parse(fmt.Sprintf("<red>No metadata found for player %s</red>", targetName)))
		}

		// Build the response message using MiniMessage
		var builder strings.Builder
		builder.WriteString(fmt.Sprintf("<gold>Metadata for <yellow>%s</yellow>:</gold>\n", targetName))
		builder.WriteString("<gold>Labels:</gold>\n")
		if len(meta.Labels) == 0 {
			builder.WriteString("<gray>  (None)</gray>\n")
		} else {
			for k, v := range meta.Labels {
				builder.WriteString(fmt.Sprintf("<gray>  %s: %s</gray>\n", k, v))
			}
		}
		builder.WriteString("<gold>Annotations:</gold>\n")
		if len(meta.Annotations) == 0 {
			builder.WriteString("<gray>  (None)</gray>\n")
		} else {
			for k, v := range meta.Annotations {
				builder.WriteString(fmt.Sprintf("<gray>  %s: %s</gray>\n", k, v))
			}
		}
		return sender.SendMessage(mini.Parse(builder.String()))
	})

	execPlayerSetAnnotation := command.Command(func(ctx *command.Context) error {
		sender, ok := ctx.Source.(proxy.Player)
		if !ok {
			return ctx.Source.SendMessage(mini.Parse("<red>Only players can use this command</red>"))
		}
		targetName := ctx.String("target_player")
		annotation := ctx.String("input")
		parts := strings.SplitN(annotation, "=", 2)
		if len(parts) != 2 {
			return sender.SendMessage(mini.Parse("<red>Invalid annotation format. Use key=value</red>"))
		}
		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])
		if key == "" || value == "" {
			return sender.SendMessage(mini.Parse("<red>Key and value cannot be empty</red>"))
		}
		players.UpdateMetadataByName(targetName, func(meta *metadata.Metadata) {
			meta.Annotations[key] = value
		})
		sender.SendMessage(mini.Parse(fmt.Sprintf("<green>Set annotation <yellow>%s=%s</yellow> for player <yellow>%s</yellow></green>", key, value, targetName)))
		return nil
	})

	return brigodier.Literal("metadata").
		Then(brigodier.Literal("player").Then(brigodier.Argument("target_player", brigodier.StringWord).Suggests(suggestNetworkPlayers()).
			Executes(executeJoin).Then(brigodier.Literal("set").Then(brigodier.Literal("annotation").Then(brigodier.Argument("input", brigodier.StringPhrase).Executes(execPlayerSetAnnotation))))))
}

func suggestNetworkPlayers() brigodier.SuggestionProvider {
	return command.SuggestFunc(func(context *command.Context, builder *brigodier.SuggestionsBuilder) *brigodier.Suggestions {
		remaining := builder.RemainingLowerCase

		for _, name := range players.GetPlayersSlice() {
			if strings.HasPrefix(strings.ToLower(name), remaining) {
				builder.Suggest(name)
			}
		}

		return builder.Build()
	})
}
