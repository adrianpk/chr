package app

import (
	"errors"
	"fmt"
	"strings"

	"forge.adrianpk.com/chr/internal/core"
)

var errRepeatedQuizQuestion = errors.New("quiz provider repeated an answered question")

func sameQuizQuestion(left, right string) bool {
	return strings.EqualFold(strings.Join(strings.Fields(left), " "), strings.Join(strings.Fields(right), " "))
}

func validateNextQuizQuestion(state core.QuizRuntimeState, entries []core.QuizEntry, question core.QuizQuestion) error {
	for _, entry := range entries {
		if quizSessionEntry(state, entry) && entry.QuestionID == question.ID {
			return fmt.Errorf("quiz provider reused an answered question identity")
		}

		if entry.Status == core.QuizEntryStatusAnswered && sameQuizQuestion(entry.Question, question.Prompt) {
			return errRepeatedQuizQuestion
		}
	}

	return nil
}

func (s service) finishQuiz(state core.QuizRuntimeState) core.QuizRuntimeState {
	state.Status = core.QuizRuntimeStatusDone
	state.ActiveQuestion, state.Candidates = nil, nil
	state.ContinuationPending, state.ContinuationError = false, ""
	state.UpdatedAt = s.utcNow()

	return state
}
