package telegram

import (
	"database/sql"
	"fmt"
	"reflect"
	"slices"
	"strings"
	"testing"
	"time"

	"forge.adrianpk.com/chr/internal/core/contenthash"
	"forge.adrianpk.com/chr/internal/core/database"
)

// Distinct same-type values expose column swaps independently of mapper order.
// The existing transition tests cover authority; these fixtures isolate SQL rows.
func TestTelegramStorageRows(t *testing.T) {
	for _, test := range []struct {
		name  string
		check func(*testing.T, *ReceiverStore, OperationRequest)
	}{
		{"action", checkTelegramActionRow}, {"notice", checkTelegramNoticeRow}, {"response", checkTelegramResponseRow},
	} {
		t.Run(test.name, func(t *testing.T) { store, request := telegramStorageRowFixture(t); test.check(t, store, request) })
	}
}

func telegramStorageRowFixture(t *testing.T) (*ReceiverStore, OperationRequest) {
	t.Helper()

	f := newTelegramConversationFixture(t)
	f.apply(t, ConversationStart, func(input *ConversationInput) { input.Project = "alpha" })

	request := telegramOperationTestRequest(t, f.store, 900)
	request.Source.Peer = PeerIdentity{BotID: 12345, ChatID: -234, TopicID: 345, UserID: 456}
	request.Target.ConversationID = f.sessions(t)[0].ID
	request.Target.ConversationRevision = 17

	return f.store, request
}

type telegramSQLExpectation struct {
	table, column, key string
	fields             map[string]any
}

func checkTelegramStoredRow(t *testing.T, db *sql.DB, want telegramSQLExpectation) {
	t.Helper()

	columns := make([]string, 0, len(want.fields))
	for column := range want.fields {
		columns = append(columns, column)
	}

	slices.Sort(columns)

	values, targets := make([]any, len(columns)), make([]any, len(columns))
	for i := range targets {
		targets[i] = &values[i]
	}

	if err := db.QueryRowContext(t.Context(), "SELECT "+strings.Join(columns, ",")+" FROM "+want.table+" WHERE "+want.column+"=?", want.key).Scan(targets...); err != nil {
		t.Fatal(err)
	}

	for i, column := range columns {
		if !reflect.DeepEqual(values[i], want.fields[column]) {
			t.Fatalf("%s.%s = %#v, want %#v", want.table, column, values[i], want.fields[column])
		}
	}
}

func checkTelegramActionRow(t *testing.T, store *ReceiverStore, request OperationRequest) {
	t.Helper()

	at := time.Date(2026, 10, 27, 10, 11, 12, 0, time.UTC)
	token, choices := strings.Repeat("a", 32), `["confirm","cancel"]`

	row := actionAdmissionRow{token: token, request: request, choices: choices, expiresAt: at.Add(time.Hour), createdAt: at}
	if _, err := store.db.ExecContext(t.Context(), insertActionSQL, row.values()...); err != nil {
		t.Fatal(err)
	}

	checkTelegramStoredRow(t, store.db, telegramSQLExpectation{table: "telegram_actions", column: "token", key: token, fields: map[string]any{
		"token": token, "request_id": request.ID, "bot_id": int64(12345), "chat_id": int64(-234), "topic_id": int64(345), "user_id": int64(456), "conversation_revision": int64(17), "guard_digest": actionGuard(request), "allowed_choices": choices, "state": "pending", "expires_at": database.FormatTime(at.Add(time.Hour)), "created_at": database.FormatTime(at), "updated_at": database.FormatTime(at), "consumed_by_request": nil, "guard_version": int64(0), "presentation_hash": nil,
	}})

	answer := telegramOperationTestRequest(t, store, 901)
	expiry := at.Add(92 * time.Minute)

	guard, manifest := strings.Repeat("b", 64), strings.Repeat("c", 64)
	if _, err := store.db.ExecContext(t.Context(), `UPDATE telegram_actions SET conversation_revision=29,guard_digest=?,allowed_choices='["pause","continue"]',state='consumed',expires_at=?,consumed_by_request=?,guard_version=1,presentation_hash=? WHERE token=?`, guard, database.FormatTime(expiry), answer.ID, manifest, token); err != nil {
		t.Fatal(err)
	}

	got, err := readAction(t.Context(), store.db, token, request.Source.Peer)

	want := Action{Token: token, RequestID: request.ID, Peer: request.Source.Peer, ConversationRevision: 29, GuardDigest: guard, GuardVersion: 1, PresentationHash: manifest, Choices: []string{"pause", "continue"}, State: ActionConsumed, ExpiresAt: expiry, AnswerRequestID: answer.ID}
	if err != nil || !reflect.DeepEqual(got, want) {
		t.Fatal("action scan mixed retained columns", got, want, err)
	}
}

