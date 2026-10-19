package pi

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"

	"forge.adrianpk.com/chr/internal/harness"
	"forge.adrianpk.com/chr/internal/harness/process"
)

const (
	workerHeartbeatInterval = 500 * time.Millisecond
	workerConnectionTimeout = 3 * time.Second
)

func exchangeWorkerInput(ctx context.Context, stdout io.Reader, stdin io.Writer, prompt string, output *[]byte) error {
	if len(prompt)+16 > harness.MaxInput {
		return process.ErrLimit
	}

	if _, err := fmt.Fprintf(stdin, "%d\n", len(prompt)); err != nil {
		return err
	}

	if _, err := io.WriteString(stdin, prompt); err != nil {
		return err
	}

	done := make(chan struct{})

	defer close(done)
	go workerHeartbeats(ctx, done, stdin)

	data, err := io.ReadAll(io.LimitReader(stdout, harness.MaxWorkerOutput+1))
	if err != nil {
		return err
	}

	if len(data) > harness.MaxWorkerOutput {
		return process.ErrLimit
	}

	*output = data

	return nil
}

func workerHeartbeats(ctx context.Context, done <-chan struct{}, stdin io.Writer) {
	ticker := time.NewTicker(workerHeartbeatInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-done:
			return
		case <-ticker.C:
			if _, err := stdin.Write([]byte{0}); err != nil {
				return
			}
		}
	}
}

// ReadWorkerInput reads one bounded raw context followed by host heartbeats.
// Losing the connection cancels the same context that owns Pi and its pipes,
// including when a host is killed without an opportunity to run cleanup.
func ReadWorkerInput(ctx context.Context, input io.Reader) (string, context.Context, context.CancelFunc, error) {
	ctx, cancel := context.WithCancel(ctx)
	if err := ctx.Err(); err != nil {
		cancel()

		return "", ctx, cancel, err
	}

	reader := bufio.NewReader(input)

	type preparedInput struct {
		prompt string
		err    error
	}

	ready := make(chan preparedInput, 1)

	go func() {
		prompt, err := readWorkerPrompt(reader)
		ready <- preparedInput{prompt, err}
	}()

	select {
	case <-ctx.Done():
		cancel()

		return "", ctx, cancel, ctx.Err()
	case <-time.After(workerConnectionTimeout):
		cancel()

		return "", ctx, cancel, errors.New("pi worker host connection expired during input")
	case prepared := <-ready:
		if prepared.err != nil {
			cancel()

			return "", ctx, cancel, prepared.err
		}

		go watchWorkerConnection(ctx, reader, cancel)

		return prepared.prompt, ctx, cancel, nil
	}
}

func readWorkerPrompt(reader *bufio.Reader) (string, error) {
	header, err := reader.ReadSlice('\n')
	if err != nil || len(header) > 9 {
		return "", harness.ErrProtocol
	}

	length, err := strconv.Atoi(strings.TrimSuffix(string(header), "\n"))
	if err != nil || length <= 0 {
		return "", harness.ErrProtocol
	}

	if length+16 > harness.MaxInput {
		return "", process.ErrLimit
	}

	data := make([]byte, length)
	if _, err := io.ReadFull(reader, data); err != nil {
		return "", err
	}

	return string(data), nil
}

func watchWorkerConnection(ctx context.Context, reader *bufio.Reader, cancel context.CancelFunc) {
	defer cancel()

	beats := make(chan bool, 1)

	go readWorkerHeartbeats(ctx, reader, beats)

	timer := time.NewTimer(workerConnectionTimeout)
	defer timer.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-timer.C:
			return
		case valid := <-beats:
			if !valid {
				return
			}

			timer.Reset(workerConnectionTimeout)
		}
	}
}

func readWorkerHeartbeats(ctx context.Context, reader *bufio.Reader, beats chan<- bool) {
	for {
		value, err := reader.ReadByte()
		select {
		case beats <- err == nil && value == 0:
		case <-ctx.Done():
			return
		}

		if err != nil || value != 0 {
			return
		}
	}
}
