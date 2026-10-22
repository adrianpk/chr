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

	view := webui.GuidedUseCaseView{Enabled: true, Namespace: result.Namespace, RequestID: rand.Text(), Policy: result.Policy, Problems: result.Problems, CommittedDump: result.CommittedDump}
	if err != nil {
		return view, err
	}

	for _, session := range result.Runs {
		view.Runs = append(view.Runs, webui.GuidedUseCaseLink{Name: session.Topic, URL: guidedUseCaseURL(session.ID, session.Namespace)})
	}

	if result.Session == nil {
		return view, nil
	}

	session := result.Session
	state := session.Guided
	view.RunID, view.Name, view.Stage, view.Status = session.ID, session.Topic, state.Stage, state.Status
	view.Revision = strconv.FormatInt(session.Revision, 10)

	view.Capture = state.Capture
	if len(state.Checkpoints) > 0 {
		view.Commit = state.Checkpoints[0].Commit
	}

	if pending := state.PendingOperation; pending != nil {
		view.PendingAction, view.PendingContent = pending.Action, pending.Content
		view.RequestID, view.Revision = pending.ID, strconv.FormatInt(pending.ExpectedRevision, 10)
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
		session, err = operations.StartGuidedUseCase(operations.ctx, request.Fields["request_id"], namespace, "local")
	} else {
		var revision int64

		revision, err = strconv.ParseInt(request.Revision, 10, 64)
		if err == nil {
			session, err = operations.ChangeGuidedUseCase(operations.ctx, core.GuidedUseCaseCommand{
				GuidedUseCaseScope: core.GuidedUseCaseScope{RunID: request.ID, Namespace: namespace}, ExpectedRevision: revision,
				OperationID: request.Fields["request_id"], Action: request.Fields["step"], Content: request.Fields["content"],
			})
		}
	}

	if errors.Is(err, core.ErrWorkflowConflict) {
		return webui.WriteActionResponse{Status: webui.WriteStatusStale, Summary: "The use case changed. Reload to continue.", ReloadURL: guidedUseCaseURL(request.ID, namespace)}, nil
	}

	if err != nil {
		return webValidationWrite(err), nil
	}

	return webui.WriteActionResponse{Status: webui.WriteStatusOK, Summary: "Use case saved.", RedirectURL: guidedUseCaseURL(session.ID, session.Namespace)}, nil
}

func guidedUseCaseURL(id, namespace string) string {
	return "/guided?" + url.Values{"id": {id}, "namespace": {namespace}}.Encode()
}
