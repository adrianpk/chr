package core

import (
	"testing"
	"time"
)

// TestTelegramWorkProjection proves revision deduplication, unsent coalescing,
// a thirty-second routine interval and immutable uncertain/terminal notices.
func TestTelegramWorkProjection(t *testing.T) {
	f := newTelegramConversationFixture(t)
	ref := telegramWorkReferenceFixture(t, f, "run-observed")
	now := time.Date(2030, 1, 1, 0, 0, 0, 0, time.UTC)
	telegramWorkDeliverInitial(t, f, now)

	project := func(revision uint64, text string, at time.Time) {
		t.Helper()

		observation := TelegramWorkObservation{Revision: revision, Phase: "executing", Kind: "progress", Text: text}
		if err := f.store.ProjectTelegramWork(t.Context(), ref, observation, f.input.Config.Digest(), at); err != nil {
			t.Fatal(err)
		}
	}
	project(1, "original observation", now)
	telegramWorkDueNotice(t, f, now, "original observation")

	project(2, "first", now)

	old := telegramWorkDueNotice(t, f, now, "first")
	project(3, "latest", now)

	if reserved, err := f.store.TryReserveDelivery(t.Context(), old.ID, now); err != nil || reserved {
		t.Fatal("coalesced stale read reserved", reserved, err)
	}

	latest := telegramWorkDueNotice(t, f, now, "latest")
	if err := f.store.ReserveDelivery(t.Context(), latest.ID, now); err != nil {
		t.Fatal(err)
	}

	if err := f.store.DeliverySent(t.Context(), latest.ID, 20001, now); err != nil {
		t.Fatal(err)
	}

	project(4, "intermediate", now.Add(time.Second))
	project(5, "next interval", now.Add(2*time.Second))

	if due, err := f.store.DueDeliveries(t.Context(), 12345, now.Add(29*time.Second)); err != nil || len(due) != 0 {
		t.Fatal("routine interval exceeded", due, err)
	}

	next := telegramWorkDueNotice(t, f, now.Add(30*time.Second), "next interval")
	telegramWorkReserveAtInterval(t, f, next, now)
	project(6, "unobserved", now.Add(31*time.Second))
	checkTelegramWorkTerminalProjection(t, f, ref, next.ID, now.Add(32*time.Second))
}

func telegramWorkReserveAtInterval(t *testing.T, f *telegramConversationFixture, next TelegramReceiverDelivery, now time.Time) {
	t.Helper()

	if reserved, err := f.store.TryReserveDelivery(t.Context(), next.ID, now.Add(29*time.Second)); err != nil || reserved {
		t.Fatal("reservation bypassed routine interval", reserved, err)
	}

	if err := f.store.ReserveDelivery(t.Context(), next.ID, now.Add(30*time.Second)); err != nil {
		t.Fatal(err)
	}

	if err := f.store.DeliveryFailed(t.Context(), next, &TelegramAPIError{Code: "network", Uncertain: true}, now.Add(30*time.Second)); err != nil {
		t.Fatal(err)
	}
}

func telegramWorkDeliverInitial(t *testing.T, f *telegramConversationFixture, now time.Time) {
	t.Helper()

	due, err := f.store.DueDeliveries(t.Context(), 12345, now)
	if err != nil {
		t.Fatal(err)
	}

	for index, delivery := range due {
		if err := f.store.ReserveDelivery(t.Context(), delivery.ID, now); err != nil {
			t.Fatal(err)
		}

		if err := f.store.DeliverySent(t.Context(), delivery.ID, int64(10000+index), now); err != nil {
			t.Fatal(err)
		}
	}
}

func telegramWorkDueNotice(t *testing.T, f *telegramConversationFixture, now time.Time, text string) TelegramReceiverDelivery {
	t.Helper()

	due, err := f.store.DueDeliveries(t.Context(), 12345, now)
	if err != nil || len(due) != 1 || due[0].Text != text {
		t.Fatal("wrong retained observation", due, err)
	}

	return due[0]
}

func checkTelegramWorkTerminalProjection(t *testing.T, f *telegramConversationFixture, ref TelegramWorkReference, sendingID string, now time.Time) {
	t.Helper()

	observation := TelegramWorkObservation{Revision: 7, Phase: "completed", Kind: "result", Text: "native receipt confirmed"}
	for range 2 {
		if err := f.store.ProjectTelegramWork(t.Context(), ref, observation, f.input.Config.Digest(), now); err != nil {
			t.Fatal(err)
		}
	}

	sending := telegramWorkNoticeCount(t, f, `SELECT COUNT(*) FROM telegram_outbox WHERE id=? AND state='uncertain' AND attempts=1`, sendingID)
	terminal := telegramWorkNoticeCount(t, f, `SELECT COUNT(*) FROM telegram_outbox WHERE kind='result' AND text='native receipt confirmed'`)
	pendingProgress := telegramWorkNoticeCount(t, f, `SELECT COUNT(*) FROM telegram_outbox WHERE kind='progress' AND attempts=0`)

	if sending != 1 || terminal != 1 || pendingProgress != 0 {
		t.Fatal("projection replaced attempted output or duplicated terminal result", sending, terminal, pendingProgress)
	}

	if refs, err := f.store.TelegramWorkReferences(t.Context(), 12345, "", 16); err != nil || len(refs) != 0 {
		t.Fatal("completed conversation still activates", refs, err)
	}
}

