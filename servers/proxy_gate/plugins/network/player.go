package network

import (
	"encoding/json"

	"github.com/go-logr/logr"
	"github.com/nats-io/nats.go"
	"go.minekube.com/gate/pkg/edition/java/proxy"
)

func watchPlayers(p *proxy.Proxy, log logr.Logger) {
	watcher, err := playersKV.WatchAll()
	if err != nil {
		log.Error(err, "Unable to start player KV watch")
		return
	}
	defer watcher.Stop()

	for entry := range watcher.Updates() {
		if entry == nil {
			continue
		}
		if entry.Operation() == nats.KeyValueDelete {
			delete(playersMetadata, entry.Key())
		} else {
			var metadata Metadata
			err := json.Unmarshal(entry.Value(), &metadata)
			if err != nil {
				log.Error(err, "Failed to unmarshal player info", "key", entry.Key())
				continue
			}
			playersMetadata[entry.Key()] = metadata
		}
	}
}

func trackPlayerServer(log logr.Logger) func(e *proxy.ServerConnectedEvent) {
	return func(e *proxy.ServerConnectedEvent) {
		player := e.Player()
		server := e.Server()
		serverInfo := server.ServerInfo()

		err := updatePlayerMetadata(player, func(meta *Metadata) {
			SetAnnotation(meta, "network/location", serverInfo.Name())
		})
		if err != nil {
			log.Error(err, "Failed to update player metadata", "player", player.Username())
			return
		}
		log.Info("Tracking player location", "player", player.Username(), "server", serverInfo.Name())
	}
}

// Track player disconnects to remove from KV
func trackPlayerDisconnect(log logr.Logger) func(e *proxy.DisconnectEvent) {
	return func(e *proxy.DisconnectEvent) {
		player := e.Player()

		removePlayerMetadata(player.GameProfile().Name)

		log.Info("Removed player tracking", "player", player.Username())
	}
}

func setPlayerMetadata(name string, meta Metadata) error {
	data, err := json.Marshal(meta)
	if err != nil {
		return err
	}
	_, err = playersKV.Put(name, data)
	if err != nil {
		return err
	}
	return nil
}

func removePlayerMetadata(name string) error {
	err := playersKV.Delete(name)
	return err
}

func updatePlayerMetadata(player proxy.Player, fn func(meta *Metadata)) error {
	name := player.GameProfile().Name
	meta, exists := playersMetadata[name]
	if !exists {
		meta = Metadata{}
	}
	fn(&meta)
	if err := setPlayerMetadata(name, meta); err != nil {
		return err
	}
	return nil
}
