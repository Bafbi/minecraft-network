package network

import (
	"github.com/bafbi/minecraft-network/servers/proxy_gate/plugins/network/players"
	"github.com/go-logr/logr"
	"github.com/nats-io/nats.go"
)

// InitPlayerSystem initializes the player system: cache, KV store, events, and watcher.
func InitPlayerSystem(js nats.JetStreamContext, log logr.Logger) error {
	players.InitCache(log.WithName("Cache"))
	if err := players.InitializeKVStore(js, log.WithName("KV")); err != nil {
		return err
	}
	players.InitEvents(log.WithName("Events"))
	go players.WatchKVStore()
	log.Info("Player system initialized")
	return nil
}