func telegramWorkNoticeCount(t *testing.T, f *telegramConversationFixture, query string, args ...any) int {
	t.Helper()

	var count int
	if err := f.store.db.QueryRowContext(t.Context(), query, args...).Scan(&count); err != nil {
		t.Fatal(err)
	}

	return count
}

// A bookkeeping revision cannot resend unchanged progress, including after
// reopening SQLite. The observed revision still advances and real changes notify.
func TestTelegramWorkProjectionUnchanged(t *testing.T) {
	f := newTelegramConversationFixture(t)
	ref := telegramWorkReferenceFixture(t, f, "run-unchanged")
	now := time.Date(2030, 1, 1, 0, 0, 0, 0, time.UTC)
	telegramWorkDeliverInitial(t, f, now)

	message := TelegramMessageBlocks{Family: "33", InitialText: "Waiting for review"}

	observation := TelegramWorkObservation{Revision: 1, Phase: "waiting", Kind: "progress", Text: RenderTelegramMessage(message).PlainFallback, Message: &message}
	if err := f.store.ProjectTelegramWork(t.Context(), ref, observation, f.input.Config.Digest(), now); err != nil {
		t.Fatal(err)
	}

	first := telegramWorkDueNotice(t, f, now, observation.Text)
	telegramWorkDeliverNotice(t, f, first.ID, now)

	if err := f.store.Close(); err != nil {
		t.Fatal(err)
	}

	f.store = openTelegramReceiverTestStore(t, f.env)

	observation.Revision = 2
	if err := f.store.ProjectTelegramWork(t.Context(), ref, observation, f.input.Config.Digest(), now.Add(time.Minute)); err != nil {
		t.Fatal(err)
	}

	if count := telegramWorkNoticeCount(t, f, `SELECT COUNT(*) FROM telegram_outbox WHERE conversation_id=? AND kind='progress'`, ref.Request.Target.ConversationID); count != 1 {
		t.Fatal("unchanged progress was repeated", count)
	}

	refs, err := f.store.TelegramWorkReferences(t.Context(), 12345, "", 16)
	if err != nil || len(refs) != 1 || refs[0].Invocation.Revision != 2 {
		t.Fatal("suppression lost the latest native revision", refs, err)
	}

	message.InitialText = "Integration confirmed"

	observation.Revision, observation.Text = 3, RenderTelegramMessage(message).PlainFallback
	if err := f.store.ProjectTelegramWork(t.Context(), ref, observation, f.input.Config.Digest(), now.Add(2*time.Minute)); err != nil {
		t.Fatal(err)
	}

	telegramWorkDueNotice(t, f, now.Add(2*time.Minute), observation.Text)
}

func telegramWorkDeliverNotice(t *testing.T, f *telegramConversationFixture, id string, now time.Time) {
	t.Helper()

	if err := f.store.ReserveDelivery(t.Context(), id, now); err != nil {
		t.Fatal(err)
	}

	if err := f.store.DeliverySent(t.Context(), id, 21001, now); err != nil {
		t.Fatal(err)
	}
}

// TestTelegramWorkProjectionFull proves terminal projection rolls back when its
// outbox reservation cannot fit, so reopen can observe and reserve it again.
func TestTelegramWorkProjectionFull(t *testing.T) {
	f := newTelegramConversationFixture(t)
	ref := telegramWorkReferenceFixture(t, f, "run-outbox-full")
	now := time.Date(2030, 1, 1, 0, 0, 0, 0, time.UTC)
	telegramWorkDeliverInitial(t, f, now)

	_, err := f.store.db.ExecContext(t.Context(), `WITH RECURSIVE n(i) AS (VALUES(1) UNION ALL SELECT i+1 FROM n WHERE i<1024) INSERT INTO telegram_outbox(id,bot_id,chat_id,topic_id,user_id,event_key,config_digest,kind,text,state,created_at,updated_at) SELECT 'fill-'||i,12345,100001,0,100001,'fill-'||i,?,'result','retained','pending','now','now' FROM n`, f.input.Config.Digest())
	if err != nil {
		t.Fatal(err)
	}

	observation := TelegramWorkObservation{Revision: 2, Phase: "completed", Kind: "result", Text: "receipt confirmed"}
	if err := f.store.ProjectTelegramWork(t.Context(), ref, observation, f.input.Config.Digest(), now); err == nil {
		t.Fatal("terminal projection ignored exhausted outbox")
	}

	refs, err := f.store.TelegramWorkReferences(t.Context(), 12345, "", 16)
	if err != nil || len(refs) != 1 || refs[0].Invocation.Revision != 1 {
		t.Fatal("failed reservation lost recoverable native observation", refs, err)
	}
}
