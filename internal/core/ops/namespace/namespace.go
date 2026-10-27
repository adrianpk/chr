// Package namespace owns operational namespace spelling.
package namespace

import (
	"regexp"
	"strings"
)

var opsNamespacePattern = regexp.MustCompile(`^[a-z0-9]+(?:-[a-z0-9]+)*$`)

// Valid reports whether value is a shared namespace name.
func Valid(value string) bool {
	return opsNamespacePattern.MatchString(strings.TrimSpace(value))
}
