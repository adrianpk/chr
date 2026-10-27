package app

import (
	"fmt"

	"forge.adrianpk.com/chr/internal/core"
	quizcore "forge.adrianpk.com/chr/internal/core/quiz"
	quizmodel "forge.adrianpk.com/chr/internal/core/quiz/model"
)

// Done asks the same clarification owner whether the saved answers suffice.
// A user control alone cannot grant the trusted completion required by sealing.
func (s service) prepareGuidedQuizDone(observed core.QuizObservation) (core.QuizTransition, error) {
	state := observed.State
	if state.AnswersRecorded == 0 {
		state.ContinuationError = "Answer at least one question before finishing clarification."

		return core.QuizTransition{State: state, Message: state.ContinuationError}, nil
	}

	check := state
	check.Mode = quizcore.ModeDeep

	request, err := s.guidedQuizModelRequest(check, observed.Entries)
	if err != nil {
		return core.QuizTransition{}, err
	}

	response, err := s.invokeQuizGeneration(check, request)
	if err != nil {
		return core.QuizTransition{}, err
	}

	if response.Status != quizmodel.ResponseReady || response.Question.Required {
		state.ContinuationError = "Clarification still needs an answer. Continue the quiz before finishing."

		return core.QuizTransition{State: state, Message: state.ContinuationError}, nil
	}

	var entry *quizcore.Entry

	if state.ActiveQuestion != nil {
		skipped := quizEntryFromState(state, quizcore.EntryStatusSkipped, "", s.utcNow())
		skipped.ID = fmt.Sprintf("%s-entry-%03d", state.SessionID, state.QuestionsAsked)
		state = updateQuizStateAfterEntry(state, skipped)
		entry = &skipped
	}

	state, err = s.completeGeneratedQuiz(state)

	return core.QuizTransition{State: state, Entry: entry, Message: "Quiz completed. The saved context is sufficient."}, err
}
