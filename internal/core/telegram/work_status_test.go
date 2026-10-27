package telegram

import (
	"errors"
	"testing"
	"time"

	projectcore "forge.adrianpk.com/chr/internal/core/project"
	tgmessage "forge.adrianpk.com/chr/internal/core/telegram/message"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

// A completed native message stays inspectable after capture is closed, without
// becoming a capture prompt or a control authorization.
func TestTelegramWorkStatusReply(t *testing.T) {
	f, input, expected := telegramWorkStatusFixture(t)

	target, err := f.store.ResolveWorkStatusTarget(t.Context(), input, time.Now())
	if err != nil || target != expected {
		t.Fatal("completed status reply lost its exact target", target, err)
	}

	if _, err := f.store.ResolveOperationTarget(t.Context(), input, time.Now()); err == nil {
		t.Fatal("status reply reopened capture")
	}

	if _, err := f.store.ResolveControlTarget(t.Context(), input, time.Now()); err == nil {
		t.Fatal("status reply authorized a control target")
	}

	checkTelegramWorkStatusQueue(t, f, input, target)
}

func checkTelegramWorkStatusQueue(t *testing.T, f *telegramConversationFixture, input ConversationInput, target OperationTarget) {
	t.Helper()

	before := f.sessions(t)
	message := tgmessage.Blocks{Family: "33", InitialText: "Work has finished."}

	if err := f.store.QueueWorkStatus(t.Context(), WorkStatusRequest{Input: input, Target: target, Invocation: "run-status", Message: message, At: time.Now()}); err != nil {
		t.Fatal(err)
	}

	acknowledgeTelegramStatusNotices(t, f, 77778)

	snapshot, err := f.store.IntentContext(t.Context(), input)
	if err != nil || snapshot.LastWorkConversation != target.ConversationID || snapshot.RepliedWorkConversation != target.ConversationID {
		t.Fatal("acknowledged status was absent from intent context", snapshot, err)
	}

	after := f.sessions(t)
	if len(before) != len(after) || before[0] != after[0] {
		t.Fatal("status read changed its execution conversation")
	}

	checkTelegramWorkReplyPrecedence(t, f, input, target)
}

func acknowledgeTelegramStatusNotices(t *testing.T, f *telegramConversationFixture, firstID int64) {
	t.Helper()

	due, err := f.store.DueDeliveries(t.Context(), 12345, time.Now())
	if err != nil {
		t.Fatal(err)
	}

	for index, delivery := range due {
		if err := f.store.ReserveDelivery(t.Context(), delivery.ID, time.Now()); err != nil {
			t.Fatal(err)
		}

		if err := f.store.DeliverySent(t.Context(), delivery.ID, firstID+int64(index), time.Now()); err != nil {
			t.Fatal(err)
		}
	}
}

func checkTelegramWorkReplyPrecedence(t *testing.T, f *telegramConversationFixture, input ConversationInput, original OperationTarget) {
	t.Helper()

	newer := telegramWorkReferenceFixture(t, f, "run-newer")
	acknowledgeTelegramStatusNotices(t, f, 90000)

	snapshot, err := f.store.IntentContext(t.Context(), input)
	if err != nil || snapshot.LastWorkConversation != newer.Request.Target.ConversationID || snapshot.RepliedWorkConversation != original.ConversationID {
		t.Fatal("a newer work message displaced the explicit reply reference", snapshot, err)
	}
}

// An acknowledged reference cannot cross peer, scope or current-root bindings.
// Undelivered message IDs cannot identify work even for an authorized peer.
func TestTelegramWorkStatusGuards(t *testing.T) {
	f, input, _ := telegramWorkStatusFixture(t)
	for _, test := range []struct {
		name   string
		change func(*ConversationInput)
	}{
		{"bot", func(i *ConversationInput) { i.Update.Peer.BotID++ }},
		{"chat", func(i *ConversationInput) { i.Update.Peer.ChatID++ }},
		{"topic", func(i *ConversationInput) { i.Update.Peer.TopicID++ }},
		{"user", func(i *ConversationInput) { i.Update.Peer.UserID++ }},
		{"unknown message", func(i *ConversationInput) { i.Update.ReplyToMessageID++ }},
		{"conversation conflict", func(i *ConversationInput) { i.Conversation = "other-session" }},
		{"project conflict", func(i *ConversationInput) { i.Project = "beta" }},
		{"namespace conflict", func(i *ConversationInput) { i.Namespace = "elsewhere" }},
		{"changed root", func(i *ConversationInput) {
			i.Projects = append([]projectcore.Project(nil), i.Projects...)
			i.Projects[0].Path = t.TempDir()
		}},
	} {
		t.Run(test.name, func(t *testing.T) {
			changed := input
			test.change(&changed)

			if _, err := f.store.ResolveWorkStatusTarget(t.Context(), changed, time.Now()); err == nil {
				t.Fatal("foreign, stale or unacknowledged work reference was accepted")
			}
		})
	}
}

func telegramWorkStatusFixture(t *testing.T) (*telegramConversationFixture, ConversationInput, OperationTarget) {
	t.Helper()

	f := newTelegramConversationFixture(t)
	ref := telegramWorkReferenceFixture(t, f, "run-status")
	now := time.Now()
	telegramWorkDeliverInitial(t, f, now)

	observation := WorkObservation{Revision: 2, Phase: "completed", Kind: "result", Text: "Work has finished."}
	if err := f.store.ProjectWork(t.Context(), ref, observation, f.input.Config.Digest(), now); err != nil {
		t.Fatal(err)
	}

	input := telegramOperationActionInput(t, f, "Show the work status.")
	input.Update.ReplyToMessageID = 77777

	if _, err := f.store.ResolveWorkStatusTarget(t.Context(), input, now); !errors.Is(err, workflowstate.ErrConflict) {
		t.Fatal("undelivered work message identified an execution", err)
	}

	delivery := telegramWorkDueNotice(t, f, now, observation.Text)
	if err := f.store.ReserveDelivery(t.Context(), delivery.ID, now); err != nil {
		t.Fatal(err)
	}

	if err := f.store.DeliverySent(t.Context(), delivery.ID, input.Update.ReplyToMessageID, now); err != nil {
		t.Fatal(err)
	}

	return f, input, ref.Request.Target
}
