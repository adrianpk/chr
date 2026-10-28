package closeout

import (
	"context"

	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

func (p Proofs) observe(ctx context.Context, input workflowstate.ClosingInput, receipt workflowstate.Record, retained workflowstate.SliceDeliveryProof) (workflowstate.IntegrationProof, error) {
	if receipt.Receipt.Native != nil {
		return p.observeNative(ctx, input, receipt)
	}

	return p.observeLegacy(ctx, input, receipt, retained)
}

func (p Proofs) observeLegacy(ctx context.Context, input workflowstate.ClosingInput, receipt workflowstate.Record, retained workflowstate.SliceDeliveryProof) (workflowstate.IntegrationProof, error) {
	pr, err := p.Source.PullRequest(ctx, receipt.Receipt.PullRequest)
	if err != nil {
		return workflowstate.IntegrationProof{}, err
	}

	if !pullRequestMatches(input, receipt, retained, pr) {
		return workflowstate.IntegrationProof{}, workflowstate.ErrConflict
	}

	query := retained.Git.IntegrationQuery

	query.SourceDirectory = p.Directory
	if query.Repository != input.Target.Repository || query.TargetBranch != input.Branch {
		return workflowstate.IntegrationProof{}, workflowstate.ErrConflict
	}

	proof, err := p.Source.Verify(ctx, query)
	if err != nil {
		return proof, err
	}

	if err := p.Source.CheckProof(query, proof); err != nil {
		return proof, err
	}

	if proof.ObservedTarget != input.Target.Commit || proof.HistoryHash != receipt.Receipt.HistoryProof || proof.VerifiedTree != receipt.Receipt.VerifiedTree {
		return proof, workflowstate.ErrConflict
	}

	return proof, nil
}

func pullRequestMatches(input workflowstate.ClosingInput, receipt workflowstate.Record, retained workflowstate.SliceDeliveryProof, pr workflowstate.PullRequestSnapshot) bool {
	return pr.Merged && pr.URL == receipt.Receipt.PullRequest && pr.Repository == input.Target.Repository && pr.HeadRepository == pr.Repository && pr.BaseBranch == input.Branch && pr.ResultCommit == receipt.Receipt.ResultCommit && pr.HeadCommit == receipt.Receipt.SourceCommit && pr.HeadBranch == retained.PullRequest.HeadBranch && pr.BaseCommit == input.Target.Commit
}

func (p Proofs) observeNative(ctx context.Context, input workflowstate.ClosingInput, receipt workflowstate.Record) (workflowstate.IntegrationProof, error) {
	observation, err := p.Source.NativeObservation(ctx, receipt)
	if err != nil {
		return workflowstate.IntegrationProof{}, err
	}

	if observation.State != workflowstate.NativeMerged || observation.Git == nil {
		return workflowstate.IntegrationProof{}, workflowstate.ErrObservationRequired
	}

	proof := *observation.Git
	if proof.Repository != input.Target.Repository || proof.TargetBranch != input.Branch || !workflowstate.ReceiptTargetProofMatches(proof, receipt, input.Target) {
		return proof, workflowstate.ErrConflict
	}

	return proof, nil
}

// CheckCanonical reobserves native receipts without accepting changed frozen images.
func (p Proofs) CheckCanonical(ctx context.Context, input workflowstate.ClosingInput) error {
	for _, slice := range input.Slices {
		receipt, err := p.Source.Record(slice.Receipt.Ref)
		if err != nil {
			return err
		}

		if receipt.Receipt == nil || receipt.SourceHash != slice.Receipt.Hash {
			return workflowstate.ErrConflict
		}

		if receipt.Receipt.Native == nil {
			continue
		}

		if _, err := p.observeNative(ctx, input, receipt); err != nil {
			return err
		}
	}

	return nil
}
