package cache

import (
	"context"
	"encoding/json"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/nats-io/nats.go"
	"google.golang.org/protobuf/types/known/structpb"
)

// MetadataCache holds cached player and server metadata.
type MetadataCache struct {
	playerCache  map[string]*structpb.Struct // Key: Player UUID
	serverCache  map[string]*structpb.Struct // Key: Server Name
	playerMu     sync.RWMutex
	serverMu     sync.RWMutex
	kv           nats.KeyValue
	playerPrefix string
	serverPrefix string
}

// NewMetadataCache creates and initializes a new MetadataCache.
func NewMetadataCache(kv nats.KeyValue, playerPrefix, serverPrefix string) *MetadataCache {
	return &MetadataCache{
		playerCache:  make(map[string]*structpb.Struct),
		serverCache:  make(map[string]*structpb.Struct),
		kv:           kv,
		playerPrefix: playerPrefix,
		serverPrefix: serverPrefix,
	}
}

// StartWatching initializes the NATS KV watchers for player and server metadata.
func (mc *MetadataCache) StartWatching(ctx context.Context) {
	go mc.watchForUpdates(ctx, mc.playerPrefix+">", mc.playerCache, &mc.playerMu, mc.extractPlayerKey)
	go mc.watchForUpdates(ctx, mc.serverPrefix+">", mc.serverCache, &mc.serverMu, mc.extractServerKey)
	log.Println("Started NATS KV watchers for metadata.")
}

// GetPlayerMetadata retrieves player metadata from the cache.
func (mc *MetadataCache) GetPlayerMetadata(uuid string) *structpb.Struct {
	mc.playerMu.RLock()
	defer mc.playerMu.RUnlock()
	return mc.playerCache[uuid]
}

// GetServerMetadata retrieves server metadata from the cache.
func (mc *MetadataCache) GetServerMetadata(name string) *structpb.Struct {
	mc.serverMu.RLock()
	defer mc.serverMu.RUnlock()
	return mc.serverCache[name]
}

// watchForUpdates is a generic function to watch for changes in NATS KV.
func (mc *MetadataCache) watchForUpdates(ctx context.Context, keyPrefix string, cache map[string]*structpb.Struct, mu *sync.RWMutex, extractKeyFunc func(string, string) string) {
	watcher, err := mc.kv.Watch(keyPrefix)
	if err != nil {
		log.Fatalf("Failed to create NATS KV watcher for %s: %v", keyPrefix, err)
	}
	defer watcher.Stop()

	log.Printf("Watching NATS KV for updates on prefix: %s", keyPrefix)

	for {
		select {
		case <-ctx.Done():
			log.Printf("Stopping NATS KV watcher for %s due to context cancellation.", keyPrefix)
			return
		case entry := <-watcher.Updates():
			if entry == nil {
				log.Printf("NATS KV watcher for %s received nil entry, likely closing.", keyPrefix)
				return
			}

			key := entry.Key()
			extractedKey := extractKeyFunc(key, keyPrefix)
			mu.Lock()

			switch entry.Operation() {
			case nats.KeyValuePut:
				var data map[string]any
				if err := json.Unmarshal(entry.Value(), &data); err != nil {
					log.Printf("Error unmarshaling JSON for key %s: %v", key, err)
					mu.Unlock()
					continue
				}
				pbStruct, err := structpb.NewStruct(data)
				if err != nil {
					log.Printf("Error converting map to protobuf Struct for key %s: %v", key, err)
					mu.Unlock()
					continue
				}
				cache[extractedKey] = pbStruct
				log.Printf("Cached PUT update for %s: %s (rev %d)", strings.TrimSuffix(keyPrefix, ">"), extractedKey, entry.Revision())
			case nats.KeyValueDelete:
				delete(cache, extractedKey)
				log.Printf("Cached DELETE update for %s: %s (rev %d)", strings.TrimSuffix(keyPrefix, ">"), extractedKey, entry.Revision())
			default:
				log.Printf("Unknown NATS KV operation for %s: %v", key, entry.Operation())
			}
			mu.Unlock()
		case <-time.After(10 * time.Second): // Periodically log heartbeats or check for no updates
			// log.Printf("NATS KV watcher for %s is active, no updates in 10s.", keyPrefix)
		}
	}
}

// Helper to extract the actual UUID from a player metadata key (e.g., "player.metadata.UUID")
func (mc *MetadataCache) extractPlayerKey(fullKey, prefix string) string {
	return strings.TrimPrefix(fullKey, strings.TrimSuffix(prefix, ">"))
}

// Helper to extract the actual server name from a server metadata key (e.g., "server.metadata.survival")
func (mc *MetadataCache) extractServerKey(fullKey, prefix string) string {
	return strings.TrimPrefix(fullKey, strings.TrimSuffix(prefix, ">"))
}
