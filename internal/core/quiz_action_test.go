package core

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"sync"
	"testing"
)

// Retrying after each publication boundary recovers the entire mutation and
// original receipt. A read cannot repair or expose a half-published session.
func TestQuizActionRecovery(t *testing.T) {
	for index := range 6 {
		t.Run(fmt.Sprint(index), func(t *testing.T) {
			checkQuizActionRecovery(t, index)
		})
	}
}

// Two independent writers of the same observed question cannot both advance;
// retries replay exactly once, and an operation ID cannot carry changed input.
func TestQuizActionBindings(t *testing.T) {
	repo, start := newQuizActionFixture(t)
	if _, err := repo.ApplyAction(t.Context(), start, startQuizActionFixture); err != nil {
		t.Fatal(err)
	}

	observed, err := repo.Observe(t.Context(), start.Scope)
	if err != nil {
		t.Fatal(err)
	}

	request := quizAnswerActionFixture(observed)

	var wait sync.WaitGroup

	results := make(chan error, 2)

	for index := range 2 {
		wait.Add(1)
		go func() {
			defer wait.Done()

			candidate := request
			candidate.Binding.OperationID = fmt.Sprintf("answer-%d", index)

			_, err := repo.ApplyAction(t.Context(), candidate, answerQuizActionFixture)
			results <- err
		}()
	}

	wait.Wait()
	close(results)

	accepted, stale := 0, 0

	for err := range results {
		switch {
		case err == nil:
			accepted++
		case errors.Is(err, ErrWorkflowConflict):
			stale++
		default:
			t.Fatal(err)
		}
	}

	if accepted != 1 || stale != 1 {
		t.Fatal("competing question writes were accepted", accepted, stale)
	}

	checkQuizActionReplay(t, repo, request)
}

func checkQuizActionReplay(t *testing.T, repo *fileQuizArtifactRepository, request QuizActionRequest) {
	t.Helper()

	observed, err := repo.Observe(t.Context(), request.Scope)
	if err != nil || len(observed.Entries) != 1 {
		t.Fatal("answer was not recorded once", observed, err)
	}

	for index := range 2 {
		request.Binding.OperationID = fmt.Sprintf("answer-%d", index)

		result, err := repo.ApplyAction(t.Context(), request, answerQuizActionFixture)
		if errors.Is(err, ErrWorkflowConflict) {
			continue
		}

		if err != nil || result.State.AnswersRecorded != 1 {
			t.Fatal("receipt failed to replay", result, err)
		}

		request.Answer = "changed"
		if _, err := repo.ApplyAction(t.Context(), request, answerQuizActionFixture); !errors.Is(err, ErrWorkflowConflict) {
			t.Fatal("operation identity accepted changed arguments", err)
		}

		return
	}

	t.Fatal("no committed receipt")
}

// Raw records preserve entry identity, timestamps and arbitrary paragraph/heading
// content rather than deriving new answers from display text on reload.
func TestQuizEntryExactRoundTrip(t *testing.T) {
	entry := validQuizEntry()
	entry.Answer = "First paragraph.\n\n## Turn\n\nSecond paragraph.\n\nAnswer:\n\nThird."
	entry.ID = "session-1-entry-1"

	entries, err := ParseQuizRawCaptureEntries(RenderQuizRawCapture(validQuizRuntimeState(), entry))
	if err != nil || len(entries) != 1 || !reflect.DeepEqual(entries[0], entry) {
		t.Fatal("entry lost identity or content", entries, err)
	}
}

// Schema upgrades happen only on explicit mutations; unsafe paths, unsupported
// versions and oversized retained files fail before any new publication.
func TestQuizObservationGuards(t *testing.T) {
	for _, change := range []string{"legacy", "version", "oversize", "symlink", "scope"} {
		t.Run(change, func(t *testing.T) {
			repo, request := newQuizActionFixture(t)

			state := validQuizRuntimeState()
			if _, err := repo.SaveRuntimeState(t.Context(), state); err != nil {
				t.Fatal(err)
			}

			paths, before := quizLegacyFixtureImages(t, repo, request.Scope)

			if change == "legacy" {
				checkQuizLegacyRead(t, repo, request.Scope, paths.RuntimePath, before)

				return
			}

			mutateQuizGuardFixture(t, paths, change, &request)

			if _, err := repo.Observe(t.Context(), request.Scope); err == nil {
				t.Fatal("invalid state accepted", change)
			}
		})
	}
}