func checkTelegramNoticeRow(t *testing.T, store *ReceiverStore, request OperationRequest) {
	t.Helper()

	at := time.Date(2026, 10, 27, 12, 13, 14, 0, time.UTC)
	digest, key := strings.Repeat("d", 64), "distinct-notice-key"
	row := workNoticeRow{ref: WorkReference{Request: request, ConversationRevision: 31}, kind: "diagnostic", key: key, text: "Distinct notice text", configDigest: digest, part: 3, parts: 7, at: at}

	if err := store.transaction(t.Context(), func(tx *sql.Tx) error { return insertWorkNotice(t.Context(), tx, row) }); err != nil {
		t.Fatal(err)
	}

	id := "tgd-" + contenthash.Sum(fmt.Appendf(nil, "%d:%s:%d", request.Source.Peer.BotID, key, 3))[:32]
	checkTelegramStoredRow(t, store.db, telegramSQLExpectation{table: "telegram_outbox", column: "id", key: id, fields: map[string]any{
		"id": id, "bot_id": int64(12345), "chat_id": int64(-234), "topic_id": int64(345), "user_id": int64(456), "conversation_id": request.Target.ConversationID, "conversation_revision": int64(31), "request_id": request.ID, "event_key": key, "config_digest": digest, "kind": "diagnostic", "text": "Distinct notice text", "part_index": int64(3), "part_count": int64(7), "state": "pending", "created_at": database.FormatTime(at), "updated_at": database.FormatTime(at),
	}})
}

func checkTelegramResponseRow(t *testing.T, store *ReceiverStore, request OperationRequest) {
	t.Helper()

	at := time.Date(2026, 10, 27, 13, 14, 15, 0, time.UTC)
	id, digest := "distinct-response", strings.Repeat("e", 64)

	row := responsePartRow{id: id, peer: request.Source.Peer, key: "distinct-response-key", configDigest: digest, text: "Distinct response text", part: 2, parts: 5, at: at}
	if _, err := store.db.ExecContext(t.Context(), insertResponsePartSQL, row.values()...); err != nil {
		t.Fatal(err)
	}

	checkTelegramStoredRow(t, store.db, telegramSQLExpectation{table: "telegram_outbox", column: "id", key: id, fields: map[string]any{
		"id": id, "bot_id": int64(12345), "chat_id": int64(-234), "topic_id": int64(345), "user_id": int64(456), "event_key": "distinct-response-key", "config_digest": digest, "kind": "result", "text": "Distinct response text", "part_index": int64(2), "part_count": int64(5), "state": "pending", "created_at": database.FormatTime(at), "updated_at": database.FormatTime(at),
	}})
	checkTelegramDeliveryRow(t, store, id, request)
}

func checkTelegramDeliveryRow(t *testing.T, store *ReceiverStore, id string, request OperationRequest) {
	t.Helper()

	peer := request.Source.Peer
	at := time.Date(2026, 10, 27, 14, 15, 16, 0, time.UTC)
	document := telegramTestDocumentResponse().Document
	document.SourceHash = strings.Repeat("2", 64)
	digest, hash, docID := strings.Repeat("f", 64), strings.Repeat("1", 64), "distinct-document"

	if err := store.transaction(t.Context(), func(tx *sql.Tx) error {
		return retainDocument(t.Context(), tx, documentSnapshotRow{id: docID, owner: id, peer: peer, configDigest: digest, document: document, createdAt: at})
	}); err != nil {
		t.Fatal(err)
	}

	token := strings.Repeat("a", 32)

	action := actionAdmissionRow{token: token, request: request, choices: `["pause","continue"]`, expiresAt: at.Add(time.Hour), createdAt: at}
	if _, err := store.db.ExecContext(t.Context(), insertActionSQL, action.values()...); err != nil {
		t.Fatal(err)
	}

	if _, err := store.db.ExecContext(t.Context(), `UPDATE telegram_outbox SET text='Delivery text',state='uncertain',attempts=2,config_digest=?,action_token=?,parse_mode='Markdown',plain_text='Distinct plain text',plain_fallback=1,presentation='{"Version":1}',presentation_hash=?,document_id=? WHERE id=?`, digest, token, hash, docID, id); err != nil {
		t.Fatal(err)
	}

	row := receiverDeliveryRow{delivery: ReceiverDelivery{Peer: PeerIdentity{BotID: peer.BotID}}}
	err := store.db.QueryRowContext(t.Context(), "SELECT "+receiverDeliveryColumns+" FROM telegram_outbox o WHERE o.id=?", id).Scan(row.scanTargets()...)

	want := receiverDeliveryRow{delivery: ReceiverDelivery{ID: id, Peer: peer, Text: "Delivery text", State: DeliveryUncertain, Attempts: 2, ConfigDigest: digest, ActionToken: token, ParseMode: "Markdown", PlainText: "Distinct plain text", PlainFallback: true, DocumentID: docID}, choices: `["pause","continue"]`, presentation: `{"Version":1}`, presentationHash: hash}
	if err != nil || !reflect.DeepEqual(row, want) {
		t.Fatal("delivery scan mixed retained columns", row, want, err)
	}
	// The document reader uses its own source/content hashes, not the presentation hash.
	var image deliveryDocumentRow
	if err := store.db.QueryRowContext(t.Context(), "SELECT "+deliveryDocumentColumns+" FROM telegram_document_snapshots d WHERE d.id=?", docID).Scan(image.scanTargets()...); err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(image.document, document) || image.hash != contenthash.Sum(document.Payload) {
		t.Fatal("document scan mixed frozen fields", image)
	}
}
