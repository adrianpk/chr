package core

import (
	"context"
	"encoding/json"
	"fmt"

	"forge.adrianpk.com/chr/internal/core/contenthash"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

func (s *WorkflowStore) observeWorkflowClosingTarget(ctx context.Context, request WorkflowSetClosureRequest, ports WorkflowFrontierPorts) (WorkflowPreviewTarget, error) {
	query := WorkflowTargetQuery{Repository: request.Repository, Branch: request.Branch, Directory: s.layout.ProjectDir()}

	target, err := ports.Target.ObserveWorkflowTarget(ctx, query)
	if err != nil {
		return target, err
	}

	if target.Repository != request.Repository || !workflowCommitID(target.Commit) {
		return target, workflowstate.ErrConflict
	}

	actual, err := ports.Provider.ReadWorkflowBranch(ctx, request.Repository, request.Branch)
	if err != nil || actual != target.Commit {
		return target, errorsOrConflict(err)
	}

	return target, s.checkWorkflowTarget(ctx, target)
}

func (s *WorkflowStore) workflowClosingSliceProof(ctx context.Context, input WorkflowClosingInput, descriptor WorkflowSlice, ports WorkflowFrontierPorts) (WorkflowClosingSlice, error) {
	if descriptor.Phase != "integrated" || descriptor.DeliveryRef == "" {
		return WorkflowClosingSlice{}, fmt.Errorf("%w: all actual slices must integrate", workflowstate.ErrConflict)
	}

	receipt, err := s.readRecord(descriptor.DeliveryRef)
	if err != nil {
		return WorkflowClosingSlice{}, err
	}

	proof, hash, err := s.workflowClosingDeliveryProof(ctx, receipt)
	if err != nil {
		return WorkflowClosingSlice{}, err
	}

	if err := s.workflowClosingSliceMembers(ctx, descriptor, receipt, proof); err != nil {
		return WorkflowClosingSlice{}, err
	}

	fresh, err := s.observeWorkflowClosingSlice(ctx, input, receipt, proof, ports)
	if err != nil {
		return WorkflowClosingSlice{}, err
	}

	refs := append([]string{descriptor.ContractRef, receipt.Path, receipt.Receipt.VerificationRef}, receipt.Receipt.ValidationRefs...)

	refs = append(refs, workflowClosingNativeEvidence(receipt, proof)...)
	if err := s.workflowClosingEvidenceHashes(input.EvidenceHashes, refs); err != nil {
		return WorkflowClosingSlice{}, err
	}

	return WorkflowClosingSlice{ID: descriptor.ID, Order: descriptor.Order, Receipt: WorkflowClosingEvidence{Ref: receipt.Path, Hash: receipt.SourceHash}, Verification: WorkflowClosingEvidence{Ref: receipt.Receipt.VerificationRef, Hash: hash}, Proof: fresh}, nil
}

func (s *WorkflowStore) workflowClosingDeliveryProof(ctx context.Context, receipt WorkflowRecord) (WorkflowSliceDeliveryProof, string, error) {
	if receipt.Receipt == nil || receipt.ProjectID != s.projectID || receipt.Namespace != s.layout.Namespace() {
		return WorkflowSliceDeliveryProof{}, "", workflowstate.ErrConflict
	}

	if receipt.Receipt.Native != nil {
		return s.workflowClosingNativeDeliveryProof(ctx, receipt)
	}

	data, err := s.workflowClosingRead(receipt.Receipt.VerificationRef)
	if err != nil {
		return WorkflowSliceDeliveryProof{}, "", err
	}

	var proof WorkflowSliceDeliveryProof
	if err := json.Unmarshal(data, &proof); err != nil {
		return proof, "", err
	}

	if proof.Schema != "chr-slice-delivery-proof/v1" || !workflowReceiptMatchesProof(receipt, proof, receipt.Receipt.VerificationRef) {
		return proof, "", workflowstate.ErrConflict
	}

	return proof, contenthash.Sum(data), nil
}

func (s *WorkflowStore) observeWorkflowClosingSlice(ctx context.Context, input WorkflowClosingInput, receipt WorkflowRecord, retained WorkflowSliceDeliveryProof, ports WorkflowFrontierPorts) (WorkflowIntegrationProof, error) {
	if receipt.Receipt.Native != nil {
		return s.observeWorkflowClosingNativeSlice(ctx, input, receipt, ports)
	}

	return s.observeWorkflowClosingLegacySlice(ctx, input, receipt, retained, ports)
}

func (s *WorkflowStore) observeWorkflowClosingLegacySlice(ctx context.Context, input WorkflowClosingInput, receipt WorkflowRecord, retained WorkflowSliceDeliveryProof, ports WorkflowFrontierPorts) (WorkflowIntegrationProof, error) {
	pr, err := ports.Provider.ReadWorkflowPullRequest(ctx, receipt.Receipt.PullRequest)
	if err != nil {
		return WorkflowIntegrationProof{}, err
	}

	if !workflowClosingPRMatches(input, receipt, retained, pr) {
		return WorkflowIntegrationProof{}, workflowstate.ErrConflict
	}

	query := retained.Git.WorkflowIntegrationQuery

	query.SourceDirectory = s.layout.ProjectDir()
	if query.Repository != input.Target.Repository || query.TargetBranch != input.Branch {
		return WorkflowIntegrationProof{}, workflowstate.ErrConflict
	}

	proof, err := ports.Observer.VerifyWorkflowIntegration(ctx, query)
	if err != nil {
		return proof, err
	}

	if err := validateWorkflowIntegrationProof(query, proof); err != nil {
		return proof, err
	}

	if proof.ObservedTarget != input.Target.Commit || proof.HistoryHash != receipt.Receipt.HistoryProof || proof.VerifiedTree != receipt.Receipt.VerifiedTree {
		return proof, workflowstate.ErrConflict
	}

	return proof, nil
}

func (s *WorkflowStore) workflowClosingSliceMembers(ctx context.Context, descriptor WorkflowSlice, receipt WorkflowRecord, proof WorkflowSliceDeliveryProof) error {
	snapshot, err := s.readSnapshot(ctx)
	if err != nil {
		return err
	}

	if descriptor.TaskInputHash != WorkflowTaskInputHash(snapshot.Records, descriptor.ID) || receipt.Receipt.Unit != (workflowstate.Ref{Kind: "slice", ID: descriptor.ID}) || descriptor.PullRequest != receipt.Receipt.PullRequest {
		return workflowstate.ErrConflict
	}

	count := 0

	for _, record := range snapshot.Records {
		if record.Task == nil || record.Task.SliceID != descriptor.ID {
			continue
		}

		count++

		if !workflowFinalizedTaskMatches(record, "refs/heads/"+proof.PullRequest.HeadBranch, receipt) {
			return workflowstate.ErrConflict
		}
	}

	if count == 0 {
		return fmt.Errorf("empty slices cannot close")
	}

	_, err = s.readWorkflowSliceContract(descriptor)

	return err
}

func workflowClosingPRMatches(input WorkflowClosingInput, receipt WorkflowRecord, retained WorkflowSliceDeliveryProof, pr WorkflowPullRequestSnapshot) bool {
	return pr.Merged && pr.URL == receipt.Receipt.PullRequest && pr.Repository == input.Target.Repository && pr.HeadRepository == pr.Repository && pr.BaseBranch == input.Branch && pr.ResultCommit == receipt.Receipt.ResultCommit && pr.HeadCommit == receipt.Receipt.SourceCommit && pr.HeadBranch == retained.PullRequest.HeadBranch && pr.BaseCommit == input.Target.Commit
}
