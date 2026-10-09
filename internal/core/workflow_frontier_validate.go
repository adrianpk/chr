package core

import (
	"bytes"
	"context"
	"fmt"
)

func validateWorkflowFrontier(run workflowInvocationFile) error {
	if run.ConfirmedCursor < 0 || run.ConfirmedCursor > run.Cursor || len(run.Deliveries) > WorkflowRecordLimit {
		return fmt.Errorf("workflow confirmed frontier exceeds bounds")
	}

	cursor := 0
	for _, delivery := range run.Deliveries {
		if err := validateWorkflowUnitDelivery(run, delivery, cursor); err != nil {
			return err
		}

		cursor = delivery.Cursor
	}

	if cursor > run.ConfirmedCursor {
		return ErrWorkflowConflict
	}

	return validateWorkflowExecutionTarget(run)
}

func validateWorkflowExecutionTarget(run workflowInvocationFile) error {
	if err := validateWorkflowInvocationClosures(run); err != nil {
		return err
	}

	var target *WorkflowPreviewTarget

	cursor := -1

	if len(run.Deliveries) != 0 {
		last := run.Deliveries[len(run.Deliveries)-1]
		target, cursor = &last.NextTarget, last.Cursor
	}

	if len(run.Closures) != 0 {
		last := run.Closures[len(run.Closures)-1]
		if last.Cursor >= cursor {
			target = &last.Target
		}
	}

	if run.ExecutionTarget != nil {
		if target == nil || *run.ExecutionTarget != *target {
			return ErrWorkflowConflict
		}
	} else if target != nil {
		return ErrWorkflowConflict
	}

	return nil
}

func validateWorkflowUnitDelivery(run workflowInvocationFile, delivery WorkflowUnitDelivery, previous int) error {
	if err := validateWorkflowUnitDeliveryIdentity(run, delivery, previous); err != nil {
		return err
	}

	if err := validateWorkflowUnitDeliveryTarget(run, delivery); err != nil {
		return err
	}

	if delivery.Lane.Unit != delivery.Unit || !validWorkflowLane(delivery.Lane) || delivery.PullRequest == nil {
		return ErrWorkflowConflict
	}

	for _, member := range run.Members[previous:delivery.Cursor] {
		if !workflowMemberOwnsUnit(member, delivery.Unit) || member.Admission.SetID != delivery.SetID {
			return ErrWorkflowConflict
		}
	}

	view := workflowDeliveredUnitView(run, delivery)
	if err := validateWorkflowInvocationBindings(view); err != nil {
		return err
	}

	query := WorkflowIntegrationQuery{Repository: delivery.PullRequest.Request.Repository, TargetBranch: delivery.PullRequest.Request.BaseBranch, Method: delivery.Proof.Method, SourceCommit: delivery.PullRequest.SourceCommit, SourceTree: delivery.PullRequest.SourceTree, OriginCommit: delivery.Target.Commit, ResultCommit: delivery.Proof.ResultCommit}

	return validateWorkflowIntegrationProof(query, delivery.Proof)
}

func validateWorkflowUnitDeliveryIdentity(run workflowInvocationFile, delivery WorkflowUnitDelivery, previous int) error {
	if delivery.Schema != "chr-unit-frontier/v1" || delivery.InvocationID != run.ID || !validWorkflowFrontierUnit(run, delivery, previous) || delivery.Cursor <= previous || delivery.Cursor > run.Cursor || delivery.Cursor > len(run.Members) {
		return ErrWorkflowConflict
	}

	if delivery.EvidenceRef != workflowFrontierEvidencePath(run.ID, delivery.Cursor) || len(delivery.EvidenceHash) != 64 {
		return ErrWorkflowConflict
	}

	return nil
}

func validWorkflowFrontierUnit(run workflowInvocationFile, delivery WorkflowUnitDelivery, previous int) bool {
	if delivery.Unit.Kind == "ticket" {
		return run.Selection.Selector.Extent == "loose" && delivery.SetID == "" && delivery.Cursor == previous+1
	}

	return delivery.Unit.Kind == "slice"
}

func validateWorkflowUnitDeliveryTarget(run workflowInvocationFile, delivery WorkflowUnitDelivery) error {
	if delivery.Target.Repository != run.Selection.Target.Repository || delivery.NextTarget.Repository != run.Selection.Target.Repository || delivery.NextTarget.Commit != delivery.Proof.ObservedTarget || !workflowCommitID(delivery.Target.Commit) {
		return ErrWorkflowConflict
	}

	return nil
}

func workflowDeliveredUnitView(run workflowInvocationFile, delivery WorkflowUnitDelivery) workflowInvocationFile {
	run.ExecutionTarget, run.Cursor, run.Lane = &delivery.Target, delivery.Cursor, &delivery.Lane
	run.CandidateRef, run.CandidateHash, run.PullRequest = delivery.CandidateRef, delivery.CandidateHash, delivery.PullRequest
	run.DeliveryProofRef, run.DeliveryProofHash = delivery.DeliveryProofRef, delivery.DeliveryProofHash
	run.DeliveryRef, run.DeliveryHash = delivery.DeliveryRef, delivery.DeliveryHash
	run.Preparation, run.Review, run.Submission = delivery.Preparation, delivery.Review, delivery.Submission

	return run
}

func (s *WorkflowStore) checkWorkflowContinuationEvidence(ctx context.Context, run workflowInvocationFile) error {
	if err := s.checkWorkflowInvocationClosureEvidence(run); err != nil {
		return err
	}

	for _, delivery := range run.Deliveries {
		if err := ctx.Err(); err != nil {
			return err
		}

		path, err := s.safePath(delivery.EvidenceRef)
		if err != nil {
			return err
		}

		data, err := readWorkflowFile(path)
		if err != nil || workflowHash(data) != delivery.EvidenceHash {
			return ErrWorkflowConflict
		}

		expected, err := encodeWorkflowUnitDelivery(delivery)
		if err != nil || !bytes.Equal(data, expected) {
			return ErrWorkflowConflict
		}

		refs := workflowUnitEvidenceHashes(run, delivery)
		if err := s.checkWorkflowPrerequisiteEvidence(refs); err != nil {
			return err
		}
	}

	return nil
}

func workflowUnitEvidenceHashes(run workflowInvocationFile, delivery WorkflowUnitDelivery) map[string]string {
	refs := map[string]string{delivery.DeliveryRef: delivery.DeliveryHash, delivery.CandidateRef: delivery.CandidateHash, delivery.DeliveryProofRef: delivery.DeliveryProofHash, delivery.PullRequest.ObservationRef: delivery.PullRequest.ObservationHash}
	for _, attempt := range run.Attempts {
		if attempt.Lane == nil || attempt.Lane.Unit != delivery.Unit {
			continue
		}

		if attempt.InputRef != "" {
			refs[attempt.InputRef] = attempt.InputHash
		}

		if attempt.ResultRef != "" {
			refs[attempt.ResultRef] = attempt.ResultHash
		}
	}

	return refs
}
