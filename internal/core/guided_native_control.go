package core

import (
	"context"
	"errors"
	"path/filepath"
	"reflect"

	authoringcore "forge.adrianpk.com/chr/internal/core/authoring"
	opscore "forge.adrianpk.com/chr/internal/core/ops"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

func (s *WorkflowStore) checkGuidedWorkflowControl(ctx context.Context, run WorkflowInvocation, request workflowstate.ControlMutationRequest) error {
	if request.Guided == nil {
		return nil
	}

	absolute, err := filepath.Abs(s.layout.ProjectDir())
	if err != nil {
		return errors.Join(workflowstate.ErrConflict, err)
	}

	session, err := authoringcore.ReadOwnedSession(ctx, absolute, request.Guided.RunID)
	if err != nil {
		return err
	}

	if err := authoringcore.ValidateGuidedScope(session, authoringcore.GuidedUseCaseScope{RunID: request.Guided.RunID, ProjectID: s.projectID, Namespace: s.layout.Namespace()}); err != nil {
		return err
	}

	return guardGuidedNativeControl(session, run, request)
}

func guardGuidedNativeControl(session authoringcore.Session, run WorkflowInvocation, request workflowstate.ControlMutationRequest) error {
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
func ReadGuidedInvocation(ctx context.Context, root string, session authoringcore.Session) (WorkflowInvocation, error) {
	if session.Guided == nil {
		return WorkflowInvocation{}, workflowstate.ErrConflict
	}

	layout, err := opscore.ResolveLayout(opscore.LayoutOptions{ProjectDir: root, Namespace: session.Namespace})
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

func guidedNativeReferenceMatches(run WorkflowInvocation, session authoringcore.Session) bool {
	if run.Admission == nil || run.Admission.Guided == nil {
		return false
	}

	proof := run.Admission.Guided

	return run.ProjectID == session.ProjectID && run.Namespace == session.Namespace && run.RequestKey == session.Guided.Work.Invocation.RequestKey && proof.RunID == session.ID && proof.Actor == session.Guided.Actor && proof.MaterializationID == session.Guided.Work.Materialization.ID
}

func guidedNativeControlAuthority(session authoringcore.Session, run WorkflowInvocation, request workflowstate.ControlMutationRequest) bool {
	state := session.Guided

	return guidedNativeReferenceMatches(run, session) && run.Status != "completed" && request.Binding.Operator == state.Actor && request.Guided.OperationDigest == state.Control.PendingOperation.Digest
}
