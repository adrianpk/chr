package telegram

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"errors"
	"strconv"
	"strings"
	"testing"
	"time"

	"forge.adrianpk.com/chr/internal/core/contenthash"
	"forge.adrianpk.com/chr/internal/core/database"
	tgmessage "forge.adrianpk.com/chr/internal/core/telegram/message"
)

func seedTelegramDocumentQuota(t *testing.T, store *ReceiverStore, count int) {
	t.Helper()

	zero := make([]byte, tgmessage.DocumentLimit)

	_, err := store.db.ExecContext(t.Context(), `WITH RECURSIVE n(x) AS (SELECT 1 UNION ALL SELECT x+1 FROM n WHERE x<?)
 INSERT INTO telegram_document_snapshots(id,owner_kind,owner_id,bot_id,chat_id,topic_id,user_id,config_digest,project_id,namespace,project_binding_digest,source_ref,source_hash,content_hash,filename,media_type,byte_count,payload,created_at)
 SELECT 'quota-'||x,'receiver','quota-owner-'||x,12345,100001,0,100001,?,'example-project','default',?,'exports/quota.md',?,?,'quota.md','text/markdown',2097152,zeroblob(2097152),? FROM n`, count, telegramReceiverTestConfig().Digest(), strings.Repeat("a", 64), contenthash.Sum(zero), contenthash.Sum(zero), database.FormatTime(time.Now()))
	if err != nil {
		t.Fatal(err)
	}
}

// Independent SQLite connections race for the final 2 MiB. Only one admission
// can commit; the loser keeps its pending input and leaves no partial snapshot.
func TestTelegramPresentationDocumentQuota(t *testing.T) {
	env := telegramReceiverTestEnv(t)
	first := openTelegramReceiverTestStore(t, env)
	second := openTelegramReceiverTestStore(t, env)
	seedTelegramDocumentQuota(t, first, 31)

	updates := []ReceivedUpdate{telegramRetainedTestUpdate(t, 1, "first"), telegramRetainedTestUpdate(t, 2, "second")}
	if err := first.RetainBatch(t.Context(), 12345, updates, time.Now()); err != nil {
		t.Fatal(err)
	}

	response := telegramTestDocumentResponse()
	response.Document.Payload = bytes.Repeat([]byte("q"), tgmessage.DocumentLimit)
	response.Document.SourceHash = contenthash.Sum(response.Document.Payload)
	response.Message.Content = tgmessage.DocumentExcerpt(response.Document.Payload)
	ready, results := make(chan struct{}), make(chan error, 2)

	for index, store := range []*ReceiverStore{first, second} {
		go func(index int, store *ReceiverStore) {
			<-ready

			results <- store.QueueDocumentResponse(t.Context(), updates[index], response, telegramReceiverTestConfig().Digest(), time.Now())
		}(index, store)
	}

	close(ready)

	successful := 0

	for range 2 {
		err := <-results
		if err == nil {
			successful++
		} else if !errors.Is(err, tgmessage.ErrDocumentUnavailable) {
			t.Fatal(err)
		}
	}

	if successful != 1 {
		t.Fatal("quota admitted the wrong number of images", successful)
	}

	checkTelegramDocumentQuota(t, first)
	checkTelegramDocumentReceipt(t, first)
}

func checkTelegramDocumentQuota(t *testing.T, store *ReceiverStore) {
	t.Helper()

	var usage int64
	if err := store.db.QueryRowContext(t.Context(), "SELECT SUM(length(payload)) FROM telegram_document_snapshots").Scan(&usage); err != nil || usage != tgmessage.DocumentQuota {
		t.Fatal("retained byte quota changed", usage, err)
	}

	pending, err := store.PendingUpdates(t.Context(), 12345)
	if err != nil || len(pending) != 1 {
		t.Fatal("quota failure acknowledged pending input", pending, err)
	}
}

