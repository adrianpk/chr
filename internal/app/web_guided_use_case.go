package app

import (
	"context"
	"crypto/rand"
	"errors"
	"net/url"
	"strconv"

	"forge.adrianpk.com/chr/internal/core"
	"forge.adrianpk.com/chr/internal/webui"
)

// GuidedUseCaseView projects the same shared run used by other adapters.
func (p webViewProvider) GuidedUseCaseView(ctx context.Context, id, namespace string) (webui.GuidedUseCaseView, error) {
	result, err := p.service(ctx).InspectGuidedUseCase(ctx, id, namespace)

	view := webui.GuidedUseCaseView{Enabled: true, Namespace: result.Namespace, RequestID: rand.Text(), Policy: result.Policy, Problems: result.Problems, CommittedDump: result.CommittedDump, Question: result.Question}
	if err != nil {
		return view, err
	}

	for _, session := range result.Runs {
		view.Runs = append(view.Runs, webui.GuidedUseCaseLink{Name: session.Topic, URL: guidedSessionURL(session), CaseSlug: session.Guided.CaseSlug})
	}

	if result.Session == nil {
		return view, nil
	}

	session := result.Session
	state := session.Guided
	view.RunID, view.Name, view.Stage, view.Status = session.ID, session.Topic, state.Stage, state.Status
	view.CaseSlug = state.CaseSlug
	projectGuidedWebQuiz(&view, result)

	view.Revision = strconv.FormatInt(session.Revision, 10)
	view.Observation = view.Revision

	if host, ok := p.config.guidedHost.(*guidedWorkHost); ok {
		host.mu.Lock()
		view.Processing = host.progress[guidedProgressKey(p.service(ctx), core.GuidedUseCaseScope{RunID: session.ID, ProjectID: session.ProjectID, Namespace: session.Namespace})]
		host.mu.Unlock()
	}

	view.Capture, view.Blocker = state.Capture, state.Blocker
	view.Artifacts = guidedWebDocuments(result)

	guidedWebInputBinding(&view, state)
	guidedWebWork(&view, result.Work)
	guidedWebControls(&view, *session, result.Work)
	guidedWebIdeas(&view, result)
	guidedWebResult(&view, result)

	if len(state.Checkpoints) > 0 {
		view.Commit = state.Checkpoints[0].Commit
	}

	return view, nil
}

func (p webViewProvider) writeGuidedUseCase(operations service, request webui.WriteActionRequest) (webui.WriteActionResponse, error) {
	if request.Action != "guided.start" && request.Action != "guided.input" {
		return webBlockedWrite("Unsupported guided action."), nil
	}

	var (
		session core.AuthoringSession
		err     error
	)

	namespace := request.Fields["namespace"]
	if request.Action == "guided.start" {
		session, err = startGuidedWebUseCase(operations, request)
	} else {
		var command core.GuidedUseCaseCommand

		command, err = guidedWebCommand(request)
		if err == nil {
			session, err = operations.ChangeGuidedUseCase(operations.ctx, command)
		}
	}

	var prerequisite *core.GuidedPrerequisiteError
	if errors.Is(err, core.ErrWorkflowConflict) && !errors.As(err, &prerequisite) {
		return webui.WriteActionResponse{Status: webui.WriteStatusStale, Summary: "The use case changed. Reload to continue.", ReloadURL: guidedUseCaseURL(request.ID, namespace)}, nil
	}

	if err != nil {
		if guidedWebPending(session) {
			return webui.WriteActionResponse{Status: webui.WriteStatusOK, Summary: "Use case retained; progression needs recovery.", RedirectURL: guidedSessionURL(session)}, nil
		}

		return webValidationWrite(err), nil
	}

	summary := "Use case saved."
	if request.Fields["step"] == "idea" {
		summary = "Idea saved for review after completion; the agreed scope is unchanged."
	}

	return webui.WriteActionResponse{Status: webui.WriteStatusOK, Summary: summary, RedirectURL: guidedSessionURL(session)}, nil
}

func guidedUseCaseURL(id, namespace string) string {
	return "/cases/workflow?" + url.Values{"id": {id}, "namespace": {namespace}}.Encode()
}

func guidedWebInputBinding(view *webui.GuidedUseCaseView, state *core.GuidedUseCaseState) {
	if state.Quiz != nil {
		view.QuestionID, view.InputDigest = state.Quiz.QuestionID, state.Quiz.InputDigest
	}

	if state.Stage == "quiz" && view.Question == "" {
		view.RetryAction = "quiz"
	}

	if state.Stage == "specification" {
		view.RetryAction = "spec"
	}

	if state.Stage == "preparation" {
		view.RetryAction = "prepare"

		for _, checkpoint := range state.Checkpoints {
			if checkpoint.Stage == "preparation" {
				view.RetryAction = "admit"
			}
		}
	}

	if pending := state.PendingOperation; pending != nil {
		view.PendingAction, view.PendingContent = pending.Action, pending.Content
		view.QuestionID, view.InputDigest = pending.QuestionID, pending.InputDigest
		view.RequestID, view.Revision = pending.ID, strconv.FormatInt(pending.ExpectedRevision, 10)
	}
}

func startGuidedWebUseCase(operations service, request webui.WriteActionRequest) (core.AuthoringSession, error) {
	if request.Fields["step"] != "name" {
		return operations.StartGuidedUseCase(operations.ctx, request.Fields["request_id"], request.Fields["namespace"], "local")
	}

	return operations.StartNamedGuidedUseCase(operations.ctx, core.GuidedUseCaseStart{GuidedUseCaseScope: core.GuidedUseCaseScope{RunID: request.Fields["request_id"], Namespace: request.Fields["namespace"]}, Actor: "local"}, request.Fields["content"])
}
