package workflow

import (
	"context"

	workflowcloseout "forge.adrianpk.com/chr/internal/core/workflow/closeout"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

func (s *Store) observeWorkflowClosingTarget(ctx context.Context, request workflowcloseout.Request, ports FrontierPorts) (workflowstate.PreviewTarget, error) {
	query := TargetQuery{Repository: request.Repository, Branch: request.Branch, Directory: s.layout.ProjectDir()}

	target, err := ports.Target.ObserveWorkflowTarget(ctx, query)
	if err != nil {
		return target, err
	}

	if target.Repository != request.Repository || !workflowstate.ValidCommitID(target.Commit) {
		return target, workflowstate.ErrConflict
	}

	actual, err := ports.Provider.ReadWorkflowBranch(ctx, request.Repository, request.Branch)
	if err != nil || actual != target.Commit {
		return target, errorsOrConflict(err)
	}

	return target, s.checkWorkflowTarget(ctx, target)
}
