package app

import (
	"context"
	"errors"
	"strings"

	"forge.adrianpk.com/chr/internal/core"
)

type telegramGuidedRoutingFailure string

func (e telegramGuidedRoutingFailure) Error() string { return string(e) }

func telegramGuidedProjectName(projects []core.Project, name string) (string, error) {
	var matches []string

	for _, project := range projects {
		if project.Enabled && (strings.EqualFold(project.Name, name) || strings.EqualFold(project.ID, name)) {
			matches = append(matches, project.ID)
		}
	}

	if len(matches) != 1 {
		return "", telegramGuidedRoutingFailure("The project name is unavailable or ambiguous. Use a unique registered project name.")
	}

	return matches[0], nil
}

func telegramGuidedRunName(text string) (string, bool) {
	const prefix = "show use case "

	trimmed := strings.TrimSpace(text)
	if !strings.HasPrefix(strings.ToLower(trimmed), prefix) {
		return "", false
	}

	return strings.TrimSpace(trimmed[len(prefix):]), true
}

func (r telegramReceiver) selectTelegramGuidedName(ctx context.Context, input core.TelegramConversationInput, name string) (string, error) {
	sessions, err := r.store.Conversations(ctx, input.Update.Peer)
	if err != nil {
		return "", err
	}

	title, projectName, _ := strings.Cut(name, " in project ")

	var matches []string

	for _, c := range sessions {
		if c.GuidedRunID == "" || !strings.EqualFold(c.Topic, title) {
			continue
		}

		if projectName != "" {
			project, err := telegramGuidedProjectName(input.Projects, projectName)
			if err != nil {
				return "", err
			}

			if project != c.ProjectID {
				continue
			}
		}

		matches = append(matches, c.ID)
	}

	if len(matches) != 1 {
		return "", telegramGuidedRoutingFailure("The use-case name is unavailable or ambiguous here. Use a unique name, optionally followed by in project and its registered project name. Saved inputs are unchanged.")
	}

	return matches[0], nil
}

func telegramGuidedStartFailure(failure error, projects []core.Project) error {
	if !errors.Is(failure, core.ErrTelegramProjectSelectionAmbiguous) {
		return failure
	}

	var names []string

	for _, project := range projects {
		if project.Enabled {
			names = append(names, project.Name)
		}
	}

	return telegramGuidedRoutingFailure("Choose a project: " + strings.Join(names, ", ") + ". Say Develop a new use case in followed by its unique project name.")
}

func (r telegramReceiver) resolveTelegramGuidedInput(ctx context.Context, user core.TelegramAllowedUser, config core.TelegramReceiverConfig, update core.TelegramReceivedUpdate) (core.TelegramConversationInput, bool, error) {
	input := core.TelegramConversationInput{Update: update, Config: config, Text: "New guided use case"}
	projects, err := r.operations.ListRegisteredProjects(user.ProjectIDs)

	input.Projects = projects
	if err != nil {
		return input, false, err
	}

	project, start := telegramGuidedStartProject(update.Text)
	if project != "" {
		input.Project, err = telegramGuidedProjectName(projects, project)
		if err != nil {
			return input, start, err
		}
	}

	if name, selected := telegramGuidedRunName(update.Text); selected {
		input.Conversation, err = r.selectTelegramGuidedName(ctx, input, name)
	}

	return input, start, err
}
