package network

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"
	"github.com/nats-io/nats.go"
	"github.com/robinbraemer/event"
	"go.minekube.com/gate/pkg/edition/java/proxy"
)

var Plugin = proxy.Plugin{
	Name: "Network",
	Init: func(ctx context.Context, p *proxy.Proxy) error {
		log := logr.FromContextOrDiscard(ctx)
		log.Info("Network plugin starting")

		// Connect to NATS
		natsURL := "nats://network-nats:4222"
		nc, err := nats.Connect(natsURL)
		if err != nil {
			return fmt.Errorf("failed to connect to NATS: %w", err)
		}

		js, err := nc.JetStream()
		if err != nil {
			return fmt.Errorf("failed to get JetStream context: %w", err)
		}

		// Initialize KV stores
		err = initializeKVStores(js, log)
		if err != nil {
			return err
		}

		go watchServers(p, log)
		go watchPlayers(p, log)

		// Register event handlers
		registerEventHandlers(p, log)

		// Register commands
		registerCommands(p, log)

		log.Info("Network plugin initialized")
		return nil
	},
}

func initializeKVStores(js nats.JetStreamContext, log logr.Logger) error {
	// Initialize servers KV
	var err error
	serversKV, err = js.KeyValue("servers")
	if err != nil {
		serversKV, err = js.CreateKeyValue(&nats.KeyValueConfig{
			Bucket: "servers",
		})
		if err != nil {
			return fmt.Errorf("failed to create servers KV store: %w", err)
		}
	}

	// Initialize players KV
	playersKV, err = js.KeyValue("players")
	if err != nil {
		playersKV, err = js.CreateKeyValue(&nats.KeyValueConfig{
			Bucket: "players",
		})
		if err != nil {
			return fmt.Errorf("failed to create players KV store: %w", err)
		}
	}
	return nil
}

func registerEventHandlers(p *proxy.Proxy, log logr.Logger) {
	// Player routing events
	event.Subscribe(p.Event(), 0, sendPlayerToServer(p, log))
	event.Subscribe(p.Event(), 0, addServerToTabList(p, log))
	event.Subscribe(p.Event(), 0, handleKickedFromServerEvent(p, log))

	// Player tracking events
	event.Subscribe(p.Event(), 0, trackPlayerServer(log))
	event.Subscribe(p.Event(), 0, trackPlayerDisconnect(log))
}
