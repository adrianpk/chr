package core

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"sync/atomic"
	"testing"
	"time"

	authoringcore "forge.adrianpk.com/chr/internal/core/authoring"
	authoringstate "forge.adrianpk.com/chr/internal/core/authoring/state"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

func telegramAuthoringOwnerRequest() AuthoringContinuationRequest {
	return AuthoringContinuationRequest{Binding: workflowstate.MutationBinding{RequestID: "authoring-fixture", Operator: "maintainer"}, ProjectID: "example-project", Namespace: "default", Action: "create", Target: authoringcore.ArtifactKindDump, Topic: "Timeout behavior", Capture: "Keep bounded retries."}
}

func telegramAuthoringOwnerDraft(context.Context, AuthoringSnapshot, AuthoringContinuationRequest) (AuthoringDraft, error) {
	return AuthoringDraft{Body: "# Curated context\n\nKeep bounded retries.\n"}, nil
}

// TestTelegramAuthoringOwnerReplay recovers each partial journal through ordinary
// reads, preserves original receipts after edits and rejects changed arguments.
func TestTelegramAuthoringOwnerReplay(t *testing.T) {
	for fault := 0; fault < 7; fault++ {
		t.Run(string(rune('a'+fault)), func(t *testing.T) {
			checkTelegramAuthoringOwnerFault(t, fault)
		})
	}
}

// TestTelegramAuthoringOwnerQuestions binds real canonical prompt content and
// multiline answers to a revision, preserving independent managed sessions.
func TestTelegramAuthoringOwnerQuestions(t *testing.T) {
	repo := NewAuthoringSessionRepository(t.TempDir())
	first := telegramAuthoringCreateOwner(t, repo, "first")
	second := telegramAuthoringCreateOwner(t, repo, "second")

	current, err := repo.ObserveContinuation(t.Context(), first.Session.ID)
	if err != nil {
		t.Fatal(err)
	}

	first.SourceHash = current.SourceHash
	question := telegramAuthoringQuestionOwner(t, repo, first)

	snapshot, err := repo.ObserveContinuation(t.Context(), first.Session.ID)
	if err != nil {
		t.Fatal(err)
	}

	if len(snapshot.Events) != 1 || snapshot.Events[0].Question.Prompt != "Which policy?\nKeep the old default?" {
		t.Fatalf("prompt lost: %+v", snapshot.Events)
	}

	answer := telegramAuthoringNextOwner(question, "answer", "answer", authoringcore.ArtifactKindQuiz)
	answer.Question = question.Question
	answer.Instructions = "Keep the old default.\n\n---\nid: not-an-event\n---\nAnd retain limits."

	answered, err := repo.ContinueAuthoring(t.Context(), answer, nil)
	if err != nil || answered.Question != nil {
		t.Fatalf("answered=%+v err=%v", answered, err)
	}

	telegramAuthoringCheckAnswer(t, repo, first.Session.ID, answer.Instructions)

	untouched, err := repo.ObserveContinuation(t.Context(), second.Session.ID)
	if err != nil || untouched.Session.Revision != second.Session.Revision {
		t.Fatalf("other session changed: %+v err=%v", untouched, err)
	}

	answer.Binding.RequestID = "stale-answer"
	if _, err := repo.ContinueAuthoring(t.Context(), answer, nil); !errors.Is(err, workflowstate.ErrConflict) {
		t.Fatalf("stale answer err=%v", err)
	}
}

func telegramAuthoringCreateOwner(t *testing.T, repo AuthoringSessionRepository, id string) AuthoringContinuationResult {
	t.Helper()

	request := telegramAuthoringOwnerRequest()
	request.Binding.RequestID = id

	result, err := repo.ContinueAuthoring(t.Context(), request, telegramAuthoringOwnerDraft)
	if err != nil {
		t.Fatal(err)
	}

	return result
}

func telegramAuthoringNextOwner(result AuthoringContinuationResult, id, action string, target authoringcore.ArtifactKind) AuthoringContinuationRequest {
	return AuthoringContinuationRequest{Binding: workflowstate.MutationBinding{RequestID: id, Operator: "maintainer"}, ProjectID: result.Session.ProjectID, Namespace: result.Session.Namespace, SessionID: result.Session.ID, Action: AuthoringContinuationAction(action), Target: target, ExpectedRevision: result.Session.Revision, ExpectedSourceHash: result.SourceHash}
}

