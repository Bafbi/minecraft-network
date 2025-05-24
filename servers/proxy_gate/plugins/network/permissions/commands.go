package permissions

import (
	"fmt"
	"strings"

	"github.com/bafbi/minecraft-network/servers/proxy_gate/plugins/network/players"
	"github.com/bafbi/minecraft-network/servers/proxy_gate/util/mini"
	"github.com/go-logr/logr"
	"go.minekube.com/brigodier"
	"go.minekube.com/common/minecraft/component"
	"go.minekube.com/gate/pkg/command"
	"go.minekube.com/gate/pkg/edition/java/proxy"
	"go.minekube.com/gate/pkg/util/uuid"
)

// RegisterCommands registers permission-specific commands.
func RegisterCommands(p *proxy.Proxy, log logr.Logger) {
	log.Info("Registering permission-specific commands")
	p.Command().Register(registerPermissionTestCommand(p, log))
}

func registerPermissionTestCommand(p *proxy.Proxy, log logr.Logger) brigodier.LiteralNodeBuilder {
	executor := command.Command(func(ctx *command.Context) error {
		sender := ctx.Source
		currentEnforcer := GetEnforcer()

		// Protect this command (same logic as before)
		if playerSender, ok := sender.(proxy.Player); ok {
			if currentEnforcer == nil {
				return sender.SendMessage(mini.Parse("<red>Casbin enforcer is not initialized. Cannot check command permission.</red>"))
			}
			isAdmin, err := currentEnforcer.HasRoleForUser(playerSender.ID().String(), "group:admin")
			if err != nil {
				log.Error(err, "Error checking admin role for permission check command", "player", playerSender.Username())
				return sender.SendMessage(mini.Parse("<red>Error checking your permissions to run this command.</red>"))
			}
			if !isAdmin {
				return sender.SendMessage(mini.Parse("<red>You must be in 'group:admin' to use this command.</red>"))
			}
		}

		subjectArg := ctx.String("subject_id")
		objectResource := ctx.String("object_resource")
		action := ctx.String("action")
		var subjectID string

		// Resolve subjectArg using players API
		parsedUUID, parseErr := uuid.Parse(subjectArg)
		if parseErr == nil {
			subjectID = parsedUUID.String()
		} else {
			// Try to resolve by name using players API
			if meta, found := players.GetMetadataByName(subjectArg); found {
				if uuidStr, ok := meta.Annotations["player/uuid"]; ok && uuidStr != "" {
					subjectID = uuidStr
				} else {
					subjectID = subjectArg
				}
			} else {
				subjectID = subjectArg
			}
		}

		if currentEnforcer == nil {
			return sender.SendMessage(mini.Parse("<red>Casbin enforcer is not initialized.</red>"))
		}

		log.Info("Executing permission check command",
			"sender", sender, "target_subject_id", subjectID, "object", objectResource, "action", action)

		allowed, err := currentEnforcer.Enforce(subjectID, objectResource, action)
		if err != nil {
			log.Error(err, "Error during Casbin Enforce check from command",
				"subject", subjectID, "object", objectResource, "action", action)
			return sender.SendMessage(mini.Parse(fmt.Sprintf("<red>Error checking permission: %s</red>", err.Error())))
		}

		var resultMsg *component.Text
		if allowed {
			resultMsg = mini.Parse(fmt.Sprintf("<green>Access <bold>GRANTED</bold> for subject '<yellow>%s</yellow>' to '<yellow>%s</yellow>' object '<yellow>%s</yellow>'</green>", subjectID, action, objectResource))
		} else {
			resultMsg = mini.Parse(fmt.Sprintf("<red>Access <bold>DENIED</bold> for subject '<yellow>%s</yellow>' to '<yellow>%s</yellow>' object '<yellow>%s</yellow>'</red>", subjectID, action, objectResource))
		}
		return sender.SendMessage(resultMsg)
	})

	return brigodier.Literal("permission").
		Then(brigodier.Literal("check").
			Then(brigodier.Argument("subject_id", brigodier.String).Suggests(suggestSubjectsDirect())).
			Then(brigodier.Argument("object_resource", brigodier.StringPhrase)).
			Then(brigodier.Argument("action", brigodier.StringWord).
				Executes(executor)))
}

func suggestSubjectsDirect() brigodier.SuggestionProvider {
	return command.SuggestFunc(func(ctx *command.Context, builder *brigodier.SuggestionsBuilder) *brigodier.Suggestions {
		currentArg := builder.RemainingLowerCase

		// Suggest sender's own UUID and username if they are a player
		if playerSender, ok := ctx.Source.(proxy.Player); ok {
			ownUUID := playerSender.ID().String()
			if strings.HasPrefix(strings.ToLower(ownUUID), currentArg) {
				builder.Suggest(ownUUID)
			}
			if strings.HasPrefix(strings.ToLower(playerSender.Username()), currentArg) {
				builder.Suggest(playerSender.Username())
			}
		}
		return builder.Build()
	})
}