func mutateQuizGuardFixture(t *testing.T, paths QuizArtifactPaths, change string, request *QuizActionRequest) {
	t.Helper()

	switch change {
	case "version":
		writeTestFile(t, paths.RuntimePath, `{"SchemaVersion":99}`)
	case "oversize":
		writeTestFile(t, paths.RuntimePath, strings.Repeat("x", WorkflowFileLimit+1))
	case "scope":
		request.Scope.ProjectID = "other-project"
	case "symlink":
		outside := filepath.Join(t.TempDir(), "external.json")
		writeTestFile(t, outside, "external")

		if err := os.Remove(paths.RuntimePath); err != nil {
			t.Fatal(err)
		}

		if err := os.Symlink(outside, paths.RuntimePath); err != nil {
			t.Fatal(err)
		}
	}
}

func newQuizActionFixture(t *testing.T) (*fileQuizArtifactRepository, QuizActionRequest) {
	t.Helper()

	repo := &fileQuizArtifactRepository{projectDir: t.TempDir()}
	scope := QuizScope{ProjectID: "project-1"}

	observed, err := repo.Observe(t.Context(), scope)
	if err != nil {
		t.Fatal(err)
	}

	return repo, QuizActionRequest{Scope: scope, Action: "start", Binding: QuizActionBinding{OperationID: "start", ExpectedRevision: observed.Revision}}
}

func startQuizActionFixture(_ context.Context, _ QuizObservation) (QuizTransition, error) {
	state := validQuizRuntimeState()
	state.SessionID = "session-1"

	return QuizTransition{State: state, Message: "Started."}, nil
}

func quizAnswerActionFixture(observed QuizObservation) QuizActionRequest {
	return QuizActionRequest{
		Scope: QuizScope{ProjectID: "project-1"}, Action: "answer", Answer: "First paragraph.\n\nSecond paragraph.",
		Binding: QuizActionBinding{SessionID: observed.State.SessionID, QuestionID: observed.State.ActiveQuestion.ID, ExpectedRevision: observed.Revision, OperationID: "answer"},
	}
}

func answerQuizActionFixture(_ context.Context, observed QuizObservation) (QuizTransition, error) {
	entry := validQuizEntry()
	entry.ID, entry.Answer = "session-1-entry-1", "First paragraph.\n\nSecond paragraph."
	state := observed.State
	state.AnswersRecorded++
	state.Status, state.ActiveQuestion = QuizRuntimeStatusIdle, nil

	return QuizTransition{State: state, Entry: &entry, Message: "Answer recorded."}, nil
}

func checkQuizLegacyRead(t *testing.T, repo *fileQuizArtifactRepository, scope QuizScope, path string, before []byte) {
	t.Helper()

	observed, err := repo.Observe(t.Context(), scope)

	after, readErr := os.ReadFile(path)
	if err != nil || readErr != nil || string(before) != string(after) || !strings.HasPrefix(observed.State.SessionID, "legacy-") {
		t.Fatal("read migrated legacy state", observed, err, readErr)
	}
}

func checkQuizActionRecovery(t *testing.T, index int) {
	t.Helper()

	repo, request := newQuizActionFixture(t)

	repo.afterMutationReplace = func(at int) error {
		if at == index {
			return errors.New("interrupted publication")
		}

		return nil
	}
	if _, err := repo.ApplyAction(t.Context(), request, startQuizActionFixture); err == nil {
		t.Fatal("fault was not reached")
	}

	if _, err := repo.Observe(t.Context(), request.Scope); !errors.Is(err, ErrWorkflowConflict) {
		t.Fatal("read exposed a pending journal", err)
	}

	fresh := &fileQuizArtifactRepository{projectDir: repo.projectDir}

	result, err := fresh.ApplyAction(t.Context(), request, func(context.Context, QuizObservation) (QuizTransition, error) {
		t.Fatal("retry repeated the transition")

		return QuizTransition{}, nil
	})
	if err != nil {
		t.Fatal(err)
	}

	observed, err := fresh.Observe(t.Context(), request.Scope)
	if err != nil || result.State.Revision != observed.Revision || observed.State.SessionID != "session-1" {
		t.Fatal("recovered state/receipt mismatch", result, observed, err)
	}
}

func quizLegacyFixtureImages(t *testing.T, repo *fileQuizArtifactRepository, scope QuizScope) (QuizArtifactPaths, []byte) {
	t.Helper()

	paths, err := repo.Paths(scope)
	if err != nil {
		t.Fatal(err)
	}

	before, err := os.ReadFile(paths.RuntimePath)
	if err != nil {
		t.Fatal(err)
	}

	return paths, before
}
