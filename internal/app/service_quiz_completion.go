package app

import (
	"errors"
	"fmt"
	"strings"

	quizcore "forge.adrianpk.com/chr/internal/core/quiz"
)

var errRepeatedQuizQuestion = errors.New("quiz provider repeated an answered question")

func sameQuizQuestion(left, right string) bool {
	return strings.EqualFold(strings.Join(strings.Fields(left), " "), strings.Join(strings.Fields(right), " "))
}

func validateNextQuizQuestion(state quizcore.RuntimeState, entries []quizcore.Entry, question quizcore.Question) error {
	for _, entry := range entries {
		if quizSessionEntry(state, entry) && entry.QuestionID == question.ID {
			return fmt.Errorf("quiz provider reused an answered question identity")
		}

		if entry.Status == quizcore.EntryStatusAnswered && sameQuizQuestion(entry.Question, question.Prompt) {
			return errRepeatedQuizQuestion
		}
	}

	return nil
}

func (s service) finishQuiz(state quizcore.RuntimeState) quizcore.RuntimeState {
	state.Status = quizcore.RuntimeStatusDone
	state.ActiveQuestion, state.Candidates = nil, nil
	state.ContinuationPending, state.ContinuationError = false, ""
	state.UpdatedAt = s.utcNow()

	return state
}
