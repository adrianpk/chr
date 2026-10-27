package core

import (
	"bufio"
	"context"
	"errors"
	"io"
	"os"
	"strings"

	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

// ReadRecentAgentEvents reads at most the final 256 KiB and retains at most
// 128 valid events/diagnostics. It never creates a missing stream.
func ReadRecentAgentEvents(ctx context.Context, path string) (stream AgentEventStream, err error) {
	stream.Path = path
	if err = ctx.Err(); err != nil {
		return stream, err
	}

	file, err := os.Open(path)
	if errors.Is(err, os.ErrNotExist) {
		return stream, nil
	}

	if err != nil {
		return stream, err
	}
	defer func() { err = errors.Join(err, file.Close()) }()

	info, err := file.Stat()
	if err != nil {
		return stream, err
	}

	if !info.Mode().IsRegular() {
		return stream, errors.New("event stream must be a regular file")
	}

	start := max(int64(0), info.Size()-256*1024)
	scanner := bufio.NewScanner(io.NewSectionReader(file, start, info.Size()-start))
	scanner.Buffer(make([]byte, 4096), agentEventScannerMaxBytes)

	if start > 0 {
		scanner.Scan()
	}

	for line := 1; scanner.Scan(); line++ {
		if err = ctx.Err(); err != nil {
			return stream, err
		}

		text := strings.TrimSpace(scanner.Text())
		if text == "" {
			continue
		}

		appendRecentAgentEvent(&stream, line, text)

		stream.Events = stream.Events[max(0, len(stream.Events)-workflowstate.WorkflowPageLimit):]
		stream.Malformed = stream.Malformed[max(0, len(stream.Malformed)-workflowstate.WorkflowPageLimit):]
	}

	return stream, scanner.Err()
}

func appendRecentAgentEvent(stream *AgentEventStream, line int, text string) {
	event, malformed := decodeAgentEventLine(line, text)
	if malformed != nil {
		stream.Malformed = append(stream.Malformed, *malformed)
	} else {
		stream.Events = append(stream.Events, event)
	}
}
