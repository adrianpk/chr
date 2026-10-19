package core

import (
	"context"
	"encoding/json"
	"fmt"
	"path/filepath"
)

// ResumePlanDelivery verifies an existing materialization from its original
// contracts and receipt without calling a provider or reallocating tickets.
func (s *WorkflowStore) ResumePlanDelivery(ctx context.Context, input GeneratedWorkflowPlan) (result WorkflowMaterializationResult, found bool, err error) {
	err = s.withRepositoryLock(ctx, func() error { return s.prepareWorkflowRead(ctx) }, func() error {
		snapshot, readErr := s.readSnapshot(ctx)
		if readErr != nil {
			return readErr
		}

		result, found, readErr = s.resumePlanDelivery(snapshot, input)

		return readErr
	})

	return result, found, err
}

func (s *WorkflowStore) resumePlanDelivery(snapshot WorkflowSnapshot, input GeneratedWorkflowPlan) (WorkflowMaterializationResult, bool, error) {
	for _, record := range snapshot.Records {
		if record.Set == nil || record.Metadata["materialization_id"] == nil {
			continue
		}

		var id string
		if err := json.Unmarshal(record.Metadata["materialization_id"], &id); err != nil {
			return WorkflowMaterializationResult{}, false, err
		}

		if id != input.MaterializationID {
			continue
		}

		plan, err := s.readPlanDeliveryContracts(input, record)
		if err != nil {
			return WorkflowMaterializationResult{}, true, err
		}

		result, err := s.readMaterializationReceipt(record, plan, snapshot)

		return result, true, err
	}

	return WorkflowMaterializationResult{}, false, nil
}

func (s *WorkflowStore) readPlanDeliveryContracts(input GeneratedWorkflowPlan, set WorkflowRecord) (GeneratedWorkflowPlan, error) {
	input.ClosingValidation = set.Set.ClosingValidation
	input.UseCaseOrder = set.Set.UseCaseOrder
	next := 0

	for i := range input.Slices {
		slice := &input.Slices[i]
		reference := filepath.ToSlash(filepath.Join("plan", set.ID, "extension", "chr", "slice-"+slice.ID+".json"))

		data, err := s.readWorkflowTrackerFile(reference)
		if err != nil {
			return input, err
		}

		var contract struct {
			Contract   string
			Validation WorkflowValidation
		}
		if err := decodeAuthoringState(data, &contract); err != nil {
			return input, err
		}

		slice.Contract, slice.Validation = contract.Contract, contract.Validation
		for j := range slice.Tasks {
			next++

			data, err := s.readWorkflowTrackerFile(filepath.ToSlash(filepath.Join("plan", set.ID, "extension", "chr", fmt.Sprintf("task-%d.json", next))))
			if err != nil {
				return input, err
			}

			if err := decodeAuthoringState(data, &slice.Tasks[j]); err != nil {
				return input, err
			}
		}
	}

	return input, ValidateGeneratedWorkflowPlan(input)
}
