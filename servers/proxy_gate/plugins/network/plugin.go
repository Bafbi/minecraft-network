package network

import (
	"context"
	// "errors" // No longer needed for HasPermission here
	"fmt"
	"os"

	"github.com/bafbi/minecraft-network/servers/proxy_gate/plugins/network/permissions"

	"github.com/go-logr/logr"
	"github.com/nats-io/nats.go"
	"github.com/robinbraemer/event"
	"go.minekube.com/gate/pkg/edition/java/proxy"
)

// casbinPolicyUpdateSubject constant removed

var (
	nc *nats.Conn
	js nats.JetStreamContext
	// globalEnforcer removed
	pluginLog logr.Logger
)

var Plugin = proxy.Plugin{
	Name: "Network",
	Init: func(ctx context.Context, p *proxy.Proxy) error {
		baseLog := logr.FromContextOrDiscard(ctx)
		pluginLog = baseLog.WithName("NetworkPlugin")
		pluginLog.Info("Network plugin starting")

		// NATS Connection (remains the same)
		natsURL := os.Getenv("NATS_URL")
		if natsURL == "" {
			natsURL = "nats://network-nats:4222"
			pluginLog.Info("NATS_URL not set, using default", "url", natsURL)
		}
		var err error
		nc, err = nats.Connect(natsURL)
		if err != nil {
			return fmt.Errorf("failed to connect to NATS: %w", err)
		}
		// NATS close goroutine (remains the same)
		go func() {
			<-ctx.Done()
			if nc != nil {
				nc.Close()
				pluginLog.Info("NATS connection closed")
			}
		}()
		pluginLog.Info("Connected to NATS", "url", nc.ConnectedUrl())

		js, err = nc.JetStream()
		if err != nil {
			return fmt.Errorf("failed to get JetStream context: %w", err)
		}
		pluginLog.Info("Obtained JetStream context")

		// Initialize Player Management using modular system
		if err := InitPlayerSystem(js, pluginLog.WithName("Players")); err != nil {
			return err
		}

		// Initialize Server Management using modular system
		if err := InitServerSystem(p, js, pluginLog.WithName("Servers")); err != nil {
			return err
		}

		// Initialize Permission System (Casbin, commands, etc.)
		if err := InitPermissionSystem(ctx, p, nc, pluginLog.WithName("Permissions")); err != nil {
			return err
		}

		// --- Chat system initialization ---
		if err = InitChatSystem(p, nc, pluginLog.WithName("Chat")); err != nil {
			return err
		}

		// Register event handlers (remains the same, but internal calls might change)
		if err = registerEventHandlers(p, pluginLog.WithName("EventHandlers")); err != nil {
			return err
		}

		// Register commands (remains the same)
		registerCommands(p, pluginLog.WithName("Commands"))

		pluginLog.Info("Network plugin initialized successfully")
		return nil
	},
}

// publishPolicyUpdate function removed, use permissions.PublishPolicyUpdate directly

// registerEventHandlers - update PostLoginEvent example
func registerEventHandlers(p *proxy.Proxy, log logr.Logger) error {
	// ... other event subscriptions ...
	event.Subscribe(p.Event(), 0, onPing(log.WithName("OnPing")))

	event.Subscribe(p.Event(), 0, addServerToTabList(p, log.WithName("AddServerToTabList")))

	event.Subscribe(p.Event(), 0, func(e *proxy.PostLoginEvent) {
		playerIDStr := e.Player().ID().String()
		casbinEnforcer := permissions.GetEnforcer() // Use new getter
		if casbinEnforcer == nil {
			log.Error(nil, "Casbin enforcer not available in PostLoginEvent")
			return
		}
		// Example: Ensure every player is part of a "default" group
		// hasDefaultGroup, _ := casbinEnforcer.HasRoleForUser(playerIDStr, "group:default")
		// if !hasDefaultGroup {
		// 	added, errAdd := casbinEnforcer.AddGroupingPolicy(playerIDStr, "group:default")
		// 	if errAdd != nil {
		// 		log.Error(errAdd, "Failed to add default group to player", "player", playerIDStr)
		// 	} else if added {
		// 		log.Info("Added default group to player", "player", playerIDStr)
		//      if errSave := casbinEnforcer.SavePolicy(); errSave != nil {
		//          log.Error(errSave, "Failed to save policy after adding default group")
		//      } else {
		//          permissions.PublishPolicyUpdate(log, nc) // Use new function
		//      }
		// 	}
		// }
		log.V(1).Info("Player post-login event processed for permissions", "player", playerIDStr)
	})

	return nil
}

// HasPermission wrapper function removed.
// All calls should now be to permissions.HasPermission(player.ID().String(), object, action, logger)
