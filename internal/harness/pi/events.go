package pi

import (
	"encoding/json"
	"fmt"
	"math"
	"strings"

	"forge.adrianpk.com/chr/internal/harness"
	"forge.adrianpk.com/chr/internal/harness/process"
)

func (s *session) checkModel(data json.RawMessage) error {
	var model struct{ Provider, ID string }
	if err := json.Unmarshal(data, &model); err != nil || model.Provider != s.request.Provider || model.ID != s.request.Model {
		return fmt.Errorf("exact provider/model unavailable: %w", ErrCapability)
	}

	return nil
}

func (s *session) checkState(data json.RawMessage) error {
	var state struct {
		Model         json.RawMessage `json:"model"`
		ThinkingLevel string          `json:"thinkingLevel"`
		IsStreaming   *bool           `json:"isStreaming"`
		IsCompacting  *bool           `json:"isCompacting"`
		MessageCount  *int            `json:"messageCount"`
		PendingCount  *int            `json:"pendingMessageCount"`
		Compaction    *bool           `json:"autoCompactionEnabled"`
	}
	if err := json.Unmarshal(data, &state); err != nil {
		return ErrProtocol
	}

	if err := s.checkModel(state.Model); err != nil {
		return err
	}

	if !idleState(state.IsStreaming, state.IsCompacting, state.MessageCount, state.PendingCount) || state.Compaction == nil || *state.Compaction {
		return fmt.Errorf("session is not private and idle: %w", ErrCapability)
	}

	if s.request.Thinking != "" && state.ThinkingLevel != s.request.Thinking {
		return fmt.Errorf("requested thinking level unavailable: %w", ErrCapability)
	}

	return nil
}

func idleState(streaming, compacting *bool, messages, pending *int) bool {
	return streaming != nil && !*streaming && compacting != nil && !*compacting && messages != nil && *messages == 0 && pending != nil && *pending == 0
}

func (s *session) accept(event record) error {
	if event.Type == "response" {
		return s.response(event)
	}

	if s.phase != "prompt" || s.settled {
		return fmt.Errorf("event outside owned run: %w", ErrProtocol)
	}

	switch event.Type {
	case "agent_start":
		s.started, s.output, s.finalErr = true, nil, nil
	case "message_end":
		return s.message(event.Message)
	case "agent_settled":
		return s.settle(event)
	default:
		return s.continuationEvent(event)
	}

	return nil
}

func (s *session) continuationEvent(event record) error {
	switch event.Type {
	case "auto_retry_end":
		if event.Success != nil && !*event.Success {
			s.finalErr = fmt.Errorf("retry exhausted: %w", ErrProvider)
		}
	case "compaction_end":
		var result struct{ Usage *observedUsage }
		if len(event.Result) != 0 {
			if err := json.Unmarshal(event.Result, &result); err != nil {
				return ErrProtocol
			}
		}

		s.usage.add(result.Usage)

		if event.ErrorMessage != "" {
			s.finalErr = fmt.Errorf("compaction failed: %w", ErrProvider)
		}
	default:
		return rejectUnexpectedCapability(event.Type)
	}

	return nil
}

func rejectUnexpectedCapability(kind string) error {
	switch kind {
	case "tool_execution_start", "tool_execution_update", "tool_execution_end", "extension_ui_request", "bash_execution_update":
		return fmt.Errorf("disabled tool or extension event: %w", ErrCapability)
	case "turn_start", "turn_end", "message_start", "message_update", "agent_end", "auto_retry_start", "compaction_start",
		"summarization_retry_scheduled", "summarization_retry_attempt_start", "summarization_retry_finished":
		return nil
	default:
		return fmt.Errorf("unknown event type: %w", ErrProtocol)
	}
}

func (s *session) settle(event record) error {
	if !s.started || event.Aborted == nil {
		return fmt.Errorf("settled event missing run state: %w", ErrProtocol)
	}

	s.settled = true
	if *event.Aborted {
		s.finalErr = ErrAborted
	}

	return nil
}

func (s *session) finalResult() error {
	if s.finalErr != nil {
		return s.finalErr
	}

	if len(s.output) == 0 {
		return fmt.Errorf("settled run has no completed assistant output: %w", ErrProtocol)
	}

	return nil
}

type assistantMessage struct {
	Role       string `json:"role"`
	Provider   string `json:"provider"`
	Model      string `json:"model"`
	StopReason string `json:"stopReason"`
	Content    []struct {
		Type string `json:"type"`
		Text string `json:"text"`
	} `json:"content"`
	Usage *observedUsage `json:"usage"`
}

type observedUsage struct {
	Input  *int64 `json:"input"`
	Output *int64 `json:"output"`
}

func (s *session) message(data json.RawMessage) error {
	var identity struct{ Role string }
	if err := json.Unmarshal(data, &identity); err != nil {
		return ErrProtocol
	}

	if identity.Role != "assistant" {
		return nil
	}

	var message assistantMessage
	if err := json.Unmarshal(data, &message); err != nil || !s.started {
		return ErrProtocol
	}

	s.usage.add(message.Usage)

	s.output = nil
	if message.StopReason == "aborted" {
		s.finalErr = ErrAborted

		return nil
	}

	if message.StopReason != "stop" {
		s.finalErr = fmt.Errorf("assistant did not complete: %w", ErrProvider)

		return nil
	}

	return s.completedMessage(message)
}

func (s *session) completedMessage(message assistantMessage) error {
	if message.Provider != s.request.Provider || message.Model != s.request.Model {
		return fmt.Errorf("assistant model changed: %w", ErrCapability)
	}

	limit := s.request.OutputLimit
	if limit <= 0 || limit > harness.MaxStructuredOutput {
		limit = harness.MaxStructuredOutput
	}

	var text strings.Builder

	for _, block := range message.Content {
		if block.Type == "toolCall" {
			return ErrCapability
		}

		if block.Type == "text" {
			if text.Len()+len(block.Text) > limit {
				return process.ErrLimit
			}

			text.WriteString(block.Text)
		}
	}

	s.output, s.observedModel, s.finalErr = []byte(text.String()), message.Provider+"/"+message.Model, nil

	return nil
}

type usageTotals struct {
	input, output int64
	inputKnown    bool
	outputKnown   bool
	seen          bool
}

func (u *usageTotals) add(usage *observedUsage) {
	if !u.seen {
		u.inputKnown, u.outputKnown, u.seen = true, true, true
	}

	if usage == nil {
		u.inputKnown, u.outputKnown = false, false

		return
	}

	u.input, u.inputKnown = addUsage(u.input, usage.Input, u.inputKnown)
	u.output, u.outputKnown = addUsage(u.output, usage.Output, u.outputKnown)
}

func addUsage(total int64, observed *int64, known bool) (int64, bool) {
	if !known || observed == nil || *observed < 0 || total > math.MaxInt64-*observed {
		return total, false
	}

	return total + *observed, true
}

func (s *session) measuredUsage() harness.Usage {
	var result harness.Usage
	if s.usage.inputKnown {
		result.Input = &s.usage.input
	}

	if s.usage.outputKnown {
		result.Output = &s.usage.output
	}

	return result
}
