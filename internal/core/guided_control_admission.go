package core

import (
	"context"
	"path/filepath"
	"reflect"
	"slices"

	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

func validateGuidedControlProof(request WorkflowControlMutationRequest) error {
	proof := request.Guided
	if proof == nil {
		return nil
	}

	if !guidedControlAction(request.Action) || validateAuthoringSessionID(proof.RunID) != nil || !validWorkflowNativeFingerprint(proof.OperationDigest) {
		return workflowstate.ErrConflict
	}

	return nil
}

// BindGuidedControlInvocation reconciles uncertain admission without admitting
// new work. Only an existing exact request can become the control target.
func BindGuidedControlInvocation(ctx context.Context, root string, session *AuthoringSession) error {
	absolute, err := filepath.Abs(root)
	if err != nil || !holdsAuthoringOwner(ctx, absolute) {
		return workflowstate.ErrConflict
	}

	state := session.Guided
	if state == nil {
		return workflowstate.ErrConflict
	}

	if !guidedControlNeedsInvocation(state) {
		return nil
	}

	if state.Control == nil || state.Control.PendingOperation == nil {
		return workflowstate.ErrConflict
	}

	lookup, err := findGuidedControlInvocation(ctx, root, *session)
	if err != nil || !lookup.Found {
		return err
	}

	run := lookup.Invocation
	state.Work.Invocation = &GuidedInvocationReference{ID: run.ID, RequestKey: run.RequestKey, Revision: run.Revision}
	operation := *state.Control.PendingOperation
	operation.NativeRevision = run.Revision
	state.Control.NativeRequest = guidedNativeControlRequest(*session, operation)

	return RetainGuidedUseCaseProgress(ctx, root, session)
}

type guidedAdmissionLookup struct {
	Invocation WorkflowInvocation
	Found      bool
}

func (s *WorkflowStore) lookupGuidedAdmission(ctx context.Context, session AuthoringSession) (result guidedAdmissionLookup, err error) {
	err = s.withLock(ctx, func() error {
		runs, err := s.repositoryWorkflowInvocations(ctx)
		if err != nil {
			return err
		}

		request := session.Guided.Work.Admission
		for _, run := range runs {
			if run.Namespace != session.Namespace || !slices.Contains(run.RequestKeys, request.RequestKey) {
				continue
			}

			if run.ProjectID != session.ProjectID || !reflect.DeepEqual(run.Admission, request.Admission) {
				return workflowstate.ErrConflict
			}

			result = guidedAdmissionLookup{Invocation: run.WorkflowInvocation, Found: true}

			return nil
		}

		return nil
	})

	return result, err
}

func findGuidedControlInvocation(ctx context.Context, root string, session AuthoringSession) (guidedAdmissionLookup, error) {
	layout, err := ResolveOpsLayout(ResolveOpsLayoutOptions{ProjectDir: root, Namespace: session.Namespace})
	if err != nil {
		return guidedAdmissionLookup{}, err
	}

	return NewWorkflowStore(layout, session.ProjectID).lookupGuidedAdmission(ctx, session)
}

func guidedControlNeedsInvocation(state *GuidedUseCaseState) bool {
	return state.Work != nil && state.Work.Admission != nil && state.Work.Invocation == nil
}
