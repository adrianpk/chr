package process

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"syscall"
	"time"

	"forge.adrianpk.com/chr/internal/harness"
)

// ErrLimit identifies input, stream or result overflow. Partial output is invalid.
var ErrLimit = errors.New("harness byte limit exceeded")

// Executor supervises one process group per call. Smaller ceilings aid fixtures.
type Executor struct {
	StdoutLimit int
	StderrLimit int
	// Explicit secrets are redacted before diagnostic excerpts are truncated.
	Secrets []string
}

type spool struct {
	file      *os.File
	remaining int
	stop      context.CancelFunc
	err       error
}

func (s *spool) Write(data []byte) (int, error) {
	if len(data) > s.remaining {
		s.err = ErrLimit
		s.stop()

		return 0, s.err
	}

	s.remaining -= len(data)

	n, err := s.file.Write(data)
	if err != nil {
		s.err = err
		s.stop()
	}

	return n, err
}

// RunAgentCommand drains both streams and kills all owned descendants on exit.
func (e Executor) RunAgentCommand(ctx context.Context, command harness.AgentCommand, stdin string, startedAt time.Time) (result harness.AgentCommandResult, err error) {
	if len(stdin) > harness.MaxInput {
		return harness.AgentCommandResult{}, fmt.Errorf("stdin: %w", ErrLimit)
	}

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	out, err := newSpool(e.StdoutLimit, harness.MaxStdout, cancel)
	if err != nil {
		return harness.AgentCommandResult{}, err
	}
	defer func() { err = errors.Join(err, out.remove()) }()

	diagnostic, err := newSpool(e.StderrLimit, harness.MaxStderr, cancel)
	if err != nil {
		return harness.AgentCommandResult{}, err
	}
	defer func() { err = errors.Join(err, diagnostic.remove()) }()

	cmd := ownedCommand(ctx, command, stdin)
	cmd.Stdout, cmd.Stderr = out, diagnostic

	err = cmd.Run()
	if cmd.Process != nil {
		err = errors.Join(err, killGroup(ctx, cmd.Process.Pid))
	}

	stdout, outErr := out.read(harness.MaxStructuredOutput + 1)
	stderr, diagnosticErr := diagnostic.read(harness.MaxStderr)
	result = harness.AgentCommandResult{Stdout: stdout, Stderr: e.diagnostic(stderr), Latency: time.Since(startedAt)}

	return result, errors.Join(err, out.err, diagnostic.err, outErr, diagnosticErr, ctx.Err())
}

func ownedCommand(ctx context.Context, command harness.AgentCommand, stdin string) *exec.Cmd {
	cmd := exec.CommandContext(ctx, command.Program, command.Args...)

	cmd.Dir = command.Dir
	if command.Env != nil {
		cmd.Env = make([]string, 0, len(command.Env))
		for _, variable := range command.Env {
			cmd.Env = append(cmd.Env, variable.Name+"="+variable.Value)
		}
	}

	cmd.Stdin = strings.NewReader(stdin)
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	cmd.WaitDelay = time.Second
	cmd.Cancel = func() error {
		err := syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL)
		if errors.Is(err, syscall.ESRCH) {
			return os.ErrProcessDone
		}

		return err
	}

	return cmd
}

func newSpool(limit, ceiling int, stop context.CancelFunc) (*spool, error) {
	file, err := os.CreateTemp("", "harness-stream-*")
	if err != nil {
		return nil, err
	}

	if limit <= 0 || limit > ceiling {
		limit = ceiling
	}

	return &spool{file: file, remaining: limit, stop: stop}, nil
}

func (s *spool) read(limit int) ([]byte, error) {
	if _, err := s.file.Seek(0, io.SeekStart); err != nil {
		return nil, err
	}

	return io.ReadAll(io.LimitReader(s.file, int64(limit)))
}

func (s *spool) remove() error {
	return errors.Join(s.file.Close(), os.Remove(s.file.Name()))
}

func killGroup(ctx context.Context, pid int) error {
	err := syscall.Kill(-pid, syscall.SIGKILL)
	if errors.Is(err, syscall.ESRCH) {
		return nil
	}

	if err != nil {
		return err
	}

	return waitKilledGroup(ctx, pid)
}

// ReadOutput rejects oversized files before exposing any partial result.
func ReadOutput(path string, limit int) (data []byte, err error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer func() { err = errors.Join(err, file.Close()) }()

	data, err = io.ReadAll(io.LimitReader(file, int64(limit)+1))
	if err != nil {
		return nil, err
	}

	if len(data) > limit {
		return nil, ErrLimit
	}

	return data, nil
}
