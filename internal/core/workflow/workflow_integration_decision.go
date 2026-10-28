package workflow

import (
	"context"
	"encoding/json"

	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

// IntegrationDecisions is an immutable revision under the existing
// queue entry. Its original merger input owns source/base, scope and questions.
type IntegrationDecisions struct {
	SchemaVersion     int                               `json:"schema_version"`
	ID                string                            `json:"id"`
	Revision          uint64                            `json:"revision"`
	CandidateRevision uint64                            `json:"candidate_revision"`
	State             string                            `json:"state"`
	Input             workflowstate.MergerInput         `json:"input"`
	Requests          []workflowstate.MergerDecision    `json:"requests"`
	Answers           []workflowstate.IntegrationAnswer `json:"answers"`
	Previous          *workflowstate.EvidenceRef        `json:"previous,omitempty"`
}

// IntegrationDecisionRequest records bounded ambiguity at a fresh
// candidate fence. Callers may use host-detected or retained merger questions.
type IntegrationDecisionRequest struct {
	IntegrationBudgetGuard
	ID        string
	Input     workflowstate.MergerInput
	Decisions []workflowstate.MergerDecision
}

// IntegrationAnswerRequest binds an answer to the displayed record
// and all existing mutation revisions; zero values cannot imply approval.
type IntegrationAnswerRequest struct {
	IntegrationBudgetGuard
	DecisionRevision uint64
	DecisionHash     string
	Answer           workflowstate.IntegrationAnswer
}

// ReadWorkflowIntegrationDecisions returns the current sealed question record.
// Stale decisions are reported without transferring old answers to new work.
func (s *Store) ReadWorkflowIntegrationDecisions(ctx context.Context, id string) (result IntegrationDecisions, err error) {
	err = s.withWorkflowInspection(ctx, func() error {
		run, owner, readErr := s.locateWorkflowInvocationPrepared(ctx, id, (*Store).requireNoWorkflowTransition)
		if readErr != nil {
			return readErr
		}

		if run.Integration == nil {
			return workflowstate.ErrConflict
		}

		entry, readErr := owner.readWorkflowIntegrationEntry(run, *run.Integration)
		if readErr != nil {
			return readErr
		}

		result, readErr = owner.readWorkflowDecisionRecord(ctx, run, entry)
		if readErr == nil && checkWorkflowDecisionFence(run, entry, result) != nil {
			result.State = "stale"
		}

		return readErr
	})

	return result, err
}

// RecordWorkflowIntegrationDecisions blocks the current candidate without
// consuming an agent call or treating elapsed time as an answer.
func (s *Store) RecordWorkflowIntegrationDecisions(ctx context.Context, request IntegrationDecisionRequest, provider PullRequestProvider) (result IntegrationDecisions, err error) {
	request, err = freezeWorkflowDecisionRequest(request)
	if err != nil {
		return result, err
	}

	err = s.withLock(ctx, func() error {
		run, owner, readErr := s.locateWorkflowInvocation(ctx, request.InvocationID)
		if readErr != nil {
			return readErr
		}

		result, readErr = owner.recordWorkflowDecisions(ctx, run, request, provider)

		return readErr
	})

	return result, err
}

// AnswerWorkflowIntegrationDecision persists only an explicit current response.
// A resolved batch returns to merger work; validation and review remain required.
func (s *Store) AnswerWorkflowIntegrationDecision(ctx context.Context, request IntegrationAnswerRequest, provider PullRequestProvider) (result IntegrationDecisions, err error) {
	data, err := json.Marshal(request)
	if err != nil || len(data) >= workflowstate.FileLimit {
		return result, workflowstate.ErrConflict
	}

	var frozen IntegrationAnswerRequest
	if err := json.Unmarshal(data, &frozen); err != nil {
		return result, err
	}

	err = s.withLock(ctx, func() error {
		run, owner, readErr := s.locateWorkflowInvocation(ctx, request.InvocationID)
		if readErr != nil {
			return readErr
		}

		result, readErr = owner.answerWorkflowDecision(ctx, run, frozen, provider, nil)

		return readErr
	})

	return result, err
}

func freezeWorkflowDecisionRequest(request IntegrationDecisionRequest) (IntegrationDecisionRequest, error) {
	data, err := json.Marshal(request)
	if err != nil || len(data) >= workflowstate.FileLimit {
		return IntegrationDecisionRequest{}, workflowstate.ErrConflict
	}

	var frozen IntegrationDecisionRequest

	err = json.Unmarshal(data, &frozen)

	return frozen, err
}
