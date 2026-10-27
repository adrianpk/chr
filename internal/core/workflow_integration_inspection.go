package core

import (
	"context"
	"time"

	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

// WorkflowIntegrationRoleObservation identifies a live role without a lease,
// command, process identity or private owner credential.
type WorkflowIntegrationRoleObservation struct {
	AttemptID string
	Role      string
	State     string
	Slot      int
}

// WorkflowIntegrationInspection is one retained image under repository exclusion.
// It does not contact providers, recover transitions or authorize execution.
type WorkflowIntegrationInspection struct {
	Entry         *WorkflowIntegrationEntry
	Budget        WorkflowIntegrationBudget
	Usage         WorkflowIntegrationUsage
	Capacity      WorkflowIntegrationCapacity
	Roles         []WorkflowIntegrationRoleObservation
	QueuePosition int
	Candidate     *WorkflowIntegrationCandidate
	Decisions     *WorkflowIntegrationDecisions
	Submission    *WorkflowSubmission
}

func (s *WorkflowStore) inspectWorkflowIntegration(ctx context.Context, run workflowInvocationFile) (*WorkflowIntegrationInspection, error) {
	if run.IntegrationBudget == nil {
		return nil, nil
	}

	snapshot, err := s.readWorkflowIntegrationSnapshot(ctx)
	if err != nil {
		return nil, err
	}

	result := &WorkflowIntegrationInspection{}

	result.Budget, err = s.readWorkflowIntegrationBudget(run)
	if err != nil {
		return nil, err
	}

	result.Usage, err = s.workflowIntegrationUsage(ctx, run)
	if err != nil {
		return nil, err
	}

	result.Capacity, err = s.workflowIntegrationCapacity(ctx, snapshot.Runs)
	if err != nil {
		return nil, err
	}

	result.Roles, err = workflowIntegrationObservedRoles(run, result.Capacity)
	if err != nil {
		return nil, err
	}

	if err := s.inspectWorkflowIntegrationEntry(ctx, run, snapshot, result); err != nil {
		return nil, err
	}

	return result, nil
}

func (s *WorkflowStore) inspectWorkflowIntegrationEntry(ctx context.Context, run workflowInvocationFile, snapshot workflowIntegrationSnapshot, result *WorkflowIntegrationInspection) error {
	if run.Integration == nil {
		return nil
	}

	entry, err := s.readWorkflowIntegrationEntry(run, *run.Integration)
	if err != nil {
		return err
	}

	result.Entry = &entry
	result.QueuePosition = workflowIntegrationQueuePosition(snapshot.Views, entry)

	result.Submission, err = workflowIntegrationObservedSubmission(snapshot.Journals, entry)
	if err != nil {
		return err
	}

	if entry.CandidateRef != nil {
		var candidate WorkflowIntegrationCandidate
		if err := s.readWorkflowNativeEvidence(run.ID, *entry.CandidateRef, &candidate); err != nil {
			return err
		}

		result.Candidate = &candidate
	}

	if entry.DecisionRef != nil {
		decisions, err := s.readWorkflowDecisionRecord(ctx, run, entry)
		if err != nil {
			return err
		}

		if checkWorkflowDecisionFence(run, entry, decisions) != nil {
			decisions.State = "stale"
		}

		result.Decisions = &decisions
	}

	return nil
}

func workflowIntegrationQueuePosition(views []WorkflowIntegrationView, entry WorkflowIntegrationEntry) int {
	position := 0

	for _, view := range views {
		if view.Entry.Repository != entry.Repository || view.Entry.TargetRef != entry.TargetRef || workflowIntegrationTerminal(view.Entry.State) {
			continue
		}

		position++
		if view.Entry.ID == entry.ID {
			return position
		}
	}

	return 0
}

func workflowIntegrationObservedSubmission(journals []workflowSubmissionJournal, entry WorkflowIntegrationEntry) (*WorkflowSubmission, error) {
	if entry.SubmissionRef == nil {
		return nil, nil
	}

	for _, journal := range journals {
		if journal.Key != entry.SubmissionRef.JournalKey {
			continue
		}

		for _, submission := range journal.Attempts {
			if submission.ID == entry.SubmissionRef.ID {
				return &submission, nil
			}
		}
	}

	return nil, workflowstate.ErrConflict
}

func workflowIntegrationObservedRoles(run workflowInvocationFile, capacity WorkflowIntegrationCapacity) ([]WorkflowIntegrationRoleObservation, error) {
	result := []WorkflowIntegrationRoleObservation{}

	for _, attempt := range run.Attempts {
		if attempt.OwnerProcess == nil || attempt.Outcome != "running" {
			continue
		}

		alive, err := workflowProcessAlive(*attempt.OwnerProcess)
		if err != nil {
			return nil, err
		}

		if !alive {
			continue
		}

		state := "running"
		if !attempt.LeaseUntil.After(time.Now().UTC()) {
			state = "reconciliation_required"
		}

		result = append(result, WorkflowIntegrationRoleObservation{AttemptID: attempt.ID, Role: workflowIntegrationOwnerRole(attempt.Purpose), State: state, Slot: attempt.CapacitySlot})
	}

	result = append(result, workflowIntegrationObservedClosingRoles(run, capacity)...)

	return result, nil
}

func workflowIntegrationObservedClosingRoles(run workflowInvocationFile, capacity WorkflowIntegrationCapacity) []WorkflowIntegrationRoleObservation {
	result := []WorkflowIntegrationRoleObservation{}

	for _, owner := range capacity.UnexpiredOwners {
		if owner.InvocationID == run.ID && owner.Namespace == run.Namespace && owner.Role == "closing" {
			result = append(result, WorkflowIntegrationRoleObservation{AttemptID: owner.AttemptID, Role: owner.Role, State: "running"})
		}
	}

	return result
}
