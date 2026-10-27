package app

import (
	"context"
	"encoding/json"
	"testing"

	"forge.adrianpk.com/chr/internal/core"
	telegramcore "forge.adrianpk.com/chr/internal/core/telegram"
)

// Even a configured model's response cannot synthesize a configuration command
// or private setup authority through the existing natural-language intent owner.
func TestTelegramHarnessSelectionModelAuthority(t *testing.T) {
	for _, test := range []struct{ name, output string }{
		{"role change", `{"operation_id":"harnesses.change","arguments":{"text":"/harness worker=beta"},"unresolved":[]}`},
		{"private setup", `{"operation_id":"harnesses.change","arguments":{"credential_ref":"foreign.json"},"unresolved":[]}`},
	} {
		t.Run(test.name, func(t *testing.T) {
			f, operations, _ := telegramHarnessFixture(t)
			calls := telegramAuthoringIntentHTTP(t, f, func(ConversationIntentContext) string { return test.output })
			f.send(t, "/project example-project", 0)

			before := webHarnessSnapshot(t, operations)
			f.send(t, "Change my project defaults and set up credentials", 0)

			if calls.Load() == 0 || webHarnessSnapshot(t, operations) != before {
				t.Fatal("model output gained configuration authority")
			}
		})
	}
}

// A separately authorized Telegram peer cannot use another actor's token,
// including when both peers can select the same registered project.
func TestTelegramHarnessSelectionForeignActor(t *testing.T) {
	f, operations, _ := telegramHarnessFixture(t)
	f.send(t, "/project example-project", 0)

	_, action, _ := telegramHarnessOffer(t, f, "/harness worker=beta")

	config, err := f.receiver.currentConfig()
	if err != nil {
		t.Fatal(err)
	}

	config.AllowedUsers = append(config.AllowedUsers, telegramcore.TelegramAllowedUser{TelegramUserID: 100002, OperatorID: "other-operator", ProjectIDs: []string{"example-project"}})
	if err := telegramcore.SaveTelegramReceiverConfig(f.receiver.env, config, true); err != nil {
		t.Fatal(err)
	}

	before := webHarnessSnapshot(t, operations)

	data, err := json.Marshal([]any{map[string]any{"update_id": f.next, "message": map[string]any{"message_id": f.next, "from": map[string]any{"id": 100002}, "chat": map[string]any{"id": 100002, "type": "private"}, "text": "/answer " + action.Token + " confirm"}}})
	if err != nil {
		t.Fatal(err)
	}

	f.next++
	f.updates <- data

	if _, err := f.receiver.step(t.Context()); err != nil || webHarnessSnapshot(t, operations) != before {
		t.Fatal("foreign actor changed the configuration", err)
	}
}

type telegramHarnessAuthorityProbe struct {
	service
	change func()
}

func (s telegramHarnessAuthorityProbe) ChangeTelegramHarnessSelections(ctx context.Context, target telegramcore.TelegramOperationTarget, request core.HarnessSelectionChangeRequest, authorize func(context.Context) error) (core.HarnessSelectionSnapshot, error) {
	calls := 0

	return s.service.ChangeTelegramHarnessSelections(ctx, target, request, func(ctx context.Context) error {
		calls++
		if calls == 2 {
			s.change()
		}

		return authorize(ctx)
	})
}

// Revocation between the initial check and the locked publication check must
// still prevent the shared configuration owner from accepting the change.
func TestTelegramHarnessSelectionLockedAuthority(t *testing.T) {
	f, operations, _ := telegramHarnessFixture(t)
	f.send(t, "/project example-project", 0)

	_, action, _ := telegramHarnessOffer(t, f, "/harness worker=beta")
	changed := false
	f.receiver.operations = telegramHarnessAuthorityProbe{service: telegramOperationConcreteService(t, f.receiver), change: func() {
		changed = true

		telegramHarnessOtherActor(t, f, operations)
	}}
	before := webHarnessSnapshot(t, operations)
	f.send(t, "/answer "+action.Token+" confirm", 0)

	if !changed || webHarnessSnapshot(t, operations) != before {
		t.Fatal("publication omitted the current mapped actor check under exclusion")
	}
}

// Unknown, unprepared and private setup choices stop before reservation. The
// existing intent provider is never consulted for deterministic slash commands.
func TestTelegramHarnessSelectionUnready(t *testing.T) {
	f, operations, _ := telegramHarnessFixture(t)
	calls := telegramAuthoringIntentHTTP(t, f, func(ConversationIntentContext) string { return `{}` })
	f.send(t, "/project example-project", 0)

	before := webHarnessSnapshot(t, operations)

	for _, text := range []string{"/harness worker=missing", "/harness worker=unavailable", "/harness worker=pi-writing", "/harness credential_ref=foreign.json", "/harness worker=/private/driver"} {
		f.send(t, text, 0)

		if webHarnessSnapshot(t, operations) != before || len(f.sessions(t)) != 0 || calls.Load() != 0 {
			t.Fatal("invalid deterministic selection created authority or called a model")
		}
	}
}
