package core

import (
	"context"
	"fmt"
	"path/filepath"
	"slices"
	"strings"

	"forge.adrianpk.com/chr/internal/core/contenthash"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

func workflowSliceReceiptID(run workflowInvocationFile) string {
	return "delivery-" + contenthash.Sum([]byte(run.ID+"\x00"+run.PullRequest.CandidateID+"\x00"+workflowDeliverySubmissionID(run)))
}

func workflowSliceReceiptPath(run workflowInvocationFile) string {
	return filepath.ToSlash(filepath.Join("run", run.ID, "evidence", workflowSliceReceiptID(run)+".json"))
}

func validateWorkflowFinalizationRef(run workflowInvocationFile) error {
	if run.DeliveryRef == "" && run.DeliveryHash == "" {
		return nil
	}

	if run.PullRequest == nil || run.DeliveryProofRef == "" || run.DeliveryRef != workflowSliceReceiptPath(run) || len(run.DeliveryHash) != 64 {
		return workflowstate.ErrConflict
	}

	return nil
}

func workflowSliceReceipt(run workflowInvocationFile, proof WorkflowSliceDeliveryProof, candidate WorkflowSliceCandidate) (WorkflowRecord, error) {
	reference, err := DetectReviewProviderReference(proof.PullRequest.URL)
	if err != nil {
		return WorkflowRecord{}, err
	}

	git := proof.Git

	facts := &WorkflowReceipt{Schema: WorkflowSchema, Unit: proof.Unit, CandidateID: proof.CandidateID, SubmissionID: proof.SubmissionID, Provider: string(reference.Provider), Repository: git.Repository, PullRequest: proof.PullRequest.URL, Method: git.Method, SourceCommit: git.SourceCommit, BaseCommit: git.BaseCommit, PreparedTree: git.PreparedTree, HistoryProof: git.HistoryHash, ResultCommit: git.ResultCommit, VerifiedTree: git.VerifiedTree, ObservedTarget: git.ObservedTarget, VerificationRef: run.DeliveryProofRef, ValidationRefs: []string{proof.CandidateRef}, ConfirmedAt: proof.ConfirmedAt}
	for _, task := range candidate.UnitInput().Tasks {
		facts.ValidationRefs = append(facts.ValidationRefs, task.EvidenceRef)
	}

	identity := WorkflowIdentity{ID: workflowSliceReceiptID(run), ProjectID: run.ProjectID, Namespace: run.Namespace, Revision: 1, CreatedAt: proof.ConfirmedAt, UpdatedAt: proof.ConfirmedAt}
	record := WorkflowRecord{WorkflowIdentity: identity, Kind: "receipt", Path: workflowSliceReceiptPath(run), Receipt: facts}

	return record, record.Validate()
}

func (s *WorkflowStore) readWorkflowSliceReceipt(ctx context.Context, run workflowInvocationFile) (WorkflowRecord, error) {
	record, err := s.readWorkflowFinalizationReceipt(ctx, run)
	if err != nil {
		return record, err
	}

	return record, s.checkWorkflowFinalizationImages(ctx, run, record)
}

func (s *WorkflowStore) readWorkflowFinalizationReceipt(ctx context.Context, run workflowInvocationFile) (WorkflowRecord, error) {
	if err := validateWorkflowFinalizationRef(run); err != nil || run.DeliveryRef == "" {
		return WorkflowRecord{}, workflowstate.ErrConflict
	}

	if run.Submission != nil {
		if err := s.checkWorkflowSubmissionProjection(ctx, run); err != nil {
			return WorkflowRecord{}, err
		}
	}

	proof, err := s.readWorkflowDeliveryProof(run)
	if err != nil {
		return WorkflowRecord{}, err
	}

	record, err := s.readRecord(run.DeliveryRef)
	if err != nil || !workflowSliceReceiptHeaderMatches(record, run) || !workflowReceiptMatchesProof(record, proof, run.DeliveryProofRef) {
		return record, workflowstate.ErrConflict
	}

	if err := s.checkWorkflowInputs(ctx, run); err != nil {
		return record, err
	}

	return record, nil
}

func workflowSliceReceiptHeaderMatches(record WorkflowRecord, run workflowInvocationFile) bool {
	return record.SourceHash == run.DeliveryHash && record.ID == workflowSliceReceiptID(run) && record.ProjectID == run.ProjectID && record.Namespace == run.Namespace
}

func workflowReceiptMatchesProof(record WorkflowRecord, proof WorkflowSliceDeliveryProof, ref string) bool {
	if record.Receipt == nil || record.Revision != 1 {
		return false
	}

	receipt := record.Receipt
	if proof.NativeReceiptRef != nil {
		return workflowNativeReceiptMatchesProof(record, proof)
	}

	return workflowLegacyReceiptMatchesProof(*receipt, proof, ref)
}

func workflowLegacyReceiptMatchesProof(receipt WorkflowReceipt, proof WorkflowSliceDeliveryProof, ref string) bool {
	if receipt.Unit != proof.Unit || receipt.CandidateID != proof.CandidateID || receipt.SubmissionID != proof.SubmissionID || receipt.VerificationRef != ref || !receipt.ConfirmedAt.Equal(proof.ConfirmedAt) {
		return false
	}

	return workflowReceiptGitMatches(receipt, proof.Git) && receipt.PullRequest == proof.PullRequest.URL && slices.Contains(receipt.ValidationRefs, proof.CandidateRef)
}

func workflowNativeReceiptMatchesProof(record WorkflowRecord, proof WorkflowSliceDeliveryProof) bool {
	receipt := record.Receipt
	if receipt.Native == nil || *proof.NativeReceiptRef != (workflowstate.EvidenceRef{Path: record.Path, Hash: record.SourceHash}) || receipt.Unit != proof.Unit || receipt.CandidateID != proof.CandidateID || receipt.SubmissionID != proof.SubmissionID {
		return false
	}

	return receipt.ConfirmedAt.Equal(proof.ConfirmedAt) && workflowReceiptGitMatches(*receipt, proof.Git) && receipt.PullRequest == proof.PullRequest.URL
}

func workflowReceiptGitMatches(receipt WorkflowReceipt, git WorkflowIntegrationProof) bool {
	if receipt.Repository != git.Repository || receipt.Method != git.Method || receipt.SourceCommit != git.SourceCommit || receipt.BaseCommit != git.BaseCommit {
		return false
	}

	return receipt.PreparedTree == git.PreparedTree && receipt.VerifiedTree == git.VerifiedTree && receipt.HistoryProof == git.HistoryHash && receipt.ResultCommit == git.ResultCommit && receipt.ObservedTarget == git.ObservedTarget
}

func validateWorkflowTaskDelivery(task WorkflowTask) error {
	if task.Delivery == nil {
		return nil
	}

	delivery := task.Delivery
	if task.Phase != "integrated" || task.DeliveryRef == "" || delivery.PullRequest == "" || !validWorkflowDeliveryBranch(delivery.Branch) || !slices.Contains([]string{"merge", "rebase", "squash"}, delivery.Method) || !workflowCommitID(delivery.ResultCommit) || delivery.ConfirmedAt.IsZero() {
		return fmt.Errorf("invalid workflow task delivery projection")
	}

	return nil
}

func validWorkflowDeliveryBranch(branch string) bool {
	return strings.HasPrefix(branch, "refs/heads/") && validWorkflowPRBranch(strings.TrimPrefix(branch, "refs/heads/"))
}
