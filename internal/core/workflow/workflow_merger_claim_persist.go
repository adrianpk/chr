package workflow

import (
	"context"
	"os"
	"strconv"
	"time"

	workflowmerge "forge.adrianpk.com/chr/internal/core/workflow/merge"
)

func workflowWorkerSlotLabel(file *os.File) string {
	return strconv.Itoa(workflowWorkerSlotNumber(file.Name()))
}

func (s *Store) persistWorkflowMergerClaim(ctx context.Context, claim *Claim, binding workflowMergerClaimInput) (workflowClaimedMerger, error) {
	run, record := binding.Observation.Run, binding.Observation.Record
	budget, input, harness := binding.Budget, binding.Input, binding.Harness
	lane, root := binding.Workspace.Lane, binding.Workspace.Root

	reservation, err := workflowMergerReservation(&run, budget, &input, claim)
	if err != nil {
		return workflowClaimedMerger{}, err
	}

	if err := chargeWorkflowIntegrationCounter(&budget, reservation); err != nil {
		return workflowClaimedMerger{}, err
	}

	budget.Reservations = append(budget.Reservations, reservation)

	input, err = workflowmerge.SealInput(input, time.Now().UTC())
	if err != nil {
		return workflowClaimedMerger{}, err
	}

	taskInput, err := s.workflowTaskInput(ctx, run, record, harness.Command)
	if err != nil {
		return workflowClaimedMerger{}, err
	}

	files, err := s.workflowClaimRecordChanges(&run, &record)
	if err != nil {
		return workflowClaimedMerger{}, err
	}

	taskInput.SourceHash, taskInput.Lane, taskInput.Merger = record.SourceHash, lane, &input
	taskInput.MergerHarnessRef = &harness.Certificate

	encodedTaskInput, err := encodeWorkflowTaskInput(&run, taskInput)
	if err != nil {
		return workflowClaimedMerger{Prepared: encodedTaskInput.Prepared}, err
	}

	budgetFile, err := s.workflowIntegrationBudgetChange(ctx, &run, &budget)
	if err != nil {
		return workflowClaimedMerger{Prepared: encodedTaskInput.Prepared}, err
	}

	encodedTaskInput.Prepared.environment, encodedTaskInput.Prepared.directory = workflowMergerEnvironment(root), root
	err = s.commitWorkflowPullRequestFiles(&run, append(files, budgetFile), encodedTaskInput.Change)

	return workflowClaimedMerger{Claim: claim, Prepared: encodedTaskInput.Prepared}, err
}
