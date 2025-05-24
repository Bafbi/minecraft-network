package network

import (
	"context"
	"fmt"

	"github.com/bafbi/minecraft-network/servers/proxy_gate/plugins/network/permissions"
	"github.com/go-logr/logr"
	"github.com/nats-io/nats.go"
	"go.minekube.com/gate/pkg/edition/java/proxy"
)

// InitPermissionSystem initializes the permission system (Casbin, commands, etc).
func InitPermissionSystem(ctx context.Context, p *proxy.Proxy, nc *nats.Conn, log logr.Logger) error {

	if err := permissions.InitCasbin(
		ctx,
		log.WithName("Casbin"),
		nc,
	); err != nil {
		log.Error(err, "Failed to initialize Casbin permissions service")
		return fmt.Errorf("failed to initialize Casbin: %w", err)
	}

	// Register permission commands
	permissions.RegisterCommands(p, log.WithName("Commands"))

	log.Info("Permission system initialized")
	return nil
}
