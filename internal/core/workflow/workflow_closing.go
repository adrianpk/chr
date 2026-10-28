package workflow

import (
	"context"
	"errors"
	"fmt"
	"time"

	workflowcloseout "forge.adrianpk.com/chr/internal/core/workflow/closeout"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

// SetClosureRequest requires the observed set and canonical target branch.
// InvocationID is optional for an explicit standalone decision, including an
// empty set. It cannot take ownership from another unfinished invocation.
type SetClosureRequest struct {
	SetPath            string
	ExpectedRevision   uint64
	ExpectedSourceHash string
	Repository         string
	Branch             string
	InvocationID       string
}

// SetClosureResult distinguishes retained failures from confirmed closure.
type SetClosureResult struct {
	Set        Record
	Receipt    workflowstate.ClosingReceipt
	ReceiptRef string
	Outcome    string
}

// CloseWorkflowSet validates every actual member on an isolated exact target.
// Failed gates retain their attempt and retry only closure. A confirmed receipt,
// set projection and owning invocation share one replayable transition.
func (s *Store) CloseWorkflowSet(ctx context.Context, request SetClosureRequest, ports FrontierPorts) (result SetClosureResult, err error) {
	if request.ExpectedRevision == 0 || request.ExpectedSourceHash == "" || !workflowstate.ValidBranch(request.Branch) || ports.Provider == nil || ports.Observer == nil || ports.Target == nil {
		return result, fmt.Errorf("closing requires an observed set and canonical observation ports")
	}

	var claim *workflowClosingClaim

	admission, cancelAdmission := context.WithTimeout(ctx, 3*time.Minute)
	err = s.withLock(admission, func() error {
		var beginErr error

		claim, result, beginErr = s.beginWorkflowClosing(admission, request, ports)

		return beginErr
	})

	cancelAdmission()

	if err != nil || claim == nil {
		return result, err
	}
	defer func() { err = errors.Join(err, claim.worker.releaseReservation()) }()

	var gate workflowcloseout.Gate = workflowClosingGate{store: s, claim: claim}

	receipt, gateErr := gate.Execute(ctx)

	cleanup, cancel := context.WithTimeout(context.WithoutCancel(ctx), 3*time.Minute)
	defer cancel()

	err = s.withLock(cleanup, func() error {
		var finishErr error

		result, finishErr = s.finishWorkflowClosing(cleanup, claim, receipt, gateErr, ports)

		return finishErr
	})

	return result, err
}

type workflowClosingClaim struct {
	worker *Claim
	set    Record
	state  workflowClosingAttempt
	input  workflowstate.ClosingInput
	data   []byte
}

type workflowClosingAttempt struct {
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
	Processes    []Process                     `json:"processes"`
	Outcome      string                        `json:"outcome"`
	FinishedAt   time.Time                     `json:"finished_at,omitempty"`
}
