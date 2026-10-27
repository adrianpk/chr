package state

import "strings"

// ValidRequestKey checks the existing opaque workflow request-key contract.
func ValidRequestKey(key string) bool {
	return strings.TrimSpace(key) == key && key != "" && len(key) <= 128 && !strings.ContainsAny(key, "\x00\n\r")
}
