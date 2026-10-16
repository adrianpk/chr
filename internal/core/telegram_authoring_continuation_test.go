package core

import (
	"context"
	"errors"
	"testing"
)

// TestTelegramAuthoringRespondRecovery journals a real answer and its requested
// artifact together, then recovers the original receipt without another model call.
func TestTelegramAuthoringRespondRecovery(t *testing.T) {
	root := t.TempDir()
	repo := &fileAuthoringSessionRepository{projectDir: root}
	request := telegramAuthoringOwnerRequest()
	request.Target = AuthoringArtifactKindSpec

	asked, err := repo.ContinueAuthoring(t.Context(), request, func(context.Context, AuthoringSnapshot, AuthoringContinuationRequest) (AuthoringDraft, error) {
		return AuthoringDraft{Question: "Which timeout?"}, nil
	})
	if err != nil {
		t.Fatal(err)
	}

	respond := telegramAuthoringNextOwner(asked, "respond", "respond", AuthoringArtifactKindSpec)
	respond.Question, respond.Instructions = asked.Question, "Sixty seconds."
	providerCalls := 0
	generate := func(context.Context, AuthoringSnapshot, AuthoringContinuationRequest) (AuthoringDraft, error) {
		providerCalls++

		return AuthoringDraft{Body: "# Requirements\n\nSixty seconds.\n"}, nil
	}
	repo.afterMutationReplace = func(index int) error {
		if index == 1 {
			return errors.New("interrupted after transcript replacement")
		}

		return nil
	}

	confirmed, err := repo.ContinueAuthoring(t.Context(), respond, generate)
	if err == nil {
		t.Fatal("journal interruption was not exercised")
	}

	reopened := NewAuthoringSessionRepository(root)

	snapshot, err := reopened.ObserveContinuation(t.Context(), asked.Session.ID)
	if err != nil || snapshot.Question != nil || len(snapshot.Session.Outputs) != 1 {
		t.Fatal("answer/artifact recovery diverged", snapshot, err)
	}

	telegramAuthoringCheckAnswer(t, reopened, asked.Session.ID, respond.Instructions)

	telegramAuthoringCheckRespondReplay(t, reopened, respond, confirmed, generate, &providerCalls)
}

func telegramAuthoringCheckRespondReplay(t *testing.T, reopened AuthoringSessionRepository, respond AuthoringContinuationRequest, confirmed AuthoringContinuationResult, generate AuthoringGenerator, providerCalls *int) {
	t.Helper()

	guardCalls := 0

	replay, err := reopened.ContinueAuthoring(t.Context(), respond, generate, func(context.Context) error {
		guardCalls++

		return ErrWorkflowConflict
	})
	if err != nil || replay.SourceHash != confirmed.SourceHash || *providerCalls != 1 || guardCalls != 0 {
		t.Fatal("receipt reconciliation repeated a new effect", replay, providerCalls, guardCalls, err)
	}
}

// TestTelegramAuthoringCommitGuard checks authority after provider preparation
// and keeps its failed request binding immutable without canonical file effects.
func TestTelegramAuthoringCommitGuard(t *testing.T) {
	repo := NewAuthoringSessionRepository(t.TempDir())
	request := telegramAuthoringOwnerRequest()
	generated := false

	_, err := repo.ContinueAuthoring(t.Context(), request, func(context.Context, AuthoringSnapshot, AuthoringContinuationRequest) (AuthoringDraft, error) {
		generated = true

		return AuthoringDraft{Body: "Unconfirmed preparation"}, nil
	}, func(context.Context) error {
		if !generated {
			t.Error("authority was not checked after generation")
		}

		return ErrWorkflowConflict
	})
	if !errors.Is(err, ErrWorkflowConflict) {
		t.Fatal("revoked authority committed", err)
	}

	sessions, err := repo.ListSessions(t.Context())
	if err != nil || len(sessions) != 0 {
		t.Fatal("failed authority wrote a session", sessions, err)
	}

	request.Capture = "Changed payload"
	if _, err := repo.ContinueAuthoring(t.Context(), request, telegramAuthoringOwnerDraft); !errors.Is(err, ErrWorkflowConflict) {
		t.Fatal("failed request changed arguments", err)
	}
}
