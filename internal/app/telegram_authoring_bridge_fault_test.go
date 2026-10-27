package app

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"sync/atomic"
	"testing"

	"forge.adrianpk.com/chr/internal/core"
	opsnamespace "forge.adrianpk.com/chr/internal/core/ops/namespace"
	projectcore "forge.adrianpk.com/chr/internal/core/project"
	telegramcore "forge.adrianpk.com/chr/internal/core/telegram"
)

var errTelegramAuthoringLostAck = errors.New("injected lost authoring acknowledgement")

type telegramAuthoringLostAck struct {
	service
	lost *atomic.Bool
}

func (s telegramAuthoringLostAck) ContinueTelegramAuthoring(ctx context.Context, target telegramcore.OperationTarget, request core.AuthoringContinuationRequest, guard core.AuthoringCommitGuard) (core.AuthoringContinuationResult, error) {
	result, err := s.service.ContinueTelegramAuthoring(ctx, target, request, guard)
	if err == nil && s.lost.CompareAndSwap(false, true) {
		return result, errTelegramAuthoringLostAck
	}

	return result, err
}

// TestTelegramAuthoringBridgeLostAck reopens SQLite after a confirmed canonical
// write, preserves an independent edit and recovers one source acknowledgement.
func TestTelegramAuthoringBridgeLostAck(t *testing.T) {
	t.Setenv(opsnamespace.UserEnv, "original")

	f := newTelegramConversationAppFixture(t)
	calls := telegramAuthoringBridgeHTTP(t, f, func(authoringModelInput) string {
		return telegramAuthoringServiceOutput(t, "Preserve five attempts.", "")
	}, telegramAuthoringBridgeCapture)
	f.send(t, "/prepare retry scope", 0)

	c := f.sessions(t)[0]
	operations := telegramOperationConcreteService(t, f.receiver)
	f.receiver.operations = telegramAuthoringLostAck{service: operations, lost: &atomic.Bool{}}

	input := telegramOperationRetainInput(t, f, "/done "+c.ID)
	if err := f.receiver.respond(t.Context(), input.Update); !errors.Is(err, errTelegramAuthoringLostAck) {
		t.Fatal("owner acknowledgement was not interrupted", err)
	}

	if f.session(t, c.ID).AuthoringSessionRef != "" {
		t.Fatal("conversation advanced before owner acknowledgement")
	}

	root, output, retained := telegramAuthoringLostAckEdit(t, operations)
	telegramAuthoringReopen(t, f)

	f.receiver.operations = operations

	t.Setenv(opsnamespace.UserEnv, "changed-after-write")

	if err := f.receiver.drain(t.Context()); err != nil {
		t.Fatal(err)
	}

	content, err := os.ReadFile(filepath.Join(root, output))
	if err != nil || string(content) != retained || calls.Load() != 1 || len(f.notices) != 1 {
		t.Fatal("recovery repeated its canonical effect", string(content), calls.Load(), len(f.notices), err)
	}

	if f.session(t, c.ID).Phase != "prepared" || f.session(t, c.ID).Namespace != "original" {
		t.Fatal("recovery lost the frozen conversation", f.session(t, c.ID))
	}

	telegramAuthoringBridgeNoWork(t, f)
}

func telegramAuthoringLostAckEdit(t *testing.T, operations service) (string, string, string) {
	t.Helper()

	projects, err := operations.ListRegisteredProjects([]string{"example-project"})
	if err != nil || len(projects) != 1 {
		t.Fatal("project unavailable", projects, err)
	}

	root := projects[0].Path
	repository := core.NewAuthoringSessionRepository(root)

	sessions, err := repository.ListSessions(t.Context())
	if err != nil || len(sessions) != 1 || len(sessions[0].Outputs) != 1 {
		t.Fatal("canonical owner did not commit before interruption", sessions, err)
	}

	operations.projectContext = commandProjectContext{ProjectID: projects[0].ID, ProjectDir: root}

	output := sessions[0].Outputs[0].Path
	if _, err := saveObservedDump(t, operations, DumpSaveRequest{RelativePath: output, Content: "Independent later edit"}); err != nil {
		t.Fatal(err)
	}

	retained, err := os.ReadFile(filepath.Join(root, output))
	if err != nil {
		t.Fatal(err)
	}

	return root, output, string(retained)
}

func telegramAuthoringReopen(t *testing.T, f *telegramConversationAppFixture) {
	t.Helper()

	if err := f.receiver.store.Close(); err != nil {
		t.Fatal(err)
	}

	store, err := telegramcore.NewReceiverStore(t.Context(), f.receiver.env, testAssetsFS())
	if err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() {
		if err := store.Close(); err != nil {
			t.Error(err)
		}
	})

	f.receiver.store = store
}

// TestTelegramAuthoringBridgeAuthority rechecks root and operator changes after
// the real model call, before canonical session/artifact writes can be committed.
func TestTelegramAuthoringBridgeAuthority(t *testing.T) {
	for _, change := range []string{"root", "operator"} {
		t.Run(change, func(t *testing.T) {
			f := newTelegramConversationAppFixture(t)
			operations := telegramOperationConcreteService(t, f.receiver)

			projects, err := operations.ListRegisteredProjects([]string{"example-project"})
			if err != nil || len(projects) != 1 {
				t.Fatal("project unavailable", projects, err)
			}

			root := projects[0].Path
			calls := telegramAuthoringBridgeHTTP(t, f, func(authoringModelInput) string {
				telegramAuthoringChangeAuthority(t, f, change)

				return telegramAuthoringServiceOutput(t, "An unconfirmed artifact.", "")
			}, telegramAuthoringBridgeCapture)
			f.send(t, "/prepare guarded generation", 0)

			c := f.sessions(t)[0]
			f.send(t, "Keep this exact captured context.", 0)

			before := f.session(t, c.ID)
			f.send(t, "/done "+c.ID, 0)

			if f.session(t, c.ID) != before || calls.Load() != 1 {
				t.Fatal("changed authority advanced the conversation")
			}

			sessions, err := core.NewAuthoringSessionRepository(root).ListSessions(t.Context())
			if err != nil || len(sessions) != 0 {
				t.Fatal("changed authority committed canonical authoring", sessions, err)
			}

			telegramAuthoringBridgeNoWork(t, f)
		})
	}
}

func telegramAuthoringChangeAuthority(t *testing.T, f *telegramConversationAppFixture, change string) {
	t.Helper()

	if change == "operator" {
		config, err := f.receiver.currentConfig()
		if err != nil {
			t.Error(err)

			return
		}

		config.AllowedUsers[0].OperatorID = "replacement-operator"
		if err := telegramcore.SaveReceiverConfig(f.receiver.env, config, true); err != nil {
			t.Error(err)
		}

		return
	}

	registry, err := projectcore.NewRegistry(t.Context(), f.receiver.env, testAssetsFS())
	if err != nil {
		t.Error(err)

		return
	}

	if _, err := registry.UpdateProjectPath(t.Context(), "example-project", t.TempDir()); err != nil {
		t.Error(err)
	}

	if err := registry.Close(); err != nil {
		t.Error(err)
	}
}
