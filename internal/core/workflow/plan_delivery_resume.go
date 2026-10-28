package workflow

import (
	"context"
	"encoding/json"
	"fmt"
	"path/filepath"

	generationcore "forge.adrianpk.com/chr/internal/core/authoring/generation"
	authoringstate "forge.adrianpk.com/chr/internal/core/authoring/state"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

// MaterializationLookup retains presence independently of a failed
// contract/receipt read, so a matching set cannot trigger a second allocation.
type MaterializationLookup struct {
	Result MaterializationResult
	Found  bool
}

// ResumePlanDelivery verifies an existing materialization from its original
// contracts and receipt without calling a provider or reallocating tickets.
func (s *Store) ResumePlanDelivery(ctx context.Context, input generationcore.GeneratedWorkflowPlan) (lookup MaterializationLookup, err error) {
	err = s.withRepositoryLock(ctx, func() error { return s.prepareWorkflowRead(ctx) }, func() error {
		snapshot, readErr := s.readSnapshot(ctx)
		if readErr != nil {
			return readErr
		}

		lookup, readErr = s.resumePlanDelivery(snapshot, input)

		return readErr
	})

	return lookup, err
}

func (s *Store) resumePlanDelivery(snapshot Snapshot, input generationcore.GeneratedWorkflowPlan) (MaterializationLookup, error) {
	for _, record := range snapshot.Records {
		if record.Set == nil || record.Metadata["materialization_id"] == nil {
			continue
		}

		var id string
		if err := json.Unmarshal(record.Metadata["materialization_id"], &id); err != nil {
			return MaterializationLookup{}, err
		}

		if id != input.MaterializationID {
			continue
		}

		plan, err := s.readPlanDeliveryContracts(input, record)
		if err != nil {
			return MaterializationLookup{Found: true}, err
		}

		result, err := s.readMaterializationReceipt(record, plan, snapshot)

		return MaterializationLookup{Result: result, Found: true}, err
	}

	return MaterializationLookup{}, nil
}

func (s *Store) readPlanDeliveryContracts(input generationcore.GeneratedWorkflowPlan, set Record) (generationcore.GeneratedWorkflowPlan, error) {
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
			Validation workflowstate.Validation
		}
		if err := authoringstate.Decode(data, &contract); err != nil {
			return input, err
		}

		slice.Contract, slice.Validation = contract.Contract, contract.Validation
		for j := range slice.Tasks {
			next++

			data, err := s.readWorkflowTrackerFile(filepath.ToSlash(filepath.Join("plan", set.ID, "extension", "chr", fmt.Sprintf("task-%d.json", next))))
			if err != nil {
				return input, err
			}

			if err := authoringstate.Decode(data, &slice.Tasks[j]); err != nil {
				return input, err
			}
		}
	}

	return input, generationcore.ValidateGeneratedWorkflowPlan(input)
}
