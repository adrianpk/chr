package pi

import "fmt"

const maxReadOnlyCalls = 64

type toolState struct {
	name    string
	started bool
}

func (s *session) registerToolMessage(message assistantMessage) error {
	if (!s.request.ReadOnlyFiles && !s.request.WorkerTools) || len(s.tools) != 0 {
		return ErrCapability
	}

	s.tools = make(map[string]toolState)

	for _, block := range message.Content {
		if block.Type != "toolCall" {
			continue
		}

		if err := s.registerToolCall(block.Name, block.ID); err != nil {
			return err
		}
	}

	if len(s.tools) == 0 {
		return ErrProtocol
	}

	s.finalErr = nil

	return nil
}

func (s *session) registerToolCall(name, id string) error {
	if !s.allowedTool(name) || id == "" || len(id) > 1024 {
		return ErrCapability
	}

	if _, exists := s.tools[id]; exists {
		return ErrProtocol
	}

	s.toolCalls++
	if !s.request.WorkerTools && s.toolCalls > maxReadOnlyCalls {
		return fmt.Errorf("pi inspection exceeds 64 tool calls: %w", ErrCapability)
	}

	s.tools[id] = toolState{name: name}

	return nil
}

func (s *session) toolEvent(event record) error {
	if !s.allowedTool(event.ToolName) {
		return ErrCapability
	}

	tool, exists := s.tools[event.ToolCallID]
	if !exists || tool.name != event.ToolName {
		return ErrProtocol
	}

	switch event.Type {
	case "tool_execution_start":
		if tool.started {
			return ErrProtocol
		}

		tool.started = true
		s.tools[event.ToolCallID] = tool
	case "tool_execution_update", "tool_execution_end":
		if !tool.started {
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

func (s *session) allowedTool(name string) bool {
	if s.request.WorkerTools {
		switch name {
		case "read", "bash", "edit", "write", "grep", "find", "ls":
			return true
		default:
			return false
		}
	}

	return s.request.ReadOnlyFiles && name == "chr_inspect"
}
