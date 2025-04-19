package network

// SetLabel sets a label on the given Metadata.
func SetLabel(meta *Metadata, key, value string) {
	if meta.Labels == nil {
		meta.Labels = make(map[string]string)
	}
	meta.Labels[key] = value
}

// SetAnnotation sets an annotation on the given Metadata.
func SetAnnotation(meta *Metadata, key, value string) {
	if meta.Annotations == nil {
		meta.Annotations = make(map[string]string)
	}
	meta.Annotations[key] = value
}

// GetLabel gets a label from the given Metadata.
func GetLabel(meta *Metadata, key string) (string, bool) {
	if meta.Labels == nil {
		return "", false
	}
	v, ok := meta.Labels[key]
	return v, ok
}

// GetAnnotation gets an annotation from the given Metadata.
func GetAnnotation(meta *Metadata, key string) (string, bool) {
	if meta.Annotations == nil {
		return "", false
	}
	v, ok := meta.Annotations[key]
	return v, ok
}

// Test if a selector matches a server's labels
func MatchesLabels(meta *Metadata, selectors map[string]string) bool {
	for k, v := range selectors {
		if val, ok := meta.Labels[k]; !ok || val != v {
			return false
		}
	}
	return true
}
