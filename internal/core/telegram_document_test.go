package core

import (
	"bytes"
	"strings"
	"testing"
	"time"
)

// Direct body excerpts omit source metadata while the complete bytes stay intact.
func TestTelegramPresentationDocumentExcerpt(t *testing.T) {
	for _, test := range []struct{ name, input, want string }{
		{"front matter", "---\nsource_request_id: private-ref\n---\n# Title\n\nBody", "# Title\nBody"},
		{"CRLF", "---\r\nstatus: draft\r\n---\r\n# Title\r\nBody", "# Title\nBody"},
		{"literal", "# Heading\n```go\nx := `a_b`\n```", "# Heading\n```go\nx := `a_b`\n```"},
		{"unterminated metadata", "---\nsource_request_id: hidden", ""},
	} {
		t.Run(test.name, func(t *testing.T) {
			payload := []byte(test.input)
			if actual := TelegramDocumentExcerpt(payload); actual != test.want || string(payload) != test.input {
				t.Fatal("excerpt changed source or exposed metadata", actual)
			}
		})
	}

	large := []byte(strings.Repeat("😀", 400) + "\nsecond\nthird\nfourth\nfifth\nsixth\nseventh")

	excerpt := TelegramDocumentExcerpt(large)
	if telegramVisibleUnits(excerpt) > 600 || !strings.HasSuffix(excerpt, "[excerpt shortened]") {
		t.Fatal("excerpt exceeds final Unicode bound", excerpt)
	}
}

func telegramTestDocumentResponse() TelegramDocumentResponse {
	payload := []byte("---\nstatus: draft\n---\n# Original\n```literal```\n")

	return TelegramDocumentResponse{Message: TelegramMessageBlocks{InitialText: "Draft ready", ContentKind: "markdown", Content: TelegramDocumentExcerpt(payload), SecondaryText: "Complete document attached."}, Document: TelegramDocument{ProjectID: "example-project", Namespace: "default", ProjectBindingDigest: strings.Repeat("a", 64), SourceRef: "ops/default/spec/example.md", SourceHash: workflowHash(payload), Filename: "example.md", MediaType: "text/markdown", Payload: payload}}
}

// Atomic source acknowledgement freezes exact bytes, destination and caption;
// reopening and replaying a mutable source cannot replace the admitted image.
func TestTelegramPresentationDocumentSnapshot(t *testing.T) {
	env := telegramReceiverTestEnv(t)
	store := openTelegramReceiverTestStore(t, env)
	now := time.Now()

	update := telegramRetainedTestUpdate(t, 1, "show document")
	if err := store.RetainBatch(t.Context(), 12345, []TelegramReceivedUpdate{update}, now); err != nil {
		t.Fatal(err)
	}

	original := telegramTestDocumentResponse()
	if err := store.QueueDocumentResponse(t.Context(), update, original, telegramReceiverTestConfig().Digest(), now); err != nil {
		t.Fatal(err)
	}

	changed := telegramTestDocumentResponse()

	changed.Document.Payload = []byte("changed after snapshot")
	if err := store.QueueDocumentResponse(t.Context(), update, changed, telegramReceiverTestConfig().Digest(), now); err != nil {
		t.Fatal(err)
	}

	due, err := store.DueDeliveries(t.Context(), 12345, now)
	if err != nil || len(due) != 1 {
		t.Fatal(due, err)
	}

	delivery := due[0]
	if err := store.ReserveDelivery(t.Context(), delivery.ID, now); err != nil {
		t.Fatal(err)
	}

	checkTelegramDocumentBytes(t, store, delivery, original.Document.Payload)

	if err := store.Close(); err != nil {
		t.Fatal(err)
	}

	store = openTelegramReceiverTestStore(t, env)
	checkTelegramDocumentBytes(t, store, delivery, original.Document.Payload)

	delivery.Peer.UserID++
	if _, err := store.ReadDeliveryDocument(t.Context(), delivery); err == nil {
		t.Fatal("snapshot crossed destinations")
	}
}

func checkTelegramDocumentBytes(t *testing.T, store *TelegramReceiverStore, delivery TelegramReceiverDelivery, want []byte) {
	t.Helper()

	document, err := store.ReadDeliveryDocument(t.Context(), delivery)
	if err != nil || !bytes.Equal(document.Payload, want) || document.Filename != "example.md" {
		t.Fatal("snapshot bytes changed", document, err)
	}

	if telegramVisibleUnits(delivery.Text) > 900 || strings.Contains(delivery.Text, "source_request_id:") {
		t.Fatal("invalid document caption", delivery.Text)
	}
}

// Invalid complete documents cannot partially acknowledge input or admit bytes.
func TestTelegramPresentationDocumentRollback(t *testing.T) {
	store := openTelegramReceiverTestStore(t, telegramReceiverTestEnv(t))
	update := telegramRetainedTestUpdate(t, 1, "oversized document")

	now := time.Now()
	if err := store.RetainBatch(t.Context(), 12345, []TelegramReceivedUpdate{update}, now); err != nil {
		t.Fatal(err)
	}

	response := telegramTestDocumentResponse()

	response.Document.Payload = bytes.Repeat([]byte("x"), telegramDocumentLimit+1)
	if err := store.QueueDocumentResponse(t.Context(), update, response, telegramReceiverTestConfig().Digest(), now); err == nil {
		t.Fatal("oversized document admitted")
	}

	pending, err := store.PendingUpdates(t.Context(), 12345)
	if err != nil || len(pending) != 1 {
		t.Fatal("source acknowledged without complete document", pending, err)
	}

	var count int
	if err := store.db.QueryRowContext(t.Context(), "SELECT COUNT(*) FROM telegram_document_snapshots").Scan(&count); err != nil || count != 0 {
		t.Fatal("partial snapshot admitted", count, err)
	}
}
