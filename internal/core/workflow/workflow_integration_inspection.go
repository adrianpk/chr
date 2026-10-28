package workflow

import (
	"context"
	"time"

	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

// IntegrationRoleObservation identifies a live role without a lease,
// command, process identity or private owner credential.
type IntegrationRoleObservation struct {
	AttemptID string
	Role      string
	State     string
	Slot      int
}

// IntegrationInspection is one retained image under repository exclusion.
// It does not contact providers, recover transitions or authorize execution.
type IntegrationInspection struct {
	Entry         *IntegrationEntry
	Budget        IntegrationBudget
	Usage         IntegrationUsage
	Capacity      IntegrationCapacity
	Roles         []IntegrationRoleObservation
	QueuePosition int
	Candidate     *IntegrationCandidate
	Decisions     *IntegrationDecisions
	Submission    *Submission
}

func (s *Store) inspectWorkflowIntegration(ctx context.Context, run workflowInvocationFile) (*IntegrationInspection, error) {
	if run.IntegrationBudget == nil {
		return nil, nil
	}

	snapshot, err := s.readWorkflowIntegrationSnapshot(ctx)
	if err != nil {
		return nil, err
	}

	result := &IntegrationInspection{}

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

func (s *Store) inspectWorkflowIntegrationEntry(ctx context.Context, run workflowInvocationFile, snapshot workflowIntegrationSnapshot, result *IntegrationInspection) error {
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
		var candidate IntegrationCandidate
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

func workflowIntegrationQueuePosition(views []IntegrationView, entry IntegrationEntry) int {
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

func workflowIntegrationObservedSubmission(journals []workflowSubmissionJournal, entry IntegrationEntry) (*Submission, error) {
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

func workflowIntegrationObservedRoles(run workflowInvocationFile, capacity IntegrationCapacity) ([]IntegrationRoleObservation, error) {
	result := []IntegrationRoleObservation{}

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

		result = append(result, IntegrationRoleObservation{AttemptID: attempt.ID, Role: workflowIntegrationOwnerRole(attempt.Purpose), State: state, Slot: attempt.CapacitySlot})
	}

	result = append(result, workflowIntegrationObservedClosingRoles(run, capacity)...)

	return result, nil
}

func workflowIntegrationObservedClosingRoles(run workflowInvocationFile, capacity IntegrationCapacity) []IntegrationRoleObservation {
	result := []IntegrationRoleObservation{}

	for _, owner := range capacity.UnexpiredOwners {
		if owner.InvocationID == run.ID && owner.Namespace == run.Namespace && owner.Role == "closing" {
			result = append(result, IntegrationRoleObservation{AttemptID: owner.AttemptID, Role: owner.Role, State: "running"})
		}
	}

	return result
}
