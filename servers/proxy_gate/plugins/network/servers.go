package network

import (
	"github.com/bafbi/minecraft-network/servers/proxy_gate/plugins/network/servers"
	"github.com/go-logr/logr"
	"github.com/nats-io/nats.go"
	"go.minekube.com/gate/pkg/edition/java/proxy"
)

// InitServerSystem initializes the server system: cache, KV store, events, and watcher.
func InitServerSystem(p *proxy.Proxy, js nats.JetStreamContext, log logr.Logger) error {
	servers.InitCache(log.WithName("Cache"))
	if err := servers.InitializeKVStore(js, log.WithName("KV")); err != nil {
		return err
	}
	servers.InitEvents(p, log.WithName("Events"))
	go servers.WatchKVStore(p)
	log.Info("Server system initialized")
	return nil
}
