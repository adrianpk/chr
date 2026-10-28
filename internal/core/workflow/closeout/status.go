package closeout

import (
	"context"

	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

// Status reports retained receipt/fanout and gate proof.
type Status struct {
	CanClose         bool
	ReceiptsVerified bool
	GateVerified     bool
	Blockers         []string
}

// StatusSource supplies ordered publication heads and live host observations.
// Capacity is observed only after receipt checks and publication blockers.
type StatusSource interface {
	Evidence
	Publications() []workflowstate.SubmissionState
	Capacity(context.Context) (workflowstate.Capacity, error)
}

// Inspect reports retained receipt, publication, capacity and final-gate proof.
func Inspect(ctx context.Context, set workflowstate.Record, proofs Proofs, source StatusSource) (Status, error) {
	status := Status{}
	if err := proofs.CheckReceipts(ctx, set); err != nil {
		status.Blockers = append(status.Blockers, "unit_receipts_or_fanout_required")

		return status, err
	}

	status.ReceiptsVerified = true

	for _, publication := range source.Publications() {
		if publication == workflowstate.SubmissionAwaitingObservation {
			status.Blockers = append(status.Blockers, "publication_requires_observation")
		}
	}

	capacity, err := source.Capacity(ctx)
	if err != nil {
		return status, err
	}

	if len(capacity.LiveProcesses) > 0 || len(capacity.UnexpiredOwners) > 0 {
		status.Blockers = append(status.Blockers, "live_process_requires_reconciliation")
	}

	if set.Set.Status == "completed" {
		if _, err := ReadClosedSet(set, source); err != nil {
			return status, err
		}

		status.GateVerified = true
	} else {
		status.Blockers = append(status.Blockers, "final_gate_required")
	}

	status.CanClose = len(status.Blockers) == 0 && status.GateVerified

	return status, nil
}
