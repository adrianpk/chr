package core

import (
	"encoding/json"
	"fmt"
	"path/filepath"

	"forge.adrianpk.com/chr/internal/core/contenthash"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

func workflowDeliveryProofPath(run workflowInvocationFile) string {
	return filepath.ToSlash(filepath.Join("run", run.ID, "extension", "chr", "delivery", run.PullRequest.CandidateID+".json"))
}

func validateWorkflowDeliveryRef(run workflowInvocationFile) error {
	if run.DeliveryProofRef == "" && run.DeliveryProofHash == "" {
		return nil
	}

	if run.PullRequest == nil || run.DeliveryProofRef != workflowDeliveryProofPath(run) || len(run.DeliveryProofHash) != 64 || run.CandidateRef != run.PullRequest.CandidateRef || run.CandidateHash != run.PullRequest.CandidateHash {
		return workflowstate.ErrConflict
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
			return proof, workflowstate.ErrConflict
		}

		workflowRetainDeliveryStatus(run, "slice_delivery_verified", false)

		return previous, s.commitWorkflowState(run)
	}

	data, err := json.MarshalIndent(proof, "", "  ")
	if err != nil || len(data) >= WorkflowFileLimit {
		return proof, fmt.Errorf("delivery proof exceeds file bounds")
	}

	data = append(data, '\n')
	run.DeliveryProofRef, run.DeliveryProofHash = workflowDeliveryProofPath(*run), contenthash.Sum(data)
	workflowRetainDeliveryStatus(run, "slice_delivery_verified", false)

	change := workflowFileChange{Path: run.DeliveryProofRef, After: data}

	return proof, s.commitWorkflowPullRequestFiles(run, nil, change)
}

func (s *WorkflowStore) readWorkflowDeliveryProof(run workflowInvocationFile) (WorkflowSliceDeliveryProof, error) {
	var proof WorkflowSliceDeliveryProof
	if run.DeliveryProofRef == "" || validateWorkflowDeliveryRef(run) != nil {
		return proof, workflowstate.ErrConflict
	}

	path, err := s.safePath(run.DeliveryProofRef)
	if err != nil {
		return proof, err
	}

	data, err := readWorkflowFile(path)
	if err != nil || contenthash.Sum(data) != run.DeliveryProofHash {
		return proof, workflowstate.ErrConflict
	}

	if err := json.Unmarshal(data, &proof); err != nil {
		return proof, err
	}

	if proof.NativeReceiptRef != nil {
		return proof, s.checkWorkflowNativeDeliveryProof(run, proof)
	}

	return proof, checkWorkflowLegacyDeliveryProof(run, proof)
}

func checkWorkflowLegacyDeliveryProof(run workflowInvocationFile, proof WorkflowSliceDeliveryProof) error {
	if !workflowDeliveryProofMatchesRun(run, proof) || !proof.PullRequest.Merged || workflowPullRequestMismatch(*run.PullRequest, []WorkflowPullRequestSnapshot{proof.PullRequest}) != "" {
		return workflowstate.ErrConflict
	}

	query := WorkflowIntegrationQuery{Repository: run.PullRequest.Request.Repository, TargetBranch: run.PullRequest.Request.BaseBranch, Method: proof.Git.Method, SourceCommit: run.PullRequest.SourceCommit, SourceTree: run.PullRequest.SourceTree, OriginCommit: run.ExecutionSelection().Target.Commit, ResultCommit: proof.PullRequest.ResultCommit}

	return validateWorkflowIntegrationProof(query, proof.Git)
}

func workflowDeliveryProofMatchesRun(run workflowInvocationFile, proof WorkflowSliceDeliveryProof) bool {
	if proof.Schema != "chr-slice-delivery-proof/v1" || proof.InvocationID != run.ID || proof.Unit != run.Lane.Unit || proof.ConfirmedAt.IsZero() {
		return false
	}

	return proof.CandidateID == run.PullRequest.CandidateID && proof.CandidateRef == run.CandidateRef && proof.CandidateHash == run.CandidateHash && proof.SubmissionID == workflowDeliverySubmissionID(run)
}
