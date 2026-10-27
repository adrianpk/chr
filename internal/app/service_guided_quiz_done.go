package app

import (
	"fmt"

	quizcore "forge.adrianpk.com/chr/internal/core/quiz"
	quizmodel "forge.adrianpk.com/chr/internal/core/quiz/model"
)

// Done asks the same clarification owner whether the saved answers suffice.
// A user control alone cannot grant the trusted completion required by sealing.
func (s service) prepareGuidedQuizDone(observed quizcore.Observation) (quizcore.Transition, error) {
	state := observed.State
	if state.AnswersRecorded == 0 {
		state.ContinuationError = "Answer at least one question before finishing clarification."

		return quizcore.Transition{State: state, Message: state.ContinuationError}, nil
	}

	check := state
	check.Mode = quizcore.ModeDeep

	request, err := s.guidedQuizModelRequest(check, observed.Entries)
	if err != nil {
		return quizcore.Transition{}, err
	}

	response, err := s.invokeQuizGeneration(check, request)
	if err != nil {
		return quizcore.Transition{}, err
	}

	if response.Status != quizmodel.ResponseReady || response.Question.Required {
		state.ContinuationError = "Clarification still needs an answer. Continue the quiz before finishing."

		return quizcore.Transition{State: state, Message: state.ContinuationError}, nil
	}

	var entry *quizcore.Entry

	if state.ActiveQuestion != nil {
		skipped := quizcore.EntryFromState(state, quizcore.EntryStatusSkipped, "", s.utcNow())
		skipped.ID = fmt.Sprintf("%s-entry-%03d", state.SessionID, state.QuestionsAsked)
		state = quizcore.StateAfterEntry(state, skipped)
		entry = &skipped
	}

	state, err = s.completeGeneratedQuiz(state)

	return quizcore.Transition{State: state, Entry: entry, Message: "Quiz completed. The saved context is sufficient."}, err
}
