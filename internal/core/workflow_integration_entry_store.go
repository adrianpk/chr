package core

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"slices"
	"time"
)

func (s *WorkflowStore) readWorkflowIntegrationEntry(run workflowInvocationFile, binding WorkflowIntegrationBinding) (WorkflowIntegrationEntry, error) {
	var entry WorkflowIntegrationEntry

	data, err := s.readWorkflowReviewImage(binding.EntryRef.Path, binding.EntryRef.Hash)
	if err != nil {
		return entry, err
	}

	if err := json.Unmarshal(data, &entry); err != nil {
		return entry, err
	}

	if err := validateWorkflowIntegrationEntry(entry); err != nil {
		return entry, err
	}

	hash, err := workflowIntegrationScopeHash(run, entry.Unit)
	if err != nil || !workflowIntegrationEntryMatches(run, binding, entry) || hash != entry.FrozenScopeHash {
		return entry, workflowIntegrationConflict("differs from its invocation owner")
	}

	return entry, nil
}

func workflowIntegrationEntryMatches(run workflowInvocationFile, binding WorkflowIntegrationBinding, entry WorkflowIntegrationEntry) bool {
	repository, err := workflowCanonicalRepository(run.ExecutionSelection().Target.Repository)
	if err != nil {
		return false
	}

	expected := [5]string{binding.EntryID, run.ID, run.ProjectID, run.Namespace, repository}
	actual := [5]string{entry.ID, entry.InvocationID, entry.ProjectID, entry.Namespace, entry.Repository}

	return expected == actual && entry.Unit == binding.Unit && entry.Revision == binding.Revision && entry.ID == workflowIntegrationEntryID(run, entry.Unit) && slices.Contains(run.RequestKeys, entry.AdmissionKey)
}

func (s *WorkflowStore) repositoryWorkflowIntegrationEntries(runs []workflowInvocationFile, coordinator workflowCoordinator) ([]WorkflowIntegrationEntry, error) {
	entries := []WorkflowIntegrationEntry{}
	sequences := map[uint64]bool{}
	ids := map[string]bool{}

	for _, run := range runs {
		owner, err := s.workflowOwnerStore(run)
		if err != nil {
			return nil, err
		}

		for _, binding := range run.IntegrationEntries {
			entry, err := owner.readWorkflowIntegrationEntry(run, binding)
			if err != nil {
				return nil, err
			}

			if entry.CoordinatorID != coordinator.ID || sequences[entry.QueueSequence] || ids[entry.ID] || len(entries) >= WorkflowRecordLimit {
				return nil, workflowIntegrationConflict("has foreign, duplicate or excessive allocation")
			}

			sequences[entry.QueueSequence], ids[entry.ID] = true, true
			entries = append(entries, entry)
		}
	}

	return entries, nil
}

func (s *WorkflowStore) admitWorkflowIntegration(ctx context.Context, runs []workflowInvocationFile, request WorkflowIntegrationAdmission) (WorkflowIntegrationEntry, error) {
	coordinator, err := s.checkWorkflowCoordinator(ctx)
	if err != nil {
		return WorkflowIntegrationEntry{}, err
	}

	entries, err := s.repositoryWorkflowIntegrationEntries(runs, coordinator)
	if err != nil {
		return WorkflowIntegrationEntry{}, err
	}

	for _, run := range runs {
		if run.ID == request.InvocationID && run.Namespace == s.layout.Namespace() {
			owner, err := s.workflowOwnerStore(run)
			if err != nil {
				return WorkflowIntegrationEntry{}, err
			}

			return owner.admitWorkflowIntegrationUnit(ctx, run, runs, entries, request, coordinator)
		}
	}

	return WorkflowIntegrationEntry{}, workflowIntegrationConflict("invocation not found")
}

func (s *WorkflowStore) admitWorkflowIntegrationUnit(ctx context.Context, run workflowInvocationFile, runs []workflowInvocationFile, entries []WorkflowIntegrationEntry, request WorkflowIntegrationAdmission, coordinator workflowCoordinator) (WorkflowIntegrationEntry, error) {
	if !slices.Contains(run.RequestKeys, request.AdmissionKey) || !validWorkflowRequestKey(request.AdmissionKey) {
		return WorkflowIntegrationEntry{}, workflowIntegrationConflict("retry key is not attached to its invocation")
	}

	for _, entry := range entries {
		if entry.ID == workflowIntegrationEntryID(run, request.Unit) {
			if entry.TargetRef != request.TargetRef {
				return WorkflowIntegrationEntry{}, workflowIntegrationConflict("retry changes its canonical target")
			}

			return entry, nil
		}
	}

	if err := s.guardWorkflowIntegrationAdmission(ctx, run, request); err != nil {
		return WorkflowIntegrationEntry{}, err
	}

	sequence, err := s.nextWorkflowIntegrationSequence(ctx, runs, entries)
	if err != nil {
		return WorkflowIntegrationEntry{}, err
	}

	entry, err := newWorkflowIntegrationEntry(run, request, coordinator, sequence)
	if err != nil {
		return entry, err
	}

	budget, err := s.workflowIntegrationBudgetAdmission(ctx, run, entry, request.TokenLimit)
	if err != nil {
		return entry, err
	}

	return entry, s.commitWorkflowIntegrationAdmission(ctx, &run, entry, &budget)
}

