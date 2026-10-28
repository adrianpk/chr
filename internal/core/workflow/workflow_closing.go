package workflow

import (
	"context"
	"errors"
	"fmt"
	"time"

	workflowcloseout "forge.adrianpk.com/chr/internal/core/workflow/closeout"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

// CloseSet validates every actual member on an isolated exact target.
// Failed gates retain their attempt and retry only closure. A confirmed receipt,
// set projection and owning invocation share one replayable transition.
func CloseSet(ctx context.Context, store *Store, request workflowcloseout.Request, ports FrontierPorts) (result workflowcloseout.Result, err error) {
	if request.ExpectedRevision == 0 || request.ExpectedSourceHash == "" || !workflowstate.ValidBranch(request.Branch) || ports.Provider == nil || ports.Observer == nil || ports.Target == nil {
		return result, fmt.Errorf("closing requires an observed set and canonical observation ports")
	}

	var claim *workflowClosingClaim

	admission, cancelAdmission := context.WithTimeout(ctx, 3*time.Minute)
	err = store.withLock(admission, func() error {
		begun, beginErr := store.beginWorkflowClosing(admission, request, ports)

		claim, result = begun.Claim, begun.Result

		return beginErr
	})

	cancelAdmission()

	if err != nil || claim == nil {
		return result, err
	}
	defer func() { err = errors.Join(err, claim.worker.releaseReservation()) }()

	receipt, gateErr := (workflowClosingHost{store: store}).execute(ctx, claim)

	cleanup, cancel := context.WithTimeout(context.WithoutCancel(ctx), 3*time.Minute)
	defer cancel()

	err = store.withLock(cleanup, func() error {
		var finishErr error

		result, finishErr = store.finishWorkflowClosing(cleanup, claim, receipt, gateErr, ports)

		return finishErr
	})

	return result, err
}

type workflowClosingClaim struct {
	worker *Claim
	set    workflowstate.Record
	state  workflowcloseout.Attempt
	input  workflowstate.ClosingInput
	data   []byte
}