func telegramAuthoringQuestionOwner(t *testing.T, repo AuthoringSessionRepository, previous AuthoringContinuationResult) AuthoringContinuationResult {
	t.Helper()

	request := telegramAuthoringNextOwner(previous, "question-"+previous.Session.ID, "ask", authoringcore.ArtifactKindQuiz)

	result, err := repo.ContinueAuthoring(t.Context(), request, func(context.Context, AuthoringSnapshot, AuthoringContinuationRequest) (AuthoringDraft, error) {
		return AuthoringDraft{Question: "Which policy?\nKeep the old default?", Rationale: "Avoid scope expansion."}, nil
	})
	if err != nil || result.Question == nil {
		t.Fatalf("result=%+v err=%v", result, err)
	}

	return result
}

// TestTelegramAuthoringOwnerConcurrent verifies one provider call and one receipt
// across independent owners applying the same request at the same time.
func TestTelegramAuthoringOwnerConcurrent(t *testing.T) {
	root := t.TempDir()
	entered, release := make(chan struct{}), make(chan struct{})

	var calls atomic.Int64

	generate := func(ctx context.Context, snapshot AuthoringSnapshot, request AuthoringContinuationRequest) (AuthoringDraft, error) {
		if calls.Add(1) == 1 {
			close(entered)
		}

		<-release

		return telegramAuthoringOwnerDraft(ctx, snapshot, request)
	}
	done := make(chan error, 2)

	for range 2 {
		go func() {
			_, err := NewAuthoringSessionRepository(root).ContinueAuthoring(t.Context(), telegramAuthoringOwnerRequest(), generate)
			done <- err
		}()
	}

	<-entered
	close(release)

	for range 2 {
		if err := <-done; err != nil {
			t.Error(err)
		}
	}

	if calls.Load() != 1 {
		t.Fatalf("provider calls=%d", calls.Load())
	}
}

// TestTelegramAuthoringOwnerDivergence preserves pending journals and foreign
// file images; neither recovery nor a new ordinary mutation overwrites them.
func TestTelegramAuthoringOwnerDivergence(t *testing.T) {
	root := t.TempDir()
	repo := &fileAuthoringSessionRepository{projectDir: root, afterMutationReplace: func(int) error { return errors.New("lost acknowledgement") }}

	result, err := repo.ContinueAuthoring(t.Context(), telegramAuthoringOwnerRequest(), telegramAuthoringOwnerDraft)
	if err == nil {
		t.Fatal("fault was not exercised")
	}

	path := filepath.Join(root, result.SessionRef)
	if err := os.WriteFile(path, []byte("independent edit"), 0o600); err != nil {
		t.Fatal(err)
	}

	reopened := NewAuthoringSessionRepository(root)
	if _, err := reopened.GetSession(t.Context(), result.Session.ID); !errors.Is(err, workflowstate.ErrConflict) {
		t.Fatalf("divergent recovery err=%v", err)
	}

	if _, err := os.Stat(filepath.Join(root, authoringRuntimeRef("pending.json"))); err != nil {
		t.Fatal("pending journal was discarded", err)
	}

	content, err := os.ReadFile(path)
	if err != nil || string(content) != "independent edit" {
		t.Fatal("divergent image overwritten", err)
	}
}

func checkTelegramAuthoringOwnerFault(t *testing.T, fault int) {
	t.Helper()

	root := t.TempDir()
	repo := &fileAuthoringSessionRepository{projectDir: root, afterMutationReplace: func(index int) error {
		if index == fault {
			return errors.New("lost acknowledgement")
		}

		return nil
	}}
	request := telegramAuthoringOwnerRequest()
	calls := 0
	generate := func(ctx context.Context, snapshot AuthoringSnapshot, request AuthoringContinuationRequest) (AuthoringDraft, error) {
		calls++

		return telegramAuthoringOwnerDraft(ctx, snapshot, request)
	}

	first, err := repo.ContinueAuthoring(t.Context(), request, generate)
	if err == nil {
		t.Fatal("fault was not exercised")
	}

	reopened := NewAuthoringSessionRepository(root)

	session, err := reopened.GetSession(t.Context(), first.Session.ID)
	if err != nil {
		t.Fatal(err)
	}

	session.Description = "Independently edited display"
	if _, err := reopened.UpdateSession(t.Context(), session); err != nil {
		t.Fatal(err)
	}

	replay, err := reopened.ContinueAuthoring(t.Context(), request, generate)
	if err != nil || replay.SourceHash != first.SourceHash || calls != 1 {
		t.Fatalf("replay=%+v calls=%d err=%v", replay, calls, err)
	}

	request.Capture = "Another payload"
	if _, err := reopened.ContinueAuthoring(t.Context(), request, generate); !errors.Is(err, workflowstate.ErrConflict) {
		t.Fatalf("changed request err=%v", err)
	}
}

