package telegram

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
	if err := f.store.FinishAuthoring(t.Context(), request, progress, "receipt", facts, "A canonical question", f.input.Config.Digest(), time.Now()); err == nil {
		t.Fatal("a ready request was confirmed before beginning application")
	}

	if current := f.sessions(t)[0]; current != c {
		t.Fatal("failed confirmation did not roll back the conversation", current)
	}

	if err := f.store.BeginOperation(t.Context(), request, time.Now()); err != nil {
		t.Fatal(err)
	}

	if err := f.store.FinishAuthoring(t.Context(), request, progress, "receipt", facts, "A canonical question", f.input.Config.Digest(), time.Now()); err != nil {
		t.Fatal(err)
	}

	telegramAuthoringCheckFinish(t, f, c, progress)
}

func telegramAuthoringFinishFixture(t *testing.T) (*telegramConversationFixture, Conversation, OperationRequest, AuthoringProgress) {
	t.Helper()

	f := newTelegramConversationFixture(t)
	f.apply(t, ConversationStart, func(input *ConversationInput) { input.Project = "alpha" })

	c := f.sessions(t)[0]
	input := f.input
	input.Project, input.Conversation = "alpha", c.ID

	input.Update = telegramRetainedTestUpdate(t, 99, "/spec "+c.ID)
	if err := f.store.RetainBatch(t.Context(), 12345, []ReceivedUpdate{input.Update}, time.Now()); err != nil {
		t.Fatal(err)
	}

	_, target, err := f.store.ResolveAuthoringTarget(t.Context(), input)
	if err != nil {
		t.Fatal(err)
	}

	request, err := f.store.ReservePreparedOperation(t.Context(), OperationRequest{Source: input.Update, Operator: "maintainer", Target: target, OperationID: "authoring.continue", Arguments: json.RawMessage(`{"payload_hash":"fixture"}`)}, json.RawMessage(`{"owner":"frozen"}`), time.Now())
	if err != nil || request.State != "ready" {
		t.Fatal("prepared request was not retained", request, err)
	}

	progress := AuthoringProgress{SessionRef: ".chr/authoring/sessions/example/session.md", Question: &ConversationQuestion{SessionRef: ".chr/authoring/sessions/example/session.md", QuestionID: "question-1", TurnID: "turn-1", Revision: 1, ContentDigest: contenthash.Sum([]byte("real prompt"))}}

	return f, c, request, progress
}

func telegramAuthoringCheckFinish(t *testing.T, f *telegramConversationFixture, c Conversation, progress AuthoringProgress) {
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
