package telegram

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"sync"
	"testing"
	"time"
)

// TestTelegramRecoveryModelOwner uses a real child that exits with a reserved
// model call. A later clock cannot steal its live lock or reset retained attempts.
func TestTelegramRecoveryModelOwner(t *testing.T) {
	if root := os.Getenv("CHR_TEST_RECOVERY_MODEL_ROOT"); root != "" {
		telegramRecoveryModelChild(t, root)

		return
	}

	root := t.TempDir()

	env := telegramTestEnvironment{root: root}

	store := openTelegramReceiverTestStore(t, env)

	update := telegramRetainedTestUpdate(t, 71, "retain this model request")
	if err := store.RetainBatch(t.Context(), 12345, []ReceivedUpdate{update}, time.Now()); err != nil {
		t.Fatal(err)
	}

	finish := startTelegramRecoveryModel(t, root)
	checkTelegramRecoveryLiveModel(t, store, env, update)
	finish()

	if err := store.Close(); err != nil {
		t.Fatal(err)
	}

	store = openTelegramReceiverTestStore(t, env)

	result, err := store.ResolveInterpretation(t.Context(), env, update, time.Now().Add(2*time.Minute), func(context.Context) (json.RawMessage, error) {
		return json.RawMessage(`{"operation_id":"help"}`), nil
	})
	if err != nil || result.Attempts != 2 || len(result.Decision) == 0 {
		t.Fatal("dead model owner did not retain its attempt history", result, err)
	}
}

func startTelegramRecoveryModel(t *testing.T, root string) func() {
	t.Helper()

	child := exec.CommandContext(t.Context(), os.Args[0], "-test.run=^TestTelegramRecoveryModelOwner$")

	child.Env = append(os.Environ(), "CHR_TEST_RECOVERY_MODEL_ROOT="+root)

	stdout, err := child.StdoutPipe()
	if err != nil {
		t.Fatal(err)
	}

	stdin, err := child.StdinPipe()
	if err != nil {
		t.Fatal(err)
	}

	var diagnostic bytes.Buffer

	child.Stderr = &diagnostic
	if err := child.Start(); err != nil {
		t.Fatal(err)
	}

	done := make(chan error, 1)
	go func() { done <- child.Wait(); close(done) }()

	release := sync.OnceFunc(func() {
		if err := stdin.Close(); err != nil {
			t.Error(err)
		}
	})
	finish := func() {
		release()

		select {
		case err := <-done:
			if err != nil {
				t.Error("owned model process did not exit", err, diagnostic.String())
			}
		case <-time.After(10 * time.Second):
			t.Error("owned model process did not finish")
		}
	}
	t.Cleanup(finish)
	waitTelegramRecoveryModelReady(t, stdout)

	return finish
}

func waitTelegramRecoveryModelReady(t *testing.T, stdout io.Reader) {
	t.Helper()

	ready := make(chan string, 1)

	go func() {
		line, err := bufio.NewReader(stdout).ReadString('\n')
		if err != nil {
			ready <- "model_ready_read_failed"

			return
		}

		ready <- strings.TrimSpace(line)
	}()

	select {
	case line := <-ready:
		if line != "reserved" {
			t.Fatal("child did not reserve a model claim", line)
		}
	case <-time.After(10 * time.Second):
		t.Fatal("child model claim unavailable")
	}
}

func checkTelegramRecoveryLiveModel(t *testing.T, store *ReceiverStore, env telegramTestEnvironment, update ReceivedUpdate) {
	t.Helper()

	called := false
	_, err := store.ResolveInterpretation(t.Context(), env, update, time.Now().Add(2*time.Minute), func(context.Context) (json.RawMessage, error) {
		called = true

		return nil, nil
	})

	var conflict *APIError
	if !errors.As(err, &conflict) || conflict.Code != "local_receiver_conflict" || called {
		t.Fatal("live process model claim was stolen", err, called)
	}

	var attempts int
	if err := store.db.QueryRowContext(t.Context(), `SELECT interpretation_attempt FROM telegram_updates WHERE bot_id=12345 AND update_id=71`).Scan(&attempts); err != nil || attempts != 1 {
		t.Fatal("live-owner rejection spent an attempt", attempts, err)
	}
}

func telegramRecoveryModelChild(t *testing.T, root string) {
	t.Helper()

	env := telegramTestEnvironment{root: root}

	store := openTelegramReceiverTestStore(t, env)
	update := telegramRetainedTestUpdate(t, 71, "retain this model request")
	_, err := store.ResolveInterpretation(t.Context(), env, update, time.Now(), func(context.Context) (json.RawMessage, error) {
		if _, err := fmt.Fprintln(os.Stdout, "reserved"); err != nil {
			return nil, err
		}

		if _, err := io.Copy(io.Discard, os.Stdin); err != nil {
			return nil, err
		}
		// Deliberately abandon the fixture claim and connection; process death, not
		// a deferred unlock or guessed lease timeout, must release OS ownership.
		os.Exit(0)

		return nil, nil
	})
	t.Fatal("fixture model process returned without abandonment", err)
}
