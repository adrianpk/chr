package closeout

import (
	"context"
	"fmt"

	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

func (p Proofs) members(ctx context.Context, descriptor workflowstate.WorkflowSlice, receipt workflowstate.Record, proof workflowstate.SliceDeliveryProof) error {
	snapshot, err := p.Source.Snapshot(ctx)
	if err != nil {
		return err
	}

	if descriptor.TaskInputHash != workflowstate.TaskInputHash(snapshot.Records, descriptor.ID) || receipt.Receipt.Unit != (workflowstate.Ref{Kind: "slice", ID: descriptor.ID}) || descriptor.PullRequest != receipt.Receipt.PullRequest {
		return workflowstate.ErrConflict
	}

	count := 0

	for _, record := range snapshot.Records {
		if record.Task == nil || record.Task.SliceID != descriptor.ID {
			continue
		}

		count++

		if !workflowstate.FinalizedTaskMatches(record, "refs/heads/"+proof.PullRequest.HeadBranch, receipt) {
			return workflowstate.ErrConflict
		}
	}

	if count == 0 {
		return fmt.Errorf("empty slices cannot close")
	}

	err = p.Source.Contract(descriptor)

	return err
}

// CheckReceipts requires every actual member and slice to retain matching delivery.
func (p Proofs) CheckReceipts(ctx context.Context, set workflowstate.Record) error {
	snapshot, err := p.Source.Snapshot(ctx)
	if err != nil {
		return err
	}

	if _, err := Members(snapshot.Records, set.ID); err != nil {
		return err
	}

	for _, descriptor := range set.Set.Slices {
		if descriptor.Phase != "integrated" || descriptor.DeliveryRef == "" {
			return workflowstate.ErrConflict
		}

		receipt, err := p.Source.Record(descriptor.DeliveryRef)
		if err != nil {
			return err
		}

		observed, err := p.delivery(ctx, receipt)
		if err != nil {
			return err
		}

		if err := p.members(ctx, descriptor, receipt, observed.Proof); err != nil {
			return err
		}
	}

	return nil
}
