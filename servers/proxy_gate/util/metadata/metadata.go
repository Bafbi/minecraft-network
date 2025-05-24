package metadata

import (
	"encoding/json"
	"fmt"
	"slices"
)

type Metadata struct {
	Labels      map[string]string
	Annotations map[string]string
}

// SetLabel sets a label on the given Metadata.
func (meta *Metadata) SetLabel(key, value string) string {
	if meta.Labels == nil {
		meta.Labels = make(map[string]string)
	}
	prevValue, _ := meta.Labels[key]
	meta.Labels[key] = value
	return prevValue
}

// SetAnnotation sets an annotation on the given Metadata.
// It initializes the Annotations map if it's nil.
// It returns the previous value associated with the key, or an empty string if the key was not present in the map.
func (meta *Metadata) SetAnnotation(key, value string) string {
	if meta.Annotations == nil {
		meta.Annotations = make(map[string]string)
	}
	prevValue, _ := meta.Annotations[key]
	meta.Annotations[key] = value
	return prevValue
}

// GetLabel gets a label from the given Metadata.
// It returns the value associated with the key and a boolean indicating if the key exists.
// If the Labels map is nil, it returns an empty string and false.
// If the key doesn't exist, it returns an empty string and false.
// If the key exists, it returns the value and true.
func (meta *Metadata) GetLabel(key string) (string, bool) {
	if meta.Labels == nil {
		return "", false
	}
	v, ok := meta.Labels[key]
	return v, ok
}

// GetAnnotation gets an annotation from the given Metadata.
func (meta *Metadata) GetAnnotation(key string) (string, bool) {
	if meta.Annotations == nil {
		return "", false
	}
	v, ok := meta.Annotations[key]
	return v, ok
}

// Test if a selector matches a server's labels
func (meta *Metadata) MatchesLabels(selectors map[string]string) bool {
	for k, v := range selectors {
		if val, ok := meta.Labels[k]; !ok || val != v {
			return false
		}
	}
	return true
}

// SetAnnotationStringSlice sets an annotation key to a JSON-encoded string slice.
func (meta *Metadata) SetAnnotationStringSlice(key string, values []string) error {
	// Ensure Annotations map exists
	if meta.Annotations == nil {
		meta.Annotations = make(map[string]string)
	}

	jsonBytes, err := json.Marshal(values)
	if err != nil {
		return fmt.Errorf("failed to marshal string slice for annotation %q: %w", key, err)
	}
	meta.Annotations[key] = string(jsonBytes)
	return nil
}

// GetAnnotationStringSlice gets a JSON-encoded string slice from an annotation key.
// It returns the slice, a boolean indicating if the key exists and contains a valid slice,
// and an error if the key exists but the value is malformed.
func (meta *Metadata) GetAnnotationStringSlice(key string) ([]string, bool, error) {
	rawValue, ok := meta.GetAnnotation(key)
	if !ok {
		return nil, false, nil // Key doesn't exist
	}

	var values []string
	// Handle empty string case explicitly, treat as empty slice
	if rawValue == "" {
		return []string{}, true, nil
	}

	err := json.Unmarshal([]byte(rawValue), &values)
	if err != nil {
		// Key exists, but value is not a valid JSON string slice
		return nil, true, fmt.Errorf("annotation %q value %q is not a valid JSON string slice: %w", key, rawValue, err)
	}

	// Key exists and value is a valid JSON string slice
	return values, true, nil
}

// AddAnnotationStringValue adds a value to a JSON-encoded string slice annotation.
// If the key doesn't exist, it creates a new slice with the value.
// If the key exists but is not a valid JSON string slice, it returns an error.
// It avoids adding duplicate values.
func (meta *Metadata) AddAnnotationStringValue(key, valueToAdd string) error {
	values, ok, err := meta.GetAnnotationStringSlice(key)
	if err != nil {
		// Existing value is malformed
		return err
	}

	if !ok {
		// Key doesn't exist, create a new slice
		return meta.SetAnnotationStringSlice(key, []string{valueToAdd})
	}

	// Key exists, check if value is already present
	found := slices.Contains(values, valueToAdd)

	if found {
		// Value already exists, nothing to do
		return nil
	}

	// Value not found, append it and update the annotation
	updatedValues := append(values, valueToAdd)
	return meta.SetAnnotationStringSlice(key, updatedValues)
}

// RemoveAnnotationStringValue removes a value from a JSON-encoded string slice annotation.
// If the key doesn't exist or the value isn't found in the slice, it does nothing and returns nil.
// If the key exists but is not a valid JSON string slice, it returns an error.
func (meta *Metadata) RemoveAnnotationStringValue(key, valueToRemove string) error {
	values, ok, err := meta.GetAnnotationStringSlice(key)
	if err != nil {
		// Existing value is malformed
		return err
	}

	if !ok {
		// Key doesn't exist, nothing to remove
		return nil
	}

	// Key exists, filter out the value to remove
	found := false
	// Allocate with estimate, might be slightly too large if value found, but avoids reallocs
	newValues := make([]string, 0, len(values))
	for _, v := range values {
		if v == valueToRemove {
			found = true // Mark that we found (and are skipping) the value
		} else {
			newValues = append(newValues, v)
		}
	}

	if !found {
		// Value wasn't in the slice, nothing changed
		return nil
	}

	// Value was removed, update the annotation with the new slice
	// This handles the case where newValues might be empty correctly (sets annotation to "[]")
	return meta.SetAnnotationStringSlice(key, newValues)
}

// HasAnnotationStringValue checks if a value exists within a JSON-encoded string slice annotation.
// Returns true if the value is present, false otherwise.
// Returns an error if the key exists but is not a valid JSON string slice.
func (meta *Metadata) HasAnnotationStringValue(key, valueToCheck string) (bool, error) {
	values, ok, err := meta.GetAnnotationStringSlice(key)
	if err != nil {
		// Existing value is malformed
		return false, err
	}

	if !ok {
		// Key doesn't exist
		return false, nil
	}

	// Key exists, check for the value
	if slices.Contains(values, valueToCheck) {
		return true, nil
	}

	// Value not found in the slice
	return false, nil
}

func (meta *Metadata) GetAnnotationBoolValue(key string) (value bool, exists bool, err error) {
	rawValue, ok := meta.GetAnnotation(key)
	if !ok {
		return false, false, nil // Key doesn't exist
	}
	if rawValue == "" {
		return false, true, nil // Key exists but value is empty
	}
	if rawValue == "true" {
		return true, true, nil // Key exists and value is "true"
	}
	if rawValue == "false" {
		return false, true, nil // Key exists and value is "false"
	}
	return false, true, fmt.Errorf("annotation %q value %q is not a valid boolean", key, rawValue)
}
