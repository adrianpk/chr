package core

import (
	"context"
	"encoding/json"
	"time"

	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

// WorkflowIntegrationAnswer is an explicit, attributable response. Resolve
// records an interpretation; defer keeps the question waiting. Neither grants
// publication, different scope, additional calls or a replacement gate.
type WorkflowIntegrationAnswer struct {
	ID             string                `json:"id"`
	DecisionID     string                `json:"decision_id"`
	Choice         string                `json:"choice"`
	Action         string                `json:"action"`
	Actor          string                `json:"actor"`
	Authority      string                `json:"authority"`
	Reason         string                `json:"reason"`
	PriorRevision  uint64                `json:"prior_revision"`
	QuestionHash   string                `json:"question_hash"`
	RecordedAt     time.Time             `json:"recorded_at"`
	OriginalCommit string                `json:"original_commit,omitempty"`
	IntegrationFix *AgenticReviewFinding `json:"integration_fix,omitempty"`
}

// WorkflowIntegrationDecisions is an immutable revision under the existing
// queue entry. Its original merger input owns source/base, scope and questions.
type WorkflowIntegrationDecisions struct {
	SchemaVersion     int                         `json:"schema_version"`
	ID                string                      `json:"id"`
	Revision          uint64                      `json:"revision"`
	CandidateRevision uint64                      `json:"candidate_revision"`
	State             string                      `json:"state"`
	Input             WorkflowMergerInput         `json:"input"`
	Requests          []WorkflowMergerDecision    `json:"requests"`
	Answers           []WorkflowIntegrationAnswer `json:"answers"`
	Previous          *WorkflowEvidenceRef        `json:"previous,omitempty"`
}

// WorkflowIntegrationDecisionRequest records bounded ambiguity at a fresh
// candidate fence. Callers may use host-detected or retained merger questions.
type WorkflowIntegrationDecisionRequest struct {
	WorkflowIntegrationBudgetGuard
	ID        string
	Input     WorkflowMergerInput
	Decisions []WorkflowMergerDecision
}

// WorkflowIntegrationAnswerRequest binds an answer to the displayed record
// and all existing mutation revisions; zero values cannot imply approval.
type WorkflowIntegrationAnswerRequest struct {
	WorkflowIntegrationBudgetGuard
	DecisionRevision uint64
	DecisionHash     string
	Answer           WorkflowIntegrationAnswer
}

// ReadWorkflowIntegrationDecisions returns the current sealed question record.
// Stale decisions are reported without transferring old answers to new work.
func (s *WorkflowStore) ReadWorkflowIntegrationDecisions(ctx context.Context, id string) (result WorkflowIntegrationDecisions, err error) {
	err = s.withWorkflowInspection(ctx, func() error {
		run, owner, readErr := s.locateWorkflowInvocationPrepared(ctx, id, (*WorkflowStore).requireNoWorkflowTransition)
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
func (s *WorkflowStore) RecordWorkflowIntegrationDecisions(ctx context.Context, request WorkflowIntegrationDecisionRequest, provider WorkflowPullRequestProvider) (result WorkflowIntegrationDecisions, err error) {
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
func (s *WorkflowStore) AnswerWorkflowIntegrationDecision(ctx context.Context, request WorkflowIntegrationAnswerRequest, provider WorkflowPullRequestProvider) (result WorkflowIntegrationDecisions, err error) {
	data, err := json.Marshal(request)
	if err != nil || len(data) >= workflowstate.WorkflowFileLimit {
		return result, workflowstate.ErrConflict
	}

	var frozen WorkflowIntegrationAnswerRequest
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

func freezeWorkflowDecisionRequest(request WorkflowIntegrationDecisionRequest) (WorkflowIntegrationDecisionRequest, error) {
	data, err := json.Marshal(request)
	if err != nil || len(data) >= workflowstate.WorkflowFileLimit {
		return WorkflowIntegrationDecisionRequest{}, workflowstate.ErrConflict
	}

	var frozen WorkflowIntegrationDecisionRequest

	err = json.Unmarshal(data, &frozen)

	return frozen, err
}
