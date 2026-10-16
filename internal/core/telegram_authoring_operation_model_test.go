package core

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"
)

func telegramAuthoringModelFixture(t *testing.T) (*telegramConversationFixture, TelegramOperationRequest) {
	t.Helper()

	f := newTelegramConversationFixture(t)
	f.apply(t, TelegramConversationStart, func(input *TelegramConversationInput) { input.Project = "alpha" })

	input := f.input
	input.Conversation = f.sessions(t)[0].ID

	input.Update = telegramRetainedTestUpdate(t, 99, "explain retries")
	if err := f.store.RetainBatch(t.Context(), 12345, []TelegramReceivedUpdate{input.Update}, time.Now()); err != nil {
		t.Fatal(err)
	}

	target, err := f.store.ResolveTelegramOperationTarget(t.Context(), input, time.Now())
	if err != nil {
		t.Fatal(err)
	}

	request, err := f.store.ReservePreparedTelegramOperation(t.Context(), TelegramOperationRequest{Source: input.Update, Operator: "maintainer", Target: target, OperationID: "authoring.explain", Arguments: json.RawMessage(`{"payload_hash":"fixture"}`)}, json.RawMessage(`{"question":"frozen"}`), time.Now())
	if err != nil {
		t.Fatal(err)
	}

	if err := f.store.BeginTelegramOperation(t.Context(), request, time.Now()); err != nil {
		t.Fatal(err)
	}

	return f, request
}

// TestTelegramAuthoringModelReceipt retains explanation and interpretation
// independently, permits another DB writer and replays without another call.
func TestTelegramAuthoringModelReceipt(t *testing.T) {
	f, request := telegramAuthoringModelFixture(t)
	env := f.env

	intent, err := f.store.ResolveTelegramInterpretation(t.Context(), env, request.Source, time.Now(), func(context.Context) (json.RawMessage, error) { return json.RawMessage(`{"operation":"explain"}`), nil })
	if err != nil {
		t.Fatal(err)
	}

	other := openTelegramReceiverTestStore(t, env)
	calls := 0
	invoke := func(ctx context.Context) (json.RawMessage, error) {
		calls++

		if err := other.RetainBatch(ctx, 12345, []TelegramReceivedUpdate{telegramRetainedTestUpdate(t, 100, "/status")}, time.Now()); err != nil {
			return nil, err
		}

		return json.RawMessage(`{"text":"The ceiling is retained."}`), nil
	}

	first, err := f.store.ResolveTelegramOperationModel(t.Context(), env, request, time.Now(), invoke)
	if err != nil || first.Attempts != 1 || len(first.Response) == 0 {
		t.Fatal(first, err)
	}

	telegramAuthoringCheckModelReplay(t, f, request, first, intent, &calls, invoke)
}

func telegramAuthoringCheckModelReplay(t *testing.T, f *telegramConversationFixture, request TelegramOperationRequest, first TelegramOperationModel, intent TelegramInterpretation, calls *int, invoke func(context.Context) (json.RawMessage, error)) {
	t.Helper()

	reopened := openTelegramReceiverTestStore(t, f.env)

	again, err := reopened.ResolveTelegramOperationModel(t.Context(), f.env, request, time.Now(), invoke)
	if err != nil || string(again.Response) != string(first.Response) || *calls != 1 {
		t.Fatal("response was regenerated", again, calls, err)
	}

	retained, err := reopened.ResolveTelegramInterpretation(t.Context(), f.env, request.Source, time.Now(), func(context.Context) (json.RawMessage, error) { t.Error("intent was regenerated"); return nil, nil })
	if err != nil || string(retained.Decision) != string(intent.Decision) {
		t.Fatal("explanation overwrote intent", retained, err)
	}
}

// TestTelegramAuthoringModelBudget preserves exhaustion and backoff across
// reopening. Only sanitized codes, rather than provider errors, are retained.
func TestTelegramAuthoringModelBudget(t *testing.T) {
	f, request := telegramAuthoringModelFixture(t)
	calls := 0
	invoke := func(context.Context) (json.RawMessage, error) {
		calls++

		return nil, errors.New("private provider detail")
	}

	for attempt := 1; attempt <= 5; attempt++ {
		store := openTelegramReceiverTestStore(t, f.env)

		result, err := store.ResolveTelegramOperationModel(t.Context(), f.env, request, time.Now().Add(time.Hour), invoke)
		if err != nil || result.Attempts != attempt || result.ErrorCode != "provider_unavailable" || len(result.Response) != 0 {
			t.Fatal(attempt, result, err)
		}

		if _, err := store.ResolveTelegramOperationModel(t.Context(), f.env, request, time.Now(), invoke); err != nil {
			t.Fatal(err)
		}
	}

	result, err := f.store.ResolveTelegramOperationModel(t.Context(), f.env, request, time.Now().Add(24*time.Hour), invoke)
	if err != nil || result.Attempts != 5 || calls != 5 {
		t.Fatal(result, calls, err)
	}
}

// TestTelegramAuthoringModelBinding rejects another source, changed input and
// forged argument bindings before a provider can run against the retained ID.
func TestTelegramAuthoringModelBinding(t *testing.T) {
	f, request := telegramAuthoringModelFixture(t)

	tests := []struct {
		name   string
		change func(*TelegramOperationRequest)
	}{
		{"source", func(r *TelegramOperationRequest) { r.Source.UpdateID++ }},
		{"peer", func(r *TelegramOperationRequest) { r.Source.Peer.UserID++ }},
		{"input", func(r *TelegramOperationRequest) { r.PreparedInput = json.RawMessage(`{"question":"changed"}`) }},
		{"arguments", func(r *TelegramOperationRequest) { r.ArgumentDigest = "changed" }},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			changed := request
			tt.change(&changed)

			_, err := f.store.ResolveTelegramOperationModel(t.Context(), f.env, changed, time.Now(), func(context.Context) (json.RawMessage, error) {
				t.Error("forged request invoked provider")

				return json.RawMessage(`{}`), nil
			})
			if err == nil {
				t.Fatal("forged request accepted")
			}
		})
	}
}

// TestTelegramAuthoringModelExclusion prevents an independent owner from
// stealing a provider call after a clock jump, without holding a SQLite TX.
func TestTelegramAuthoringModelExclusion(t *testing.T) {
	f, request := telegramAuthoringModelFixture(t)
	entered, release := make(chan struct{}), make(chan struct{})
	done := make(chan error, 1)

	go func() {
		_, err := f.store.ResolveTelegramOperationModel(t.Context(), f.env, request, time.Now(), func(context.Context) (json.RawMessage, error) {
			close(entered)
			<-release

			return json.RawMessage(`{}`), nil
		})
		done <- err
	}()

	<-entered

	other := openTelegramReceiverTestStore(t, f.env)
	_, err := other.ResolveTelegramOperationModel(t.Context(), f.env, request, time.Now().Add(24*time.Hour), func(context.Context) (json.RawMessage, error) { return nil, errors.New("must not run") })

	close(release)

	if err == nil {
		t.Error("provider ownership was stolen")
	}

	if err := <-done; err != nil {
		t.Fatal(err)
	}
}
