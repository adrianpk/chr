package core

import (
	"encoding/json"
	"fmt"
	"path/filepath"
)

func workflowDeliveryProofPath(run workflowInvocationFile) string {
	return filepath.ToSlash(filepath.Join("run", run.ID, "extension", "chr", "delivery", run.PullRequest.CandidateID+".json"))
}

func validateWorkflowDeliveryRef(run workflowInvocationFile) error {
	if run.DeliveryProofRef == "" && run.DeliveryProofHash == "" {
		return nil
	}

	if run.PullRequest == nil || run.DeliveryProofRef != workflowDeliveryProofPath(run) || len(run.DeliveryProofHash) != 64 || run.CandidateRef != run.PullRequest.CandidateRef || run.CandidateHash != run.PullRequest.CandidateHash {
		return ErrWorkflowConflict
	}

	return nil
}

func (s *WorkflowStore) persistWorkflowDeliveryProof(run *workflowInvocationFile, proof WorkflowSliceDeliveryProof) (WorkflowSliceDeliveryProof, error) {
	if run.DeliveryProofRef != "" {
		previous, err := s.readWorkflowDeliveryProof(*run)
		if err != nil {
			return proof, err
		}

		if previous.Git.HistoryHash != proof.Git.HistoryHash || previous.Git.ResultCommit != proof.Git.ResultCommit || previous.Git.TargetBranch != proof.Git.TargetBranch {
			return proof, ErrWorkflowConflict
		}

		workflowRetainDeliveryStatus(run, "slice_delivery_verified", false)

		return previous, s.commitWorkflowState(run)
	}

	data, err := json.MarshalIndent(proof, "", "  ")
	if err != nil || len(data) >= WorkflowFileLimit {
		return proof, fmt.Errorf("delivery proof exceeds file bounds")
	}

	data = append(data, '\n')
	run.DeliveryProofRef, run.DeliveryProofHash = workflowDeliveryProofPath(*run), workflowHash(data)
	workflowRetainDeliveryStatus(run, "slice_delivery_verified", false)

	change := workflowFileChange{Path: run.DeliveryProofRef, After: data}

	return proof, s.commitWorkflowPullRequestFiles(run, nil, change)
}

func (s *WorkflowStore) readWorkflowDeliveryProof(run workflowInvocationFile) (WorkflowSliceDeliveryProof, error) {
	var proof WorkflowSliceDeliveryProof
	if run.DeliveryProofRef == "" || validateWorkflowDeliveryRef(run) != nil {
		return proof, ErrWorkflowConflict
	}

	path, err := s.safePath(run.DeliveryProofRef)
	if err != nil {
		return proof, err
	}

	data, err := readWorkflowFile(path)
	if err != nil || workflowHash(data) != run.DeliveryProofHash {
		return proof, ErrWorkflowConflict
	}

	if err := json.Unmarshal(data, &proof); err != nil {
		return proof, err
	}

	if !workflowDeliveryProofMatchesRun(run, proof) || !proof.PullRequest.Merged || workflowPullRequestMismatch(*run.PullRequest, []WorkflowPullRequestSnapshot{proof.PullRequest}) != "" {
		return proof, ErrWorkflowConflict
	}

	query := WorkflowIntegrationQuery{Repository: run.PullRequest.Request.Repository, TargetBranch: run.PullRequest.Request.BaseBranch, Method: proof.Git.Method, SourceCommit: run.PullRequest.SourceCommit, SourceTree: run.PullRequest.SourceTree, OriginCommit: run.Selection.Target.Commit, ResultCommit: proof.PullRequest.ResultCommit}

	return proof, validateWorkflowIntegrationProof(query, proof.Git)
}

func workflowDeliveryProofMatchesRun(run workflowInvocationFile, proof WorkflowSliceDeliveryProof) bool {
	if proof.Schema != "chr-slice-delivery-proof/v1" || proof.InvocationID != run.ID || proof.Unit != run.Lane.Unit || proof.ConfirmedAt.IsZero() {
		return false
	}

	return proof.CandidateID == run.PullRequest.CandidateID && proof.CandidateRef == run.CandidateRef && proof.CandidateHash == run.CandidateHash && proof.SubmissionID == workflowDeliverySubmissionID(run)
}
