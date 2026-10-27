package core

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	opsfile "forge.adrianpk.com/chr/internal/core/ops/file"
)

type workflowPublicationBuffer struct {
	bytes.Buffer
	limit int
}

func (buffer *workflowPublicationBuffer) Write(data []byte) (int, error) {
	if len(data) > buffer.limit-buffer.Len() {
		return 0, fmt.Errorf("publication output exceeds bounds")
	}

	return buffer.Buffer.Write(data)
}

func workflowPublicationCommand(ctx context.Context, name string, args ...string) *exec.Cmd {
	command := exec.CommandContext(ctx, name, args...)
	command.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}

	command.Env = append(os.Environ(), "GIT_TERMINAL_PROMPT=0", "GIT_NO_REPLACE_OBJECTS=1")
	command.WaitDelay = time.Second
	command.Stderr = &opsfile.CommandOutput{}
	command.Cancel = func() error {
		err := syscall.Kill(-command.Process.Pid, syscall.SIGKILL)
		if errors.Is(err, syscall.ESRCH) {
			return os.ErrProcessDone
		}

		return err
	}

	return command
}

func workflowPublicationPack(ctx context.Context, request WorkflowGitPublication) ([]byte, error) {
	command := workflowPublicationCommand(ctx, "git", "--no-replace-objects", "-C", request.Directory, "pack-objects", "--stdout", "--revs")
	output := &workflowPublicationBuffer{limit: workflowPublicationPackLimit}
	command.Stdout = output
	command.Stdin = bytes.NewBufferString(request.Git.ResultCommit + "\n")

	if err := command.Run(); err != nil {
		return nil, fmt.Errorf("bounded publication pack generation failed: %w", err)
	}

	return output.Bytes(), nil
}

type workflowReceivePackProcess struct {
	command *exec.Cmd
	input   io.WriteCloser
	output  io.Reader
	cancel  context.CancelFunc
	closed  bool
	waited  bool
}

func (git OSWorkflowGitPublisher) openReceivePack(ctx context.Context, directory string) (workflowReceivePack, error) {
	if strings.HasPrefix(git.transport, "https://") {
		return git.openHTTPReceivePack(ctx, directory)
	}

	if !filepath.IsAbs(git.transport) {
		return git.openSSHReceivePack(ctx)
	}

	owned, cancel := context.WithCancel(ctx)
	command := workflowPublicationCommand(owned, "git", "receive-pack", "--", git.transport)

	return startWorkflowReceivePack(command, cancel)
}

func startWorkflowReceivePack(command *exec.Cmd, cancel context.CancelFunc) (workflowReceivePack, error) {
	input, err := command.StdinPipe()
	if err != nil {
		cancel()

		return nil, err
	}

	output, err := command.StdoutPipe()
	if err != nil {
		cancel()

		return nil, errors.Join(err, input.Close())
	}

	if err := command.Start(); err != nil {
		cancel()

		return nil, errors.Join(err, input.Close(), output.Close())
	}

	return &workflowReceivePackProcess{command: command, input: input, output: output, cancel: cancel}, nil
}

func (session *workflowReceivePackProcess) Advertisement() io.Reader { return session.output }

func (session *workflowReceivePackProcess) Send(data []byte) (io.Reader, error) {
	if _, err := session.input.Write(data); err != nil {
		return nil, err
	}

	err := session.input.Close()
	session.closed = true

	return session.output, err
}

func (session *workflowReceivePackProcess) Finish() error {
	err := session.command.Wait()
	session.waited = true

	return err
}

func (session *workflowReceivePackProcess) Close() error {
	session.cancel()

	if session.waited {
		return nil
	}

	var closeErr error
	if !session.closed {
		closeErr = session.input.Close()
	}

	waitErr := session.command.Wait()
	session.waited = true

	if errors.Is(waitErr, context.Canceled) {
		waitErr = nil
	}

	return errors.Join(closeErr, waitErr)
}
