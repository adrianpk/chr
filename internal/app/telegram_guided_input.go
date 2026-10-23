package app

import (
	"encoding/json"
	"strings"

	"forge.adrianpk.com/chr/internal/core"
)

type telegramGuidedBinding struct {
	Revision    int64  `json:"revision"`
	Stage       string `json:"stage"`
	QuestionID  string `json:"question_id,omitempty"`
	InputDigest string `json:"input_digest,omitempty"`
}

// These explicit phrases carry the user's development instruction. A model
// proposal or incidental mention of development cannot authorize a guided start.
func telegramGuidedStart(text string) bool {
	switch strings.TrimSuffix(strings.ToLower(strings.TrimSpace(text)), ".") {
	case "develop a new use case", "start guided development", "guide use-case development":
		return true
	default:
		return false
	}
}

func telegramGuidedCommand(c core.TelegramConversation, update core.TelegramReceivedUpdate, actor string) (telegramGuidedInput, error) {
	if strings.HasPrefix(strings.TrimSpace(update.Text), "/") {
		return telegramGuidedInput{}, core.ErrWorkflowConflict
	}

	var binding telegramGuidedBinding
	if err := decodeOperationJSON(c.QuestionBinding, &binding); err != nil {
		return telegramGuidedInput{}, err
	}

	id, err := telegramAuthoringSessionID(c.AuthoringSessionRef)
	if err != nil {
		return telegramGuidedInput{}, err
	}

	command := core.GuidedUseCaseCommand{GuidedUseCaseScope: core.GuidedUseCaseScope{RunID: id, ProjectID: c.ProjectID, Namespace: c.Namespace}, ExpectedRevision: binding.Revision, OperationID: core.TelegramOperationID(update, 0), Content: update.Text, QuestionID: binding.QuestionID, InputDigest: binding.InputDigest}
	switch binding.Stage {
	case "name":
		command.Action = "name"
	case "dump":
		command.Action = "append"
		if strings.EqualFold(strings.TrimSpace(update.Text), "finish capture") {
			command.Action, command.Content = "finish", ""
		}
	case "quiz":
		command.Action = "answer"

		if binding.QuestionID == "" {
			return telegramGuidedInput{}, core.ErrWorkflowConflict
		}
	default:
		return telegramGuidedInput{}, core.ErrWorkflowConflict
	}

	return telegramGuidedInput{Actor: actor, Command: command}, nil
}

func telegramGuidedProgress(session core.AuthoringSession, ref string) (core.TelegramGuidedProgress, error) {
	binding := telegramGuidedBinding{Revision: session.Revision, Stage: session.Guided.Stage}
	if session.Guided.Quiz != nil {
		binding.QuestionID, binding.InputDigest = session.Guided.Quiz.QuestionID, session.Guided.Quiz.InputDigest
	}

	data, err := json.Marshal(binding)

	return core.TelegramGuidedProgress{SessionRef: ref, Topic: session.Topic, Binding: string(data), Phase: "collecting"}, err
}
