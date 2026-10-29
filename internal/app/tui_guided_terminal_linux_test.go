package app

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
	"sync"
	"syscall"
	"testing"
	"time"

	"golang.org/x/sys/unix"

	"forge.adrianpk.com/chr/internal/core"
	authoringcore "forge.adrianpk.com/chr/internal/core/authoring"
)

// TestGuidedUseCaseTUITerminal drives the production Bubble Tea loop in an owned
// PTY child, then restarts it over the same retained store without advancing work.
func TestGuidedUseCaseTUITerminal(t *testing.T) {
	t.Parallel()

	operations, _ := guidedTUIFixture(t)
	terminal := startGuidedTerminal(t, operations)
	terminal.keys(t, "w")
	terminal.see(t, "[Create]")
	terminal.keys(t, "Terminal process audit\x13")
	terminal.see(t, "Add the known context")
	terminal.keys(t, "e")
	terminal.see(t, "ctrl+s Continue to quiz")
	terminal.keys(t, "Preserve finite work and original inputs.")
	terminal.keys(t, " EXACT_FINAL_CAPTURE\x13")
	terminal.see(t, "What should the workflow preserve")
	terminal.keys(t, "e")
	terminal.see(t, "alt+x cancels quiz")
	terminal.keys(t, "Preserve the original agreed scope.\r")
	terminal.see(t, "Wizard: 5/7 preparation")

	before := onlyGuidedTUIRun(t, operations)
	if before.Guided.SpecApproval == nil || len(before.Guided.Checkpoints) != 3 || before.Guided.Capture != "Preserve finite work and original inputs. EXACT_FINAL_CAPTURE" {
		t.Fatal("terminal process did not persist exact approved inputs")
	}

	terminal.keys(t, "c")
	terminal.see(t, "Current stage · r Refresh")
	pauseAuthoringForReentry(t, terminal, operations, before)
	terminal.keys(t, "i")
	terminal.see(t, "ctrl+s submit")
	terminal.keys(t, "Retain a separate follow-up idea.\x13")
	terminal.see(t, "Later idea, pending review")
	terminal.keys(t, "1")
	terminal.see(t, "Original commit:")
	terminal.see(t, "Preserve finite work and original inputs.")
	terminal.stop(t)

	paused := onlyGuidedTUIRun(t, operations)
	reopened := startGuidedTerminal(t, operations)
	reopened.catalog(t)
	reopened.see(t, "Terminal process audit")
	reopened.see(t, "Dispatch stopped: pause")
	reopened.stop(t)

	after := onlyGuidedTUIRun(t, operations)
	if after.ID != paused.ID || after.Revision != paused.Revision || len(after.Guided.LaterIdeas) != 1 {
		t.Fatal("terminal restart changed the paused run or lost its later idea")
	}
}

// The child assembles the actual managed TUI host and fresh application adapters;
// only generation responses are deterministic, and no service instance is shared.
func TestGuidedUseCaseTUIProcess(t *testing.T) {
	project := os.Getenv("CHR_GUIDED_TUI_PROJECT")
	if project == "" {
		t.Skip("launched by the owned terminal interaction fixture")
	}

	env, err := core.NewUserEnvironment(core.UserEnvironmentOptions{DevMode: true, DevRoot: os.Getenv("CHR_GUIDED_TUI_ROOT")})
	if err != nil {
		t.Fatal(err)
	}

	operations := newService(t.Context(), env, commandProjectContext{ProjectDir: project, ProjectID: os.Getenv("CHR_GUIDED_TUI_ID")}, testAssetsFS())
	guidedTelegramModel(t, &operations)

	options := newTUIOptions(tuiOptionsData{ctx: t.Context(), env: env, projectContext: operations.projectContext, stdin: os.Stdin, stdout: os.Stdout, assetsFS: testAssetsFS(), operations: operations})
	if err := runManagedTUI(t.Context(), options, operations); err != nil {
		t.Fatal(err)
	}
}

type guidedTerminal struct {
	master   *os.File
	done     chan error
	readDone chan error
	output   guidedTerminalOutput
	closed   bool
	seen     int
}

type guidedTerminalOutput struct {
	mu   sync.Mutex
	data bytes.Buffer
}

func (b *guidedTerminalOutput) Write(data []byte) (int, error) {
	b.mu.Lock()
	defer b.mu.Unlock()

	return b.data.Write(data)
}

func (b *guidedTerminalOutput) text() string {
	b.mu.Lock()
	defer b.mu.Unlock()

	return b.data.String()
}

func startGuidedTerminal(t *testing.T, operations service) *guidedTerminal {
	t.Helper()

	return startTUITerminal(t, operations, "TestGuidedUseCaseTUIProcess")
}

