package app

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"slices"
	"strings"

	quizcore "forge.adrianpk.com/chr/internal/core/quiz"
	quizmodel "forge.adrianpk.com/chr/internal/core/quiz/model"
)

func (s service) nextQuizQuestion(ctx context.Context, state quizcore.RuntimeState, entries []quizcore.Entry) (quizcore.RuntimeState, error) {
	if state.Mode == quizcore.ModeFast {
		return s.nextFastQuizQuestion(ctx, state, entries)
	}

	if len(state.Candidates) > 0 {
		state = quizcore.ActivateCandidate(state, state.Candidates[0], s.utcNow())
		state.Candidates = state.Candidates[1:]

		return state, nil
	}

	request, err := s.newQuizModelRequestWithEntries(ctx, state, entries)
	if err != nil {
		return state, err
	}

	response, err := s.invokeQuizGeneration(ctx, state, request)
	if err != nil {
		return state, err
	}

	if response.Status == quizmodel.ResponseBlocked {
		state.Status, state.ActiveQuestion = quizcore.RuntimeStatusBlocked, nil
		state.ContinuationPending = true
		state.ContinuationError = "More project context is required. Resume to retry."
		state.UpdatedAt = s.utcNow()

		return state, nil
	}

	if !response.Question.Required {
		return s.completeGeneratedQuiz(state)
	}

	if err := validateNextQuizQuestion(state, entries, response.Question); err != nil {
		return state, err
	}

	state = quizcore.ActivateQuestion(state, response.Question, response.CreatedAt.UTC())
	state.Candidates = response.Candidates

	return state, nil
}

func (s service) nextFastQuizQuestion(ctx context.Context, state quizcore.RuntimeState, entries []quizcore.Entry) (quizcore.RuntimeState, error) {
	if state.FastBatchID == "" {
		var err error

		state, err = s.generateFastQuizBatch(ctx, state, entries)
		if err != nil {
			return state, err
		}
	}

	state = retainUnansweredFastQuestions(state, entries)

	return s.activateNextFastQuestion(state), nil
}

func (s service) activateNextFastQuestion(state quizcore.RuntimeState) quizcore.RuntimeState {
	if state.ActiveQuestion != nil {
		return state
	}

	state.Status, state.UpdatedAt = quizcore.RuntimeStatusIdle, s.utcNow()

	state.ContinuationPending, state.ContinuationError = false, ""
	if len(state.FastQueue) == 0 || state.QuestionsAsked >= state.QuestionBudget {
		return state
	}

	candidate := state.FastQueue[0]
	state.FastQueue = state.FastQueue[1:]
	state.FastConsumed = append(state.FastConsumed, candidate.ID)

	return quizcore.ActivateCandidate(state, candidate, s.utcNow())
}

func (s service) generateFastQuizBatch(ctx context.Context, state quizcore.RuntimeState, entries []quizcore.Entry) (quizcore.RuntimeState, error) {
	request, err := s.newQuizModelRequestWithEntries(ctx, state, entries)
	if err != nil {
		return state, err
	}

	response, err := s.invokeQuizGeneration(ctx, state, request)
	if err != nil {
		return state, err
	}

	if response.Status != quizmodel.ResponseReady {
		return state, fmt.Errorf("fast question batch is blocked; retry after adding context")
	}

	encoded, err := json.Marshal(request)
	if err != nil {
		return state, err
	}

	digest := sha256.Sum256(encoded)
	state.FastBatchID = state.SessionID + "-fast"
	state.FastSourceHash = hex.EncodeToString(digest[:])
	state.FastQueue = fastQuizBatchQuestions(state, entries, response)
	state.UpdatedAt = s.utcNow()

	return state, nil
}

func fastQuizBatchQuestions(state quizcore.RuntimeState, entries []quizcore.Entry, response quizmodel.Response) []quizcore.Candidate {
	question := response.Question
	questions := append([]quizcore.Candidate{{ID: question.ID, Prompt: question.Prompt, Topic: question.Topic, Depth: 1, CreatedAt: question.CreatedAt}}, response.Candidates...)

	seen := map[string]bool{}
	if state.ActiveQuestion != nil {
		seen[state.ActiveQuestion.ID], seen[state.ActiveQuestion.Prompt] = true, true
	}

	for _, entry := range entries {
		if quizSessionEntry(state, entry) {
			seen[entry.QuestionID], seen[entry.Question] = true, true
		}
	}

	remaining := min(quizcore.DefaultQuestionBudget, state.QuestionBudget-state.QuestionsAsked)

	queue := make([]quizcore.Candidate, 0, remaining)
	for _, candidate := range questions {
		if len(queue) >= remaining {
			break
		}

		if seen[candidate.ID] || seen[candidate.Prompt] {
			continue
		}

		candidate.Depth, candidate.ParentID = 1, ""
		queue = append(queue, candidate)
		seen[candidate.ID], seen[candidate.Prompt] = true, true
	}

	return queue
}

func quizSessionEntry(state quizcore.RuntimeState, entry quizcore.Entry) bool {
	return strings.HasPrefix(entry.ID, state.SessionID+"-entry-") || strings.HasPrefix(state.SessionID, "legacy-")
}

func retainUnansweredFastQuestions(state quizcore.RuntimeState, entries []quizcore.Entry) quizcore.RuntimeState {
	seen := map[string]bool{}
	if state.ActiveQuestion != nil {
		seen[state.ActiveQuestion.ID], seen[state.ActiveQuestion.Prompt] = true, true
	}

	for _, entry := range entries {
		if quizSessionEntry(state, entry) {
			seen[entry.QuestionID], seen[entry.Question] = true, true
		}
	}

	queue := make([]quizcore.Candidate, 0, len(state.FastQueue))
	for _, candidate := range state.FastQueue {
		if !seen[candidate.ID] && !seen[candidate.Prompt] && !slices.Contains(state.FastConsumed, candidate.ID) {
			queue = append(queue, candidate)

			continue
		}

		if !slices.Contains(state.FastConsumed, candidate.ID) {
			state.FastConsumed = append(state.FastConsumed, candidate.ID)
		}
	}

	state.FastQueue = queue

	return state
}

func (s service) completeGeneratedQuiz(state quizcore.RuntimeState) (quizcore.RuntimeState, error) {
	if s.guidedContext == nil {
		return s.finishQuiz(state), nil
	}

	if state.AnswersRecorded == 0 {
		return state, fmt.Errorf("guided quiz generated no answered clarification")
	}

	state.ClarificationComplete = true

	return s.finishQuiz(state), nil
}
