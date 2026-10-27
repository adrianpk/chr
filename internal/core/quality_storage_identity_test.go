package core

import (
	"database/sql"
	"reflect"
	"strings"
	"testing"
)

// TestQualityStorageIdentity uses distinct peer integers and distinct hashes,
// including source and content hashes that represent different frozen inputs.
func TestQualityStorageIdentity(t *testing.T) {
	t.Run("conversation", qualityConversationIdentity)
	t.Run("receiver", qualityReceiverIdentity)
	t.Run("notification", qualityNotificationIdentity)
}

func qualityConversationIdentity(t *testing.T) {
	db := openQualityLegacyDB(t)
	want := qualityLegacyExpected(t).Conversation
	want.Peer.ChatID = -202
	want.Peer.TopicID = 303
	want.Peer.UserID = 404

	qualityStorageExec(t, db, `INSERT INTO telegram_peers(bot_id,chat_id,topic_id,user_id,updated_at) VALUES(12345,-202,303,404,'2026-10-26T11:12:13Z')`)
	qualityStorageExec(t, db, `UPDATE telegram_conversations SET chat_id=-202,topic_id=303,user_id=404`)

	got, err := scanTelegramConversation(db.QueryRowContext(t.Context(), "SELECT "+telegramConversationColumns+" FROM telegram_conversations WHERE id=?", want.ID))
	if err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("conversation peer fields changed: %#v", got)
	}
}

func qualityReceiverIdentity(t *testing.T) {
	db := openQualityLegacyDB(t)
	store := &TelegramReceiverStore{db: db}
	update := telegramRetainedTestUpdate(t, 505, "distinct retained identity")
	update.Peer.ChatID = -202
	update.Peer.TopicID = 303
	update.MessageID = 404

	update.ReplyToMessageID = 606
	if err := store.RetainBatch(t.Context(), 12345, []TelegramReceivedUpdate{update}, qualityStorageNow()); err != nil {
		t.Fatal(err)
	}

	checkQualityInboxIdentity(t, db)

	response := telegramTestDocumentResponse()
	response.Document.SourceHash = strings.Repeat("b", 64)

	digest := strings.Repeat("c", 64)
	if err := store.QueueDocumentResponse(t.Context(), update, response, digest, qualityStorageNow()); err != nil {
		t.Fatal(err)
	}

	deliveries, err := store.DueDeliveries(t.Context(), 12345, qualityStorageNow())
	if err != nil {
		t.Fatal(err)
	}

	for _, delivery := range deliveries {
		if delivery.Peer.ChatID != -202 {
			continue
		}

		if err := store.ReserveDelivery(t.Context(), delivery.ID, qualityStorageNow()); err != nil {
			t.Fatal(err)
		}

		document, err := store.ReadDeliveryDocument(t.Context(), delivery)
		if err != nil {
			t.Fatal(err)
		}

		if !reflect.DeepEqual(document, response.Document) {
			t.Fatalf("document fields changed: %#v", document)
		}

		return
	}

	t.Fatal("distinct receiver delivery missing")
}

func checkQualityInboxIdentity(t *testing.T, db *sql.DB) {
	t.Helper()

	var values [7]int64

	err := db.QueryRowContext(t.Context(), "SELECT bot_id,update_id,chat_id,topic_id,user_id,message_id,reply_to_message_id FROM telegram_updates WHERE update_id=505").Scan(&values[0], &values[1], &values[2], &values[3], &values[4], &values[5], &values[6])
	if err != nil {
		t.Fatal(err)
	}

	if values != [7]int64{12345, 505, -202, 303, 100001, 404, 606} {
		t.Fatalf("inbox identity columns changed: %v", values)
	}
}

func qualityNotificationIdentity(t *testing.T) {
	db := openQualityLegacyDB(t)
	input := qualityNotificationInput(t, db)
	input.Record.Event.ID = "distinct-notification"
	input.Document.SourceHash = strings.Repeat("b", 64)

	repository := qualityNotificationRepository(t, db)
	if err := repository.CreateTelegram(t.Context(), input); err != nil {
		t.Fatal(err)
	}

	lookup, err := repository.ReadTelegram(t.Context(), input.Record.Event.ID, input.Target)
	delivery, frozen := lookup.Delivery, lookup.Frozen

	if err != nil {
		t.Fatal(err)
	}

	if !frozen || !reflect.DeepEqual(delivery.Document, input.Document) {
		t.Fatalf("notification document fields changed: %#v", delivery)
	}
}
