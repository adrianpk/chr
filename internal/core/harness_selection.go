package core

import (
	"forge.adrianpk.com/chr/internal/core/contenthash"
)

// Revision identifies the complete persisted configuration image.
func (c WorkspaceConfig) Revision() string {
	if c.workflowRevision == "" {
		return contenthash.Sum(nil)
	}

	return c.workflowRevision
}
