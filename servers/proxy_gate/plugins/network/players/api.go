package players

import (
	"fmt"

	"github.com/bafbi/minecraft-network/servers/proxy_gate/util/metadata"
	"github.com/nats-io/nats.go" // For nats.ErrKeyNotFound
	"go.minekube.com/gate/pkg/util/uuid"
)

// GetMetadataByUUID retrieves player metadata by UUID, checking cache first, then KV.
func GetMetadataByUUID(playerUUID uuid.UUID) (metadata.Metadata, bool) {
	return getMetadataByUUIDInternal(playerUUID)
}

// GetMetadataByName retrieves player metadata by name from cache.
func GetMetadataByName(name string) (metadata.Metadata, bool) {
	return getMetadataByNameInternal(name)
}

// UpdateMetadataByUUID updates a player's metadata in NATS KV only.
// Cache updates are handled by the KV watcher.
func UpdateMetadataByUUID(playerUUID uuid.UUID, modFunc func(meta *metadata.Metadata)) error {
	// Read from cache with lock
	playerMu.RLock()
	currentMeta, exists := playersMetadata[playerUUID]
	playerMu.RUnlock()

	// Initialize new metadata if it doesn't exist
	if !exists {
		currentMeta = metadata.Metadata{}
		if currentMeta.Annotations == nil {
			currentMeta.Annotations = make(map[string]string)
		}
		if currentMeta.Labels == nil {
			currentMeta.Labels = make(map[string]string)
		}
	}

	// Make a copy to pass to modFunc
	metaToModify := currentMeta // If metadata.Metadata is a struct, this is a copy.

	modFunc(&metaToModify) // Apply modifications

	// Persist to KV only - let the watcher update the cache
	if err := putMetadataToKV(playerUUID, metaToModify); err != nil {
		return fmt.Errorf("failed to persist player metadata to KV for %s: %w", playerUUID, err)
	}

	cacheLog.V(1).Info("API updated player metadata in KV", "uuid", playerUUID)
	return nil
}

// UpdateMetadataByName updates a player's metadata in NATS KV only.
// Cache updates are handled by the KV watcher.
func UpdateMetadataByName(name string, modFunc func(meta *metadata.Metadata)) error {
	playerUUID, exist := nameToUUID[name]
	if !exist {
		return fmt.Errorf("player name %s not found", name)
	}
	return UpdateMetadataByUUID(playerUUID, modFunc)
}

// RemovePlayer removes a player's metadata from KV only.
// Cache updates are handled by the KV watcher.
func RemovePlayer(playerUUID uuid.UUID) error {
	// Remove from KV only
	err := deleteMetadataFromKV(playerUUID)
	if err != nil && err != nats.ErrKeyNotFound { // Don't error if already not in KV
		kvLog.Error(err, "Failed to delete player metadata from KV", "uuid", playerUUID)
		// Decide if this is a fatal error for the operation
		// return fmt.Errorf("failed to delete player metadata from KV for %s: %w", playerUUID, err)
	}
	// Let the watcher handle cache updates
	return nil // Or return KV error if critical
}

func GetPlayersSlice() []string {
	// Read from cache with lock
	playerMu.RLock()
	defer playerMu.RUnlock()

	// Create a slice to hold the names
	names := make([]string, 0, len(playersMetadata))

	// Use the map to get the names name>uuid
	for name := range nameToUUID {
		names = append(names, name)
	}

	return names
}

func GetPlayerCount() int {
	// Read from cache with lock
	playerMu.RLock()
	defer playerMu.RUnlock()

	count := 0
	for _, meta := range playersMetadata {
		if online, exists, err := meta.GetAnnotationBoolValue("player/online"); err == nil && exists && online {
			count++
		}
	}

	return count
}
