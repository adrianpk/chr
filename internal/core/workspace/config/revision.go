package config

import (
	"forge.adrianpk.com/chr/internal/core/contenthash"
)

// Revision identifies the complete persisted configuration image.
func (c Config) Revision() string {
	if c.workflowRevision == "" {
		return contenthash.Sum(nil)
	}

	return c.workflowRevision
}

// Explicit reports whether the parsed image supplied a key, without exposing mutable parser metadata.
func (c Config) Explicit(key string) bool { return c.workflowExplicit[key] }
