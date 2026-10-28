package workflow

import (
	"context"

	authoringcore "forge.adrianpk.com/chr/internal/core/authoring"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

type guidedWorkflowObserver struct{ root string }

// NewGuidedWorkflowObserver supplies the native owner's read-only guided ports.
// Construction performs no IO; each operation keeps its existing native guard.
func NewGuidedWorkflowObserver(root string) authoringcore.GuidedWorkflowObserver {
	return guidedWorkflowObserver{root: root}
}

func (observer guidedWorkflowObserver) ObserveStop(ctx context.Context, session authoringcore.Session, intent string) (authoringcore.GuidedStopObservation, error) {
	run, err := ReadGuidedInvocation(ctx, observer.root, session)

	return authoringcore.GuidedStopObservation{
		Status:       string(run.Status),
		Revision:     run.Revision,
		Acknowledged: run.Status == workflowstate.StopStatus(intent) && run.StopIntent == intent,
	}, err
}

func (observer guidedWorkflowObserver) FindAdmission(ctx context.Context, session authoringcore.Session) (authoringcore.GuidedAdmissionObservation, error) {
	lookup, err := findGuidedControlInvocation(ctx, observer.root, session)
	run := lookup.Invocation

	return authoringcore.GuidedAdmissionObservation{
		Invocation: authoringcore.GuidedInvocationReference{ID: run.ID, RequestKey: run.RequestKey, Revision: run.Revision},
		Found:      lookup.Found,
	}, err
}

func (observer guidedWorkflowObserver) InspectResult(ctx context.Context, session authoringcore.Session) (authoringcore.GuidedResultObservation, error) {
	native, err := ReadGuidedResult(ctx, observer.root, session)
	result := authoringcore.GuidedResultObservation{
		Reference: authoringcore.GuidedResultReference{InvocationID: native.Invocation.ID, Revision: native.Invocation.Revision},
		UpdatedAt: native.Invocation.UpdatedAt,
		Complete:  native.Complete,
	}

	// Preserve partial observations on errors; only the native owner verifies proof.
	for _, receipt := range native.Receipts {
		result.Reference.Receipts = append(result.Reference.Receipts, receipt.Reference)
	}

	for _, closure := range native.Closures {
		result.Reference.Closures = append(result.Reference.Closures, closure.Reference)
	}

	return result, err
}
