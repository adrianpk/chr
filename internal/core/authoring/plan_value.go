package authoring

import workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"

// PlanPublicationResult identifies the retained revision from one committed
// operation. Reference is project-relative; binding paths are namespace-relative.
type PlanPublicationResult struct {
	OperationID string
	Reference   string
	Binding     workflowstate.PlanRevisionBinding
}
