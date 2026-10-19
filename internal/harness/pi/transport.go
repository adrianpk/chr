// Package pi implements the pinned Pi subprocess RPC transport.
package pi

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"
	"time"
	"unicode/utf8"

	"forge.adrianpk.com/chr/internal/harness"
	"forge.adrianpk.com/chr/internal/harness/process"
)

// ProtocolRevision identifies the supported wire lifecycle, not a capability certificate.
const ProtocolRevision = "pi-rpc/1.3.0"

// Transport failures remain distinguishable from cancellation and process exit.
var (
	ErrProtocol   = fmt.Errorf("pi RPC: %w", harness.ErrProtocol)
	ErrEarlyExit  = errors.New("pi exited before settled completion")
	ErrProvider   = errors.New("pi provider failed")
	ErrAborted    = fmt.Errorf("pi run: %w", harness.ErrAborted)
	ErrCapability = errors.New("unsupported Pi capability")
)

// Transport uses the shared process owner; smaller record ceilings aid fixtures.
type Transport struct {
	Executor    process.Executor
	RecordLimit int
}

// Request describes one independently owned RPC call with exact model identity.
type Request struct {
	Prompt      string
	Provider    string
	Model       string
	Thinking    string
	OutputLimit int
}

// Result contains only a settled final assistant message and observed counters.
type Result struct {
	Output        []byte
	Usage         harness.Usage
	ObservedModel string
	Process       harness.AgentCommandResult
}

// Run does not add retries or capacity. The caller supplies its retained deadline.
func (t Transport) Run(ctx context.Context, command harness.AgentCommand, request Request) (Result, error) {
	if err := validateRequest(request); err != nil {
		return Result{}, err
	}

	session := newSession(request)
	session.secrets = t.Executor.Secrets
	result, err := t.Executor.Exchange(ctx, command, time.Now(), func(stdout io.Reader, stdin io.WriteCloser, stop context.CancelFunc) error {
		return session.exchange(ctx, stdout, stdin, recordLimit(t.RecordLimit), stop)
	})

	response := Result{Process: result, Usage: session.measuredUsage()}
	if err != nil {
		return response, err
	}

	response.Output = session.output
	response.ObservedModel = session.observedModel

	return response, nil
}

func validateRequest(request Request) error {
	if strings.TrimSpace(request.Prompt) == "" || request.Provider == "" || request.Model == "" {
		return fmt.Errorf("pi requires input and an exact provider/model: %w", ErrCapability)
	}

	for _, value := range []string{request.Prompt, request.Provider, request.Model, request.Thinking} {
		if !utf8.ValidString(value) {
			return ErrProtocol
		}
	}

	return validateRequestBounds(request)
}

func validateRequestBounds(request Request) error {
	if len(request.Prompt) > harness.MaxInput {
		return process.ErrLimit
	}

	if len(request.Provider) > 1024 || len(request.Model) > 1024 || len(request.Thinking) > 16 {
		return process.ErrLimit
	}

	data, err := json.Marshal(command{ID: "prompt", Type: "prompt", Message: request.Prompt})
	if err != nil || len(data) > harness.MaxRecord {
		return process.ErrLimit
	}

	return nil
}

func recordLimit(limit int) int {
	if limit > 0 && limit < harness.MaxRecord {
		return limit
	}

	return harness.MaxRecord
}

func readRecord(reader *bufio.Reader, limit int) (record, error) {
	data, err := reader.ReadSlice('\n')
	if errors.Is(err, bufio.ErrBufferFull) || framedLength(data) > limit {
		return record{}, process.ErrLimit
	}

	if err != nil {
		if len(data) != 0 {
			return record{}, fmt.Errorf("unterminated record: %w", ErrProtocol)
		}

		return record{}, errors.Join(ErrEarlyExit, err)
	}

	data = data[:len(data)-1]
	if !utf8.Valid(data) {
		return record{}, fmt.Errorf("record is not UTF-8: %w", ErrProtocol)
	}

	var event record
	if err := json.Unmarshal(data, &event); err != nil || event.Type == "" {
		return record{}, fmt.Errorf("record is not a typed JSON object: %w", ErrProtocol)
	}

	return event, nil
}

func framedLength(data []byte) int {
	if len(data) != 0 && data[len(data)-1] == '\n' {
		return len(data) - 1
	}

	return len(data)
}
