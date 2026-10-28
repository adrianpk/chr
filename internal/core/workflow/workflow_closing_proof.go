package workflow

import (
	"context"
	"encoding/json"
	"fmt"

	"forge.adrianpk.com/chr/internal/core/contenthash"
	workflowintegrate "forge.adrianpk.com/chr/internal/core/workflow/integrate"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

func (s *Store) observeWorkflowClosingTarget(ctx context.Context, request SetClosureRequest, ports FrontierPorts) (workflowstate.PreviewTarget, error) {
	query := TargetQuery{Repository: request.Repository, Branch: request.Branch, Directory: s.layout.ProjectDir()}

	target, err := ports.Target.ObserveWorkflowTarget(ctx, query)
	if err != nil {
		return target, err
	}

	if target.Repository != request.Repository || !workflowstate.ValidCommitID(target.Commit) {
		return target, workflowstate.ErrConflict
	}

	actual, err := ports.Provider.ReadWorkflowBranch(ctx, request.Repository, request.Branch)
	if err != nil || actual != target.Commit {
		return target, errorsOrConflict(err)
	}

	return target, s.checkWorkflowTarget(ctx, target)
}

func (s *Store) workflowClosingSliceProof(ctx context.Context, input workflowstate.ClosingInput, descriptor workflowstate.WorkflowSlice, ports FrontierPorts) (workflowstate.ClosingSlice, error) {
	if descriptor.Phase != "integrated" || descriptor.DeliveryRef == "" {
		return workflowstate.ClosingSlice{}, fmt.Errorf("%w: all actual slices must integrate", workflowstate.ErrConflict)
	}

	receipt, err := s.readRecord(descriptor.DeliveryRef)
	if err != nil {
		return workflowstate.ClosingSlice{}, err
	}

	proof, hash, err := s.workflowClosingDeliveryProof(ctx, receipt)
	if err != nil {
		return workflowstate.ClosingSlice{}, err
	}

	if err := s.workflowClosingSliceMembers(ctx, descriptor, receipt, proof); err != nil {
		return workflowstate.ClosingSlice{}, err
	}

	fresh, err := s.observeWorkflowClosingSlice(ctx, input, receipt, proof, ports)
	if err != nil {
		return workflowstate.ClosingSlice{}, err
	}

	refs := append([]string{descriptor.ContractRef, receipt.Path, receipt.Receipt.VerificationRef}, receipt.Receipt.ValidationRefs...)

	refs = append(refs, workflowClosingNativeEvidence(receipt, proof)...)
	if err := s.workflowClosingEvidenceHashes(input.EvidenceHashes, refs); err != nil {
		return workflowstate.ClosingSlice{}, err
	}

	return workflowstate.ClosingSlice{ID: descriptor.ID, Order: descriptor.Order, Receipt: workflowstate.ClosingEvidence{Ref: receipt.Path, Hash: receipt.SourceHash}, Verification: workflowstate.ClosingEvidence{Ref: receipt.Receipt.VerificationRef, Hash: hash}, Proof: fresh}, nil
}

func (s *Store) workflowClosingDeliveryProof(ctx context.Context, receipt Record) (SliceDeliveryProof, string, error) {
	if receipt.Receipt == nil || receipt.ProjectID != s.projectID || receipt.Namespace != s.layout.Namespace() {
		return SliceDeliveryProof{}, "", workflowstate.ErrConflict
	}

	if receipt.Receipt.Native != nil {
		closingProof, err := s.workflowClosingNativeDeliveryProof(ctx, receipt)

		return closingProof.Proof, closingProof.VerificationHash, err
	}

	data, err := s.workflowClosingRead(receipt.Receipt.VerificationRef)
	if err != nil {
		return SliceDeliveryProof{}, "", err
	}

	var proof SliceDeliveryProof
	if err := json.Unmarshal(data, &proof); err != nil {
		return proof, "", err
	}

	if proof.Schema != "chr-slice-delivery-proof/v1" || !workflowReceiptMatchesProof(receipt, proof, receipt.Receipt.VerificationRef) {
		return proof, "", workflowstate.ErrConflict
	}

	return proof, contenthash.Sum(data), nil
}

func (s *Store) observeWorkflowClosingSlice(ctx context.Context, input workflowstate.ClosingInput, receipt Record, retained SliceDeliveryProof, ports FrontierPorts) (workflowstate.IntegrationProof, error) {
	if receipt.Receipt.Native != nil {
		return s.observeWorkflowClosingNativeSlice(ctx, input, receipt, ports)
	}

	return s.observeWorkflowClosingLegacySlice(ctx, input, receipt, retained, ports)
}

func (s *Store) observeWorkflowClosingLegacySlice(ctx context.Context, input workflowstate.ClosingInput, receipt Record, retained SliceDeliveryProof, ports FrontierPorts) (workflowstate.IntegrationProof, error) {
	pr, err := ports.Provider.ReadWorkflowPullRequest(ctx, receipt.Receipt.PullRequest)
	if err != nil {
		return workflowstate.IntegrationProof{}, err
	}

	if !workflowClosingPRMatches(input, receipt, retained, pr) {
		return workflowstate.IntegrationProof{}, workflowstate.ErrConflict
	}

	query := retained.Git.IntegrationQuery

	query.SourceDirectory = s.layout.ProjectDir()
	if query.Repository != input.Target.Repository || query.TargetBranch != input.Branch {
		return workflowstate.IntegrationProof{}, workflowstate.ErrConflict
	}

	proof, err := ports.Observer.VerifyWorkflowIntegration(ctx, query)
	if err != nil {
		return proof, err
	}

	if err := workflowintegrate.CheckProof(query, proof, workflowIntegrationRepairs{}); err != nil {
		return proof, err
	}

	if proof.ObservedTarget != input.Target.Commit || proof.HistoryHash != receipt.Receipt.HistoryProof || proof.VerifiedTree != receipt.Receipt.VerifiedTree {
		return proof, workflowstate.ErrConflict
	}

	return proof, nil
}

func (s *Store) workflowClosingSliceMembers(ctx context.Context, descriptor workflowstate.WorkflowSlice, receipt Record, proof SliceDeliveryProof) error {
	snapshot, err := s.readSnapshot(ctx)
	if err != nil {
		return err
	}

	if descriptor.TaskInputHash != TaskInputHash(snapshot.Records, descriptor.ID) || receipt.Receipt.Unit != (workflowstate.Ref{Kind: "slice", ID: descriptor.ID}) || descriptor.PullRequest != receipt.Receipt.PullRequest {
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

func workflowClosingPRMatches(input workflowstate.ClosingInput, receipt Record, retained SliceDeliveryProof, pr workflowstate.PullRequestSnapshot) bool {
	return pr.Merged && pr.URL == receipt.Receipt.PullRequest && pr.Repository == input.Target.Repository && pr.HeadRepository == pr.Repository && pr.BaseBranch == input.Branch && pr.ResultCommit == receipt.Receipt.ResultCommit && pr.HeadCommit == receipt.Receipt.SourceCommit && pr.HeadBranch == retained.PullRequest.HeadBranch && pr.BaseCommit == input.Target.Commit
}
