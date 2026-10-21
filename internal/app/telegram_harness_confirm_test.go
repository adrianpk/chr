package app

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"forge.adrianpk.com/chr/internal/core"
)

func telegramHarnessOffer(t *testing.T, f *telegramConversationAppFixture, text string) (core.TelegramOperationRequest, core.TelegramAction, telegramConversationNotice) {
	t.Helper()

	input := telegramOperationRetainInput(t, f, text)
	if err := f.receiver.drain(t.Context()); err != nil {
		t.Fatal(err)
	}

	request, found, err := f.receiver.store.TelegramOperation(t.Context(), input.Update, 0)
	if err != nil || !found {
		t.Fatal("selection request was not retained", err)
	}

	action, err := f.receiver.store.TelegramRequestAction(t.Context(), request)
	if err != nil {
		t.Fatal("selection confirmation was not offered", err)
	}

	return request, action, telegramOperationNotice(t, f)
}

// The real receiver freezes both role choices before publication, records the
// mapped Telegram actor, and releases prompt capacity after confirm or cancel.
func TestTelegramHarnessSelectionConfirm(t *testing.T) {
	f, operations, marker := telegramHarnessFixture(t)
	f.send(t, "/project example-project", 0)

	before := webHarnessSnapshot(t, operations)
	request, action, prompt := telegramHarnessOffer(t, f, "/harness worker=beta generation=pi-writing")

	if webHarnessSnapshot(t, operations) != before || !strings.Contains(prompt.Text, "beta · Codex · model-b") || !strings.Contains(prompt.Text, "pi-writing · Pi · example/model-pi") {
		t.Fatal("preparation changed defaults or omitted exact choices", prompt)
	}

	if reply := f.send(t, "confirm", prompt.ID); !strings.Contains(reply.Text, "Harness defaults saved") {
		t.Fatal("reply confirmation failed", reply)
	}

	saved := telegramHarnessCheckSaved(t, operations, request.ID)
	checkTelegramHarnessReplay(t, f, operations, action, saved)

	if reply := f.send(t, "/answer "+action.Token+" cancel", 0); !strings.Contains(reply.Text, "unavailable") || webHarnessSnapshot(t, operations) != saved {
		t.Fatal("consumed confirmation exchanged its choice", reply)
	}

	telegramHarnessClearAndCancel(t, f, operations)

	if _, err := os.Stat(marker); !os.IsNotExist(err) {
		t.Fatal("confirmation launched a driver")
	}

	checkTelegramNoWork(t, f.receiver.env)
}

func telegramHarnessCheckSaved(t *testing.T, operations service, requestID string) core.HarnessSelectionSnapshot {
	t.Helper()

	saved := webHarnessSnapshot(t, operations)
	if saved.Worker != "beta" || saved.Generation != "pi-writing" || saved.RequestID != requestID || saved.ChangedBy != "maintainer" {
		t.Fatal("selection lost mapped actor or atomic choices", saved)
	}

	return saved
}

func telegramHarnessClearAndCancel(t *testing.T, f *telegramConversationAppFixture, operations service) {
	t.Helper()

	for range 9 {
		_, action, _ := telegramHarnessOffer(t, f, "/harness worker=- generation=-")
		if notice := f.send(t, "/answer "+action.Token+" confirm", 0); !strings.Contains(notice.Text, "Harness defaults saved") {
			t.Fatal("terminal prompts exhausted confirmation capacity", notice)
		}
	}

	cleared := webHarnessSnapshot(t, operations)
	if cleared.Worker != "" || cleared.Generation != "" {
		t.Fatal("explicit clear did not restore legacy dispatch", cleared)
	}

	_, action, _ := telegramHarnessOffer(t, f, "/harness worker=alpha")
	if notice := f.send(t, "/answer "+action.Token+" cancel", 0); !strings.Contains(notice.Text, "cancelled") || webHarnessSnapshot(t, operations) != cleared {
		t.Fatal("cancel changed selections", notice)
	}
}

// Authority and preparation are checked again at publication. Neither a changed
// configuration nor a different selected project can inherit an old approval.
func TestTelegramHarnessSelectionReject(t *testing.T) {
	for _, test := range []struct {
		name   string
		change func(*testing.T, *telegramConversationAppFixture, service)
	}{
		{"stale revision", telegramHarnessStaleConfig},
		{"selected project", telegramHarnessOtherProject},
		{"registry root", telegramHarnessOtherRoot},
		{"disabled project", telegramHarnessDisableProject},
		{"mapped actor", telegramHarnessOtherActor},
		{"revoked project", telegramHarnessRevokeProject},
		{"unprepared profile", telegramHarnessBreakProfiles},
	} {
		t.Run(test.name, func(t *testing.T) {
			f, operations, _ := telegramHarnessFixture(t)
			f.send(t, "/project example-project", 0)

			_, action, _ := telegramHarnessOffer(t, f, "/harness worker=beta generation=pi-writing")
			test.change(t, f, operations)

			before := webHarnessSnapshot(t, operations)

			notice := f.send(t, "/answer "+action.Token+" confirm", 0)
			if strings.Contains(notice.Text, "Harness defaults saved") || webHarnessSnapshot(t, operations) != before {
				t.Fatal("invalid confirmation changed the configuration", notice)
			}
		})
	}
}

