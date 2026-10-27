package core

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"slices"
	"time"

	"forge.adrianpk.com/chr/internal/core/contenthash"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

func (s *WorkflowStore) readWorkflowIntegrationBudget(run workflowInvocationFile) (WorkflowIntegrationBudget, error) {
	var budget WorkflowIntegrationBudget

	binding := run.IntegrationBudget
	if binding == nil || binding.Ref.Path != workflowIntegrationBudgetPath(run.ID) {
		return budget, workflowstate.ErrConflict
	}

	data, err := s.readWorkflowReviewImage(binding.Ref.Path, binding.Ref.Hash)
	if err != nil {
		return budget, err
	}

	if err := json.Unmarshal(data, &budget); err != nil {
		return budget, err
	}

	if budget.InvocationID != run.ID || budget.Revision != binding.Revision {
		return budget, workflowstate.ErrConflict
	}

	if err := validateWorkflowIntegrationBudget(budget); err != nil {
		return budget, err
	}

	return budget, s.checkWorkflowIntegrationBudgetEvidence(run, budget)
}

func (s *WorkflowStore) workflowIntegrationUsage(ctx context.Context, run workflowInvocationFile) (WorkflowIntegrationUsage, error) {
	usage := WorkflowIntegrationUsage{ReviewCount: run.ReviewCount, PreparationCount: run.PreparationCount, ReviewRepairs: slices.Clone(run.ReviewRepairs)}
	for _, attempt := range run.Attempts {
		usage.Attempts = append(usage.Attempts, fmt.Sprintf("%s:%d", attempt.ID, attempt.Generation))
		if attempt.Purpose == "" {
			usage.WorkerCount++
		}
	}

	snapshot, err := s.readSnapshot(ctx)
	if err != nil {
		return usage, err
	}

	usage.ClosingAttempts, err = s.workflowIntegrationClosingUsage(snapshot, run.ID)

	return usage, err
}

func (s *WorkflowStore) workflowIntegrationClosingUsage(snapshot WorkflowSnapshot, id string) ([]string, error) {
	attempts := []string{}

	for _, record := range snapshot.Records {
		if record.Set == nil || record.Set.Closing == nil {
			continue
		}

		for i := range record.Set.Closing.Attempts {
			attempt, err := s.readWorkflowClosingAttempt(record, i)
			if err != nil {
				return nil, err
			}

			if attempt.BudgetID == id {
				attempts = append(attempts, record.ID+":"+fmt.Sprint(attempt.Generation))
			}
		}
	}

	return attempts, nil
}

func (s *WorkflowStore) workflowIntegrationBudgetAdmission(ctx context.Context, run workflowInvocationFile, entry WorkflowIntegrationEntry, tokenLimit *int64) (WorkflowIntegrationBudget, error) {
	if run.Policy.AgentTokenBudget > 0 {
		if tokenLimit != nil && *tokenLimit != run.Policy.AgentTokenBudget {
			return WorkflowIntegrationBudget{}, workflowstate.ErrConflict
		}

		limit := run.Policy.AgentTokenBudget
		tokenLimit = &limit
	}

	if run.IntegrationBudget != nil {
		return s.extendWorkflowIntegrationBudget(run, entry, tokenLimit)
	}

	if tokenLimit != nil {
		if err := s.checkWorkflowAgentCeilingAdmission(run); err != nil {
			return WorkflowIntegrationBudget{}, err
		}
	}

	counters, err := s.workflowIntegrationLegacySubmissions(ctx, run)
	if err != nil {
		return WorkflowIntegrationBudget{}, err
	}

	if !slices.ContainsFunc(counters, func(c WorkflowIntegrationCounter) bool { return c.Unit == entry.Unit }) {
		counters = append(counters, WorkflowIntegrationCounter{Unit: entry.Unit})
	}

	limits := workflowIntegrationAdmissionLimits(run.Policy)

	return WorkflowIntegrationBudget{SchemaVersion: 1, InvocationID: run.ID, TokenLimit: tokenLimit, AdmissionTokenLimit: tokenLimit, AdmissionLimits: limits, LimitsSnapshot: limits, UnitCounters: counters}, nil
}

func (s *WorkflowStore) extendWorkflowIntegrationBudget(run workflowInvocationFile, entry WorkflowIntegrationEntry, tokenLimit *int64) (WorkflowIntegrationBudget, error) {
	budget, err := s.readWorkflowIntegrationBudget(run)
	if err != nil {
		return budget, err
	}

	if tokenLimit != nil && (budget.AdmissionTokenLimit == nil || *budget.AdmissionTokenLimit != *tokenLimit) {
		return budget, workflowstate.ErrConflict
	}

	if !slices.ContainsFunc(budget.UnitCounters, func(c WorkflowIntegrationCounter) bool { return c.Unit == entry.Unit }) {
		budget.UnitCounters = append(budget.UnitCounters, WorkflowIntegrationCounter{Unit: entry.Unit})
	}

	return budget, nil
}

