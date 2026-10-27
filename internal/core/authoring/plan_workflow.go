package authoring

import (
	"forge.adrianpk.com/chr/internal/core/contenthash"
)

// WorkflowContentHash binds an approved source image without deriving identity.
func WorkflowContentHash(data []byte) string { return contenthash.Sum(data) }
