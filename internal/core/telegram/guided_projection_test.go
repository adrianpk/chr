package telegram

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	tgmessage "forge.adrianpk.com/chr/internal/core/telegram/message"
)

// TestGuidedUseCaseTelegramProjection prevents background display changes from
// overtaking a pending owner acknowledgement and coalesces unchanged human views.
func TestGuidedUseCaseTelegramProjection(t *testing.T) {
	f, request, progress := guidedTelegramProjectionFixture(t)

	refs, err := f.store.TelegramGuidedReferences(t.Context(), 12345, "", 16)
	if err != nil || len(refs) != 1 {
		t.Fatal("confirmed association missing", err)
	}

	ref := refs[0]

	next := telegramRetainedTestUpdate(t, 2, "next current input")
	if err := f.store.RetainBatch(t.Context(), 12345, []TelegramReceivedUpdate{next}, time.Now()); err != nil {
		t.Fatal(err)
	}

	pending := TelegramOperationRequest{Source: next, Operator: request.Operator, Target: ref.Request.Target, OperationID: "guided.continue", Arguments: json.RawMessage(`{}`)}

	pending, err = f.store.ReservePreparedTelegramOperation(t.Context(), pending, json.RawMessage(`{}`), time.Now())
	if err != nil {
		t.Fatal(err)
	}

	progress.Binding = `{"revision":2,"view_digest":"second view"}`
	if err := f.store.ProjectTelegramGuided(t.Context(), ref, progress, tgmessage.Blocks{InitialText: "Second view"}, nil, f.input.Config.Digest(), time.Now()); err != nil {
		t.Fatal(err)
	}

	guarded, _, err := f.store.TelegramConversationReference(t.Context(), ref.Conversation.Peer, ref.Conversation.ID)
	if err != nil || guarded.Revision != ref.Conversation.Revision {
		t.Fatal("projection overtook an unacknowledged owner", err)
	}

	if err := f.store.FinishTelegramOperation(t.Context(), pending, "blocked", "", json.RawMessage(`{}`), "Pending input rejected.", f.input.Config.Digest(), time.Now()); err != nil {
		t.Fatal(err)
	}

	if err := f.store.ProjectTelegramGuided(t.Context(), ref, progress, tgmessage.Blocks{InitialText: "Second view"}, nil, f.input.Config.Digest(), time.Now()); err != nil {
		t.Fatal(err)
	}

	assertGuidedTelegramCoalesced(t, f, progress)
}

func guidedTelegramProjectionFixture(t *testing.T) (*telegramConversationFixture, TelegramOperationRequest, TelegramGuidedProgress) {
	t.Helper()

	f := newTelegramConversationFixture(t)
	input := f.input
	input.Project, input.Text = "alpha", "Guided capture"

	input.Update = telegramRetainedTestUpdate(t, 1, "Develop a new use case")
	if err := f.store.RetainBatch(t.Context(), 12345, []TelegramReceivedUpdate{input.Update}, time.Now()); err != nil {
		t.Fatal(err)
	}

	c, target, err := f.store.ReserveTelegramGuidedConversation(t.Context(), input, time.Now())
	if err != nil {
		t.Fatal(err)
	}

	request, err := f.store.ReservePreparedTelegramOperation(t.Context(), TelegramOperationRequest{Source: input.Update, Operator: "operator", Target: target, OperationID: "guided.continue", Arguments: json.RawMessage(`{}`)}, json.RawMessage(`{}`), time.Now())
	if err != nil {
		t.Fatal(err)
	}

	if err := f.store.BeginTelegramOperation(t.Context(), request, time.Now()); err != nil {
		t.Fatal(err)
	}

	progress := TelegramGuidedProgress{SessionRef: c.AuthoringSessionRef, Topic: c.Topic, Binding: `{"revision":1,"view_digest":"first view"}`, Phase: "collecting"}
	if err := f.store.FinishTelegramGuided(t.Context(), request, progress, json.RawMessage(`{}`), "First view", input.Config.Digest(), time.Now()); err != nil {
		t.Fatal(err)
	}

	return f, request, progress
}

func assertGuidedTelegramCoalesced(t *testing.T, f *telegramConversationFixture, progress TelegramGuidedProgress) {
	t.Helper()

	refs, err := f.store.TelegramGuidedReferences(t.Context(), 12345, "", 16)
	if err != nil || len(refs) != 1 {
		t.Fatal(err)
	}

	before := refs[0].Conversation

	progress.Binding = `{"revision":2,"native_revision":99,"view_digest":"second view"}`
	if err := f.store.ProjectTelegramGuided(t.Context(), refs[0], progress, tgmessage.Blocks{InitialText: "Second view"}, nil, f.input.Config.Digest(), time.Now()); err != nil {
		t.Fatal(err)
	}

	after, _, err := f.store.TelegramConversationReference(t.Context(), before.Peer, before.ID)
	if err != nil || after.Revision != before.Revision || !strings.Contains(after.QuestionBinding, "second view") {
		t.Fatal("unchanged projection replaced its prompt", err)
	}
}
