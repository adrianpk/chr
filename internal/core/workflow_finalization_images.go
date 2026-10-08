package core

import (
	"context"
)

func (s *WorkflowStore) commitWorkflowSliceFinalization(ctx context.Context, run *workflowInvocationFile, observation workflowSliceObservation, receipt WorkflowRecord) error {
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

	run.DeliveryRef, run.DeliveryHash = receipt.Path, workflowHash(files[0].After)
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

	if err := snapshot.Validate(); err != nil {
		return err
	}

	run.Status, run.Outcome = "waiting_integration", "slice_integrated"
	if run.Cursor == len(run.Members) {
		run.Status = "completed"
	}

	return s.commitWorkflowPullRequestFiles(run, append(files, changes...), workflowFileChange{})
}

func (s *WorkflowStore) workflowFinalizationTaskImages(run *workflowInvocationFile, snapshot *WorkflowSnapshot, tasks []WorkflowRecord, receipt WorkflowRecord) ([]workflowFileChange, error) {
	files := []workflowFileChange{}

	for _, record := range tasks {
		record.Task.Phase, record.Task.Status, record.Task.DeliveryRef = "integrated", TicketStatusDone, receipt.Path
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

func workflowTaskDelivery(branch string, receipt WorkflowReceipt) *WorkflowTaskDelivery {
	return &WorkflowTaskDelivery{PullRequest: receipt.PullRequest, Branch: branch, Method: receipt.Method, ResultCommit: receipt.ResultCommit, ConfirmedAt: receipt.ConfirmedAt}
}

func (s *WorkflowStore) workflowFinalizationSetImages(run *workflowInvocationFile, snapshot *WorkflowSnapshot, set WorkflowRecord) ([]workflowFileChange, error) {
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

func replaceWorkflowFinalizationRecord(snapshot *WorkflowSnapshot, record WorkflowRecord) {
	for i := range snapshot.Records {
		if snapshot.Records[i].Path == record.Path {
			snapshot.Records[i] = record

			return
		}
	}
}

func (s *WorkflowStore) checkWorkflowFinalizationImages(ctx context.Context, run workflowInvocationFile, receipt WorkflowRecord) error {
	for _, member := range run.Members {
		if member.Admission.SliceID != receipt.Receipt.Unit.ID {
			continue
		}

		record, err := s.readRecord(member.Admission.Path)
		if err != nil || !workflowFinalizedTaskMatches(record, run.Lane.Branch, receipt) {
			return ErrWorkflowConflict
		}
	}

	selection, err := s.readWorkflowSelection(ctx)
	if err != nil {
		return err
	}

	unit := selection.slices[receipt.Receipt.Unit.ID]
	if unit.Phase != "integrated" || unit.DeliveryRef != receipt.Path || unit.PullRequest != receipt.Receipt.PullRequest {
		return ErrWorkflowConflict
	}

	return nil
}

func workflowFinalizedTaskMatches(record WorkflowRecord, branch string, receipt WorkflowRecord) bool {
	return record.Task != nil && record.Task.Phase == "integrated" && record.Task.Status == TicketStatusDone && record.Task.DeliveryRef == receipt.Path && workflowTaskDeliveryMatches(record.Task.Delivery, branch, *receipt.Receipt)
}

func workflowTaskDeliveryMatches(delivery *WorkflowTaskDelivery, branch string, receipt WorkflowReceipt) bool {
	return delivery != nil && delivery.PullRequest == receipt.PullRequest && delivery.Branch == branch && delivery.Method == receipt.Method && delivery.ResultCommit == receipt.ResultCommit && delivery.ConfirmedAt.Equal(receipt.ConfirmedAt)
}
