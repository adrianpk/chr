package workflow

import (
	"context"

	workflowcloseout "forge.adrianpk.com/chr/internal/core/workflow/closeout"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

func (s *Store) workflowClosingInvocationFiles(claim *workflowClosingClaim, previous workflowstate.InputImage, files []workflowstate.FileChange) ([]workflowstate.FileChange, error) {
	if claim.state.InvocationID == "" {
		return files, nil
	}

	run, err := s.readWorkflowInvocation(claim.state.InvocationID)
	if err != nil {
		return nil, err
	}

	if workflowstate.InvocationTerminal(run.Status) && claim.state.Outcome != "completed" {
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

	workflowcloseout.ProjectInvocation(&run.Invocation, workflowcloseout.Projection{State: claim.state, SetID: claim.set.ID, Target: &claim.input.Target})

	run.CapacityResumeStatus = workflowstate.InvocationUnspecified

	if run.StopIntent != "" && claim.state.Outcome != "running" {
		run.Status = workflowstate.StopStatus(run.StopIntent)
	}

	run.Revision++
	run.UpdatedAt = claim.set.UpdatedAt

	change, err := workflowInvocationChange(run)

	return append(files, change), err
}

func workflowClosingProjectionFence(run workflowInvocationFile, claim *workflowClosingClaim, previous workflowstate.InputImage, root string) error {
	if run.OwnerRoot != root || !workflowcloseout.AtFrontier(run.Invocation, claim.set.ID) {
		return workflowstate.ErrOwned
	}

	image := run.SetImages[len(run.Closures)]
	if image.Path != previous.Path || image.Revision != previous.Revision || image.SourceHash != previous.SourceHash {
		return workflowstate.ErrConflict
	}

	if run.ClosingSet != "" && run.ClosingSet != claim.set.ID {
		return workflowstate.ErrOwned
	}

	return nil
}

// CloseInvocationSet continues only the next frozen set's closing role.
func CloseInvocationSet(ctx context.Context, store *Store, id, branch string, ports FrontierPorts) (workflowcloseout.Result, error) {
	run, err := store.ReadWorkflow(ctx, id)
	if err != nil {
		return workflowcloseout.Result{}, err
	}

	if run.ClosingSet != "" {
		if _, err := store.ReconcileWorkflowClosing(ctx, run.SetImages[len(run.Closures)].Path); err != nil {
			return workflowcloseout.Result{}, err
		}

		run, err = store.ReadWorkflow(ctx, id)
		if err != nil {
			return workflowcloseout.Result{}, err
		}
	}

	if workflowcloseout.NextSet(run) == "" || workflowstate.InvocationTerminal(run.Status) {
		return workflowcloseout.Result{}, workflowstate.ErrStop
	}

	image := run.SetImages[len(run.Closures)]
	if len(run.Deliveries) > 0 {
		retained := run.Deliveries[len(run.Deliveries)-1].PullRequest.Request.BaseBranch
		if branch != "" && branch != retained {
			return workflowcloseout.Result{}, workflowstate.ErrConflict
		}

		branch = retained
	}

	request := workflowcloseout.Request{SetPath: image.Path, ExpectedRevision: image.Revision, ExpectedSourceHash: image.SourceHash, Repository: run.Selection.Target.Repository, Branch: branch, InvocationID: run.ID}

	return CloseSet(ctx, store, request, ports)
}
