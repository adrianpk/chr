package core

import (
	"context"
	"fmt"
	"time"
)

// WorkflowUnitDelivery retains a completed slice before its active bindings move.
// EvidenceRef binds the exact frontier, target and receipt to an immutable image.
type WorkflowUnitDelivery struct {
	Schema            string                      `json:"schema"`
	InvocationID      string                      `json:"invocation_id"`
	Unit              WorkflowRef                 `json:"unit"`
	SetID             string                      `json:"set_id"`
	Cursor            int                         `json:"cursor"`
	Target            WorkflowPreviewTarget       `json:"target"`
	NextTarget        WorkflowPreviewTarget       `json:"next_target"`
	Lane              WorkflowLane                `json:"lane"`
	CandidateRef      string                      `json:"candidate_ref"`
	CandidateHash     string                      `json:"candidate_hash"`
	PullRequest       *WorkflowPullRequestBinding `json:"pull_request_binding"`
	DeliveryProofRef  string                      `json:"delivery_proof_ref"`
	DeliveryProofHash string                      `json:"delivery_proof_hash"`
	DeliveryRef       string                      `json:"delivery_ref"`
	DeliveryHash      string                      `json:"delivery_hash"`
	Preparation       *WorkflowPreparationBinding `json:"preparation_binding,omitempty"`
	Review            *WorkflowReviewBinding      `json:"review_binding,omitempty"`
	Submission        *WorkflowSubmissionBinding  `json:"submission_binding,omitempty"`
	Proof             WorkflowIntegrationProof    `json:"proof"`
	EvidenceRef       string                      `json:"evidence_ref,omitempty"`
	EvidenceHash      string                      `json:"evidence_hash,omitempty"`
}

// WorkflowFrontierPorts supplies trusted canonical observation and target hydration.
type WorkflowFrontierPorts struct {
	Provider WorkflowPullRequestProvider
	Observer WorkflowIntegrationGit
	Target   WorkflowTargetGit
}

// ExecutionSelection preserves admission identity while refreshing execution base.
func (run WorkflowInvocation) ExecutionSelection() WorkflowPreviewRequest {
	selection := run.Selection
	if run.ExecutionTarget != nil {
		selection.Target = *run.ExecutionTarget
	}

	return selection
}

// AdvanceWorkflowFrontier releases one finalized slice's barrier without work.
// It verifies canonical delivery, journals retained proof, and hydrates the next
// base. Set boundaries wait for closing validation, including the last set.
func (s *WorkflowStore) AdvanceWorkflowFrontier(ctx context.Context, id string, ports WorkflowFrontierPorts) (result WorkflowInvocation, err error) {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Minute)
	defer cancel()

	err = s.withLock(ctx, func() error {
		run, owner, readErr := s.locateWorkflowInvocation(ctx, id)
		if readErr != nil {
			return readErr
		}

		readErr = owner.advanceWorkflowFrontier(ctx, &run, ports)
		result = run.WorkflowInvocation

		return readErr
	})

	return result, err
}

func (s *WorkflowStore) advanceWorkflowFrontier(ctx context.Context, run *workflowInvocationFile, ports WorkflowFrontierPorts) error {
	ready, err := s.workflowFrontierReady(ctx, *run)
	if err != nil || !ready {
		return err
	}

	receipt, err := s.readWorkflowSliceReceipt(ctx, *run)
	if err != nil {
		return err
	}

	candidate, err := s.readWorkflowFrontierCandidate(ctx, *run)
	if err != nil {
		return err
	}

	target, proof, err := s.observeWorkflowFrontier(ctx, *run, receipt, candidate, ports)
	if err != nil {
		return err
	}

	if err := s.checkWorkflowInputs(ctx, *run); err != nil {
		return err
	}

	return s.commitWorkflowFrontier(run, target, proof)
}

func (s *WorkflowStore) workflowFrontierReady(ctx context.Context, run workflowInvocationFile) (bool, error) {
	if run.StopIntent != "" || run.Status == "paused" || run.Status == "cancelled" {
		return false, ErrWorkflowStop
	}

	if err := s.checkWorkflowInputs(ctx, run); err != nil {
		return false, err
	}

	if run.Selection.Selector.Extent != "slice_set" && run.Selection.Selector.Extent != "use_case" {
		return false, nil
	}

	if run.Lane == nil {
		return false, nil
	}

	if run.DeliveryRef == "" {
		return false, ErrWorkflowObservationRequired
	}

	return true, nil
}

func (s *WorkflowStore) readWorkflowFrontierCandidate(ctx context.Context, run workflowInvocationFile) (WorkflowSliceCandidate, error) {
	selection, err := s.readWorkflowSelection(ctx)
	if err != nil {
		return WorkflowSliceCandidate{}, err
	}

	dependencies, err := newWorkflowDependencies(selection)
	if err != nil {
		return WorkflowSliceCandidate{}, err
	}

	tasks := dependencies.tasks[run.Lane.Unit.ID]
	if len(tasks) == 0 {
		return WorkflowSliceCandidate{}, fmt.Errorf("%w: finalized slice has no members", ErrWorkflowConflict)
	}

	observation := workflowSliceObservation{descriptor: selection.slices[run.Lane.Unit.ID], set: selection.sets[tasks[0].Task.SetID], tasks: tasks}

	return s.readWorkflowSliceCandidate(ctx, run, observation)
}
