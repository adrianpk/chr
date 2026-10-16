package app

import (
	"context"
	"reflect"

	"forge.adrianpk.com/chr/internal/core"
)

// WorkflowRetainedReference names original admission, independently of an adapter.
type WorkflowRetainedReference struct {
	ID         string
	RequestKey string
}

// ResumeBoundWork observes original scope and delegates one guarded checkpoint
// to the shared retained-work owner. It never admits or widens native work.
func (s service) ResumeBoundWork(ctx context.Context, reference WorkflowRetainedReference, preview WorkflowBoundPreview, check core.WorkflowAdmissionCheck) (RetainedWorkProgress, error) {
	run, err := s.ReadWorkflow(ctx, reference.ID)
	if err != nil {
		return RetainedWorkProgress{}, err
	}

	expected := &core.WorkflowAdmissionGuard{SnapshotHash: preview.Request.SnapshotHash, Selected: preview.Selected, Policy: preview.View.Policy, ExecutionPolicy: preview.View.ExecutionPolicy}
	if run.RequestKey != reference.RequestKey || run.ProjectID != s.projectContext.ProjectID || run.Selection.Selector != preview.Request.Selector || run.Selection.Target != preview.Request.Target || !reflect.DeepEqual(run.Admission, expected) {
		return RetainedWorkProgress{}, core.ErrWorkflowConflict
	}

	return s.ResumeRetainedWorkGuarded(ctx, run.ID, run.Revision, check)
}
