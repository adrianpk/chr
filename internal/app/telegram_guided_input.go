package app

import (
	"encoding/json"
	"strings"

	"forge.adrianpk.com/chr/internal/core"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

type telegramGuidedBinding struct {
	ViewDigest     string `json:"view_digest,omitempty"`
	Revision       int64  `json:"revision"`
	NativeRevision uint64 `json:"native_revision,omitempty"`
	Stage          string `json:"stage"`
	QuestionID     string `json:"question_id,omitempty"`
	InputDigest    string `json:"input_digest,omitempty"`
}

// These explicit phrases carry the user's development instruction. A model
// proposal or incidental mention of development cannot authorize a guided start.
func telegramGuidedStart(text string) bool {
	_, start := telegramGuidedStartProject(text)

	return start
}

func telegramGuidedStartProject(text string) (string, bool) {
	trimmed := strings.TrimSuffix(strings.TrimSpace(text), ".")
	switch strings.ToLower(trimmed) {
	case "develop a new use case", "start guided development", "guide use-case development":
		return "", true
	}

	const prefix = "develop a new use case in "
	if strings.HasPrefix(strings.ToLower(trimmed), prefix) && len(trimmed) > len(prefix) {
		return strings.TrimSpace(trimmed[len(prefix):]), true
	}

	return "", false
}

func telegramGuidedCommand(c core.TelegramConversation, update core.TelegramReceivedUpdate, actor string) (telegramGuidedInput, error) {
	var binding telegramGuidedBinding
	if err := decodeOperationJSON(c.QuestionBinding, &binding); err != nil {
		return telegramGuidedInput{}, err
	}

	command := core.GuidedUseCaseCommand{GuidedUseCaseScope: core.GuidedUseCaseScope{RunID: c.GuidedRunID, ProjectID: c.ProjectID, Namespace: c.Namespace}, ExpectedRevision: binding.Revision, NativeRevision: binding.NativeRevision, OperationID: core.TelegramOperationID(update, 0), Content: update.Text, QuestionID: binding.QuestionID, InputDigest: binding.InputDigest}
	input := telegramGuidedInput{Actor: actor, Command: command}
	text := strings.ToLower(strings.TrimSpace(update.Text))
	input.Document = telegramGuidedDocumentKind(text)
	input.Inspect = text == "show current step" || text == "show use case" || input.Document != ""
	input.Retry = text == "retry" || text == "continue"

	input.Offer = text == "show controls"
	if input.Inspect || input.Retry || input.Offer {
		input.Command.Content = ""

		return input, nil
	}

	action, content, err := telegramGuidedTextAction(binding, update.Text)
	input.Command.Action, input.Command.Content = action, content

	return input, err
}

func telegramGuidedTextAction(binding telegramGuidedBinding, text string) (string, string, error) {
	normalized := strings.ToLower(strings.TrimSpace(text))

	controls := map[string]string{"pause use case": "pause", "resume use case": "resume", "cancel use case": "cancel", "/pause": "pause", "/resume": "resume", "/cancel": "cancel"}
	if action := controls[normalized]; action != "" {
		return action, "", nil
	}

	if strings.HasPrefix(normalized, "idea:") {
		return "idea", strings.TrimSpace(strings.TrimSpace(text)[len("idea:"):]), nil
	}

	if strings.HasPrefix(normalized, "/") {
		return "", "", workflowstate.ErrConflict
	}

	return telegramGuidedStageAction(binding, text)
}

func telegramGuidedStageAction(binding telegramGuidedBinding, text string) (string, string, error) {
	switch binding.Stage {
	case "name":
		return "name", text, nil
	case "dump":
		if strings.EqualFold(strings.TrimSpace(text), "finish capture") {
			return "finish", "", nil
		}

		return "append", text, nil
	case "quiz":
		if binding.QuestionID != "" {
			return "answer", text, nil
		}
	}

	return "", "", workflowstate.ErrConflict
}

func telegramGuidedProgress(observed GuidedUseCaseInspection) (core.TelegramGuidedProgress, error) {
	session := *observed.Session

	binding := telegramGuidedBinding{Revision: session.Revision, Stage: session.Guided.Stage, ViewDigest: recordDigest([]byte(telegramGuidedPrompt(observed)))}
	if session.Guided.Quiz != nil {
		binding.QuestionID, binding.InputDigest = session.Guided.Quiz.QuestionID, session.Guided.Quiz.InputDigest
	}

	if observed.Work != nil {
		binding.NativeRevision = observed.Work.Revision
	}

	data, err := json.Marshal(binding)

	return core.TelegramGuidedProgress{SessionRef: core.DirectoryName + "/authoring/sessions/" + session.ID + "/session.md", Topic: session.Topic, Binding: string(data), Phase: telegramGuidedTransportPhase(observed)}, err
}

func telegramGuidedDocumentKind(text string) string {
	kinds := map[string]string{"show dump": "dump", "show quiz": "quiz", "show specification": "specification", "show plan": "plan", "show tracker": "tracker", "show results": "results"}

	return kinds[text]
}

func telegramGuidedTransportPhase(observed GuidedUseCaseInspection) string {
	if observed.Result != nil && observed.Result.Complete {
		return "completed"
	}

	switch observed.Session.Guided.Status {
	case "cancelled":
		return "canceled"
	case "failed":
		return "failed"
	default:
		return "collecting"
	}
}