func telegramHarnessStaleConfig(t *testing.T, _ *telegramConversationAppFixture, operations service) {
	t.Helper()

	handler, cookies := webHarnessAuthenticated(t, operations)
	form := webHarnessForm(t, handler, cookies)
	form.Set("worker", "alpha")
	webHarnessAssertWrite(t, handler, cookies, form, "ok")
}

func telegramHarnessOtherProject(t *testing.T, f *telegramConversationAppFixture, _ service) {
	t.Helper()

	config, err := f.receiver.currentConfig()
	if err != nil {
		t.Fatal(err)
	}

	config.AllowedUsers[0].ProjectIDs = append(config.AllowedUsers[0].ProjectIDs, "private-project")
	if err := core.SaveTelegramReceiverConfig(f.receiver.env, config, true); err != nil {
		t.Fatal(err)
	}

	f.send(t, "/project private-project", 0)
}

func telegramHarnessOtherRoot(t *testing.T, f *telegramConversationAppFixture, _ service) {
	t.Helper()

	registry, err := core.NewProjectRegistry(t.Context(), f.receiver.env, testAssetsFS())
	if err != nil {
		t.Fatal(err)
	}

	_, updateErr := registry.UpdateProjectPath(t.Context(), "example-project", t.TempDir())

	closeErr := registry.Close()
	if updateErr != nil || closeErr != nil {
		t.Fatal(updateErr, closeErr)
	}
}

func telegramHarnessDisableProject(t *testing.T, f *telegramConversationAppFixture, _ service) {
	t.Helper()

	registry, err := core.NewProjectRegistry(t.Context(), f.receiver.env, testAssetsFS())
	if err != nil {
		t.Fatal(err)
	}

	_, disableErr := registry.DisableProject(t.Context(), "example-project")

	closeErr := registry.Close()
	if disableErr != nil || closeErr != nil {
		t.Fatal(disableErr, closeErr)
	}
}

func telegramHarnessOtherActor(t *testing.T, f *telegramConversationAppFixture, _ service) {
	t.Helper()

	config, err := f.receiver.currentConfig()
	if err != nil {
		t.Fatal(err)
	}

	config.AllowedUsers[0].OperatorID = "different-operator"
	if err := core.SaveTelegramReceiverConfig(f.receiver.env, config, true); err != nil {
		t.Fatal(err)
	}
}

func telegramHarnessRevokeProject(t *testing.T, f *telegramConversationAppFixture, _ service) {
	t.Helper()

	config, err := f.receiver.currentConfig()
	if err != nil {
		t.Fatal(err)
	}

	config.AllowedUsers[0].ProjectIDs = []string{"disabled-project"}
	if err := core.SaveTelegramReceiverConfig(f.receiver.env, config, true); err != nil {
		t.Fatal(err)
	}
}

func telegramHarnessBreakProfiles(t *testing.T, _ *telegramConversationAppFixture, operations service) {
	t.Helper()
	harnessSelectionWrite(t, filepath.Join(operations.env.ConfigDir(), "harnesses.yml"), "profiles: [invalid\n")
}

// The retained envelope must agree with the original deterministic command;
// synthetic model proposals and changed arguments cannot create authority.
func TestTelegramHarnessSelectionEnvelope(t *testing.T) {
	f, _, _ := telegramHarnessFixture(t)
	f.send(t, "/project example-project", 0)

	request, _, _ := telegramHarnessOffer(t, f, "/harness worker=beta")

	for _, test := range []struct {
		name   string
		change func(*core.TelegramOperationRequest)
	}{
		{"model only", func(r *core.TelegramOperationRequest) { r.Source.Text = "Switch the worker to beta" }},
		{"changed arguments", func(r *core.TelegramOperationRequest) { r.Arguments = json.RawMessage(`{"worker":"alpha"}`) }},
		{"changed actor", func(r *core.TelegramOperationRequest) { r.Operator = "foreign" }},
		{"changed request", func(r *core.TelegramOperationRequest) { r.ID = "foreign" }},
		{"changed project", func(r *core.TelegramOperationRequest) { r.Target.ProjectID = "private-project" }},
	} {
		t.Run(test.name, func(t *testing.T) {
			changed := request
			test.change(&changed)

			if _, err := retainedTelegramHarnessChange(changed); err == nil {
				t.Fatal("altered confirmation retained configuration authority")
			}
		})
	}
}

func checkTelegramHarnessReplay(t *testing.T, f *telegramConversationAppFixture, operations service, action core.TelegramAction, saved core.HarnessSelectionSnapshot) {
	t.Helper()

	notice := f.send(t, "/answer "+action.Token+" confirm", 0)
	for _, part := range []string{"already confirmed", "Original defaults saved", "Worker: beta", "Generation: pi-writing"} {
		if !strings.Contains(notice.Text, part) {
			t.Fatal("replay omitted the original saved roles", part, notice)
		}
	}

	if webHarnessSnapshot(t, operations) != saved {
		t.Fatal("replay changed the saved defaults")
	}
}
