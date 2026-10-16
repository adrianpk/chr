package core

import (
	"context"
	"reflect"
	"slices"
	"time"
)

func (s *WorkflowStore) answerWorkflowDecision(ctx context.Context, run workflowInvocationFile, request WorkflowIntegrationAnswerRequest, provider WorkflowPullRequestProvider, control *workflowDecisionControlCommit) (WorkflowIntegrationDecisions, error) {
	entry, budget, err := s.currentWorkflowIntegrationBudget(ctx, run, request.WorkflowIntegrationBudgetGuard)
	if err != nil {
		return WorkflowIntegrationDecisions{}, err
	}

	record, err := s.readWorkflowDecisionRecord(ctx, run, entry)
	if err != nil {
		return record, err
	}

	if err := s.checkWorkflowDecisionCurrent(ctx, run, entry, record, provider); err != nil {
		return record, err
	}

	if found, err := workflowDecisionAnswerRetry(record, request); found || err != nil {
		return record, s.completeWorkflowControlRetry(ctx, run, control, err)
	}

	answer := request.Answer

	answer.PriorRevision, answer.QuestionHash, answer.RecordedAt = record.Revision, entry.DecisionRef.Hash, time.Now().UTC()
	if err := s.checkWorkflowFreshAnswer(run, entry, budget, record, request, answer); err != nil {
		return record, err
	}

	record.Answers = append(record.Answers, answer)
	record.Previous = entry.DecisionRef
	record.Revision++
	entry.Revision++
	entry.UpdatedAt = answer.RecordedAt

	if workflowDecisionsResolved(record) {
		record.State = "resolved"
		entry.State, entry.Blocker = "merging", nil
		entry.StateRevision++
		record.CandidateRevision = entry.StateRevision
		run.Outcome = "merger_decision_resolved"
	}

	file, err := workflowDecisionFile(&entry, record)
	if err != nil {
		return record, err
	}

	files, err := control.files(ctx)
	if err != nil {
		return record, err
	}

	return record, s.commitWorkflowDecisionRecord(&run, entry, file, files...)
}

func (s *WorkflowStore) checkWorkflowDecisionCurrent(ctx context.Context, run workflowInvocationFile, entry WorkflowIntegrationEntry, record WorkflowIntegrationDecisions, provider WorkflowPullRequestProvider) error {
	if run.StopIntent != "" || entry.StopIntent != "" {
		return ErrWorkflowStop
	}

	if err := checkWorkflowDecisionFence(run, entry, record); err != nil {
		return err
	}

	if err := s.checkWorkflowInputs(ctx, run); err != nil {
		return err
	}

	return checkWorkflowMergerRemote(ctx, provider, record.Input)
}

func checkWorkflowDecisionFence(run workflowInvocationFile, entry WorkflowIntegrationEntry, record WorkflowIntegrationDecisions) error {
	if run.Lane == nil || entry.StateRevision != record.CandidateRevision || entry.ClaimGeneration != record.Input.ClaimGeneration {
		return ErrWorkflowConflict
	}

	input := record.Input
	input.CandidateRevision = entry.StateRevision

	return checkWorkflowMergerEntry(run, entry, input)
}

func workflowDecisionsResolved(record WorkflowIntegrationDecisions) bool {
	return len(record.Requests) > 0 && !slices.ContainsFunc(record.Requests, func(d WorkflowMergerDecision) bool { return !workflowDecisionResolved(record.Answers, d.ID) })
}

func workflowDecisionResolved(answers []WorkflowIntegrationAnswer, id string) bool {
	for i := len(answers) - 1; i >= 0; i-- {
		if answers[i].DecisionID == id {
			return answers[i].Action == "resolve"
		}
	}

	return false
}

func workflowDecisionAnswerRetry(record WorkflowIntegrationDecisions, request WorkflowIntegrationAnswerRequest) (bool, error) {
	for _, old := range record.Answers {
		if old.ID != request.Answer.ID {
			continue
		}

		answer := request.Answer

		answer.PriorRevision, answer.QuestionHash, answer.RecordedAt = old.PriorRevision, old.QuestionHash, old.RecordedAt
		if !reflect.DeepEqual(answer, old) || request.DecisionRevision != old.PriorRevision || request.DecisionHash != old.QuestionHash {
			return false, ErrWorkflowConflict
		}

		return true, nil
	}

	return false, nil
}

func (s *WorkflowStore) checkWorkflowFreshAnswer(run workflowInvocationFile, entry WorkflowIntegrationEntry, budget WorkflowIntegrationBudget, record WorkflowIntegrationDecisions, request WorkflowIntegrationAnswerRequest, answer WorkflowIntegrationAnswer) error {
	if err := guardWorkflowIntegrationBudget(run, entry, budget, request.WorkflowIntegrationBudgetGuard); err != nil {
		return err
	}

	if request.DecisionRevision != record.Revision || request.DecisionHash != entry.DecisionRef.Hash || entry.State != "blocked" || record.State != "waiting" {
		return ErrWorkflowConflict
	}

	if err := validateWorkflowDecisionAnswer(record, answer); err != nil {
		return err
	}

	if err := s.checkWorkflowDecisionReplayAnswer(record, answer); err != nil {
		return err
	}

	return nil
}
