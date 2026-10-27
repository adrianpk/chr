package authoring

import (
	generationcore "forge.adrianpk.com/chr/internal/core/authoring/generation"
)

// PlanPublicationResult identifies the retained revision from one committed
// operation. Reference is project-relative; binding paths are namespace-relative.
type PlanPublicationResult struct {
	OperationID string
	Reference   string
	Binding     generationcore.PlanRevisionBinding
}
