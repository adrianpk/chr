package app

import (
	"context"
	"testing"

	"forge.adrianpk.com/chr/internal/core"
)

type telegramWorkChangeAuthority struct {
	service
	change func()
}

func (s telegramWorkChangeAuthority) StartTelegramWork(ctx context.Context, target core.TelegramOperationTarget, preview WorkflowBoundPreview, binding WorkflowBoundAdmission) (AgentWorkResult, error) {
	s.change()

	return s.service.StartTelegramWork(ctx, target, preview, binding)
}

// TestTelegramWorkAuthority rejects a changed operator or registry root between
// the prompt claim and the native effect, without allocating work at either root.
func TestTelegramWorkAuthority(t *testing.T) {
	for _, change := range []string{"operator", "root"} {
		t.Run(change, func(t *testing.T) {
			f, operations, record := telegramWorkFixture(t)
			action, _ := telegramWorkOffer(t, f, "/work ticket "+record.Task.Key)
			f.receiver.operations = telegramWorkChangeAuthority{service: operations, change: func() { telegramWorkChangeActor(t, f, change) }}
			f.send(t, "/answer "+action.Token+" start", 0)
			checkTelegramWorkRunCount(t, operations, 0)
		})
	}
}

func telegramWorkChangeActor(t *testing.T, f *telegramConversationAppFixture, change string) {
	t.Helper()

	if change == "operator" {
		config, err := f.receiver.currentConfig()
		if err != nil {
			t.Fatal(err)
		}

		config.AllowedUsers[0].OperatorID = "another-operator"
		if err := core.SaveTelegramReceiverConfig(f.receiver.env, config, true); err != nil {
			t.Fatal(err)
		}

		return
	}

	registry, err := core.NewProjectRegistry(t.Context(), f.receiver.env, testAssetsFS())
	if err != nil {
		t.Fatal(err)
	}

	if _, err := registry.UpdateProjectPath(t.Context(), "Workflow project", t.TempDir()); err != nil {
		t.Fatal(err)
	}

	if err := registry.Close(); err != nil {
		t.Fatal(err)
	}
}
