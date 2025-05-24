package players

import (
	"sync"

	"github.com/bafbi/minecraft-network/servers/proxy_gate/util/metadata"
	"github.com/go-logr/logr" // For logging within cache updates
	"go.minekube.com/gate/pkg/util/uuid"
)

var (
	playerMu        sync.RWMutex
	playersMetadata = make(map[uuid.UUID]metadata.Metadata) // uuid → player metadata
	nameToUUID      = make(map[string]uuid.UUID)            // name → uuid
	cacheLog        logr.Logger                             // Logger for cache operations
)

// InitCache initializes the cache module, primarily for setting up logging.
func InitCache(log logr.Logger) {
	cacheLog = log.WithName("PlayerCache")
	cacheLog.Info("Player cache initialized")
}

// --- Internal cache update functions ---

func updateLocalCache(playerUUID uuid.UUID, meta metadata.Metadata) {
	playerMu.Lock()
	defer playerMu.Unlock()

	name, _ := meta.GetAnnotation("player/name") // Assuming GetAnnotation is a method on metadata.Metadata

	// If name changed or player is new, handle nameToUUID map update
	if oldMeta, ok := playersMetadata[playerUUID]; ok {
		if oldName, _ := oldMeta.GetAnnotation("player/name"); oldName != "" && oldName != name {
			delete(nameToUUID, oldName)
		}
	}
	playersMetadata[playerUUID] = meta
	if name != "" {
		nameToUUID[name] = playerUUID
	}
	cacheLog.V(1).Info("Updated local cache", "uuid", playerUUID, "name", name)
}

func deleteFromLocalCache(playerUUID uuid.UUID) {
	playerMu.Lock()
	defer playerMu.Unlock()

	var name string
	if oldMeta, ok := playersMetadata[playerUUID]; ok {
		name, _ = oldMeta.GetAnnotation("player/name")
		delete(playersMetadata, playerUUID)
	}
	if name != "" {
		delete(nameToUUID, name)
	}
	cacheLog.V(1).Info("Deleted from local cache", "uuid", playerUUID)
}

// --- Internal cache getter functions (used by api.go) ---

func getMetadataByUUIDInternal(playerUUID uuid.UUID) (metadata.Metadata, bool) {
	playerMu.RLock()
	defer playerMu.RUnlock()
	meta, exists := playersMetadata[playerUUID]
	return meta, exists
}

func getMetadataByNameInternal(name string) (metadata.Metadata, bool) {
	playerMu.RLock()
	defer playerMu.RUnlock()
	playerUUID, exists := nameToUUID[name]
	if !exists {
		return metadata.Metadata{}, false
	}
	meta, exists := playersMetadata[playerUUID]
	return meta, exists
}
