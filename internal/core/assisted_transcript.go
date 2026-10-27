package core

import (
	"bytes"
	"fmt"
	"strconv"
	"strings"
	"time"

	opsmetadata "forge.adrianpk.com/chr/internal/core/ops/metadata"
	ticketmetadata "forge.adrianpk.com/chr/internal/core/ticket/metadata"
)

// AssistedTranscriptEventType defines a package boundary type.
type AssistedTranscriptEventType string

const (
	// AssistedTranscriptEventQuestion defines a package boundary constant.
	AssistedTranscriptEventQuestion AssistedTranscriptEventType = "question"
	// AssistedTranscriptEventAnswer defines a package boundary constant.
	AssistedTranscriptEventAnswer AssistedTranscriptEventType = "answer"
	// AssistedTranscriptEventResult defines a package boundary constant.
	AssistedTranscriptEventResult AssistedTranscriptEventType = "result"
	// AssistedTranscriptEventState defines a package boundary constant.
	AssistedTranscriptEventState AssistedTranscriptEventType = "state"
)

// AssistedTranscriptEvent describes a package boundary value.
type AssistedTranscriptEvent struct {
	ID        string
	SessionID string
	TurnID    string
	TurnIndex int
	Revision  int64
	Type      AssistedTranscriptEventType
	Intent    AssistedAuthoringIntent
	Status    AssistedAuthoringTurnStatus
	Question  AssistedAuthoringQuestion
	Answer    AssistedAuthoringAnswer
	Result    AssistedAuthoringTurnResult
	CreatedAt time.Time
}

// ValidateAssistedTranscriptEvent validates the corresponding package boundary value.
func ValidateAssistedTranscriptEvent(event AssistedTranscriptEvent) error {
	if event.ID == "" {
		return fmt.Errorf("assisted transcript event id is required")
	}

	if event.SessionID == "" {
		return fmt.Errorf("assisted transcript event session id is required")
	}

	if event.TurnID == "" {
		return fmt.Errorf("assisted transcript event turn id is required")
	}

	if event.TurnIndex <= 0 {
		return fmt.Errorf("assisted transcript event turn index must be greater than zero")
	}

	if !event.Type.Valid() {
		return fmt.Errorf("assisted transcript event type is invalid: %s", event.Type)
	}

	if !event.Intent.Valid() {
		return fmt.Errorf("assisted transcript event intent is invalid: %s", event.Intent)
	}

	if !event.Status.Valid() {
		return fmt.Errorf("assisted transcript event status is invalid: %s", event.Status)
	}

	err := validateAssistedTranscriptPayload(event)
	if err != nil {
		return err
	}

	if event.CreatedAt.IsZero() {
		return fmt.Errorf("assisted transcript event created time is required")
	}

	return nil
}

// Valid reports whether the value is one of the supported constants.
func (t AssistedTranscriptEventType) Valid() bool {
	switch t {
	case AssistedTranscriptEventQuestion,
		AssistedTranscriptEventAnswer,
		AssistedTranscriptEventResult,
		AssistedTranscriptEventState:
		return true
	default:
		return false
	}
}

// AssistedTranscriptEventsFromTurn executes the corresponding package operation.
func AssistedTranscriptEventsFromTurn(turn AssistedAuthoringTurn) ([]AssistedTranscriptEvent, error) {
	err := ValidateAssistedAuthoringTurn(turn)
	if err != nil {
		return nil, err
	}

	events := []AssistedTranscriptEvent{}

	if turn.Status == AssistedAuthoringTurnPending {
		events = append(events, assistedTranscriptEventFromTurn(turn, AssistedTranscriptEventState))

		return events, nil
	}

	events = append(events, assistedTranscriptEventFromTurn(turn, AssistedTranscriptEventQuestion))

	if turn.Status == AssistedAuthoringTurnAnswered {
		events = append(events, assistedTranscriptEventFromTurn(turn, AssistedTranscriptEventAnswer))
	}

	if turn.Status == AssistedAuthoringTurnCompleted || turn.Status == AssistedAuthoringTurnBlocked {
		events = append(events, assistedTranscriptEventFromTurn(turn, AssistedTranscriptEventResult))
	}

	return events, nil
}

// RenderAssistedTranscriptEvent renders the corresponding package data.
func RenderAssistedTranscriptEvent(event AssistedTranscriptEvent) (string, error) {
	err := ValidateAssistedTranscriptEvent(event)
	if err != nil {
		return "", err
	}

	var output bytes.Buffer

	renderAssistedTranscriptEventFrontmatter(&output, event)
	renderAssistedTranscriptEventBody(&output, event)

	return output.String(), nil
}

