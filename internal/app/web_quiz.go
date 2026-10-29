package app

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"net/url"
	"strings"
	"time"

	quizcore "forge.adrianpk.com/chr/internal/core/quiz"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
	"forge.adrianpk.com/chr/internal/webui"
)

func quizFocusURL(reference string) string {
	return "/quizzes/detail?id=" + url.QueryEscape(reference)
}

func webQuizScope(operations service, id string) (QuizView, error) {
	quizzes, err := operations.ListQuizzes()
	if err != nil {
		return QuizView{}, err
	}

	for _, quiz := range quizzes.Views {
		if quiz.CuratedPath == id {
			return quiz, nil
		}
	}

	return QuizView{}, fmt.Errorf("quiz scope not found")
}

func (p webViewProvider) quizArtifactDetail(ctx context.Context, request webui.ArtifactDetailRequest) (webui.ArtifactDetailView, error) {
	operations := p.service(ctx)

	quiz, err := webQuizScope(operations, request.ID)
	if err != nil {
		return webui.ArtifactDetailView{}, err
	}

	observed, err := operations.quizRepository().Observe(ctx, quiz.Scope)
	if err != nil {
		return webui.ArtifactDetailView{}, err
	}

	view := webQuizAnswerView(quiz, observed)
	view.DraftContext = webStableRevision(struct{ Root, Scope string }{operations.projectContext.ProjectDir, request.ID})

	return webui.ArtifactDetailView{Family: "quizzes", ID: request.ID, Title: quiz.Title, Scope: webScopeLabel(quiz.Scope.CaseSlug, quiz.Title), Status: view.Status, Revision: observed.Revision, Quiz: view}, nil
}

func webQuizAnswerView(quiz QuizView, observed quizcore.Observation) *webui.QuizAnswerView {
	state := observed.State
	if observed.Exists {
		quiz.Mode, quiz.Status = state.Mode, string(state.Status)
	}

	view := &webui.QuizAnswerView{ScopeSlug: quiz.Scope.CaseSlug, SessionID: state.SessionID, Revision: observed.Revision, OperationID: rand.Text(), Mode: string(quiz.Mode), Status: quiz.Status, Started: observed.Exists, Error: state.ContinuationError}
	view.Open = observed.Exists && state.Status != quizcore.RuntimeStatusDone && state.Status != quizcore.RuntimeStatusCancelled
	view.CanAnswer = view.Open && state.Status == quizcore.RuntimeStatusActive && state.ActiveQuestion != nil

	view.CanResume = webQuizCanResume(state, observed.Exists)
	view.CanStart = !observed.Exists || state.Status == quizcore.RuntimeStatusDone

	if state.ActiveQuestion != nil {
		view.QuestionID, view.Question = state.ActiveQuestion.ID, state.ActiveQuestion.Prompt
	}

	view.History = webQuizHistory(observed.Entries)

	return view
}

func webQuizHistory(entries []quizcore.Entry) []webui.QuizAnswerHistoryView {
	answered := quizAnswerHistory(entries)

	history := make([]webui.QuizAnswerHistoryView, 0, len(answered))
	for _, entry := range answered {
		history = append(history, webui.QuizAnswerHistoryView{ID: entry.ID, Question: entry.Question, Answer: entry.Answer, AnsweredAt: entry.AnsweredAt.UTC().Format(time.RFC3339)})
	}

	return history
}

func (p webViewProvider) writeQuizAction(operations service, request webui.WriteActionRequest) (webui.WriteActionResponse, error) {
	if request.Fields["guided_run"] != "" {
		return p.writeGuidedQuizAction(operations, request)
	}

	scope, err := webQuizActionScope(operations, request)
	if err != nil {
		return webQuizWriteFailure(request.ID, err), nil
	}

	binding := quizcore.ActionBinding{SessionID: request.Fields["session_id"], QuestionID: request.Fields["question_id"], ExpectedRevision: request.Revision, OperationID: request.Fields["request_id"]}
	if binding.ExpectedRevision == "" || binding.OperationID == "" {
		return webValidationWrite(fmt.Errorf("complete quiz action binding is required")), nil
	}

	action := strings.TrimPrefix(request.Action, "quiz.")
	if !strings.HasPrefix(request.Action, "quiz.") {
		return webBlockedWrite("Unsupported quiz action."), nil
	}

	answer := ""
	if action == "answer" {
		answer = strings.TrimSpace(request.Fields["answer"])
	}

	result, err := operations.applyQuizAction(operations.ctx, quizcore.ActionRequest{Scope: scope, Binding: binding, Action: action, Mode: quizcore.Mode(request.Fields["mode"]), Answer: answer})
	if err != nil {
		return webQuizWriteFailure(request.ID, err), nil
	}

	location := quizFocusURL(request.ID)
	if action == "done" || action == "cancel" {
		location = "/quizzes"
	}

	return webui.WriteActionResponse{
		Status: webui.WriteStatusOK, Summary: result.Message, RedirectURL: location,
		Quiz: &webui.QuizActionAcknowledgement{Action: action, SessionID: binding.SessionID, QuestionID: binding.QuestionID},
	}, nil
}

func webQuizWriteFailure(id string, err error) webui.WriteActionResponse {
	if errors.Is(err, workflowstate.ErrConflict) {
		return webui.WriteActionResponse{Status: webui.WriteStatusStale, Summary: "This question changed. Your draft is retained; reload the current question.", ReloadURL: quizFocusURL(id)}
	}

	result := webValidationWrite(err)
	result.ReloadURL = quizFocusURL(id)

	return result
}

func webQuizActionScope(operations service, request webui.WriteActionRequest) (quizcore.Scope, error) {
	scope := operations.quizScope(quizcore.Scope{CaseSlug: request.Fields["case_slug"]})

	paths, err := operations.quizRepository().Paths(scope)
	if err != nil {
		return scope, err
	}

	if paths.CuratedRelativePath != request.ID {
		return scope, workflowstate.ErrConflict
	}

	if request.Action != "quiz.start" {
		return scope, nil
	}

	scopes, err := operations.quizReadScopes()
	if err != nil {
		return scope, err
	}

	for _, known := range scopes {
		if known.scope == scope {
			return scope, nil
		}
	}

	return scope, fmt.Errorf("quiz scope not found")
}

func webQuizCanResume(state quizcore.RuntimeState, exists bool) bool {
	if !exists || state.Status == quizcore.RuntimeStatusActive || state.Status == quizcore.RuntimeStatusDone || state.QuestionsAsked >= state.QuestionBudget {
		return false
	}

	if state.Status == quizcore.RuntimeStatusCancelled || state.ContinuationPending {
		return true
	}

	return state.Mode != quizcore.ModeFast || state.FastBatchID == "" || len(state.FastQueue) != 0
}
