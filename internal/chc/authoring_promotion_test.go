package chc

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestPlanQuizToDraftPromotionCreatesDraftWithProvenance(t *testing.T) {
	projectDir := t.TempDir()
	now := time.Date(2026, 5, 21, 16, 0, 0, 0, time.UTC)
	session, quizPath := writePromotionQuizFixture(t, projectDir, now)

	plan, err := PlanQuizToDraftPromotion(QuizToDraftPromotionOptions{
		ProjectDir: projectDir,
		Session:    session,
		QuizPath:   quizPath,
		Now:        now,
		DryRun:     true,
	})
	if err != nil {
		t.Fatalf("PlanQuizToDraftPromotion returned error: %v", err)
	}

	if plan.WouldWrite {
		t.Fatal("WouldWrite = true, want false for dry run")
	}

	if plan.Layout.Ref.Kind != AuthoringArtifactKindDraft {
		t.Fatalf("Layout.Ref.Kind = %q, want draft", plan.Layout.Ref.Kind)
	}

	if !strings.Contains(plan.Content, "# Draft") {
		t.Fatal("Content does not contain Draft heading")
	}

	if !strings.Contains(plan.Content, "The API must retry failed requests.") {
		t.Fatal("Content does not include quiz answer content")
	}

	if len(plan.Session.Provenance) != len(session.Provenance)+1 {
		t.Fatalf("len(Provenance) = %d, want %d", len(plan.Session.Provenance), len(session.Provenance)+1)
	}
}

func TestWriteQuizToDraftPromotionWritesDraft(t *testing.T) {
	projectDir := t.TempDir()
	now := time.Date(2026, 5, 21, 16, 0, 0, 0, time.UTC)
	session, quizPath := writePromotionQuizFixture(t, projectDir, now)
	plan := mustPlanQuizToDraftPromotion(t, projectDir, session, quizPath, now, false)

	err := WriteQuizToDraftPromotion(plan)
	if err != nil {
		t.Fatalf("WriteQuizToDraftPromotion returned error: %v", err)
	}

	content, err := os.ReadFile(plan.Layout.FullPath)
	if err != nil {
		t.Fatalf("ReadFile returned error: %v", err)
	}

	if !strings.Contains(string(content), session.Topic) {
		t.Fatal("written draft does not contain session topic")
	}
}

func TestWriteQuizToDraftPromotionRejectsUnallowedCollision(t *testing.T) {
	projectDir := t.TempDir()
	now := time.Date(2026, 5, 21, 16, 0, 0, 0, time.UTC)
	session, quizPath := writePromotionQuizFixture(t, projectDir, now)
	initial := mustPlanQuizToDraftPromotion(t, projectDir, session, quizPath, now, false)

	err := os.MkdirAll(filepath.Dir(initial.Layout.FullPath), 0o755)
	if err != nil {
		t.Fatalf("MkdirAll returned error: %v", err)
	}

	err = os.WriteFile(initial.Layout.FullPath, []byte("# Existing\n"), 0o644)
	if err != nil {
		t.Fatalf("WriteFile returned error: %v", err)
	}

	colliding := mustPlanQuizToDraftPromotion(t, projectDir, session, quizPath, now, false)

	err = WriteQuizToDraftPromotion(colliding)
	if err == nil {
		t.Fatal("WriteQuizToDraftPromotion returned nil error")
	}
}

func TestPlanQuizToDraftPromotionRejectsInvalidInput(t *testing.T) {
	now := time.Date(2026, 5, 21, 16, 0, 0, 0, time.UTC)
	session := validContextualQuizSession(now)

	for _, test := range []struct {
		name string
		opts QuizToDraftPromotionOptions
	}{
		{name: "project dir", opts: QuizToDraftPromotionOptions{Session: session, QuizPath: "quiz.md", Now: now}},
		{name: "session", opts: QuizToDraftPromotionOptions{ProjectDir: "/tmp/project", QuizPath: "quiz.md", Now: now}},
		{name: "quiz path", opts: QuizToDraftPromotionOptions{ProjectDir: "/tmp/project", Session: session, Now: now}},
		{name: "time", opts: QuizToDraftPromotionOptions{ProjectDir: "/tmp/project", Session: session, QuizPath: "quiz.md"}},
	} {
		t.Run(test.name, func(t *testing.T) {
			_, err := PlanQuizToDraftPromotion(test.opts)
			if err == nil {
				t.Fatal("PlanQuizToDraftPromotion returned nil error")
			}
		})
	}
}

func writePromotionQuizFixture(t *testing.T, projectDir string, now time.Time) (AuthoringSession, string) {
	t.Helper()

	session := validContextualQuizSession(now)
	quizPlan := mustPlanContextualQuiz(t, projectDir, session, now, false)

	err := WriteContextualQuiz(quizPlan)
	if err != nil {
		t.Fatalf("WriteContextualQuiz returned error: %v", err)
	}

	answerContent := strings.ReplaceAll(quizPlan.Content, "Answer:\n\n", "Answer:\n\nThe API must retry failed requests.\n\n")

	err = os.WriteFile(quizPlan.Layout.FullPath, []byte(answerContent), 0o644)
	if err != nil {
		t.Fatalf("WriteFile returned error: %v", err)
	}

	return quizPlan.Session, quizPlan.Layout.FullPath
}

func mustPlanContextualQuiz(
	t *testing.T,
	projectDir string,
	session AuthoringSession,
	now time.Time,
	allowReplace bool,
) ContextualQuizGenerationPlan {
	t.Helper()

	plan, err := PlanContextualQuizGeneration(context.Background(), ContextualQuizGenerationOptions{
		ProjectDir:   projectDir,
		Session:      session,
		Now:          now,
		AllowReplace: allowReplace,
		Provider: &FakeAuthoringQuizQuestionProvider{
			Result: validContextualQuizGenerationResult(),
		},
	})
	if err != nil {
		t.Fatalf("PlanContextualQuizGeneration returned error: %v", err)
	}

	return plan
}

func mustPlanQuizToDraftPromotion(
	t *testing.T,
	projectDir string,
	session AuthoringSession,
	quizPath string,
	now time.Time,
	allowReplace bool,
) QuizToDraftPromotionPlan {
	t.Helper()

	plan, err := PlanQuizToDraftPromotion(QuizToDraftPromotionOptions{
		ProjectDir:   projectDir,
		Session:      session,
		QuizPath:     quizPath,
		Now:          now,
		AllowReplace: allowReplace,
	})
	if err != nil {
		t.Fatalf("PlanQuizToDraftPromotion returned error: %v", err)
	}

	return plan
}
