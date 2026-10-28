package workflow

import (
	"context"

	"forge.adrianpk.com/chr/internal/core/contenthash"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

func (s *Store) replaceWorkflowNativePreparation(ctx context.Context, run *workflowInvocationFile, entry IntegrationEntry, previous workflowstate.PreparationInput, source workflowstate.ReviewInput) error {
	next, err := s.workflowNativeReplacementEntry(ctx, *run, entry)
	if err != nil {
		return err
	}

	budget, err := s.readWorkflowIntegrationBudget(*run)
	if err != nil {
		return err
	}

	if previous.Git.BaseCommit != source.TargetCommit {
		if err := reserveWorkflowNativeDrift(&budget, next, source.TargetCommit); err != nil {
			return err
		}
	}

	budgetFile, err := s.workflowIntegrationBudgetChange(ctx, run, &budget)
	if err != nil {
		return err
	}

	queueFile, err := workflowIntegrationEntryChange(next, run.Integration.EntryRef.Hash)
	if err != nil {
		return err
	}

	updateWorkflowIntegrationBinding(run, workflowstate.IntegrationBinding{EntryID: next.ID, Unit: next.Unit, Revision: next.Revision, EntryRef: workflowstate.EvidenceRef{Path: queueFile.Path, Hash: contenthash.Sum(queueFile.After)}})
	staleWorkflowNativePreparation(run)
	projectWorkflowIntegrationPhase(run, next)

	run.Outcome = "integration_inputs_changed"

	if err := s.commitWorkflowPullRequestFiles(run, []workflowstate.FileChange{budgetFile, queueFile}, workflowstate.FileChange{}); err != nil {
		return err
	}

	*run, err = s.readWorkflowInvocation(run.ID)

	return err
}

func (s *Store) workflowNativeReplacementEntry(ctx context.Context, run workflowInvocationFile, entry IntegrationEntry) (IntegrationEntry, error) {
	request := workflowNativeReplacementTransition(entry, "queued")
	if err := s.checkWorkflowCandidateRepairTransition(ctx, run, entry, request); err != nil {
		return entry, err
	}

	updateWorkflowIntegrationEntry(&entry, request)

	for _, state := range []IntegrationState{IntegrationPreparing, IntegrationValidating} {
		request = workflowNativeReplacementTransition(entry, state)
		if err := guardWorkflowIntegrationTransition(entry, request); err != nil {
			return entry, err
		}

		if err := s.guardWorkflowPreparationTransition(ctx, run, request); err != nil {
			return entry, err
		}

		updateWorkflowIntegrationEntry(&entry, request)
	}

	entry.CandidateRef = nil

	return entry, validateWorkflowIntegrationEntry(entry)
}

func workflowNativeReplacementTransition(entry IntegrationEntry, state IntegrationState) IntegrationTransition {
	request := IntegrationTransition{EntryRevision: entry.Revision, ClaimGeneration: entry.ClaimGeneration, State: state}
	if entry.CandidateRef != nil {
		request.CandidateHash = entry.CandidateRef.Hash
	}

	return request
}

func reserveWorkflowNativeDrift(budget *IntegrationBudget, entry IntegrationEntry, base string) error {
	id := contenthash.Sum([]byte(entry.ID + ":drift:" + base))

	reservation := IntegrationReservation{ID: id, Kind: "drift", EntryID: entry.ID, Unit: entry.Unit, ClaimGeneration: entry.ClaimGeneration, AttemptID: id, BaseCommit: base, State: "reserved"}
	if workflowIntegrationBaseCharged(budget.Reservations, reservation) {
		return nil
	}

	if len(budget.Reservations) >= workflowstate.RecordLimit {
		return ErrWorkflowIntegrationBudget
	}

	if err := chargeWorkflowIntegrationCounter(budget, reservation); err != nil {
		return err
	}

	budget.Reservations = append(budget.Reservations, reservation)

	return nil
}

func staleWorkflowNativePreparation(run *workflowInvocationFile) {
	if run.Preparation != nil {
		run.Preparation.State = "stale"
	}

	if run.Review != nil {
		run.Review.State = "stale"
	}

	run.CandidateRepair = nil
}
