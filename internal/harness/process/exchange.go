package process

import (
	"context"
	"errors"
	"io"
	"os"
	"time"

	"forge.adrianpk.com/chr/internal/harness"
)

// Exchange runs a bidirectional protocol within the ordinary process owner.
// The handler must keep reading while writing. Returning closes stdin; the
// process then has one second to drain and exit before owned cancellation.
func (e Executor) Exchange(ctx context.Context, command harness.AgentCommand, startedAt time.Time, handle func(io.Reader, io.WriteCloser, context.CancelFunc) error) (result harness.AgentCommandResult, err error) {
	callerCtx := ctx

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	out, err := newSpool(e.StdoutLimit, harness.MaxStdout, cancel)
	if err != nil {
		return result, err
	}
	defer func() { err = errors.Join(err, out.remove()) }()

	diagnostic, err := newSpool(e.StderrLimit, harness.MaxStderr, cancel)
	if err != nil {
		return result, err
	}
	defer func() { err = errors.Join(err, diagnostic.remove()) }()

	err = exchangeCommand(ctx, command, out, diagnostic, cancel, handle)
	// Keep one record of lookahead so sanitization precedes excerpt truncation.
	stdout, outErr := out.read(harness.MaxDiagnostic + harness.MaxRecord)
	stderr, diagnosticErr := diagnostic.read(harness.MaxStderr)
	result = harness.AgentCommandResult{Stdout: e.diagnostic(stdout), Stderr: e.diagnostic(stderr), Latency: time.Since(startedAt)}

	return result, errors.Join(err, out.err, diagnostic.err, outErr, diagnosticErr, callerCtx.Err())
}

func exchangeCommand(ctx context.Context, command harness.AgentCommand, out, diagnostic *spool, cancel context.CancelFunc, handle func(io.Reader, io.WriteCloser, context.CancelFunc) error) (err error) {
	cmd := ownedCommand(ctx, command, "")
	cmd.Stdin = nil
	cmd.Stderr = diagnostic

	stdin, err := cmd.StdinPipe()
	if err != nil {
		return err
	}
	defer func() { err = errors.Join(err, closeExchangePipe(stdin)) }()

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}
	defer func() { err = errors.Join(err, closeExchangePipe(stdout)) }()

	if err := cmd.Start(); err != nil {
		return err
	}

	reader := io.TeeReader(stdout, out)
	err = handle(reader, &exchangeInput{WriteCloser: stdin, remaining: harness.MaxInput}, cancel)
	closeErr := closeExchangePipe(stdin)

	if err != nil {
		cancel()
	}

	grace := time.AfterFunc(time.Second, cancel)
	_, drainErr := io.Copy(io.Discard, reader)
	waitErr := cmd.Wait()
	// CommandContext can report the supervisor's private stop as cancellation.
	// Exchange reports caller cancellation separately through its original context.
	if errors.Is(waitErr, context.Canceled) {
		waitErr = errors.New("owned process stopped before orderly exit")
	}

	grace.Stop()

	return errors.Join(err, closeErr, drainErr, waitErr, killGroup(cmd.Process.Pid))
}

func closeExchangePipe(pipe io.Closer) error {
	err := pipe.Close()
	if errors.Is(err, os.ErrClosed) {
		return nil
	}

	return err
}

type exchangeInput struct {
	io.WriteCloser
	remaining int
}

func (w *exchangeInput) Write(data []byte) (int, error) {
	if len(data) > w.remaining {
		return 0, ErrLimit
	}

	w.remaining -= len(data)

	return w.WriteCloser.Write(data)
}
