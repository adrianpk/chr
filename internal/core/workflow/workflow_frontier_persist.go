package workflow

import (
	"encoding/json"
	"fmt"
	"path/filepath"

	"forge.adrianpk.com/chr/internal/core/contenthash"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

func (s *Store) commitWorkflowFrontier(run *workflowInvocationFile, target workflowstate.PreviewTarget, proof workflowstate.IntegrationProof) error {
	delivery := workflowUnitDelivery(*run, target, proof)

	data, err := encodeWorkflowUnitDelivery(delivery)
	if err != nil {
		return err
	}

	delivery.EvidenceRef = workflowFrontierEvidencePath(run.ID, delivery.Cursor)
	delivery.EvidenceHash = contenthash.Sum(data)
	run.Deliveries = append(run.Deliveries, delivery)
	run.ExecutionTarget = &target
	run.ConfirmedCursor = run.Cursor
	clearWorkflowSliceBindings(run)

	run.Status, run.Outcome, run.ResumeStatus, run.Blocker = workflowstate.InvocationPending, "slice_integrated", workflowstate.InvocationUnspecified, nil
	if run.Selection.Selector.Extent == "loose" {
		run.Outcome = "ticket_integrated"
	} else if run.Cursor == len(run.Members) || run.Members[run.Cursor].Admission.SetID != delivery.SetID {
		run.Status, run.Outcome = workflowstate.InvocationWaitingIntegration, "set_closure_required"
	}

	return s.commitWorkflowPullRequestFiles(run, nil, workflowstate.FileChange{Path: delivery.EvidenceRef, After: data})
}

func workflowUnitDelivery(run workflowInvocationFile, target workflowstate.PreviewTarget, proof workflowstate.IntegrationProof) UnitDelivery {
	return UnitDelivery{
		Schema: "chr-unit-frontier/v1", InvocationID: run.ID, Unit: run.Lane.Unit,
		SetID: run.Members[run.Cursor-1].Admission.SetID, Cursor: run.Cursor,
		Target: run.ExecutionSelection().Target, NextTarget: target, Lane: *run.Lane,
		CandidateRef: run.CandidateRef, CandidateHash: run.CandidateHash, PullRequest: run.PullRequest,
		DeliveryProofRef: run.DeliveryProofRef, DeliveryProofHash: run.DeliveryProofHash,
		DeliveryRef: run.DeliveryRef, DeliveryHash: run.DeliveryHash,
		Preparation: run.Preparation, Review: run.Review, Submission: run.Submission, Proof: proof,
	}
}

func clearWorkflowSliceBindings(run *workflowInvocationFile) {
	run.Lane, run.PullRequest, run.Preparation, run.Review, run.Submission = nil, nil, nil, nil, nil
	run.CandidateRef, run.CandidateHash = "", ""
	run.DeliveryProofRef, run.DeliveryProofHash = "", ""
	run.DeliveryRef, run.DeliveryHash = "", ""
	run.Integration = nil
}

func encodeWorkflowUnitDelivery(delivery UnitDelivery) ([]byte, error) {
	delivery.EvidenceRef, delivery.EvidenceHash = "", ""

	data, err := json.MarshalIndent(delivery, "", "  ")
	if err != nil || len(data) >= workflowstate.FileLimit {
		return nil, fmt.Errorf("workflow frontier evidence exceeds file bounds: %w", errorsOrConflict(err))
	}

	return append(data, '\n'), nil
}

func workflowFrontierEvidencePath(id string, cursor int) string {
	return filepath.ToSlash(filepath.Join("run", id, "evidence", fmt.Sprintf("frontier-%04d.json", cursor)))
}

// Known optional fields must disappear when their authoritative typed value is
// cleared. Unknown fields survive; stale serialized PR/owner bindings must not.
func workflowOptionalInvocationField(key string) bool {
	if key == "capacity_resume_status" || key == "integration_binding" || key == "integration_entries" || key == "integration_budget" {
		return true
	}

	switch key {
	case "blocker", "lane", "pull_request_binding", "review_binding", "preparation_binding", "submission_binding", "review_restriction", "review_count", "review_repairs", "preparation_count", "execution_target", "confirmed_cursor", "integrated_units", "continuation_checkpoint", "set_closures", "closing_set", "loose_queue", "dependency_targets":
		return true
	default:
		return false
	}
}
