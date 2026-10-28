package workflow

import (
	"context"

	"forge.adrianpk.com/chr/internal/core/contenthash"
	ticketcore "forge.adrianpk.com/chr/internal/core/ticket"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

func (s *Store) commitWorkflowSliceFinalization(ctx context.Context, run *workflowInvocationFile, observation workflowSliceObservation, receipt workflowstate.Record) error {
	snapshot, err := s.readSnapshot(ctx)
	if err != nil {
		return err
	}

	if err := s.guardWrite(&receipt, 0); err != nil {
		return err
	}

	files, err := s.encodeChanges(receipt)
	if err != nil {
		return err
	}

	run.DeliveryRef, run.DeliveryHash = receipt.Path, contenthash.Sum(files[0].After)
	snapshot.Records = append(snapshot.Records, receipt)

	changes, err := s.workflowFinalizationTaskImages(run, &snapshot, observation.tasks, receipt)
	if err != nil {
		return err
	}

	files = append(files, changes...)

	changes, err = s.workflowFinalizationSetImages(run, &snapshot, observation.set)
	if err != nil {
		return err
	}

	if err := validateWorkflowSnapshot(snapshot); err != nil {
		return err
	}

	run.ConfirmedCursor = run.Cursor

	outcome := "slice_integrated"
	if run.Lane.Unit.Kind == "ticket" {
		outcome = "ticket_integrated"
	}

	workflowRetainDeliveryStatus(run, outcome, run.Cursor == len(run.Members) && (run.Selection.Selector.Extent == "slice" || run.Selection.Selector.Extent == "ticket"))

	return s.commitWorkflowPullRequestFiles(run, append(files, changes...), workflowstate.FileChange{})
}

func (s *Store) workflowFinalizationTaskImages(run *workflowInvocationFile, snapshot *workflowstate.Snapshot, tasks []workflowstate.Record, receipt workflowstate.Record) ([]workflowstate.FileChange, error) {
	files := []workflowstate.FileChange{}

	for _, record := range tasks {
		record.Task.Phase, record.Task.Status, record.Task.DeliveryRef = "integrated", string(ticketcore.StatusDone), receipt.Path
		record.Task.Delivery = workflowTaskDelivery(run.Lane.Branch, *receipt.Receipt)

		changes, err := s.workflowClaimRecordChanges(run, &record)
		if err != nil {
			return nil, err
		}

		files = append(files, changes...)

		replaceWorkflowFinalizationRecord(snapshot, record)
	}

	return files, nil
}

func workflowTaskDelivery(branch string, receipt workflowstate.Receipt) *workflowstate.TaskDelivery {
	return &workflowstate.TaskDelivery{PullRequest: receipt.PullRequest, Branch: branch, Method: receipt.Method, ResultCommit: receipt.ResultCommit, ConfirmedAt: receipt.ConfirmedAt}
}

func (s *Store) workflowFinalizationSetImages(run *workflowInvocationFile, snapshot *workflowstate.Snapshot, set workflowstate.Record) ([]workflowstate.FileChange, error) {
	if set.Set == nil {
		return nil, nil
	}

	for i := range set.Set.Slices {
		descriptor := &set.Set.Slices[i]
		if descriptor.ID == run.Lane.Unit.ID {
			descriptor.Phase, descriptor.DeliveryRef = "integrated", run.DeliveryRef
		}
	}

	files, err := s.workflowSetRecordChanges(run, &set)
	if err != nil {
		return nil, err
	}

	replaceWorkflowFinalizationRecord(snapshot, set)

	return files, nil
}

func replaceWorkflowFinalizationRecord(snapshot *workflowstate.Snapshot, record workflowstate.Record) {
	for i := range snapshot.Records {
		if snapshot.Records[i].Path == record.Path {
			snapshot.Records[i] = record

			return
		}
	}
}

func (s *Store) checkWorkflowFinalizationImages(ctx context.Context, run workflowInvocationFile, receipt workflowstate.Record) error {
	for _, member := range run.Members {
		if !workflowMemberOwnsUnit(member, receipt.Receipt.Unit) {
			continue
		}

		record, err := s.readRecord(member.Admission.Path)
		if err != nil || !workflowstate.FinalizedTaskMatches(record, run.Lane.Branch, receipt) {
			return workflowstate.ErrConflict
		}
	}

	if receipt.Receipt.Unit.Kind == "ticket" {
		return nil
	}

	selection, err := s.readWorkflowSelection(ctx)
	if err != nil {
		return err
	}

	unit := selection.slices[receipt.Receipt.Unit.ID]
	if unit.Phase != "integrated" || unit.DeliveryRef != receipt.Path || unit.PullRequest != receipt.Receipt.PullRequest {
		return workflowstate.ErrConflict
	}

	return nil
}
