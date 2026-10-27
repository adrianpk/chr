package state

import "strings"

// ValidOperationIdentity retains the generation identity bound and forbidden bytes.
func ValidOperationIdentity(value string) bool {
	return strings.TrimSpace(value) != "" && len(value) <= 256 && !strings.ContainsAny(value, "\x00\r\n")
}