func telegramAuthoringCheckAnswer(t *testing.T, repo AuthoringSessionRepository, id, expected string) {
	t.Helper()

	after, err := repo.ObserveContinuation(t.Context(), id)
	if err != nil || len(after.Events) != 2 || after.Events[1].Answer.Text != expected {
		t.Fatalf("answer lost: %+v err=%v", after.Events, err)
	}
}

// TestTelegramAuthoringOwnerLifetime prevents an inherited context from bypassing
// exclusion after its original mutation has returned and released the OS lock.
func TestTelegramAuthoringOwnerLifetime(t *testing.T) {
	root := t.TempDir()

	var retained context.Context

	if err := WithAuthoringMutation(t.Context(), root, func(ctx context.Context) error { retained = ctx; return nil }); err != nil {
		t.Fatal(err)
	}

	entered, release := make(chan struct{}), make(chan struct{})

	done := make(chan error, 1)
	go func() {
		done <- WithAuthoringMutation(t.Context(), root, func(context.Context) error { close(entered); <-release; return nil })
	}()

	<-entered

	ctx, cancel := context.WithTimeout(retained, 20*time.Millisecond)
	defer cancel()

	err := WithAuthoringMutation(ctx, root, func(context.Context) error { t.Error("expired owner context bypassed exclusion"); return nil })

	close(release)

	if !errors.Is(err, context.DeadlineExceeded) {
		t.Errorf("lock wait err=%v", err)
	}

	if err := <-done; err != nil {
		t.Fatal(err)
	}
}

// TestTelegramAuthoringOwnerBudget retains provider attempts and failed argument
// identity, while retries preserve all prior canonical state and emit no artifact.
func TestTelegramAuthoringOwnerBudget(t *testing.T) {
	root := t.TempDir()
	repo := &fileAuthoringSessionRepository{projectDir: root}
	request := telegramAuthoringOwnerRequest()
	calls := 0

	generate := func(context.Context, AuthoringSnapshot, AuthoringContinuationRequest) (AuthoringDraft, error) {
		calls++

		return AuthoringDraft{}, errors.New("provider unavailable")
	}
	for attempt := 1; attempt <= 5; attempt++ {
		if _, err := repo.ContinueAuthoring(t.Context(), request, generate); err == nil {
			t.Fatal("provider failure became success")
		}

		binding := telegramAuthoringTestBinding(t, repo, request.Binding.RequestID)
		if binding.Attempts != attempt {
			t.Fatalf("attempts=%d expected=%d", binding.Attempts, attempt)
		}

		if attempt < 5 {
			if _, err := repo.ContinueAuthoring(t.Context(), request, generate); !errors.Is(err, ErrAuthoringProviderWait) {
				t.Fatalf("retry deadline err=%v", err)
			}
		}

		binding.RetryAt = time.Time{}
		if err := repo.saveAuthoringRequest(binding); err != nil {
			t.Fatal(err)
		}
	}

	if _, err := NewAuthoringSessionRepository(root).ContinueAuthoring(t.Context(), request, generate); !errors.Is(err, ErrAuthoringProviderExhausted) {
		t.Fatalf("reopened exhaustion err=%v", err)
	}

	if calls != 5 {
		t.Fatalf("provider calls=%d", calls)
	}

	request.Capture = "changed input"
	if _, err := repo.ContinueAuthoring(t.Context(), request, generate); !errors.Is(err, workflowstate.ErrConflict) {
		t.Fatalf("changed failed request err=%v", err)
	}
}

func telegramAuthoringTestBinding(t *testing.T, repo *fileAuthoringSessionRepository, id string) authoringRequestBinding {
	t.Helper()

	data, err := repo.authoringRead(authoringRequestRef(id))
	if err != nil {
		t.Fatal(err)
	}

	var binding authoringRequestBinding
	if err := authoringstate.Decode(data, &binding); err != nil {
		t.Fatal(err)
	}

	return binding
}

// TestTelegramAuthoringOwnerPhysicalBinding rejects a session file whose embedded
// identity changed; observing its path cannot authorize writing another session.
func TestTelegramAuthoringOwnerPhysicalBinding(t *testing.T) {
	root := t.TempDir()
	repo := NewAuthoringSessionRepository(root)
	result := telegramAuthoringCreateOwner(t, repo, "physical-binding")
	path := filepath.Join(root, result.SessionRef)

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}

	session, err := ParseAuthoringSession(string(data))
	if err != nil {
		t.Fatal(err)
	}

	session.ID = "another-session"

	content, err := RenderAuthoringSession(session, AuthoringSessionLayout(root, result.Session.ID))
	if err != nil {
		t.Fatal(err)
	}

	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}

	if _, err := repo.ObserveContinuation(t.Context(), result.Session.ID); !errors.Is(err, workflowstate.ErrConflict) {
		t.Fatalf("mismatched physical binding err=%v", err)
	}
}