func checkTelegramDocumentReceipt(t *testing.T, store *ReceiverStore) {
	t.Helper()

	request := applyingTelegramDocumentTestRequest(t, store)
	response := telegramTestDocumentResponse()

	request.Delivery = &response
	for range 2 {
		if err := store.FinishOperation(t.Context(), request, OperationCompletion{State: "confirmed", ReceiptRef: "receipts/original.json", Facts: json.RawMessage(`{"saved":true}`), Text: "Result saved.", ConfigDigest: telegramReceiverTestConfig().Digest(), At: time.Now()}); err != nil {
			t.Fatal(err)
		}
	}

	operationLookup, err := store.Operation(t.Context(), request.Source, 0)

	retained, found := operationLookup.Request, operationLookup.Found
	if err != nil || !found || retained.State != "confirmed" || retained.ReceiptRef != "receipts/original.json" {
		t.Fatal("document failure lost the original effect receipt", retained, err)
	}

	var count int
	if err := store.db.QueryRowContext(t.Context(), "SELECT COUNT(*) FROM telegram_outbox WHERE request_id=? AND document_id IS NULL AND text LIKE '%document storage is full%'", request.ID).Scan(&count); err != nil || count != 1 {
		t.Fatal("presentation failure duplicated the result or falsely offered a document", count, err)
	}
}

