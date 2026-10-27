package core

import (
	"context"
	"errors"
	"path/filepath"
	"reflect"

	opscore "forge.adrianpk.com/chr/internal/core/ops"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

// GuidedWorkflowControlProof binds a native control to retained guided authority.
type GuidedWorkflowControlProof struct {
	RunID, OperationDigest string
}

func (s *WorkflowStore) checkGuidedWorkflowControl(ctx context.Context, run WorkflowInvocation, request WorkflowControlMutationRequest) error {
	if request.Guided == nil {
		return nil
	}

	absolute, err := filepath.Abs(s.layout.ProjectDir())
	if err != nil || !holdsAuthoringOwner(ctx, absolute) {
		return errors.Join(workflowstate.ErrConflict, err)
	}

	r := &fileAuthoringSessionRepository{projectDir: absolute}

	session, err := r.getGuardedSession(ctx, request.Guided.RunID)
	if err != nil {
		return err
	}

	if err := guardGuidedScope(session, GuidedUseCaseScope{RunID: request.Guided.RunID, ProjectID: s.projectID, Namespace: s.layout.Namespace()}); err != nil {
		return err
	}

	return guardGuidedNativeControl(session, run, request)
}

func guardGuidedNativeControl(session AuthoringSession, run WorkflowInvocation, request WorkflowControlMutationRequest) error {
	state := session.Guided
	if state.Work == nil || state.Work.Invocation == nil || state.Work.Materialization == nil || state.Control == nil || state.Control.PendingOperation == nil {
		return workflowstate.ErrConflict
	}

	if !guidedNativeControlAuthority(session, run, request) {
		return workflowstate.ErrConflict
	}

	if !reflect.DeepEqual(state.Control.NativeRequest, &request) {
		return workflowstate.ErrConflict
	}

	return nil
}

// ReadGuidedInvocation checks the exact native reference without claiming work.
func ReadGuidedInvocation(ctx context.Context, root string, session AuthoringSession) (WorkflowInvocation, error) {
	if session.Guided == nil {
		return WorkflowInvocation{}, workflowstate.ErrConflict
	}

	layout, err := opscore.ResolveOpsLayout(opscore.ResolveOpsLayoutOptions{ProjectDir: root, Namespace: session.Namespace})
	if err != nil {
		return WorkflowInvocation{}, err
	}

	work := session.Guided.Work
	if work == nil || work.Invocation == nil || work.Materialization == nil {
		return WorkflowInvocation{}, workflowstate.ErrConflict
	}

	run, err := NewWorkflowStore(layout, session.ProjectID).ReadWorkflow(ctx, work.Invocation.ID)
	if err != nil {
		return run, err
	}

	if !guidedNativeReferenceMatches(run, session) {
		return run, workflowstate.ErrConflict
	}

	return run, nil
}

func guidedNativeReferenceMatches(run WorkflowInvocation, session AuthoringSession) bool {
	if run.Admission == nil || run.Admission.Guided == nil {
		return false
	}

	proof := run.Admission.Guided

	return run.ProjectID == session.ProjectID && run.Namespace == session.Namespace && run.RequestKey == session.Guided.Work.Invocation.RequestKey && proof.RunID == session.ID && proof.Actor == session.Guided.Actor && proof.MaterializationID == session.Guided.Work.Materialization.ID
}

// ReconcileGuidedStop acknowledges quiescence only from the original native owner.
func ReconcileGuidedStop(ctx context.Context, root string, scope GuidedUseCaseScope) (session AuthoringSession, err error) {
	r := &fileAuthoringSessionRepository{projectDir: root}
	err = WithAuthoringMutation(ctx, root, func(ctx context.Context) error {
		session, err = r.getGuardedSession(ctx, scope.RunID)
		if err != nil {
			return err
		}

		if err := guardGuidedScope(session, scope); err != nil {
			return err
		}

		control := session.Guided.Control
		if !guidedStopNeedsAcknowledgement(control) {
			return nil
		}

		run, err := ReadGuidedInvocation(ctx, root, session)
		if err != nil {
			return err
		}

		if run.Status != workflowStopStatus(control.StopIntent) || run.StopIntent != control.StopIntent {
			return nil
		}

		control.AcknowledgedRevision = run.Revision
		session.Guided.Status, session.Guided.Blocker = run.Status, ""

		return RetainGuidedUseCaseProgress(ctx, root, &session)
	})

	return session, err
}

func guidedStopNeedsAcknowledgement(control *GuidedControlState) bool {
	return control != nil && control.PendingOperation == nil && control.StopIntent != "" && control.NativeReceipt != nil && control.AcknowledgedRevision == 0
}

func guidedNativeControlAuthority(session AuthoringSession, run WorkflowInvocation, request WorkflowControlMutationRequest) bool {
	state := session.Guided

	return guidedNativeReferenceMatches(run, session) && run.Status != "completed" && request.Binding.Operator == state.Actor && request.Guided.OperationDigest == state.Control.PendingOperation.Digest
}
