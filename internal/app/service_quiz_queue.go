package app

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"slices"
	"strings"

	"forge.adrianpk.com/chr/internal/core"
	quizcore "forge.adrianpk.com/chr/internal/core/quiz"
)

func (s service) nextQuizQuestion(state quizcore.QuizRuntimeState, entries []quizcore.QuizEntry) (quizcore.QuizRuntimeState, error) {
	if state.Mode == quizcore.QuizModeFast {
		return s.nextFastQuizQuestion(state, entries)
	}

	if len(state.Candidates) > 0 {
		state = activateQuizCandidate(state, state.Candidates[0], s.utcNow())
		state.Candidates = state.Candidates[1:]

		return state, nil
	}

	request, err := s.newQuizModelRequestWithEntries(state, entries)
	if err != nil {
		return state, err
	}

	response, err := s.invokeQuizGeneration(state, request)
	if err != nil {
		return state, err
	}

	if response.Status == core.QuizModelResponseBlocked {
		state.Status, state.ActiveQuestion = quizcore.QuizRuntimeStatusBlocked, nil
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

	state = activateQuizQuestion(state, response.Question, response.CreatedAt.UTC())
	state.Candidates = response.Candidates

	return state, nil
}

func (s service) nextFastQuizQuestion(state quizcore.QuizRuntimeState, entries []quizcore.QuizEntry) (quizcore.QuizRuntimeState, error) {
	if state.FastBatchID == "" {
		var err error

		state, err = s.generateFastQuizBatch(state, entries)
		if err != nil {
			return state, err
		}
	}

	state = retainUnansweredFastQuestions(state, entries)

	return s.activateNextFastQuestion(state), nil
}

func (s service) activateNextFastQuestion(state quizcore.QuizRuntimeState) quizcore.QuizRuntimeState {
	if state.ActiveQuestion != nil {
		return state
	}

	state.Status, state.UpdatedAt = quizcore.QuizRuntimeStatusIdle, s.utcNow()

	state.ContinuationPending, state.ContinuationError = false, ""
	if len(state.FastQueue) == 0 || state.QuestionsAsked >= state.QuestionBudget {
		return state
	}

	candidate := state.FastQueue[0]
	state.FastQueue = state.FastQueue[1:]
	state.FastConsumed = append(state.FastConsumed, candidate.ID)

	return activateQuizCandidate(state, candidate, s.utcNow())
}

func (s service) generateFastQuizBatch(state quizcore.QuizRuntimeState, entries []quizcore.QuizEntry) (quizcore.QuizRuntimeState, error) {
	request, err := s.newQuizModelRequestWithEntries(state, entries)
	if err != nil {
		return state, err
	}

	response, err := s.invokeQuizGeneration(state, request)
	if err != nil {
		return state, err
	}

	if response.Status != core.QuizModelResponseReady {
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

func fastQuizBatchQuestions(state quizcore.QuizRuntimeState, entries []quizcore.QuizEntry, response core.QuizModelResponse) []quizcore.QuizCandidate {
	question := response.Question
	questions := append([]quizcore.QuizCandidate{{ID: question.ID, Prompt: question.Prompt, Topic: question.Topic, Depth: 1, CreatedAt: question.CreatedAt}}, response.Candidates...)

	seen := map[string]bool{}
	if state.ActiveQuestion != nil {
		seen[state.ActiveQuestion.ID], seen[state.ActiveQuestion.Prompt] = true, true
	}

	for _, entry := range entries {
		if quizSessionEntry(state, entry) {
			seen[entry.QuestionID], seen[entry.Question] = true, true
		}
	}

	remaining := min(quizcore.DefaultQuizQuestionBudget, state.QuestionBudget-state.QuestionsAsked)

	queue := make([]quizcore.QuizCandidate, 0, remaining)
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

func quizSessionEntry(state quizcore.QuizRuntimeState, entry quizcore.QuizEntry) bool {
	return strings.HasPrefix(entry.ID, state.SessionID+"-entry-") || strings.HasPrefix(state.SessionID, "legacy-")
}

func retainUnansweredFastQuestions(state quizcore.QuizRuntimeState, entries []quizcore.QuizEntry) quizcore.QuizRuntimeState {
	seen := map[string]bool{}
	if state.ActiveQuestion != nil {
		seen[state.ActiveQuestion.ID], seen[state.ActiveQuestion.Prompt] = true, true
	}

	for _, entry := range entries {
		if quizSessionEntry(state, entry) {
			seen[entry.QuestionID], seen[entry.Question] = true, true
		}
	}

	queue := make([]quizcore.QuizCandidate, 0, len(state.FastQueue))
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

func (s service) completeGeneratedQuiz(state quizcore.QuizRuntimeState) (quizcore.QuizRuntimeState, error) {
	if s.guidedContext == nil {
		return s.finishQuiz(state), nil
	}

	if state.AnswersRecorded == 0 {
		return state, fmt.Errorf("guided quiz generated no answered clarification")
	}

	state.ClarificationComplete = true

	return s.finishQuiz(state), nil
}