// ParseAssistedTranscriptEvents parses the corresponding package data.
func ParseAssistedTranscriptEvents(content string) ([]AssistedTranscriptEvent, error) {
	blocks := splitAssistedTranscriptBlocks(content)
	events := make([]AssistedTranscriptEvent, 0, len(blocks))

	for _, block := range blocks {
		event, err := parseAssistedTranscriptEvent(block)
		if err != nil {
			return nil, err
		}

		events = append(events, event)
	}

	return events, nil
}

func validateAssistedTranscriptPayload(event AssistedTranscriptEvent) error {
	switch event.Type {
	case AssistedTranscriptEventQuestion:
		return validateAssistedAuthoringQuestion(event.Question)
	case AssistedTranscriptEventAnswer:
		return validateAssistedAuthoringAnswer(event.Question, event.Answer)
	case AssistedTranscriptEventResult:
		return validateAssistedAuthoringTurnResult(event.Status, event.Result)
	case AssistedTranscriptEventState:
		return nil
	default:
		return nil
	}
}

func assistedTranscriptEventFromTurn(
	turn AssistedAuthoringTurn,
	eventType AssistedTranscriptEventType,
) AssistedTranscriptEvent {
	return AssistedTranscriptEvent{
		ID:        fmt.Sprintf("%s-%s", turn.ID, eventType),
		SessionID: turn.SessionID,
		TurnID:    turn.ID,
		TurnIndex: turn.Index,
		Type:      eventType,
		Intent:    turn.Intent,
		Status:    turn.Status,
		Question:  turn.Question,
		Answer:    turn.Answer,
		Result:    turn.Result,
		CreatedAt: turn.UpdatedAt,
	}
}

func renderAssistedTranscriptEventFrontmatter(output *bytes.Buffer, event AssistedTranscriptEvent) {
	output.WriteString("---\n")
	output.WriteString("id: " + opsmetadata.Quote(event.ID) + "\n")
	output.WriteString("session_id: " + opsmetadata.Quote(event.SessionID) + "\n")
	output.WriteString("turn_id: " + opsmetadata.Quote(event.TurnID) + "\n")
	output.WriteString("turn_index: " + strconv.Itoa(event.TurnIndex) + "\n")

	if event.Revision > 0 {
		output.WriteString("revision: " + strconv.FormatInt(event.Revision, 10) + "\n")
	}

	output.WriteString("type: " + opsmetadata.Quote(string(event.Type)) + "\n")
	output.WriteString("intent: " + opsmetadata.Quote(string(event.Intent)) + "\n")
	output.WriteString("status: " + opsmetadata.Quote(string(event.Status)) + "\n")
	output.WriteString("created_at: " + opsmetadata.Timestamp(event.CreatedAt) + "\n")
	output.WriteString("---\n\n")
}

func renderAssistedTranscriptEventBody(output *bytes.Buffer, event AssistedTranscriptEvent) {
	output.WriteString("## Event " + event.ID + "\n\n")

	if event.Question.ID != "" {
		output.WriteString("### Question\n\n")
		output.WriteString("id: " + event.Question.ID + "\n\n")
		output.WriteString("prompt: " + renderTranscriptText(event, event.Question.Prompt) + "\n\n")
		output.WriteString("rationale: " + renderTranscriptText(event, event.Question.Rationale) + "\n\n")
		fmt.Fprintf(output, "required: %t\n\n", event.Question.Required)
	}

	if event.Answer.QuestionID != "" {
		output.WriteString("### Answer\n\n")
		output.WriteString("question_id: " + event.Answer.QuestionID + "\n\n")
		output.WriteString("text: " + renderTranscriptText(event, event.Answer.Text) + "\n\n")
		output.WriteString("answered_at: " + event.Answer.AnsweredAt.UTC().Format(time.RFC3339) + "\n\n")
	}

	if event.Result.CompletionReason != "" || event.Result.BlockedReason != "" || event.Result.Summary != "" {
		output.WriteString("### Result\n\n")
		output.WriteString("completion_reason: " + string(event.Result.CompletionReason) + "\n\n")
		output.WriteString("blocked_reason: " + string(event.Result.BlockedReason) + "\n\n")
		output.WriteString("summary: " + renderTranscriptText(event, event.Result.Summary) + "\n\n")
		output.WriteString("next_intent: " + string(event.Result.NextIntent) + "\n\n")
	}
}

func splitAssistedTranscriptBlocks(content string) []string {
	blocks := []string{}
	remaining := strings.TrimSpace(content)

	for {
		start := strings.Index(remaining, "---\n")
		if start < 0 {
			return blocks
		}

		remaining = remaining[start:]

		closing := strings.Index(remaining[len("---\n"):], "\n---\n")
		if closing < 0 {
			return blocks
		}

		bodyStart := len("---\n") + closing + len("\n---\n")

		next := strings.Index(remaining[bodyStart:], "\n---\n")
		if next < 0 {
			blocks = append(blocks, strings.TrimSpace(remaining))

			return blocks
		}

		blocks = append(blocks, strings.TrimSpace(remaining[:bodyStart+next]))
		remaining = remaining[bodyStart+next+1:]
	}
}

