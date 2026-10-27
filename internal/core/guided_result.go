package core

import (
	"context"
	"reflect"

	authoringcore "forge.adrianpk.com/chr/internal/core/authoring"
	opscore "forge.adrianpk.com/chr/internal/core/ops"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

// ReadGuidedResult verifies the exact native binding and its original evidence.
func ReadGuidedResult(ctx context.Context, root string, session authoringcore.Session) (WorkflowResultInspection, error) {
	if session.Guided == nil || session.Guided.Work == nil || session.Guided.Work.Invocation == nil {
		return WorkflowResultInspection{}, workflowstate.ErrConflict
	}

	layout, err := opscore.ResolveLayout(opscore.LayoutOptions{ProjectDir: root, Namespace: session.Namespace})
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

func guardGuidedResultBinding(session authoringcore.Session, run WorkflowInvocation) error {
	work := session.Guided.Work
	if work.Materialization == nil || work.Admission == nil {
		return workflowstate.ErrConflict
	}

	if !guidedNativeReferenceMatches(run, session) || !reflect.DeepEqual(run.Admission, work.Admission.Admission) {
		return workflowstate.ErrConflict
	}

	if run.Selection.Selector.Extent != "slice_set" || run.Selection.Selector.Identity != work.Materialization.SetID || len(run.Members) != len(work.Materialization.Members) {
		return workflowstate.ErrConflict
	}

	return guardGuidedResultMembers(work, run)
}

func guardGuidedResultMembers(work *authoringcore.GuidedWorkState, run WorkflowInvocation) error {
	for i, member := range run.Members {
		frozen := work.Materialization.Members[i]
		if member.Admission.Key != frozen.Key || member.Admission.ContractHash != frozen.ContractHash {
			return workflowstate.ErrConflict
		}
	}

	return nil
}

// ReconcileGuidedResult is an owned lifecycle transition. Only complete native
// receipts can finish a run; pending operations and explicit stops take priority.
func ReconcileGuidedResult(ctx context.Context, root string, scope authoringcore.GuidedUseCaseScope) (session authoringcore.Session, err error) {
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

		reference := authoringcore.GuidedResultReference{InvocationID: result.Invocation.ID, Revision: result.Invocation.Revision}
		for _, receipt := range result.Receipts {
			reference.Receipts = append(reference.Receipts, receipt.Reference)
		}

		for _, closure := range result.Closures {
			reference.Closures = append(reference.Closures, closure.Reference)
		}

		session.Guided.Result = &reference
		session.Guided.Stage, session.Guided.Status, session.Guided.Blocker = "results", "completed", ""
		session.Status, session.FinishedAt = authoringcore.SessionStatusFinalized, &result.Invocation.UpdatedAt

		return RetainGuidedUseCaseProgress(ctx, root, &session)
	})

	return session, err
}

func guidedResultReady(session authoringcore.Session) bool {
	state := session.Guided
	if !GuidedDispatchAllowed(session) || state.PendingOperation != nil || state.Stage != "execution" {
		return false
	}

	return state.Control == nil || state.Control.PendingOperation == nil
}
