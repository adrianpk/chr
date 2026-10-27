package app

import (
	"context"
	"strings"
	"time"

	"forge.adrianpk.com/chr/internal/core"
	telegramcore "forge.adrianpk.com/chr/internal/core/telegram"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

func (r telegramReceiver) attachTelegramGuidedName(ctx context.Context, input telegramcore.ConversationInput, title, projectName string) (string, error) {
	operations, ok := r.operations.(telegramGuidedService)
	if !ok {
		return "", ErrRecordUnsupported
	}

	if projectName != "" {
		project, err := telegramGuidedProjectName(input.Projects, projectName)
		if err != nil {
			return "", err
		}

		input.Project = project
	}

	target, err := r.store.ResolveOperationTarget(ctx, input, time.Now())
	if err != nil {
		return "", telegramGuidedStartFailure(err, input.Projects)
	}

	observed, err := operations.InspectTelegramGuided(ctx, target, "")
	if err != nil {
		return "", err
	}

	var matches []core.AuthoringSession

	for _, session := range observed.Runs {
		if strings.EqualFold(session.Topic, title) {
			matches = append(matches, session)
		}
	}

	if len(matches) != 1 {
		return "", telegramGuidedRoutingFailure("The guided use-case name is unavailable or ambiguous in the selected project. Use a unique guided name, optionally followed by in project and its registered project name. Saved work is unchanged.")
	}

	association, err := telegramGuidedAssociation(matches[0])
	if err != nil {
		return "", err
	}

	input.Project, input.Namespace = target.ProjectID, target.Namespace
	conversation, _, err := r.store.BindGuidedConversation(ctx, input, association, time.Now())

	return conversation.ID, err
}

func telegramGuidedAssociation(session core.AuthoringSession) (telegramcore.GuidedAssociation, error) {
	if session.Guided == nil || core.ValidateGuidedUseCase(session) != nil {
		return telegramcore.GuidedAssociation{}, workflowstate.ErrConflict
	}

	return telegramcore.GuidedAssociation{SessionID: session.ID, ProjectID: session.ProjectID, Namespace: session.Namespace, Topic: session.Topic}, nil
}
