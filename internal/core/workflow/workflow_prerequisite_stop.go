package workflow

import (
	"context"
	"errors"
	"fmt"
	"time"

	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

type workflowPrerequisiteError struct {
	blocker Blocker
}

func (failure workflowPrerequisiteError) Error() string {
	return fmt.Sprintf("workflow prerequisite %s: %s", failure.blocker.Dependency.ID, failure.blocker.Reason)
}

func workflowPrerequisitePaused(run workflowInvocationFile) bool {
	return run.Status == "paused" && run.Outcome == "prerequisite_blocked" && run.Blocker != nil && run.Blocker.NextActionOwner == "prerequisite_owner"
}

func workflowFailedTaskPaused(run workflowInvocationFile) bool {
	return run.Selection.Selector.Extent == "loose" && run.Status == "paused" && len(run.Attempts) > 0 && run.Attempts[len(run.Attempts)-1].Outcome == "failed"
}

func (s *Store) retainWorkflowPrerequisiteStop(run *workflowInvocationFile, failure error) error {
	var prerequisite workflowPrerequisiteError
	if !errors.As(failure, &prerequisite) {
		return failure
	}

	run.Status, run.Outcome, run.Blocker = "paused", "prerequisite_blocked", &prerequisite.blocker

	return errors.Join(failure, s.commitWorkflowState(run))
}

// ResumePorts can hydrate a descendant canonical target after an
// outside prerequisite integrates. It never supplies members or execution input.
type ResumePorts struct {
	AuthorizationCheck AdmissionCheck `json:"-"`
	ExpectedRevision   uint64
	Provider           PullRequestProvider
	Target             TargetGit
	TargetBranch       string
}

func (s *Store) resumeWorkflowPrerequisite(ctx context.Context, run *workflowInvocationFile, ports []ResumePorts) ([]workflowstate.FileChange, error) {
	if !workflowPrerequisitePaused(*run) {
		return nil, nil
	}

	_, err := s.observeWorkflowClaim(ctx, *run)
	if err == nil {
		run.Blocker = nil

		return nil, nil
	}

	var prerequisite workflowPrerequisiteError
	if !errors.As(err, &prerequisite) || len(ports) == 0 {
		return nil, err
	}

	return s.resumeWorkflowDependencyTarget(ctx, run, ports[0])
}

func (s *Store) commitWorkflowResume(run *workflowInvocationFile, changes []workflowstate.FileChange) error {
	run.Revision++
	run.UpdatedAt = time.Now().UTC()

	state, err := workflowInvocationChange(*run)
	if err != nil {
		return err
	}

	return s.commitFiles(append(changes, state))
}
