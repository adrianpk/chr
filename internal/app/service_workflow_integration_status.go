package app

import (
	"forge.adrianpk.com/chr/internal/core"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

// WorkflowIntegrationRoleStatus reports actual role occupancy without leases.
type WorkflowIntegrationRoleStatus struct {
	AttemptID string `json:"attempt_id"`
	Role      string `json:"role"`
	State     string `json:"state"`
	Slot      int    `json:"slot,omitempty"`
}

// WorkflowIntegrationDecisionStatus carries only displayed questions and answers.
type WorkflowIntegrationDecisionStatus struct {
	ID       string                           `json:"id"`
	Revision uint64                           `json:"revision"`
	Hash     string                           `json:"hash"`
	State    string                           `json:"state"`
	Requests []core.WorkflowMergerDecision    `json:"questions"`
	Answers  []core.WorkflowIntegrationAnswer `json:"answers"`
}

// WorkflowIntegrationStatus projects retained integration facts for every adapter.
// Revisions are action fences; observations confer no claim or publication grant.
type WorkflowIntegrationStatus struct {
	EntryID        string                             `json:"entry_id,omitempty"`
	EntryRevision  uint64                             `json:"entry_revision"`
	Unit           workflowstate.Ref                  `json:"unit"`
	Repository     string                             `json:"repository,omitempty"`
	TargetRef      string                             `json:"target_ref,omitempty"`
	Phase          string                             `json:"phase"`
	QueuePosition  int                                `json:"queue_position"`
	CandidateHash  string                             `json:"candidate_hash,omitempty"`
	OriginalSource string                             `json:"original_source,omitempty"`
	ValidatedBase  string                             `json:"validated_base,omitempty"`
	PreparedResult string                             `json:"prepared_result,omitempty"`
	Method         string                             `json:"method,omitempty"`
	Publication    string                             `json:"publication"`
	Blocker        *core.WorkflowIntegrationBlocker   `json:"blocker,omitempty"`
	MaximumRoles   int                                `json:"maximum_roles"`
	AvailableRoles int                                `json:"available_roles"`
	Roles          []WorkflowIntegrationRoleStatus    `json:"roles"`
	Budget         WorkflowIntegrationBudgetStatus    `json:"budget"`
	Decision       *WorkflowIntegrationDecisionStatus `json:"decision,omitempty"`
}

func workflowIntegrationStatus(observed *core.WorkflowIntegrationInspection) *WorkflowIntegrationStatus {
	if observed == nil {
		return nil
	}

	result := &WorkflowIntegrationStatus{Phase: "not_admitted", Publication: "not_started", MaximumRoles: observed.Capacity.Maximum, AvailableRoles: observed.Capacity.Available, Roles: []WorkflowIntegrationRoleStatus{}}
	for _, role := range observed.Roles {
		result.Roles = append(result.Roles, WorkflowIntegrationRoleStatus{AttemptID: role.AttemptID, Role: role.Role, State: role.State, Slot: role.Slot})
	}

	if entry := observed.Entry; entry != nil {
		result.EntryID, result.EntryRevision, result.Unit = entry.ID, entry.Revision, entry.Unit
		result.Repository, result.TargetRef, result.Phase = entry.Repository, entry.TargetRef, entry.State

		result.QueuePosition, result.Blocker = observed.QueuePosition, entry.Blocker
		if entry.CandidateRef != nil {
			result.CandidateHash = entry.CandidateRef.Hash
		}
	}

	workflowIntegrationCandidateStatus(result, observed.Candidate)

	result.Publication = workflowIntegrationPublicationStatus(observed.Submission)

	result.Budget = workflowIntegrationBudgetStatus(observed, result.Unit)
	if decision := observed.Decisions; decision != nil && observed.Entry != nil {
		result.Decision = &WorkflowIntegrationDecisionStatus{ID: decision.ID, Revision: decision.Revision, Hash: observed.Entry.DecisionRef.Hash, State: decision.State, Requests: decision.Requests, Answers: decision.Answers}
		if observed.Candidate == nil {
			result.OriginalSource, result.ValidatedBase, result.Method = decision.Input.Source, decision.Input.Base, decision.Input.Method
			result.CandidateHash = decision.Input.CandidateHash
		}
	}

	result.CandidateHash = core.WorkflowIntegrationDisplayedCandidate(observed)

	return result
}

func workflowIntegrationCandidateStatus(result *WorkflowIntegrationStatus, candidate *core.WorkflowIntegrationCandidate) {
	if candidate == nil {
		return
	}

	git := candidate.Input.Git

	result.OriginalSource, result.ValidatedBase, result.PreparedResult, result.Method = git.SourceCommit, git.BaseCommit, git.ResultCommit, git.Method
	if candidate.Native != nil {
		result.CandidateHash = candidate.Native.CandidateHash
	}
}

func workflowIntegrationPublicationStatus(submission *core.WorkflowSubmission) string {
	if submission == nil {
		return "not_started"
	}

	if submission.State == "confirmed" {
		return "verified"
	}

	if submission.State == "unpublished" {
		return "unpublished"
	}

	if submission.Native != nil {
		if submission.Native.ApplicationState != "not_started" {
			return submission.Native.ApplicationState
		}

		if submission.Native.DispatchState != "not_started" {
			return "in_flight"
		}
	}

	return "awaiting_observation"
}
