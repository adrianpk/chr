package app

import (
	"encoding/json"
	"testing"

	"forge.adrianpk.com/chr/internal/core"
)

// TestGuidedUseCaseTelegramSourceGuards rejects another allowed peer's button
// and replies/buttons made stale by an independent Web-side canonical mutation.
func TestGuidedUseCaseTelegramSourceGuards(t *testing.T) {
	for _, name := range []string{"foreign button", "foreign reply", "stale reply", "stale button"} {
		t.Run(name, func(t *testing.T) { checkGuidedTelegramSourceGuard(t, name) })
	}
}

func checkGuidedTelegramSourceGuard(t *testing.T, fault string) {
	t.Helper()

	f, operations := guidedTelegramFixture(t)

	config, err := core.LoadTelegramReceiverConfig(f.receiver.env)
	if err != nil {
		t.Fatal(err)
	}

	config.AllowedUsers = append(config.AllowedUsers, core.TelegramAllowedUser{TelegramUserID: 100002, OperatorID: "other", ProjectIDs: []string{"example-project"}})
	if err := core.SaveTelegramReceiverConfig(f.receiver.env, config, true); err != nil {
		t.Fatal(err)
	}

	start := f.send(t, "Develop a new use case", 0)
	name := f.send(t, "Guarded capture", start.ID)

	before := readGuidedTelegramSession(t, operations, f.sessions(t)[0])

	switch fault {
	case "foreign reply":
		sendForeignGuidedTelegramReply(t, f, name.ID)
	case "stale reply":
		before = changeGuidedTelegramFromWeb(t, operations, before)
		f.send(t, "Must not replace the current Web input.", name.ID)
	default:
		controls := f.send(t, "Show controls", name.ID)
		_, action := guidedTelegramAction(t, f, controls)
		peer := int64(100002)

		if fault == "stale button" {
			before = changeGuidedTelegramFromWeb(t, operations, before)
			peer = 100001
		}

		sendGuidedTelegramCallback(t, f, action, "finish", peer)
	}

	after := readGuidedTelegramSession(t, operations, f.sessions(t)[0])
	if after.Revision != before.Revision || after.Guided.Capture != before.Guided.Capture || after.Guided.Stage != "dump" {
		t.Fatal("stale or foreign source changed the canonical run")
	}
}

func sendForeignGuidedTelegramReply(t *testing.T, f *telegramConversationAppFixture, reply int64) {
	t.Helper()

	message := map[string]any{"message_id": f.next, "from": map[string]any{"id": 100002}, "chat": map[string]any{"id": 100002, "type": "private"}, "text": "Must not become another person's capture.", "reply_to_message": map[string]any{"message_id": reply}}

	data, err := json.Marshal([]any{map[string]any{"update_id": f.next, "message": message}})
	if err != nil {
		t.Fatal(err)
	}

	f.next++
	f.updates <- data

	if _, err := f.receiver.step(t.Context()); err != nil {
		t.Fatal(err)
	}
}

func changeGuidedTelegramFromWeb(t *testing.T, operations service, before core.AuthoringSession) core.AuthoringSession {
	t.Helper()

	operations.projectContext = commandProjectContext{ProjectID: before.ProjectID, ProjectDir: guidedTelegramProject(t, operations, before.ProjectID)}

	after, err := operations.ChangeGuidedUseCase(t.Context(), core.GuidedUseCaseCommand{GuidedUseCaseScope: core.GuidedUseCaseScope{RunID: before.ID, ProjectID: before.ProjectID, Namespace: before.Namespace}, ExpectedRevision: before.Revision, OperationID: "web-independent-capture", Action: "append", Content: "Independent Web capture."})
	if err != nil {
		t.Fatal(err)
	}

	return after
}