func (s *WorkflowStore) workflowIntegrationBudgetChange(ctx context.Context, run *workflowInvocationFile, budget *WorkflowIntegrationBudget) (workflowstate.FileChange, error) {
	if budget.Revision == math.MaxUint64 {
		return workflowstate.FileChange{}, workflowstate.ErrConflict
	}

	usage, err := s.workflowIntegrationUsage(ctx, *run)
	if err != nil {
		return workflowstate.FileChange{}, err
	}

	if err := checkWorkflowIntegrationUsage(budget.Usage, usage); err != nil {
		return workflowstate.FileChange{}, err
	}

	budget.Usage = usage
	budget.Revision++

	budget.TokensSettled, budget.TokensReserved, err = workflowIntegrationTokenTotals(budget.Reservations)
	if err != nil {
		return workflowstate.FileChange{}, err
	}

	if err := validateWorkflowIntegrationBudget(*budget); err != nil {
		return workflowstate.FileChange{}, err
	}

	data, err := json.MarshalIndent(budget, "", "  ")
	if err != nil || len(data) >= workflowstate.FileLimit {
		return workflowstate.FileChange{}, workflowstate.ErrConflict
	}

	path, before := workflowIntegrationBudgetPath(run.ID), ""
	if run.IntegrationBudget != nil {
		before = run.IntegrationBudget.Ref.Hash
	}

	file := workflowstate.FileChange{Path: path, Before: before, After: append(data, '\n')}
	run.IntegrationBudget = &WorkflowIntegrationBudgetBinding{Revision: budget.Revision, Ref: WorkflowEvidenceRef{Path: path, Hash: contenthash.Sum(file.After)}}

	return file, nil
}

func (s *WorkflowStore) commitWorkflowIntegrationBudget(ctx context.Context, run *workflowInvocationFile, budget *WorkflowIntegrationBudget, extra ...workflowstate.FileChange) error {
	file, err := s.workflowIntegrationBudgetChange(ctx, run, budget)
	if err != nil {
		return err
	}

	run.Revision++
	run.UpdatedAt = time.Now().UTC()

	invocation, err := workflowInvocationChange(*run)
	if err != nil {
		return err
	}

	return s.commitFiles(append([]workflowstate.FileChange{file, invocation}, extra...))
}

func (s *WorkflowStore) currentWorkflowIntegrationBudget(ctx context.Context, run workflowInvocationFile, guard WorkflowIntegrationBudgetGuard) (WorkflowIntegrationEntry, WorkflowIntegrationBudget, error) {
	if err := s.checkWorkflowIntegrationRetainedRecords(ctx, run); err != nil {
		return WorkflowIntegrationEntry{}, WorkflowIntegrationBudget{}, err
	}

	if run.Integration == nil || run.Integration.EntryID != guard.EntryID {
		return WorkflowIntegrationEntry{}, WorkflowIntegrationBudget{}, workflowstate.ErrConflict
	}

	entry, err := s.readWorkflowIntegrationEntry(run, *run.Integration)
	if err != nil {
		return entry, WorkflowIntegrationBudget{}, err
	}

	budget, err := s.readWorkflowIntegrationBudget(run)
	if err != nil {
		return entry, budget, err
	}

	usage, err := s.workflowIntegrationUsage(ctx, run)
	if err == nil {
		err = checkWorkflowIntegrationUsage(budget.Usage, usage)
	}

	if err != nil {
		return entry, budget, err
	}

	budget.Usage = usage

	counters, err := s.workflowIntegrationLegacySubmissions(ctx, run)
	if err == nil {
		err = checkWorkflowIntegrationLegacySubmissions(budget, counters)
	}

	return entry, budget, err
}

func (s *WorkflowStore) checkWorkflowIntegrationBudgetEvidence(run workflowInvocationFile, budget WorkflowIntegrationBudget) error {
	if run.Policy.AgentTokenBudget > 0 && (budget.AdmissionTokenLimit == nil || *budget.AdmissionTokenLimit != run.Policy.AgentTokenBudget) {
		return workflowstate.ErrConflict
	}

	if budget.AdmissionLimits != workflowIntegrationAdmissionLimits(run.Policy) {
		return workflowIntegrationConflict("admission limits differ from frozen execution policy")
	}

	for _, reservation := range budget.Reservations {
		if err := s.checkWorkflowBudgetReservationEvidence(run, reservation); err != nil {
			return err
		}
	}

	return nil
}

func (s *WorkflowStore) checkWorkflowBudgetReservationEvidence(run workflowInvocationFile, reservation WorkflowIntegrationReservation) error {
	if reservation.EntryID != workflowIntegrationEntryID(run, reservation.Unit) {
		return workflowstate.ErrConflict
	}

	if reservation.AttemptInput != nil {
		if err := s.checkWorkflowAgentReservationEvidence(run, reservation); err != nil {
			return err
		}
	}

	if reservation.State != "settled" {
		return nil
	}

	return s.checkWorkflowTokenSettlement(run.ID, reservation)
}
