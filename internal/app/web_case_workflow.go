package app

import (
	"context"
	"errors"
	"net/url"

	authoringcore "forge.adrianpk.com/chr/internal/core/authoring"
	"forge.adrianpk.com/chr/internal/webui"
)

// CaseWorkflowView resolves a named use case without adopting manual artifacts.
func (p webViewProvider) CaseWorkflowView(ctx context.Context, slug, namespace string) (webui.GuidedUseCaseView, error) {
	operations := p.service(ctx)

	scope, err := operations.guidedUseCaseScope("", namespace)
	if err != nil {
		return webui.GuidedUseCaseView{}, err
	}

	sessions, err := authoringcore.ListGuidedUseCases(ctx, operations.projectContext.ProjectDir, scope.ProjectID, scope.Namespace)
	if err != nil {
		return webui.GuidedUseCaseView{}, err
	}

	id := ""

	for _, session := range sessions {
		if session.Guided.CaseSlug == slug {
			if id != "" {
				return webui.GuidedUseCaseView{}, errors.New("multiple workflows belong to this use case")
			}

			id = session.ID
		}
	}

	if id == "" {
		return webui.GuidedUseCaseView{}, errors.New("no workflow belongs to this use case")
	}

	return p.GuidedUseCaseView(ctx, id, scope.Namespace)
}

func guidedSessionURL(session authoringcore.Session) string {
	if session.Guided == nil || session.Guided.CaseSlug == "" {
		return guidedUseCaseURL(session.ID, session.Namespace)
	}

	return webui.GuidedUseCasePath(session.Guided.Stage) + "?" + url.Values{"case": {session.Guided.CaseSlug}, "namespace": {session.Namespace}}.Encode()
}
