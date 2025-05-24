package network

import (
	"github.com/bafbi/minecraft-network/servers/proxy_gate/plugins/network/chat"
	"github.com/go-logr/logr"
	"github.com/nats-io/nats.go"
	"github.com/robinbraemer/event"
	"go.minekube.com/gate/pkg/edition/java/proxy"
)

// InitChatSystem initializes the chat system: registers the chat event handler and starts the NATS chat subscriber.
func InitChatSystem(p *proxy.Proxy, nc *nats.Conn, log logr.Logger) error {
	// Register chat event handler
	event.Subscribe(p.Event(), 0, chat.CreatePlayerChatEventHandler(nc, log.WithName("Handler")))

	// Start NATS chat subscriber
	_, err := chat.StartNatsSubscriber(p, nc, log.WithName("Subscriber"))
	if err != nil {
		log.Error(err, "Failed to start NATS chat subscriber")
		return err
	}
	log.Info("Chat system initialized")
	return nil
}
