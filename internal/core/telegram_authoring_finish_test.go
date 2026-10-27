package core

import (
	"encoding/json"
	"testing"
	"time"

	"forge.adrianpk.com/chr/internal/core/contenthash"
)

// TestTelegramAuthoringAtomicFinish rolls back conversation/outbox changes when
// request confirmation fails, then binds the real question and source together.
func TestTelegramAuthoringAtomicFinish(t *testing.T) {
	f, c, request, progress := telegramAuthoringFinishFixture(t)

	facts := json.RawMessage(`{"request":"confirmed"}`)
	if err := f.store.FinishTelegramAuthoring(t.Context(), request, progress, "receipt", facts, "A canonical question", f.input.Config.Digest(), time.Now()); err == nil {
		t.Fatal("a ready request was confirmed before beginning application")
	}

	if current := f.sessions(t)[0]; current != c {
		t.Fatal("failed confirmation did not roll back the conversation", current)
	}

	if err := f.store.BeginTelegramOperation(t.Context(), request, time.Now()); err != nil {
		t.Fatal(err)
	}

	if err := f.store.FinishTelegramAuthoring(t.Context(), request, progress, "receipt", facts, "A canonical question", f.input.Config.Digest(), time.Now()); err != nil {
		t.Fatal(err)
	}

	telegramAuthoringCheckFinish(t, f, c, progress)
}

func telegramAuthoringFinishFixture(t *testing.T) (*telegramConversationFixture, TelegramConversation, TelegramOperationRequest, TelegramAuthoringProgress) {
	t.Helper()

	f := newTelegramConversationFixture(t)
	f.apply(t, TelegramConversationStart, func(input *TelegramConversationInput) { input.Project = "alpha" })

	c := f.sessions(t)[0]
	input := f.input
	input.Project, input.Conversation = "alpha", c.ID

	input.Update = telegramRetainedTestUpdate(t, 99, "/spec "+c.ID)
	if err := f.store.RetainBatch(t.Context(), 12345, []TelegramReceivedUpdate{input.Update}, time.Now()); err != nil {
		t.Fatal(err)
	}

	_, target, err := f.store.ResolveTelegramAuthoringTarget(t.Context(), input)
	if err != nil {
		t.Fatal(err)
	}

	request, err := f.store.ReservePreparedTelegramOperation(t.Context(), TelegramOperationRequest{Source: input.Update, Operator: "maintainer", Target: target, OperationID: "authoring.continue", Arguments: json.RawMessage(`{"payload_hash":"fixture"}`)}, json.RawMessage(`{"owner":"frozen"}`), time.Now())
	if err != nil || request.State != "ready" {
		t.Fatal("prepared request was not retained", request, err)
	}

	progress := TelegramAuthoringProgress{SessionRef: ".chr/authoring/sessions/example/session.md", Question: &TelegramConversationQuestion{SessionRef: ".chr/authoring/sessions/example/session.md", QuestionID: "question-1", TurnID: "turn-1", Revision: 1, ContentDigest: contenthash.Sum([]byte("real prompt"))}}

	return f, c, request, progress
}

func telegramAuthoringCheckFinish(t *testing.T, f *telegramConversationFixture, c TelegramConversation, progress TelegramAuthoringProgress) {
	t.Helper()

	current := f.sessions(t)[0]
	if current.Phase != "clarifying" || current.AuthoringSessionRef != progress.SessionRef || current.QuestionBinding == "" || current.Revision != c.Revision+1 {
		t.Fatal("confirmed question was not bound", current)
	}

	pending, err := f.store.PendingUpdates(t.Context(), 12345)
	if err != nil || len(pending) != 0 {
		t.Fatal("confirmation did not acknowledge the source", pending, err)
	}
}
