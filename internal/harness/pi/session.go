package pi

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"

	"forge.adrianpk.com/chr/internal/harness"
	"forge.adrianpk.com/chr/internal/harness/process"
)

type command struct {
	ID       string `json:"id"`
	Type     string `json:"type"`
	Message  string `json:"message,omitempty"`
	Provider string `json:"provider,omitempty"`
	Model    string `json:"modelId,omitempty"`
	Level    string `json:"level,omitempty"`
	Enabled  bool   `json:"enabled"`
}

type record struct {
	ID           string          `json:"id"`
	Type         string          `json:"type"`
	Command      string          `json:"command"`
	Success      *bool           `json:"success"`
	Error        string          `json:"error"`
	Data         json.RawMessage `json:"data"`
	Message      json.RawMessage `json:"message"`
	Aborted      *bool           `json:"aborted"`
	ErrorMessage string          `json:"errorMessage"`
	FinalError   string          `json:"finalError"`
	Result       json.RawMessage `json:"result"`
	ToolName     string          `json:"toolName"`
	ToolCallID   string          `json:"toolCallId"`
}

type session struct {
	request       Request
	pending       map[string]string
	writes        chan command
	phase         string
	promptAck     bool
	settled       bool
	started       bool
	output        []byte
	observedModel string
	finalErr      error
	usage         usageTotals
	secrets       []string
	tools         map[string]bool
	toolCalls     int
}

func newSession(request Request) *session {
	return &session{request: request, pending: make(map[string]string), writes: make(chan command, 6), phase: "setup"}
}

func (s *session) exchange(ctx context.Context, stdout io.Reader, stdin io.WriteCloser, limit int, stop context.CancelFunc) error {
	writerDone := make(chan error, 1)

	go func() {
		err := s.writeCommands(ctx, stdin)
		if err != nil {
			stop()
		}

		writerDone <- err
	}()

	s.enqueue(command{ID: "retry", Type: "set_auto_retry"})
	s.enqueue(command{ID: "compaction", Type: "set_auto_compaction"})
	s.enqueue(command{ID: "model", Type: "set_model", Provider: s.request.Provider, Model: s.request.Model})

	err := s.readEvents(bufio.NewReaderSize(stdout, limit+2), limit)
	close(s.writes)

	closeErr := stdin.Close()

	return errors.Join(err, closeErr, <-writerDone)
}

func (s *session) enqueue(value command) {
	s.pending[value.ID] = value.Type
	s.writes <- value
}

func (s *session) writeCommands(ctx context.Context, stdin io.Writer) error {
	for value := range s.writes {
		if err := ctx.Err(); err != nil {
			return err
		}

		data, err := json.Marshal(value)
		if err != nil {
			return err
		}

		if len(data) > harness.MaxRecord {
			return process.ErrLimit
		}

		if _, err := stdin.Write(append(data, '\n')); err != nil {
			return err
		}
	}

	return nil
}

func (s *session) readEvents(reader *bufio.Reader, limit int) error {
	for {
		event, err := readRecord(reader, limit)
		if err != nil {
			return err
		}

		if err := s.accept(event); err != nil {
			return err
		}

		if s.settled && s.promptAck {
			return s.finalResult()
		}
	}
}

func (s *session) response(event record) error {
	expected, found := s.pending[event.ID]
	if !found || expected != event.Command || event.Success == nil {
		return fmt.Errorf("uncorrelated command response: %w", ErrProtocol)
	}

	delete(s.pending, event.ID)

	if !*event.Success {
		return fmt.Errorf("command %s rejected: %w: %s", expected, ErrCapability, process.DiagnosticWithSecrets([]byte(event.Error), s.secrets...))
	}

	if err := s.responseData(event); err != nil {
		return err
	}

	return s.advance()
}

func (s *session) responseData(event record) error {
	switch event.Command {
	case "set_model":
		if err := s.checkModel(event.Data); err != nil {
			return err
		}

		if s.request.Thinking != "" {
			s.enqueue(command{ID: "thinking", Type: "set_thinking_level", Level: s.request.Thinking})
		}
	case "get_state":
		return s.checkState(event.Data)
	case "prompt":
		var data struct{ Disposition string }
		if err := json.Unmarshal(event.Data, &data); err != nil || data.Disposition != "started" {
			return fmt.Errorf("owned prompt did not start: %w", ErrProtocol)
		}

		s.promptAck = true
	}

	return nil
}

func (s *session) advance() error {
	if len(s.pending) != 0 {
		return nil
	}

	switch s.phase {
	case "setup":
		s.phase = "verify"
		s.enqueue(command{ID: "state", Type: "get_state"})
	case "verify":
		s.phase = "prompt"
		s.enqueue(command{ID: "prompt", Type: "prompt", Message: s.request.Prompt})
	}

	return nil
}
