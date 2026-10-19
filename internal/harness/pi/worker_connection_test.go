package pi

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"strings"
	"testing"
	"time"

	"forge.adrianpk.com/chr/internal/harness"
	"forge.adrianpk.com/chr/internal/harness/process"
)

// A lost host pipe cancels the runtime independently of the model's completion.
// Keeping the pipe open must not be mistaken for a prompt/session boundary.
func TestPiWorkerConnection(t *testing.T) {
	reader, writer := io.Pipe()

	t.Cleanup(func() { _ = reader.Close(); _ = writer.Close() }) //nolint:errcheck // Owned fixture pipe cleanup.

	go func() { _, _ = fmt.Fprint(writer, "4\ntask") }() //nolint:errcheck // Reader assertions report a broken fixture.

	prompt, ctx, cancel, err := ReadWorkerInput(t.Context(), reader)
	if err != nil || prompt != "task" || ctx.Err() != nil {
		t.Fatalf("host input failed: %q, %v", prompt, err)
	}
	defer cancel()

	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}

	select {
	case <-ctx.Done():
	case <-time.After(time.Second):
		t.Fatal("lost host did not cancel its runtime")
	}
}

// Framing and length failures reject input before starting a model or allocating
// unbounded context. Already canceled ownership cannot start a call.
func TestPiWorkerConnectionBounds(t *testing.T) {
	for _, scenario := range []struct {
		name string
		data string
		want error
	}{
		{"oversized", "4194304\n", process.ErrLimit},
		{"malformed", "not-a-length\n", harness.ErrProtocol},
		{"unterminated", "4", harness.ErrProtocol},
		{"short payload", "4\nab", io.ErrUnexpectedEOF},
	} {
		t.Run(scenario.name, func(t *testing.T) {
			if _, err := readWorkerPrompt(bufio.NewReader(strings.NewReader(scenario.data))); !errors.Is(err, scenario.want) {
				t.Fatalf("input boundary: %v, want %v", err, scenario.want)
			}
		})
	}

	ctx, cancel := context.WithCancel(t.Context())
	cancel()

	_, _, release, err := ReadWorkerInput(ctx, strings.NewReader("4\ntask"))
	release()

	if !errors.Is(err, context.Canceled) {
		t.Fatalf("canceled host admitted input: %v", err)
	}
}
