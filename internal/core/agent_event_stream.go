package core

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const agentEventScannerMaxBytes = 1024 * 1024

// AgentEventStreamFileName is the durable event stream filename.
const AgentEventStreamFileName = "events.jsonl"

// AgentEventMalformedLine describes an event stream line that could not be loaded.
type AgentEventMalformedLine struct {
	Line  int
	Text  string
	Error string
}

// AgentEventStream describes loaded events and degraded read diagnostics.
type AgentEventStream struct {
	Path      string
	Events    []AgentEvent
	Malformed []AgentEventMalformedLine
}

// FileAgentEventSink appends agent events to one JSONL stream.
type FileAgentEventSink struct {
	Path string
}

// NewFileAgentEventSink constructs the corresponding package service or value.
func NewFileAgentEventSink(path string) FileAgentEventSink {
	return FileAgentEventSink{Path: path}
}

// AgentEventStreamPath returns the event stream path for a runtime run directory.
func AgentEventStreamPath(runDir string) string {
	if strings.TrimSpace(runDir) == "" {
		return ""
	}

	return filepath.Join(runDir, AgentEventStreamFileName)
}

// EmitAgentEvent appends the requested event to the configured stream.
func (s FileAgentEventSink) EmitAgentEvent(ctx context.Context, event AgentEvent) error {
	return AppendAgentEvent(ctx, s.Path, event)
}

// AppendAgentEvent appends one event to an append-only JSONL stream.
func AppendAgentEvent(ctx context.Context, path string, event AgentEvent) error {
	err := ctx.Err()
	if err != nil {
		return err
	}

	path = strings.TrimSpace(path)
	if path == "" {
		return fmt.Errorf("agent event stream path is required")
	}

	event.Timestamp = event.Timestamp.UTC()

	err = ValidateAgentEvent(event)
	if err != nil {
		return err
	}

	err = os.MkdirAll(filepath.Dir(path), 0o755)
	if err != nil {
		return fmt.Errorf("create agent event stream directory: %w", err)
	}

	file, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		return fmt.Errorf("open agent event stream: %w", err)
	}

	encoder := json.NewEncoder(file)

	err = encoder.Encode(event)
	if err != nil {
		closeErr := file.Close()
		if closeErr != nil {
			return fmt.Errorf("write agent event: %w; close agent event stream: %w", err, closeErr)
		}

		return fmt.Errorf("write agent event: %w", err)
	}

	err = file.Close()
	if err != nil {
		return fmt.Errorf("close agent event stream: %w", err)
	}

	return nil
}

// ReadAgentEventStream loads a JSONL event stream and keeps malformed-line diagnostics.
func ReadAgentEventStream(ctx context.Context, path string) (AgentEventStream, error) {
	err := ctx.Err()
	if err != nil {
		return AgentEventStream{}, err
	}

	path = strings.TrimSpace(path)
	if path == "" {
		return AgentEventStream{}, fmt.Errorf("agent event stream path is required")
	}

	file, err := os.Open(path)
	if errors.Is(err, os.ErrNotExist) {
		return AgentEventStream{Path: path}, nil
	}

	if err != nil {
		return AgentEventStream{}, fmt.Errorf("open agent event stream: %w", err)
	}

	defer func() {
		//nolint:errcheck // Read-only stream; scanner errors are checked below.
		_ = file.Close()
	}()

	stream := AgentEventStream{Path: path}
	scanner := bufio.NewScanner(file)
	scanner.Buffer(make([]byte, 0, 64*1024), agentEventScannerMaxBytes)

	lineNumber := 0

	for scanner.Scan() {
		lineNumber++

		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		event, malformed := decodeAgentEventLine(lineNumber, line)
		if malformed != nil {
			stream.Malformed = append(stream.Malformed, *malformed)

			continue
		}

		stream.Events = append(stream.Events, event)
	}

	err = scanner.Err()
	if err != nil {
		return AgentEventStream{}, fmt.Errorf("scan agent event stream: %w", err)
	}

	return stream, nil
}

func decodeAgentEventLine(lineNumber int, line string) (AgentEvent, *AgentEventMalformedLine) {
	var event AgentEvent

	err := json.Unmarshal([]byte(line), &event)
	if err != nil {
		return AgentEvent{}, &AgentEventMalformedLine{
			Line:  lineNumber,
			Text:  line,
			Error: err.Error(),
		}
	}

	err = ValidateAgentEvent(event)
	if err != nil {
		return AgentEvent{}, &AgentEventMalformedLine{
			Line:  lineNumber,
			Text:  line,
			Error: err.Error(),
		}
	}

	event.Timestamp = event.Timestamp.UTC()

	return event, nil
}
