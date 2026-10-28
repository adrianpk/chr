package workflow

import (
	"context"
	"os"
	"strconv"
	"time"

	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

func workflowWorkerSlotLabel(file *os.File) string {
	return strconv.Itoa(workflowWorkerSlotNumber(file.Name()))
}

func (s *Store) persistWorkflowMergerClaim(ctx context.Context, run workflowInvocationFile, record Record, claim *Claim, budget IntegrationBudget, input workflowstate.MergerInput, harness MergerHarness, lane workflowstate.Lane, root string) (*Claim, workflowPreparedTask, error) {
	reservation, err := workflowMergerReservation(&run, budget, &input, claim)
	if err != nil {
		return nil, workflowPreparedTask{}, err
	}

	if err := chargeWorkflowIntegrationCounter(&budget, reservation); err != nil {
		return nil, workflowPreparedTask{}, err
	}

	budget.Reservations = append(budget.Reservations, reservation)

	input, err = SealWorkflowMergerInput(input, time.Now().UTC())
	if err != nil {
		return nil, workflowPreparedTask{}, err
	}

	taskInput, err := s.workflowTaskInput(ctx, run, record, harness.Command)
	if err != nil {
		return nil, workflowPreparedTask{}, err
	}

	files, err := s.workflowClaimRecordChanges(&run, &record)
	if err != nil {
		return nil, workflowPreparedTask{}, err
	}

	taskInput.SourceHash, taskInput.Lane, taskInput.Merger = record.SourceHash, lane, &input
	taskInput.MergerHarnessRef = &harness.Certificate

	prepared, file, err := encodeWorkflowTaskInput(&run, taskInput)
	if err != nil {
		return nil, prepared, err
	}

	budgetFile, err := s.workflowIntegrationBudgetChange(ctx, &run, &budget)
	if err != nil {
		return nil, prepared, err
	}

	prepared.environment, prepared.directory = workflowMergerEnvironment(root), root
	err = s.commitWorkflowPullRequestFiles(&run, append(files, budgetFile), file)

	return claim, prepared, err
}
