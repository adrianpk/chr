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

	refs, err := f.store.GuidedReferences(t.Context(), 12345, "", 16)
	if err != nil || len(refs) != 1 {
		t.Fatal("confirmed association missing", err)
	}

	ref := refs[0]

	next := telegramRetainedTestUpdate(t, 2, "next current input")
	if err := f.store.RetainBatch(t.Context(), 12345, []ReceivedUpdate{next}, time.Now()); err != nil {
		t.Fatal(err)
	}

	pending := OperationRequest{Source: next, Operator: request.Operator, Target: ref.Request.Target, OperationID: "guided.continue", Arguments: json.RawMessage(`{}`)}

	pending, err = f.store.ReservePreparedOperation(t.Context(), pending, json.RawMessage(`{}`), time.Now())
	if err != nil {
		t.Fatal(err)
	}

	progress.Binding = `{"revision":2,"view_digest":"second view"}`
	if err := f.store.ProjectGuided(t.Context(), GuidedProjectionRequest{Reference: ref, Progress: progress, Message: tgmessage.Blocks{InitialText: "Second view"}, Delivery: nil, ConfigDigest: f.input.Config.Digest(), At: time.Now()}); err != nil {
		t.Fatal(err)
	}

	conversationLookup, err := f.store.ConversationReference(t.Context(), ref.Conversation.Peer, ref.Conversation.ID)

	guarded := conversationLookup.Conversation
	if err != nil || guarded.Revision != ref.Conversation.Revision {
		t.Fatal("projection overtook an unacknowledged owner", err)
	}

	if err := f.store.FinishOperation(t.Context(), pending, OperationCompletion{State: "blocked", ReceiptRef: "", Facts: json.RawMessage(`{}`), Text: "Pending input rejected.", ConfigDigest: f.input.Config.Digest(), At: time.Now()}); err != nil {
		t.Fatal(err)
	}

	if err := f.store.ProjectGuided(t.Context(), GuidedProjectionRequest{Reference: ref, Progress: progress, Message: tgmessage.Blocks{InitialText: "Second view"}, Delivery: nil, ConfigDigest: f.input.Config.Digest(), At: time.Now()}); err != nil {
		t.Fatal(err)
	}

	assertGuidedTelegramCoalesced(t, f, progress)
}

func guidedTelegramProjectionFixture(t *testing.T) (*telegramConversationFixture, OperationRequest, GuidedProgress) {
	t.Helper()

	f := newTelegramConversationFixture(t)
	input := f.input
	input.Project, input.Text = "alpha", "Guided capture"

	input.Update = telegramRetainedTestUpdate(t, 1, "Develop a new use case")
	if err := f.store.RetainBatch(t.Context(), 12345, []ReceivedUpdate{input.Update}, time.Now()); err != nil {
		t.Fatal(err)
	}

	bindingResult, err := f.store.ReserveGuidedConversation(t.Context(), input, time.Now())
	c, target := bindingResult.Conversation, bindingResult.Target

	if err != nil {
		t.Fatal(err)
	}

	request, err := f.store.ReservePreparedOperation(t.Context(), OperationRequest{Source: input.Update, Operator: "operator", Target: target, OperationID: "guided.continue", Arguments: json.RawMessage(`{}`)}, json.RawMessage(`{}`), time.Now())
	if err != nil {
		t.Fatal(err)
	}

	if err := f.store.BeginOperation(t.Context(), request, time.Now()); err != nil {
		t.Fatal(err)
	}

	progress := GuidedProgress{SessionRef: c.AuthoringSessionRef, Topic: c.Topic, Binding: `{"revision":1,"view_digest":"first view"}`, Phase: "collecting"}
	if err := f.store.FinishGuided(t.Context(), request, GuidedCompletion{Progress: progress, Facts: json.RawMessage(`{}`), Text: "First view", ConfigDigest: input.Config.Digest(), At: time.Now()}); err != nil {
		t.Fatal(err)
	}

	return f, request, progress
}

func assertGuidedTelegramCoalesced(t *testing.T, f *telegramConversationFixture, progress GuidedProgress) {
	t.Helper()

	refs, err := f.store.GuidedReferences(t.Context(), 12345, "", 16)
	if err != nil || len(refs) != 1 {
		t.Fatal(err)
	}

	before := refs[0].Conversation

	progress.Binding = `{"revision":2,"native_revision":99,"view_digest":"second view"}`
	if err := f.store.ProjectGuided(t.Context(), GuidedProjectionRequest{Reference: refs[0], Progress: progress, Message: tgmessage.Blocks{InitialText: "Second view"}, Delivery: nil, ConfigDigest: f.input.Config.Digest(), At: time.Now()}); err != nil {
		t.Fatal(err)
	}

	conversationLookup2, err := f.store.ConversationReference(t.Context(), before.Peer, before.ID)

	after := conversationLookup2.Conversation
	if err != nil || after.Revision != before.Revision || !strings.Contains(after.QuestionBinding, "second view") {
		t.Fatal("unchanged projection replaced its prompt", err)
	}
}