func (s *WorkflowStore) guardWorkflowIntegrationAdmission(ctx context.Context, run workflowInvocationFile, request WorkflowIntegrationAdmission) error {
	if !workflowIntegrationAdmissionCurrent(run, request) {
		return workflowIntegrationConflict("admission is stale or outside the current frozen unit")
	}

	if run.PullRequest != nil && "refs/heads/"+run.PullRequest.Request.BaseBranch != request.TargetRef {
		return workflowIntegrationConflict("target differs from its canonical PR binding")
	}

	if err := s.guardWorkflowIntegrationAdmissionHistory(run); err != nil {
		return err
	}

	return s.checkWorkflowInputs(ctx, run)
}

func (s *WorkflowStore) guardWorkflowIntegrationAdmissionHistory(run workflowInvocationFile) error {
	for _, binding := range run.IntegrationEntries {
		current, err := s.readWorkflowIntegrationEntry(run, binding)
		if err != nil {
			return err
		}

		if !workflowIntegrationTerminal(current.State) {
			return ErrWorkflowOwned
		}
	}

	return nil
}

func workflowIntegrationAdmissionCurrent(run workflowInvocationFile, request WorkflowIntegrationAdmission) bool {
	return run.Revision == request.ExpectedRevision && !workflowInvocationTerminal(run.Status) && run.StopIntent == "" && run.ConfirmedCursor < len(run.Members) && workflowMemberOwnsUnit(run.Members[run.ConfirmedCursor], request.Unit)
}

func (s *WorkflowStore) nextWorkflowIntegrationSequence(ctx context.Context, runs []workflowInvocationFile, entries []WorkflowIntegrationEntry) (uint64, error) {
	limits, err := s.workflowRepositoryCapacity(ctx, runs)
	if err != nil {
		return 0, err
	}

	maximum, active := uint64(0), 0
	for _, entry := range entries {
		maximum = max(maximum, entry.QueueSequence)
		if !workflowIntegrationTerminal(entry.State) {
			active++
		}
	}

	if active >= limits.Queue {
		return 0, ErrWorkflowQueueFull
	}

	if maximum == math.MaxUint64 {
		return 0, workflowIntegrationConflict("queue sequence exhausted")
	}

	return maximum + 1, nil
}

func workflowIntegrationEntryChange(entry WorkflowIntegrationEntry, before string) (workflowFileChange, error) {
	if err := validateWorkflowIntegrationEntry(entry); err != nil {
		return workflowFileChange{}, err
	}

	data, err := json.MarshalIndent(entry, "", "  ")
	if err != nil || len(data) >= WorkflowFileLimit {
		return workflowFileChange{}, fmt.Errorf("integration entry exceeds file bounds: %w", errorsOrConflict(err))
	}

	return workflowFileChange{Path: workflowIntegrationEntryPath(entry.InvocationID, entry.ID), Before: before, After: append(data, '\n')}, nil
}

func (s *WorkflowStore) commitWorkflowIntegrationEntry(run *workflowInvocationFile, entry WorkflowIntegrationEntry, before string) error {
	if run.Revision == math.MaxUint64 {
		return workflowIntegrationConflict("invocation revision exhausted")
	}

	file, err := workflowIntegrationEntryChange(entry, before)
	if err != nil {
		return err
	}

	binding := WorkflowIntegrationBinding{EntryID: entry.ID, Unit: entry.Unit, Revision: entry.Revision, EntryRef: WorkflowEvidenceRef{Path: file.Path, Hash: workflowHash(file.After)}}
	updateWorkflowIntegrationBinding(run, binding)

	run.Revision++
	run.UpdatedAt = time.Now().UTC()

	invocation, err := workflowInvocationChange(*run)
	if err != nil {
		return err
	}

	return s.commitFiles([]workflowFileChange{file, invocation})
}

func updateWorkflowIntegrationBinding(run *workflowInvocationFile, binding WorkflowIntegrationBinding) {
	for i, old := range run.IntegrationEntries {
		if old.EntryID == binding.EntryID {
			run.IntegrationEntries[i] = binding
			run.Integration = &binding

			return
		}
	}

	run.IntegrationEntries = append(run.IntegrationEntries, binding)
	run.Integration = &binding
}

func (s *WorkflowStore) commitWorkflowIntegrationAdmission(ctx context.Context, run *workflowInvocationFile, entry WorkflowIntegrationEntry, budget *WorkflowIntegrationBudget) error {
	if run.Revision == math.MaxUint64 {
		return ErrWorkflowConflict
	}

	budgetFile, err := s.workflowIntegrationBudgetChange(ctx, run, budget)
	if err != nil {
		return err
	}

	file, err := workflowIntegrationEntryChange(entry, "")
	if err != nil {
		return err
	}

	updateWorkflowIntegrationBinding(run, WorkflowIntegrationBinding{EntryID: entry.ID, Unit: entry.Unit, Revision: entry.Revision, EntryRef: WorkflowEvidenceRef{Path: file.Path, Hash: workflowHash(file.After)}})

	run.Revision++
	run.UpdatedAt = time.Now().UTC()

	invocation, err := workflowInvocationChange(*run)
	if err != nil {
		return err
	}

	return s.commitFiles([]workflowFileChange{budgetFile, file, invocation})
}
