package pi

import "fmt"

const maxReadOnlyCalls = 64

func (s *session) readOnlyCalls(message assistantMessage) error {
	if !s.request.ReadOnlyFiles || len(s.tools) != 0 {
		return ErrCapability
	}

	s.tools = make(map[string]bool)

	for _, block := range message.Content {
		if block.Type != "toolCall" {
			continue
		}

		if err := s.registerInspectionCall(block.Name, block.ID); err != nil {
			return err
		}
	}

	if len(s.tools) == 0 {
		return ErrProtocol
	}

	s.finalErr = nil

	return nil
}

func (s *session) registerInspectionCall(name, id string) error {
	if name != "chr_inspect" || id == "" || len(id) > 1024 {
		return ErrCapability
	}

	if _, exists := s.tools[id]; exists {
		return ErrProtocol
	}

	s.toolCalls++
	if s.toolCalls > maxReadOnlyCalls {
		return fmt.Errorf("pi inspection exceeds 64 tool calls: %w", ErrCapability)
	}

	s.tools[id] = false

	return nil
}

func (s *session) toolEvent(event record) error {
	if !s.request.ReadOnlyFiles || event.ToolName != "chr_inspect" {
		return ErrCapability
	}

	started, exists := s.tools[event.ToolCallID]
	if !exists {
		return ErrProtocol
	}

	switch event.Type {
	case "tool_execution_start":
		if started {
			return ErrProtocol
		}

		s.tools[event.ToolCallID] = true
	case "tool_execution_update", "tool_execution_end":
		if !started {
			return ErrProtocol
		}

		if event.Type == "tool_execution_end" {
			delete(s.tools, event.ToolCallID)
		}
	default:
		return ErrProtocol
	}

	return nil
}
