package workflow

import (
	"context"
	"fmt"
	"path/filepath"
	"slices"

	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

// InvocationClosure retains a confirmed set and its refreshed execution
// base without changing earlier unit delivery images or admission identity.
type InvocationClosure struct {
	SetID   string                        `json:"set_id"`
	Cursor  int                           `json:"cursor"`
	Receipt workflowstate.ClosingEvidence `json:"receipt"`
	Target  workflowstate.PreviewTarget   `json:"target"`
}

func workflowInvocationSetID(image InputImage) string {
	return filepath.Base(filepath.Dir(image.Path))
}

func workflowNextClosingSet(run Invocation) string {
	if len(run.Closures) >= len(run.SetImages) {
		return ""
	}

	return workflowInvocationSetID(run.SetImages[len(run.Closures)])
}

func (s *Store) workflowClosingRunReady(ctx context.Context, run workflowInvocationFile, setID string) error {
	if workflowClosingRunStopped(run, s.layout.Root()) {
		return ErrWorkflowStop
	}

	if !workflowClosingRunFrontier(run, setID) {
		return fmt.Errorf("%w: invocation must reach this set's closing frontier", ErrWorkflowOwned)
	}

	if run.Cursor < len(run.Members) && run.Members[run.Cursor].Admission.SetID == setID {
		return fmt.Errorf("%w: selected set work remains", workflowstate.ErrConflict)
	}

	for _, attempt := range run.Attempts {
		if slices.Contains([]string{"running", "checkpointed"}, attempt.Outcome) {
			return ErrWorkflowOwned
		}

		if err := workflowAttemptQuiescent(attempt); err != nil {
			return err
		}
	}

	return s.checkWorkflowInputs(ctx, run)
}

func workflowClosingRunStopped(run workflowInvocationFile, root string) bool {
	return workflowInvocationTerminal(run.Status) || run.StopIntent != "" || run.OwnerRoot != root || run.Status == "paused"
}

func workflowClosingRunFrontier(run workflowInvocationFile, setID string) bool {
	return run.Lane == nil && run.Submission == nil && run.Cursor == run.ConfirmedCursor && workflowNextClosingSet(run.Invocation) == setID
}

func (s *Store) workflowClosingInvocationFiles(ctx context.Context, claim *workflowClosingClaim, previous InputImage, files []workflowstate.FileChange) ([]workflowstate.FileChange, error) {
	if claim.state.InvocationID == "" {
		return files, nil
	}

	run, err := s.readWorkflowInvocation(claim.state.InvocationID)
	if err != nil {
		return nil, err
	}

	if workflowInvocationTerminal(run.Status) && claim.state.Outcome != "completed" {
		return files, nil
	}

	if err := workflowClosingProjectionFence(run, claim, previous, s.layout.Root()); err != nil {
		return nil, err
	}

	for i := range run.SetImages {
		if run.SetImages[i].Path == claim.set.Path {
			run.SetImages[i].Revision, run.SetImages[i].SourceHash = claim.set.Revision, claim.set.SourceHash
		}
	}

	projectWorkflowClosing(&run, claim)

	run.CapacityResumeStatus = ""

	if run.StopIntent != "" && claim.state.Outcome != "running" {
		run.Status = workflowstate.StopStatus(run.StopIntent)
	}

	run.Revision++
	run.UpdatedAt = claim.set.UpdatedAt

	change, err := workflowInvocationChange(run)

	return append(files, change), err
}

func workflowClosingProjectionFence(run workflowInvocationFile, claim *workflowClosingClaim, previous InputImage, root string) error {
	if run.OwnerRoot != root || !workflowClosingRunFrontier(run, claim.set.ID) {
		return ErrWorkflowOwned
	}

	image := run.SetImages[len(run.Closures)]
	if image.Path != previous.Path || image.Revision != previous.Revision || image.SourceHash != previous.SourceHash {
		return workflowstate.ErrConflict
	}

	if run.ClosingSet != "" && run.ClosingSet != claim.set.ID {
		return ErrWorkflowOwned
	}

	return nil
}

func projectWorkflowClosing(run *workflowInvocationFile, claim *workflowClosingClaim) {
	run.ClosingSet = ""

	run.Status, run.Outcome = "waiting_integration", "set_closure_required"
	switch claim.state.Outcome {
	case "running":
		run.ClosingSet, run.Status, run.Outcome = claim.set.ID, "running", "set_closing"
	case "completed":
		run.Closures = append(run.Closures, InvocationClosure{SetID: claim.set.ID, Cursor: run.Cursor, Receipt: claim.state.Result, Target: claim.input.Target})
		run.ExecutionTarget = &claim.input.Target

		run.Status, run.Outcome = "pending", "set_closed"
		if len(run.Closures) == len(run.SetImages) {
			run.Status, run.Outcome = "completed", "sets_closed"
		} else if workflowInvocationNeedsClosing(run.Invocation) {
			run.Status, run.Outcome = "waiting_integration", "set_closure_required"
		}
	default:
		run.Outcome = claim.state.Outcome
	}
}

// CloseWorkflowInvocationSet continues only the next frozen set's closing role.
func (s *Store) CloseWorkflowInvocationSet(ctx context.Context, id, branch string, ports FrontierPorts) (SetClosureResult, error) {
	run, err := s.ReadWorkflow(ctx, id)
	if err != nil {
		return SetClosureResult{}, err
	}

	if run.ClosingSet != "" {
		if _, err := s.ReconcileWorkflowClosing(ctx, run.SetImages[len(run.Closures)].Path); err != nil {
			return SetClosureResult{}, err
		}

		run, err = s.ReadWorkflow(ctx, id)
		if err != nil {
			return SetClosureResult{}, err
		}
	}

	if workflowNextClosingSet(run) == "" || workflowInvocationTerminal(run.Status) {
		return SetClosureResult{}, ErrWorkflowStop
	}

	image := run.SetImages[len(run.Closures)]
	if len(run.Deliveries) > 0 {
		retained := run.Deliveries[len(run.Deliveries)-1].PullRequest.Request.BaseBranch
		if branch != "" && branch != retained {
			return SetClosureResult{}, workflowstate.ErrConflict
		}

		branch = retained
	}

	request := SetClosureRequest{SetPath: image.Path, ExpectedRevision: image.Revision, ExpectedSourceHash: image.SourceHash, Repository: run.Selection.Target.Repository, Branch: branch, InvocationID: run.ID}

	return s.CloseWorkflowSet(ctx, request, ports)
}
