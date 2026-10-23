package app

import (
	"strconv"
	"strings"

	"forge.adrianpk.com/chr/internal/core"
	"forge.adrianpk.com/chr/internal/webui"
)

func projectGuidedWebQuiz(view *webui.GuidedUseCaseView, observed GuidedUseCaseInspection) {
	if observed.Quiz == nil {
		return
	}

	quiz := QuizView{Scope: core.QuizScope{ProjectID: observed.Session.ProjectID, Namespace: observed.Namespace, CaseSlug: observed.Session.Guided.CaseSlug}, Mode: core.QuizModeDeep}
	view.Quiz = webQuizAnswerView(quiz, *observed.Quiz)
	view.Quiz.DraftContext = observed.Namespace + ":" + observed.Session.ID

	view.QuizID = observed.Quiz.State.CuratedPath
}

// The shared form retains both owners' observed bindings. Only the application
// workflow owner may advance a guided quiz or its subsequent stages.
func (p webViewProvider) writeGuidedQuizAction(operations service, request webui.WriteActionRequest) (webui.WriteActionResponse, error) {
	observed, err := operations.InspectGuidedUseCase(operations.ctx, request.Fields["guided_run"], request.Fields["namespace"])
	if err != nil {
		return webQuizWriteFailure(request.ID, err), nil
	}

	if observed.Session == nil || observed.Quiz == nil {
		return webQuizWriteFailure(request.ID, core.ErrWorkflowConflict), nil
	}

	view := webui.GuidedUseCaseView{}
	projectGuidedWebQuiz(&view, observed)

	if request.ID != view.QuizID || request.Fields["session_id"] != observed.Quiz.State.SessionID || request.Revision != observed.Quiz.Revision || request.Fields["case_slug"] != observed.Session.Guided.CaseSlug {
		return webGuidedQuizFailure(observed, core.ErrWorkflowConflict), nil
	}

	command, err := guidedQuizWebCommand(request)
	if err != nil {
		return webValidationWrite(err), nil
	}

	session, err := operations.ChangeGuidedUseCase(operations.ctx, command)
	if err != nil {
		return webGuidedQuizFailure(observed, err), nil
	}

	return webui.WriteActionResponse{
		Status: webui.WriteStatusOK, Summary: "Quiz updated.", RedirectURL: guidedSessionURL(session),
		Quiz: &webui.QuizActionAcknowledgement{Action: strings.TrimPrefix(request.Action, "quiz."), SessionID: request.Fields["session_id"], QuestionID: command.QuestionID},
	}, nil
}

func guidedQuizWebCommand(request webui.WriteActionRequest) (core.GuidedUseCaseCommand, error) {
	command := core.GuidedUseCaseCommand{
		GuidedUseCaseScope: core.GuidedUseCaseScope{RunID: request.Fields["guided_run"], Namespace: request.Fields["namespace"]},
		OperationID:        request.Fields["request_id"], QuestionID: request.Fields["question_id"], InputDigest: request.Fields["input_digest"],
		Action: strings.TrimPrefix(request.Action, "quiz."),
	}
	switch request.Action {
	case "quiz.answer":
		command.Action, command.Content = "answer", request.Fields["answer"]
	case "quiz.mode":
		command.Action, command.Content = "quiz-mode", request.Fields["mode"]
	case "quiz.skip":
		command.Action = "quiz-skip"
	case "quiz.done":
		command.Action = "quiz-done"
	case "quiz.resume":
		command.Action = "quiz"
	case "quiz.cancel":
		command.Action = "cancel"
	default:
		return command, core.ErrWorkflowConflict
	}

	var err error

	command.ExpectedRevision, err = strconv.ParseInt(request.Fields["guided_revision"], 10, 64)

	return command, err
}

func webGuidedQuizFailure(observed GuidedUseCaseInspection, err error) webui.WriteActionResponse {
	result := webQuizWriteFailure("", err)
	result.ReloadURL = guidedSessionURL(*observed.Session)

	return result
}
