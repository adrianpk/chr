package core

import projectcore "forge.adrianpk.com/chr/internal/core/project"

import (
	"context"
	"os"
	"testing"
	"time"
)

func telegramControlRegistryFixture(t *testing.T) (UserEnvironment, *TelegramReceiverStore) {
	t.Helper()

	env := telegramReceiverTestEnv(t)
	store := openTelegramReceiverTestStore(t, env)

	registry, err := projectcore.NewProjectRegistry(t.Context(), env, os.DirFS("../../cmd/chr"))
	if err != nil {
		t.Fatal(err)
	}

	defer func() {
		if err := registry.Close(); err != nil {
			t.Error(err)
		}
	}()

	if err := registry.AddProject(t.Context(), projectcore.Project{ID: "registered", Name: "Registered project", Path: t.TempDir(), Enabled: true, RegisteredAt: time.Now()}); err != nil {
		t.Fatal(err)
	}

	return env, store
}

func telegramControlExclusiveWrite(t *testing.T, store *TelegramReceiverStore) func() {
	t.Helper()

	conn, err := store.db.Conn(t.Context())
	if err != nil {
		t.Fatal(err)
	}

	released := false
	release := func() {
		if released {
			return
		}

		if _, err := conn.ExecContext(context.Background(), "ROLLBACK"); err != nil {
			t.Error(err)
		}

		released = true
	}

	t.Cleanup(func() {
		release()

		if err := conn.Close(); err != nil {
			t.Error(err)
		}
	})

	if _, err := conn.ExecContext(t.Context(), "BEGIN EXCLUSIVE"); err != nil {
		t.Fatal(err)
	}

	return release
}

// TestTelegramControlRegistryWait keeps read-only authority lookup available
// across a short concurrent inbox write, without changing database access mode.
func TestTelegramControlRegistryWait(t *testing.T) {
	env, store := telegramControlRegistryFixture(t)
	ctx, cancel := context.WithCancel(t.Context())
	finished := make(chan error, 1)
	pending := false

	t.Cleanup(func() {
		cancel()

		if pending {
			select {
			case <-finished:
			case <-time.After(3 * time.Second):
				t.Error("owned reader did not stop")
			}
		}
	})

	release := telegramControlExclusiveWrite(t, store)
	pending = true

	go func() { finished <- telegramControlReadProject(ctx, env) }()

	select {
	case err := <-finished:
		pending = false

		t.Fatal("authority read failed to wait for the writer", err)
	case <-time.After(100 * time.Millisecond):
	}

	release()

	select {
	case err := <-finished:
		pending = false

		if err != nil {
			t.Fatal("authority lookup did not recover after the write", err)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("authority lookup stayed blocked")
	}
}

func telegramControlReadProject(ctx context.Context, env UserEnvironment) error {
	projects, err := projectcore.ReadRegisteredProjects(ctx, env, []string{"registered"})
	if err == nil && len(projects) != 1 {
		return projectcore.ErrProjectNotFound
	}

	return err
}
