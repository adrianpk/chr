package app

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"forge.adrianpk.com/chr/internal/core"
)

var errConversationIntentInvalid = errors.New("invalid conversation intent")

// ConversationIntentArguments is the bounded argument envelope. RecordJSON is
// decoded to the same concrete record request used by deterministic controls.
type ConversationIntentArguments struct {
	Project      string `json:"project"`
	Conversation string `json:"conversation"`
	Namespace    string `json:"namespace"`
	Text         string `json:"text"`
	Action       string `json:"action"`
	RecordJSON   string `json:"record_json"`
}

// ConversationIntent is a proposal; current targets and authority are checked at dispatch.
type ConversationIntent struct {
	OperationID string                      `json:"operation_id"`
	Arguments   ConversationIntentArguments `json:"arguments"`
	Unresolved  []string                    `json:"unresolved"`
	Explanation string                      `json:"explanation"`
}

// ConversationIntentOperation contains only allowed registry-owned descriptions.
type ConversationIntentOperation struct {
	ID          string `json:"id"`
	Description string `json:"description"`
	Arguments   string `json:"arguments"`
}

// ConversationIntentSession excludes peer identities and hidden transcripts.
type ConversationIntentSession struct {
	ID, ProjectID, Namespace, Intent, Phase, Topic, QuestionBinding string
	Revision                                                        int64
}

// ConversationIntentProject exposes an allowed ID and exact registered name.
type ConversationIntentProject struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// ConversationIntentContext is a bounded current-state projection, never authority.
type ConversationIntentContext struct {
	Message             string                        `json:"message"`
	Projects            []ConversationIntentProject   `json:"projects"`
	SelectedProject     string                        `json:"selected_project"`
	FocusedConversation string                        `json:"focused_conversation"`
	Sessions            []ConversationIntentSession   `json:"sessions"`
	Operations          []ConversationIntentOperation `json:"operations"`
}

// ResolveConversationIntent invokes the configured provider and validates the
// response envelope. Dispatch validates each operation's concrete arguments.
func (s service) ResolveConversationIntent(ctx context.Context, requestID string, input ConversationIntentContext) (ConversationIntent, error) {
	var result ConversationIntent
	if !validConversationIntentContext(input) {
		return result, errors.New("invalid conversation intent context")
	}

	var err error

	s, ctx, err = s.bindIntentGeneration(ctx, requestID, input)
	if err != nil {
		return result, err
	}

	project := s.projectContext.ProjectDir

	payload, err := json.Marshal(input)
	if err != nil {
		return result, err
	}

	provider := s.conversationModel
	if provider == nil {
		return result, errors.New("conversation provider is unavailable")
	}

	modelRequest := core.StructuredModelRequest{ID: requestID, Purpose: "conversation_intent", ProjectDir: project, Input: payload, Schema: conversationIntentSchema(input.Operations), Instructions: conversationIntentInstructions}

	output, err := provider.InvokeStructuredModel(ctx, modelRequest)
	if err != nil {
		return result, err
	}

	if err := decodeConversationIntent(output, &result); err != nil {
		return result, core.ValidateStructuredModelOutput(ctx, modelRequest, fmt.Errorf("%w: %w", errConversationIntentInvalid, err))
	}

	if err := validateConversationIntent(result, input.Operations); err != nil {
		return result, core.ValidateStructuredModelOutput(ctx, modelRequest, fmt.Errorf("%w: %w", errConversationIntentInvalid, err))
	}

	return result, core.ValidateStructuredModelOutput(ctx, modelRequest, nil)
}

const conversationIntentInstructions = `Choose one allowed operation using the current message and bounded context. Never invent a target, actor, revision, credential or authority. Use capture for an answer or captured context in an existing collecting conversation. Explicit explanation/preparation does not authorize development. Return unresolved fields and one focused correction question when unsupported, ambiguous or missing information; use operation_id=clarify and empty arguments. Do not guess record guards. For record operations, record_json contains the strict JSON payload accepted by that operation, never a command. The original message, not your explanation, is the retained user instruction.`

func validateConversationIntent(result ConversationIntent, operations []ConversationIntentOperation) error {
	if err := validateConversationIntentPayload(result); err != nil {
		return err
	}

	if result.OperationID == "clarify" {
		if result.Arguments != (ConversationIntentArguments{}) || strings.TrimSpace(result.Explanation) == "" {
			return errors.New("invalid clarification")
		}

		return nil
	}

	for _, operation := range operations {
		if result.OperationID == operation.ID && len(result.Unresolved) == 0 {
			return nil
		}
	}

	return fmt.Errorf("unsupported or unresolved intent operation")
}

func conversationIntentSchema(operations []ConversationIntentOperation) map[string]any {
	ids := []string{"clarify"}
	for _, operation := range operations {
		ids = append(ids, operation.ID)
	}

	fields := map[string]any{}

	names := []string{"project", "conversation", "namespace", "text", "action", "record_json"}
	for _, name := range names {
		fields[name] = map[string]any{"type": "string"}
	}

	return map[string]any{"type": "object", "additionalProperties": false, "required": []string{"operation_id", "arguments", "unresolved", "explanation"}, "properties": map[string]any{
		"operation_id": map[string]any{"type": "string", "enum": ids},
		"arguments":    map[string]any{"type": "object", "additionalProperties": false, "properties": fields, "required": names},
		"unresolved":   map[string]any{"type": "array", "items": map[string]any{"type": "string"}, "maxItems": 8},
		"explanation":  map[string]any{"type": "string", "maxLength": 512},
	}}
}

func validConversationIntentContext(input ConversationIntentContext) bool {
	return len(input.Message) > 0 && len(input.Message) <= 16384 && !strings.HasPrefix(strings.TrimSpace(input.Message), "/") && len(input.Projects) <= 50 && len(input.Sessions) <= 50 && len(input.Operations) <= 64
}

func validateConversationIntentPayload(result ConversationIntent) error {
	if len(result.Explanation) > 512 || len(result.Unresolved) > 8 || len(result.Arguments.Text) > 16384 || len(result.Arguments.Project) > 256 || len(result.Arguments.Conversation) > 128 || len(result.Arguments.Namespace) > 128 || len(result.Arguments.RecordJSON) > 49152 {
		return errors.New("oversized intent decision")
	}

	return validateConversationIntentUnresolved(result.Unresolved)
}

func validateConversationIntentUnresolved(fields []string) error {
	for _, field := range fields {
		if field == "" || len(field) > 128 {
			return errors.New("invalid unresolved field")
		}
	}

	return nil
}

func decodeConversationIntent(output json.RawMessage, result *ConversationIntent) error {
	if err := decodeOperationJSON(string(output), result); err != nil {
		return err
	}

	fields, err := requiredOperationFields(output, []string{"operation_id", "arguments", "unresolved", "explanation"})
	if err != nil {
		return err
	}

	_, err = requiredOperationFields(fields["arguments"], []string{"project", "conversation", "namespace", "text", "action", "record_json"})

	return err
}