func parseAssistedTranscriptEvent(block string) (AssistedTranscriptEvent, error) {
	metadata, err := ticketmetadata.Parse(block)
	if err != nil {
		return AssistedTranscriptEvent{}, err
	}

	event, err := assistedTranscriptEventFromMetadata(metadata.Metadata)
	if err != nil {
		return AssistedTranscriptEvent{}, err
	}

	frontmatterEnd := strings.Index(block[len("---\n"):], "\n---\n")
	bodyStart := len("---\n") + frontmatterEnd + len("\n---\n")
	body := block[bodyStart:]
	event.Question = parseAssistedTranscriptQuestion(body)
	event.Answer = parseAssistedTranscriptAnswer(body)

	event.Result = parseAssistedTranscriptResult(body)
	if event.Revision > 0 {
		if err := decodeManagedTranscriptText(&event); err != nil {
			return AssistedTranscriptEvent{}, err
		}
	}

	err = ValidateAssistedTranscriptEvent(event)
	if err != nil {
		return AssistedTranscriptEvent{}, err
	}

	return event, nil
}

func assistedTranscriptEventFromMetadata(metadata map[string]string) (AssistedTranscriptEvent, error) {
	revision, err := optionalAuthoringRevision(metadata)
	if err != nil {
		return AssistedTranscriptEvent{}, err
	}

	turnIndex, err := strconv.Atoi(opsmetadata.Unquote(metadata["turn_index"]))
	if err != nil {
		return AssistedTranscriptEvent{}, fmt.Errorf("parse assisted transcript turn index: %w", err)
	}

	createdAt, err := requiredAuthoringTime(metadata["created_at"], "created_at")
	if err != nil {
		return AssistedTranscriptEvent{}, err
	}

	return AssistedTranscriptEvent{
		ID:        opsmetadata.Unquote(metadata["id"]),
		SessionID: opsmetadata.Unquote(metadata["session_id"]),
		TurnID:    opsmetadata.Unquote(metadata["turn_id"]),
		TurnIndex: turnIndex,
		Revision:  revision,
		Type:      AssistedTranscriptEventType(opsmetadata.Unquote(metadata["type"])),
		Intent:    AssistedAuthoringIntent(opsmetadata.Unquote(metadata["intent"])),
		Status:    AssistedAuthoringTurnStatus(opsmetadata.Unquote(metadata["status"])),
		CreatedAt: createdAt,
	}, nil
}

func parseAssistedTranscriptQuestion(body string) AssistedAuthoringQuestion {
	section := sectionContent(body, "### Question")

	return AssistedAuthoringQuestion{
		ID:        transcriptValue(section, "id"),
		Prompt:    transcriptValue(section, "prompt"),
		Rationale: transcriptValue(section, "rationale"),
		Required:  transcriptValue(section, "required") == "true",
	}
}

func parseAssistedTranscriptAnswer(body string) AssistedAuthoringAnswer {
	section := sectionContent(body, "### Answer")

	answeredAt, err := parseAuthoringTableTime(transcriptValue(section, "answered_at"))
	if err != nil {
		answeredAt = time.Time{}
	}

	return AssistedAuthoringAnswer{
		QuestionID: transcriptValue(section, "question_id"),
		Text:       transcriptValue(section, "text"),
		AnsweredAt: answeredAt,
	}
}

func parseAssistedTranscriptResult(body string) AssistedAuthoringTurnResult {
	section := sectionContent(body, "### Result")

	return AssistedAuthoringTurnResult{
		CompletionReason: AssistedAuthoringCompletionReason(transcriptValue(section, "completion_reason")),
		BlockedReason:    AssistedAuthoringBlockedReason(transcriptValue(section, "blocked_reason")),
		Summary:          transcriptValue(section, "summary"),
		NextIntent:       AssistedAuthoringIntent(transcriptValue(section, "next_intent")),
	}
}

func transcriptValue(section, key string) string {
	for _, line := range strings.Split(section, "\n") {
		foundKey, value, ok := strings.Cut(line, ":")
		if !ok || strings.TrimSpace(foundKey) != key {
			continue
		}

		return strings.TrimSpace(value)
	}

	return ""
}

func renderTranscriptText(event AssistedTranscriptEvent, value string) string {
	if event.Revision > 0 {
		return strconv.Quote(value)
	}

	return value
}

func decodeManagedTranscriptText(event *AssistedTranscriptEvent) error {
	for _, value := range []*string{&event.Question.Prompt, &event.Question.Rationale, &event.Answer.Text, &event.Result.Summary} {
		if *value == "" {
			continue
		}

		decoded, err := strconv.Unquote(*value)
		if err != nil {
			return err
		}

		*value = decoded
	}

	return nil
}
