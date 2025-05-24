package players

import (
	"encoding/json"
	"fmt"

	"github.com/bafbi/minecraft-network/servers/proxy_gate/util/metadata"
	"github.com/go-logr/logr"
	"github.com/nats-io/nats.go"
	"go.minekube.com/gate/pkg/util/uuid"
)

var (
	playersKV nats.KeyValue
	kvLog     logr.Logger
)

// InitializeKVStore initializes the NATS KV store for players.
func InitializeKVStore(js nats.JetStreamContext, log logr.Logger) error {
	kvLog = log.WithName("PlayerKV")
	var err error
	playersKV, err = js.KeyValue("players")
	if err != nil {
		kvLog.Info("Players KV store not found, attempting to create.")
		playersKV, err = js.CreateKeyValue(&nats.KeyValueConfig{Bucket: "players"})
		if err != nil {
			kvLog.Error(err, "Failed to create players KV store")
			return fmt.Errorf("failed to create players KV store: %w", err)
		}
	}
	kvLog.Info("Player KV store initialized")
	return nil
}

// WatchKVStore watches for changes in the player KV store and updates local cache.
func WatchKVStore() {
	if playersKV == nil {
		kvLog.Error(nil, "Players KV store not initialized. Cannot start watcher.")
		return
	}
	watcher, err := playersKV.WatchAll(nats.IgnoreDeletes()) // Optionally ignore deletes if local cache handles removal via events
	if err != nil {
		kvLog.Error(err, "Unable to start player KV watch")
		return
	}
	kvLog.Info("Starting player KV watcher")
	// This loop should be in a goroutine managed by the caller (e.g., in plugin.Init)
	// defer watcher.Stop() // Belongs in the goroutine

	for entry := range watcher.Updates() {
		if entry == nil { // Should not happen with WatchAll but good practice
			continue
		}
		playerUUID, err := uuid.Parse(entry.Key())
		if err != nil {
			kvLog.Error(err, "Failed to parse player UUID from KV key", "key", entry.Key())
			continue
		}

		// If operation is delete, NATS KV WatchAll with IgnoreDeletes() won't send it.
		// Deletes from KV would be handled by TrackPlayerDisconnect calling a KV delete function.
		// If not ignoring deletes:
		// if entry.Operation() == nats.KeyValueDelete {
		//    deleteFromLocalCache(playerUUID) // Update cache
		//    continue
		// }

		var meta metadata.Metadata
		if err := json.Unmarshal(entry.Value(), &meta); err != nil {
			kvLog.Error(err, "Failed to unmarshal player metadata from KV", "key", entry.Key(), "value", string(entry.Value()))
			continue
		}
		updateLocalCache(playerUUID, meta) // Update cache
	}
	kvLog.Info("Player KV watcher stopped.")
}

// --- Internal KV interaction functions (used by api.go) ---

func getMetadataFromKV(playerUUID uuid.UUID) (metadata.Metadata, uint64, error) {
	if playersKV == nil {
		return metadata.Metadata{}, 0, fmt.Errorf("players KV not initialized")
	}
	entry, err := playersKV.Get(playerUUID.String())
	if err != nil {
		return metadata.Metadata{}, 0, err // Handles nats.ErrKeyNotFound
	}
	var meta metadata.Metadata
	if err := json.Unmarshal(entry.Value(), &meta); err != nil {
		return metadata.Metadata{}, 0, fmt.Errorf("failed to unmarshal KV data for %s: %w", playerUUID, err)
	}
	return meta, entry.Revision(), nil
}

func putMetadataToKV(playerUUID uuid.UUID, meta metadata.Metadata) error {
	if playersKV == nil {
		return fmt.Errorf("players KV not initialized")
	}
	data, err := json.Marshal(meta)
	if err != nil {
		return fmt.Errorf("failed to marshal player metadata for KV: %w", err)
	}
	_, err = playersKV.Put(playerUUID.String(), data)
	return err
}

func deleteMetadataFromKV(playerUUID uuid.UUID) error {
	if playersKV == nil {
		return fmt.Errorf("players KV not initialized")
	}
	return playersKV.Delete(playerUUID.String())
}
