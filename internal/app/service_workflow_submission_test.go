package app

import (
	"context"
	"errors"
	"testing"

	"forge.adrianpk.com/chr/internal/core"
)

type serviceWorkflowSubmissionFixture struct {
	*serviceWorkflowPreparationFixture
	submissions []string
	unpublished bool
}

func (p *serviceWorkflowSubmissionFixture) SubmitWorkflowIntegration(_ context.Context, submission core.WorkflowSubmission) (core.WorkflowSubmissionResponse, error) {
	p.submissions = append(p.submissions, submission.ID)

	return core.WorkflowSubmissionResponse{}, context.DeadlineExceeded
}

func (p *serviceWorkflowSubmissionFixture) ObserveWorkflowSubmission(_ context.Context, submission core.WorkflowSubmission) (core.WorkflowSubmissionObservation, error) {
	return core.WorkflowSubmissionObservation{SubmissionID: submission.ID, Unpublished: p.unpublished}, nil
}

// TestServiceWorkflowSubmission exercises the central operation's durable intent,
// lost response, observation-only retry and explicit sealed-result recovery.
func TestServiceWorkflowSubmission(t *testing.T) {
	operations, id, guarded := serviceWorkflowPreparationReady(t)
	provider := &serviceWorkflowSubmissionFixture{serviceWorkflowPreparationFixture: guarded}

	operations.workflowPullRequests = provider
	if _, err := operations.PrepareWorkflowSliceIntegration(t.Context(), id); err != nil {
		t.Fatal(err)
	}

	review, err := operations.ReviewWorkflowSlice(t.Context(), id, nil)
	if err != nil {
		t.Fatal(err)
	}

	if _, err := operations.ApproveWorkflowSlice(t.Context(), id, review.InputHash); err != nil {
		t.Fatal(err)
	}

	assertServiceWorkflowSubmission(t, operations, id, provider)
}

func assertServiceWorkflowSubmission(t *testing.T, operations service, id string, provider *serviceWorkflowSubmissionFixture) {
	t.Helper()

	first, err := operations.SubmitWorkflowSliceIntegration(t.Context(), id)
	if !errors.Is(err, context.DeadlineExceeded) || first.State != "awaiting_observation" {
		t.Fatalf("submission: %+v, %v", first, err)
	}

	if _, err := operations.SubmitWorkflowSliceIntegration(t.Context(), id); !errors.Is(err, core.ErrWorkflowObservationRequired) {
		t.Fatal("central service retried an unknown result", err)
	}

	provider.unpublished = true

	last, err := operations.ReconcileWorkflowSubmission(t.Context(), id)
	if err != nil || last.ID != first.ID || last.State != "unpublished" || len(provider.submissions) != 1 {
		t.Fatalf("sealed observation: %+v, %v", last, err)
	}

	assertServiceWorkflowSubmissionCheckpoint(t, operations, id, first.ID)
}

func assertServiceWorkflowSubmissionCheckpoint(t *testing.T, operations service, id, submissionID string) {
	t.Helper()

	retained, err := operations.ReadWorkflowSubmission(t.Context(), id)
	if err != nil || retained.ID != submissionID || retained.Attempt != 1 {
		t.Fatal("central service lost the durable submission checkpoint")
	}
}
