package app

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"reflect"
	"syscall"
	"testing"
	"time"

	"forge.adrianpk.com/chr/internal/core"
)

// TestTelegramRecoverySignal sends SIGTERM only to its own receiver process.
// Replacing the installed executable leaves that running build and retained
// conversation intact; shutdown cancels long polling and releases OS exclusion.
func TestTelegramRecoverySignal(t *testing.T) {
	if root := os.Getenv("CHR_TEST_RECOVERY_SERVE_ROOT"); root != "" {
		telegramRecoveryServeChild(t, root)

		return
	}

	f := newTelegramConversationAppFixture(t)
	f.send(t, "/prepare retained context", 0)

	before := f.sessions(t)
	polling, release := make(chan struct{}), make(chan struct{})

	server := telegramShutdownTestServer(t, polling, release)
	defer server.Close()
	defer close(release)

	installed := filepath.Join(t.TempDir(), "receiver")
	copyTelegramRecoveryExecutable(t, installed, "")

	child := exec.CommandContext(t.Context(), installed, "-test.run=^TestTelegramRecoverySignal$")
	root := filepath.Dir(filepath.Dir(f.receiver.env.ConfigDir()))
	child.Env = append(os.Environ(), "CHR_TEST_RECOVERY_SERVE_ROOT="+root, "CHR_TEST_RECOVERY_SERVE_API="+server.URL)

	var output bytes.Buffer

	child.Stdout, child.Stderr = &output, &output
	if err := child.Start(); err != nil {
		t.Fatal(err)
	}

	finished := make(chan error, 1)
	go func() { finished <- child.Wait(); close(finished) }()

	t.Cleanup(func() { waitTelegramRecoveryReceiver(t, finished) })

	select {
	case <-polling:
	case err := <-finished:
		t.Fatal("fixture receiver did not start", err, output.String())
	case <-time.After(10 * time.Second):
		t.Fatal("fixture receiver did not poll")
	}

	checkTelegramRecoveryFixedBuild(t, child.Process.Pid, installed)

	if err := child.Process.Signal(syscall.SIGTERM); err != nil {
		t.Fatal(err)
	}

	waitTelegramRecoveryReceiver(t, finished)
	checkTelegramShutdownReleased(t, f.receiver.env)

	if after := f.sessions(t); !reflect.DeepEqual(before, after) {
		t.Fatal("SIGTERM changed retained conversation", after)
	}
}

func telegramRecoveryServeChild(t *testing.T, root string) {
	t.Helper()

	env, err := core.NewUserEnvironment(core.UserEnvironmentOptions{DevMode: true, DevRoot: root})
	if err != nil {
		t.Fatal(err)
	}

	ctx, cancel := signal.NotifyContext(t.Context(), syscall.SIGTERM)
	defer cancel()

	api := core.TelegramReceiverAPI{Client: &http.Client{}, BaseURL: os.Getenv("CHR_TEST_RECOVERY_SERVE_API")}
	if err := serveTelegramReceiver(ctx, env, testAssetsFS(), api); err != nil {
		t.Fatal(err)
	}
}

func copyTelegramRecoveryExecutable(t *testing.T, target, footer string) {
	t.Helper()

	source, err := os.Open(os.Args[0])
	if err != nil {
		t.Fatal(err)
	}

	defer func() {
		if err := source.Close(); err != nil {
			t.Error(err)
		}
	}()

	candidate, err := os.OpenFile(target, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o700)
	if err != nil {
		t.Fatal(err)
	}

	_, copyErr := io.Copy(candidate, source)

	_, writeErr := io.WriteString(candidate, footer)
	if err := errors.Join(copyErr, writeErr, candidate.Close()); err != nil {
		t.Fatal(err)
	}
}

func checkTelegramRecoveryFixedBuild(t *testing.T, pid int, installed string) {
	t.Helper()

	running := fmt.Sprintf("/proc/%d/exe", pid)

	original, err := os.Stat(running)
	if err != nil {
		t.Fatal(err)
	}

	candidate := installed + ".candidate"
	copyTelegramRecoveryExecutable(t, candidate, "\nfixture candidate build\n")

	if err := os.Rename(candidate, installed); err != nil {
		t.Fatal(err)
	}

	current, err := os.Stat(running)
	if err != nil {
		t.Fatal(err)
	}

	replacement, err := os.Stat(installed)
	if err != nil {
		t.Fatal(err)
	}

	if !os.SameFile(original, current) || os.SameFile(current, replacement) {
		t.Fatal("source/install change replaced the live coordinating image")
	}
}
