package app

import (
	"context"
	"database/sql"
	"errors"
	"net/http"
	"strings"
	"testing"
	"time"

	"forge.adrianpk.com/chr/internal/core"
)

func reopenTelegramOperationReceiver(t *testing.T, f *telegramConversationAppFixture) {
	t.Helper()

	if err := f.receiver.store.Close(); err != nil {
		t.Fatal(err)
	}

	store, err := core.NewTelegramReceiverStore(t.Context(), f.receiver.env, testAssetsFS())
	if err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() {
		if err := store.Close(); err != nil {
			t.Error(err)
		}
	})

	f.receiver.store = store
}

// TestTelegramOperationActionLostAcknowledgement observes the same action, answer and owner receipt after restart.
func TestTelegramOperationActionLostAcknowledgement(t *testing.T) {
	f := telegramOperationAppFixture(t)
	action, _ := telegramOperationOfferRecord(t, f)
	operations := telegramOperationConcreteService(t, f.receiver)
	lost := false
	f.receiver.operations = telegramLostRecordAck{service: operations, lost: &lost}
	queueTelegramOperationCallback(t, f, action)

	if _, err := f.receiver.step(t.Context()); !errors.Is(err, errTelegramLostRecordAck) {
		t.Fatal("action owner acknowledgement was not interrupted", err)
	}

	reopenTelegramOperationReceiver(t, f)

	f.receiver.operations = operations
	if err := f.receiver.drain(t.Context()); err != nil {
		t.Fatal(err)
	}

	notice := telegramOperationNotice(t, f)
	if !strings.Contains(notice.Text, "Record saved") {
		t.Fatal("action recovery did not confirm receipt", notice)
	}

	checkTelegramOperationOneRecord(t, f)
}

type telegramDropSendAck struct {
	client  core.TelegramHTTPClient
	dropped bool
	skip    int
}

func (c *telegramDropSendAck) Do(request *http.Request) (*http.Response, error) {
	response, err := c.client.Do(request)
	if err != nil {
		return response, err
	}

	if (strings.HasSuffix(request.URL.Path, "/sendMessage") || strings.HasSuffix(request.URL.Path, "/sendDocument")) && !c.dropped {
		if c.skip > 0 {
			c.skip--

			return response, nil
		}

		c.dropped = true

		return nil, errors.Join(errors.New("injected lost send acknowledgement"), response.Body.Close())
	}

	return response, nil
}

type telegramCountRecordApply struct {
	service
	calls *int
}

func (s telegramCountRecordApply) ApplyTelegramRecord(ctx context.Context, target core.TelegramOperationTarget, prepared PreparedRecordMutation) (RecordMutationResult, error) {
	*s.calls++

	return s.service.ApplyTelegramRecord(ctx, target, prepared)
}

// TestTelegramOperationUncertainActionSend repeats only a labelled notice after lost send acknowledgement and database reopen.
func TestTelegramOperationUncertainActionSend(t *testing.T) {
	f := telegramOperationAppFixture(t)
	action, _ := telegramOperationOfferRecord(t, f)
	operations := telegramOperationConcreteService(t, f.receiver)
	calls := 0
	f.receiver.operations = telegramCountRecordApply{service: operations, calls: &calls}
	f.receiver.api.Client = &telegramDropSendAck{client: f.receiver.api.Client}
	queueTelegramOperationCallback(t, f, action)

	if _, err := f.receiver.step(t.Context()); err != nil {
		t.Fatal(err)
	}

	first := telegramOperationNotice(t, f)
	reopenTelegramOperationReceiver(t, f)
	setTelegramOperationRetryAt(t, f.receiver.env, time.Now().Add(-time.Minute))

	if err := f.receiver.drain(t.Context()); err != nil {
		t.Fatal(err)
	}

	retry := telegramOperationNotice(t, f)
	if calls != 1 || !strings.Contains(first.Text, action.RequestID) || !strings.Contains(retry.Text, action.RequestID) || !strings.Contains(retry.Text, "Uncertain delivery retry") {
		t.Fatal("send retry changed application or uncertainty", calls, first, retry)
	}

	f.receiver.operations = operations
	checkTelegramOperationOneRecord(t, f)
}

func setTelegramOperationRetryAt(t *testing.T, env core.UserEnvironment, deadline time.Time) {
	t.Helper()

	db, err := sql.Open("sqlite", core.UserDatabasePath(env))
	if err != nil {
		t.Fatal(err)
	}

	_, changeErr := db.ExecContext(t.Context(), `UPDATE telegram_outbox SET retry_at=? WHERE state='uncertain'`, deadline.UTC().Format(time.RFC3339))
	if err := errors.Join(changeErr, db.Close()); err != nil {
		t.Fatal(err)
	}
}

// TestTelegramOperationUncertainPrompt preserves its token when a callback arrives before the prompt's send acknowledgement.
func TestTelegramOperationUncertainPrompt(t *testing.T) {
	f := telegramOperationAppFixture(t)
	f.receiver.api.Client = &telegramDropSendAck{client: f.receiver.api.Client, skip: 1}
	action, prompt := telegramOperationOfferRecord(t, f)
	setTelegramOperationRetryAt(t, f.receiver.env, time.Now().Add(time.Hour))
	queueTelegramOperationCallback(t, f, action)

	if _, err := f.receiver.step(t.Context()); err != nil {
		t.Fatal(err)
	}

	confirmed := telegramOperationNotice(t, f)
	if !strings.Contains(confirmed.Text, "Record saved") {
		t.Fatal("unacknowledged prompt callback failed", confirmed)
	}

	reopenTelegramOperationReceiver(t, f)
	setTelegramOperationRetryAt(t, f.receiver.env, time.Now().Add(-time.Minute))

	if err := f.receiver.drain(t.Context()); err != nil {
		t.Fatal(err)
	}

	repeat := telegramOperationNotice(t, f)
	if repeat.ID == prompt.ID || !strings.Contains(repeat.Text, action.Token) || !strings.Contains(repeat.Text, "Uncertain delivery retry") {
		t.Fatal("uncertain prompt lost its original binding", repeat)
	}

	if notice := f.send(t, "confirm", repeat.ID); !strings.Contains(notice.Text, "unavailable") {
		t.Fatal("consumed prompt became capture or approval", notice)
	}

	checkTelegramOperationOneRecord(t, f)
}