func startTUITerminal(t *testing.T, operations service, processTest string) *guidedTerminal {
	t.Helper()

	master, slave := openGuidedPTY(t)

	program, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}

	ctx, cancel := context.WithTimeout(context.WithoutCancel(t.Context()), 45*time.Second)
	command := exec.CommandContext(ctx, program, "-test.run=^"+processTest+"$", "-test.timeout=40s")

	command.Env = append(os.Environ(), "TERM=xterm-256color", "CHR_GUIDED_TUI_PROJECT="+operations.projectContext.ProjectDir, "CHR_GUIDED_TUI_ID="+operations.projectContext.ProjectID, "CHR_GUIDED_TUI_ROOT="+filepath.Dir(filepath.Dir(operations.env.ConfigDir())))
	command.Stdin, command.Stdout, command.Stderr = slave, slave, slave
	command.SysProcAttr = &syscall.SysProcAttr{Setsid: true, Setctty: true, Ctty: 0}

	terminal := &guidedTerminal{master: master, done: make(chan error, 1), readDone: make(chan error, 1)}

	if err := command.Start(); err != nil {
		cancel()
		t.Fatal(err)
	}

	if err := slave.Close(); err != nil {
		t.Fatal(err)
	}

	go func() { _, err := io.Copy(&terminal.output, master); terminal.readDone <- err }()
	go func() { terminal.done <- command.Wait() }()

	t.Cleanup(func() { defer cancel(); terminal.stop(t) })
	terminal.see(t, "\x1b[?1049h")

	return terminal
}

func openGuidedPTY(t *testing.T) (*os.File, *os.File) {
	t.Helper()

	master, err := os.OpenFile("/dev/ptmx", os.O_RDWR|syscall.O_NOCTTY, 0)
	if err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() { closeGuidedPTY(t, master) })

	if err := unix.IoctlSetPointerInt(int(master.Fd()), unix.TIOCSPTLCK, 0); err != nil {
		t.Fatal(err)
	}

	number, err := unix.IoctlGetInt(int(master.Fd()), unix.TIOCGPTN)
	if err != nil {
		t.Fatal(err)
	}

	slave, err := os.OpenFile(fmt.Sprintf("/dev/pts/%d", number), os.O_RDWR|syscall.O_NOCTTY, 0)
	if err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() { closeGuidedPTY(t, slave) })

	if err := unix.IoctlSetWinsize(int(slave.Fd()), unix.TIOCSWINSZ, &unix.Winsize{Row: 44, Col: 120}); err != nil {
		t.Fatal(err)
	}

	return master, slave
}

func (p *guidedTerminal) keys(t *testing.T, value string) {
	t.Helper()

	if _, err := io.WriteString(p.master, value); err != nil {
		t.Fatal(err)
	}
}

func (p *guidedTerminal) catalog(t *testing.T) {
	t.Helper()

	p.keys(t, "\x1bOP")
	p.keys(t, "o")
	p.see(t, "Case: ")
}

func (p *guidedTerminal) see(t *testing.T, value string) {
	t.Helper()

	deadline := time.NewTimer(10 * time.Second)
	defer deadline.Stop()

	tick := time.NewTicker(10 * time.Millisecond)
	defer tick.Stop()

	for {
		if index := strings.Index(p.output.text()[p.seen:], value); index >= 0 {
			p.seen += index + len(value)

			return
		}

		select {
		case <-deadline.C:
			t.Fatalf("terminal did not render %q: %s", value, terminalOutputTail(p.output.text()))
		case <-tick.C:
		}
	}
}

func (p *guidedTerminal) stop(t *testing.T) {
	t.Helper()

	if p.closed {
		return
	}

	p.closed = true
	p.keys(t, "\x03")

	select {
	case err := <-p.done:
		if err != nil {
			t.Errorf("managed terminal did not shut down cleanly: %v: %s", err, terminalOutputTail(p.output.text()))
		}
	case <-time.After(5 * time.Second):
		t.Error("managed terminal did not join its owned host")
	}

	closeGuidedPTY(t, p.master)

	if err := <-p.readDone; err != nil && !errors.Is(err, syscall.EIO) && !errors.Is(err, os.ErrClosed) {
		t.Errorf("read terminal output: %v", err)
	}
}

func terminalOutputTail(text string) string {
	return text[max(0, len(text)-6000):]
}

func closeGuidedPTY(t *testing.T, file *os.File) {
	t.Helper()

	if err := file.Close(); err != nil && !errors.Is(err, os.ErrClosed) {
		t.Errorf("close terminal: %v", err)
	}
}

func pauseAuthoringForReentry(t *testing.T, terminal *guidedTerminal, operations service, before authoringcore.Session) {
	t.Helper()
	// Authoring has no Pause control. Seed the retained paused state through
	// its existing owner; the execution terminal test verifies the visible Pause.
	if _, err := operations.ControlGuidedUseCase(t.Context(), guidedQuizCommand(before, "paused-reentry", "pause", "")); err != nil {
		t.Fatal(err)
	}

	terminal.keys(t, "r")
	terminal.see(t, "Dispatch stopped: pause")
}