// A lost acknowledgement survives reopen with the same source, file bytes and
// attempt allowance. Only the definite original upload acknowledgement completes it.
func TestTelegramPresentationDeliveryRecovery(t *testing.T) {
	env := telegramReceiverTestEnv(t)
	store := openTelegramReceiverTestStore(t, env)
	now := time.Now()

	update := telegramRetainedTestUpdate(t, 1, "document")
	if err := store.RetainBatch(t.Context(), 12345, []ReceivedUpdate{update}, now); err != nil {
		t.Fatal(err)
	}

	response := telegramTestDocumentResponse()
	if err := store.QueueDocumentResponse(t.Context(), update, response, telegramReceiverTestConfig().Digest(), now); err != nil {
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

	if err := store.DeliveryFailed(t.Context(), delivery, &APIError{Code: "transport_failure", Uncertain: true}, now); err != nil {
		t.Fatal(err)
	}

	if err := store.Close(); err != nil {
		t.Fatal(err)
	}

	store = openTelegramReceiverTestStore(t, env)
	checkTelegramDocumentRecovery(t, store, delivery.ID, response.Document.Payload, now.Add(time.Minute))
}

func checkTelegramDocumentRecovery(t *testing.T, store *ReceiverStore, id string, payload []byte, now time.Time) {
	t.Helper()

	due, err := store.DueDeliveries(t.Context(), 12345, now)
	if err != nil || len(due) != 1 || due[0].ID != id || due[0].Attempts != 1 || due[0].State != "uncertain" {
		t.Fatal("reopen reset uncertainty or allowance", due, err)
	}

	if err := store.ReserveDelivery(t.Context(), id, now); err != nil {
		t.Fatal(err)
	}

	checkTelegramDocumentBytes(t, store, due[0], payload)

	acknowledgeTelegramDocumentTest(t, store, id, now)
}

// Cleanup purges only aged acknowledged bytes. Uncertain deliveries and a live
// confirmation retain their images, while content hashes remain after byte purge.
func TestTelegramPresentationDocumentRetention(t *testing.T) {
	f, _, action := telegramOperationActionFixture(t)
	now := time.Now()

	var ids []string

	for index := range 3 {
		ids = append(ids, queueRetentionTestDocument(t, f, index, now))
	}

	if _, err := f.store.db.ExecContext(t.Context(), "UPDATE telegram_outbox SET action_token=? WHERE id=?", action.Token, ids[1]); err != nil {
		t.Fatal(err)
	}

	if _, err := f.store.db.ExecContext(t.Context(), "UPDATE telegram_actions SET expires_at=? WHERE token=?", database.FormatTime(now.Add(14*24*time.Hour)), action.Token); err != nil {
		t.Fatal(err)
	}

	if err := f.store.transaction(t.Context(), func(tx *sql.Tx) error { return purgeDocuments(t.Context(), tx, 12345, now.Add(8*24*time.Hour)) }); err != nil {
		t.Fatal(err)
	}

	assertTelegramDocumentRetention(t, f.store, ids)
}

func assertTelegramDocumentRetention(t *testing.T, store *ReceiverStore, ids []string) {
	t.Helper()

	for index, id := range ids {
		var (
			purged bool
			hash   string
		)
		if err := store.db.QueryRowContext(t.Context(), "SELECT payload IS NULL,content_hash FROM telegram_document_snapshots WHERE owner_id=?", id).Scan(&purged, &hash); err != nil {
			t.Fatal(err)
		}

		if purged != (index == 0) || len(hash) != 64 {
			t.Fatal("cleanup removed live bytes or lost content identity", index, purged, hash)
		}
	}
}

// Admission rejects unsafe display names, nonlocal source references and unknown
// encodings before it can acknowledge input or create an attachment.
func TestTelegramPresentationDocumentGuards(t *testing.T) {
	for _, test := range []struct {
		name   string
		change func(*tgmessage.Document)
	}{
		{"path escape", func(d *tgmessage.Document) { d.SourceRef = "../../secret" }},
		{"absolute source", func(d *tgmessage.Document) { d.SourceRef = "/private/config" }},
		{"header injection", func(d *tgmessage.Document) { d.Filename = "file.md\r\nX: value" }},
		{"source digest", func(d *tgmessage.Document) { d.SourceHash = strings.Repeat("z", 64) }},
		{"future type", func(d *tgmessage.Document) { d.MediaType = "application/unknown" }},
	} {
		t.Run(test.name, func(t *testing.T) {
			document := telegramTestDocumentResponse().Document
			test.change(&document)

			if err := tgmessage.ValidateDocument(document); !errors.Is(err, tgmessage.ErrDocumentUnavailable) {
				t.Fatal("invalid image accepted", err)
			}
		})
	}
}

func applyingTelegramDocumentTestRequest(t *testing.T, store *ReceiverStore) OperationRequest {
	t.Helper()

	request := telegramOperationTestRequest(t, store, 3)

	request.PreparedInput = json.RawMessage(`{"original":"owner input"}`)
	if err := store.PrepareOperation(t.Context(), request, request.PreparedInput, time.Now()); err != nil {
		t.Fatal(err)
	}

	if err := store.BeginOperation(t.Context(), request, time.Now()); err != nil {
		t.Fatal(err)
	}

	return request
}

func acknowledgeTelegramDocumentTest(t *testing.T, store *ReceiverStore, id string, now time.Time) {
	t.Helper()

	if err := store.DeliveryDocumentSent(t.Context(), id, tgmessage.Acknowledgement{MessageID: 77, FileID: "original-file"}, now); err != nil {
		t.Fatal(err)
	}

	var (
		attempts int
		fileID   string
	)
	if err := store.db.QueryRowContext(t.Context(), "SELECT attempts,telegram_file_id FROM telegram_outbox WHERE id=?", id).Scan(&attempts, &fileID); err != nil || attempts != 2 || fileID != "original-file" {
		t.Fatal("upload acknowledgement lost original identity", attempts, fileID, err)
	}
}

func queueRetentionTestDocument(t *testing.T, f *telegramConversationFixture, index int, now time.Time) string {
	t.Helper()

	update := telegramRetainedTestUpdate(t, int64(100+index), "document")
	if err := f.store.RetainBatch(t.Context(), 12345, []ReceivedUpdate{update}, now); err != nil {
		t.Fatal(err)
	}

	if err := f.store.QueueDocumentResponse(t.Context(), update, telegramTestDocumentResponse(), f.input.Config.Digest(), now); err != nil {
		t.Fatal(err)
	}

	var id string
	if err := f.store.db.QueryRowContext(t.Context(), "SELECT id FROM telegram_outbox WHERE event_key=?", "update:"+strconv.FormatInt(update.UpdateID, 10)).Scan(&id); err != nil {
		t.Fatal(err)
	}

	if err := f.store.ReserveDelivery(t.Context(), id, now); err != nil {
		t.Fatal(err)
	}

	if index < 2 {
		if err := f.store.DeliveryDocumentSent(t.Context(), id, tgmessage.Acknowledgement{MessageID: int64(200 + index), FileID: "file"}, now); err != nil {
			t.Fatal(err)
		}
	}

	return id
}
