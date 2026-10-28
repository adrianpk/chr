package closeout

import (
	"time"

	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

// Attempt retains one bounded closing owner, evidence binding and terminal outcome.
type Attempt struct {
	Schema       string                        `json:"schema"`
	SetID        string                        `json:"set_id"`
	Generation   int                           `json:"generation"`
	CapacitySlot int                           `json:"capacity_slot,omitempty"`
	BudgetID     string                        `json:"budget_id"`
	OwnerRoot    string                        `json:"owner_root"`
	InvocationID string                        `json:"invocation_id,omitempty"`
	Owner        workflowstate.ProcessIdentity `json:"owner"`
	TokenHash    string                        `json:"token_hash"`
	LeaseUntil   time.Time                     `json:"lease_until"`
	Input        workflowstate.ClosingEvidence `json:"input"`
	Result       workflowstate.ClosingEvidence `json:"result"`
	Processes    []workflowstate.Process       `json:"processes"`
	Outcome      string                        `json:"outcome"`
	FinishedAt   time.Time                     `json:"finished_at,omitempty"`
}
