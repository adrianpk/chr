package app

import (
	quizcore "forge.adrianpk.com/chr/internal/core/quiz"
	"forge.adrianpk.com/chr/internal/tui"
)

func tuiGuidedQuiz(observed GuidedUseCaseInspection) *tui.QuizView {
	if observed.Quiz == nil {
		return nil
	}

	state := observed.Quiz.State

	view := &tui.QuizView{
		Title:  observed.Session.Guided.CaseName,
		Scope:  tui.QuizScope{CaseSlug: observed.Session.Guided.CaseSlug},
		Status: string(state.Status), Mode: string(state.Mode),
		Binding:        tui.QuizActionBinding{SessionID: state.SessionID, QuestionID: observed.Session.Guided.Quiz.QuestionID, ExpectedRevision: observed.Quiz.Revision},
		History:        tuiQuizHistory(quizAnswerHistory(observed.Quiz.Entries)),
		QuestionsAsked: state.QuestionsAsked, AnswersRecorded: state.AnswersRecorded, SkipsRecorded: state.SkipsRecorded,
	}
	if state.ActiveQuestion != nil {
		view.ActiveQuestionSummary = observed.Question
	}

	if state.Status == quizcore.RuntimeStatusDone {
		view.Status = "done"
	}

	return view
}
