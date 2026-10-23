package core

import (
	"context"
	"reflect"
)

// GuidedResultReference retains only native completion references, not new receipts.
type GuidedResultReference struct {
	InvocationID string
	Revision     uint64
	Receipts     []WorkflowEvidenceRef
	Closures     []WorkflowClosingEvidence
}

// ReadGuidedResult verifies the exact native binding and its original evidence.
func ReadGuidedResult(ctx context.Context, root string, session AuthoringSession) (WorkflowResultInspection, error) {
	if session.Guided == nil || session.Guided.Work == nil || session.Guided.Work.Invocation == nil {
		return WorkflowResultInspection{}, ErrWorkflowConflict
	}

	layout, err := ResolveOpsLayout(ResolveOpsLayoutOptions{ProjectDir: root, Namespace: session.Namespace})
	if err != nil {
		return WorkflowResultInspection{}, err
	}

	result, err := NewWorkflowStore(layout, session.ProjectID).InspectWorkflowResult(ctx, session.Guided.Work.Invocation.ID)
	if err != nil {
		return result, err
	}

	if err := guardGuidedResultBinding(session, result.Invocation); err != nil {
		return WorkflowResultInspection{}, err
	}

	return result, nil
}

func guardGuidedResultBinding(session AuthoringSession, run WorkflowInvocation) error {
	work := session.Guided.Work
	if work.Materialization == nil || work.Admission == nil {
		return ErrWorkflowConflict
	}

	if !guidedNativeReferenceMatches(run, session) || !reflect.DeepEqual(run.Admission, work.Admission.Admission) {
		return ErrWorkflowConflict
	}

	if run.Selection.Selector.Extent != "slice_set" || run.Selection.Selector.Identity != work.Materialization.SetID || len(run.Members) != len(work.Materialization.Members) {
		return ErrWorkflowConflict
	}

	return guardGuidedResultMembers(work, run)
}

func guardGuidedResultMembers(work *GuidedWorkState, run WorkflowInvocation) error {
	for i, member := range run.Members {
		frozen := work.Materialization.Members[i]
		if member.Admission.Key != frozen.Key || member.Admission.ContractHash != frozen.ContractHash {
			return ErrWorkflowConflict
		}
	}

	return nil
}

// ReconcileGuidedResult is an owned lifecycle transition. Only complete native
// receipts can finish a run; pending operations and explicit stops take priority.
func ReconcileGuidedResult(ctx context.Context, root string, scope GuidedUseCaseScope) (session AuthoringSession, err error) {
	r := &fileAuthoringSessionRepository{projectDir: root}
	err = WithAuthoringMutation(ctx, root, func(ctx context.Context) error {
		session, err = r.getGuardedSession(ctx, scope.RunID)
		if err != nil {
			return err
		}

		if err := guardGuidedScope(session, scope); err != nil {
			return err
		}

		if !guidedResultReady(session) {
			return nil
		}

		result, err := ReadGuidedResult(ctx, root, session)
		if err != nil || !result.Complete {
			return err
		}

		reference := GuidedResultReference{InvocationID: result.Invocation.ID, Revision: result.Invocation.Revision}
		for _, receipt := range result.Receipts {
			reference.Receipts = append(reference.Receipts, receipt.Reference)
		}

		for _, closure := range result.Closures {
			reference.Closures = append(reference.Closures, closure.Reference)
		}

		session.Guided.Result = &reference
		session.Guided.Stage, session.Guided.Status, session.Guided.Blocker = "results", "completed", ""
		session.Status, session.FinishedAt = AuthoringSessionStatusFinalized, &result.Invocation.UpdatedAt

		return RetainGuidedUseCaseProgress(ctx, root, &session)
	})

	return session, err
}

func guidedResultReady(session AuthoringSession) bool {
	state := session.Guided
	if !GuidedDispatchAllowed(session) || state.PendingOperation != nil || state.Stage != "execution" {
		return false
	}

	return state.Control == nil || state.Control.PendingOperation == nil
}
