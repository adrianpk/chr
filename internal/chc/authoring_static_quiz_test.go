package chc

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestPlanStaticQuizGenerationCreatesEditableMarkdown(t *testing.T) {
	now := time.Date(2026, 5, 21, 15, 0, 0, 0, time.UTC)
	session := validStaticQuizSession(now)

	plan, err := PlanStaticQuizGeneration(StaticQuizGenerationOptions{
		ProjectDir:    t.TempDir(),
		Session:       session,
		Now:           now,
		DryRun:        true,
		QuestionCount: 2,
	})
	if err != nil {
		t.Fatalf("PlanStaticQuizGeneration returned error: %v", err)
	}

	if plan.WouldWrite {
		t.Fatal("WouldWrite = true, want false for dry run")
	}

	if len(plan.Questions) != 2 {
		t.Fatalf("len(Questions) = %d, want 2", len(plan.Questions))
	}

	if !strings.Contains(plan.Content, "# Static Quiz") {
		t.Fatal("Content does not contain Static Quiz heading")
	}

	if !strings.Contains(plan.Content, "Answer:") {
		t.Fatal("Content does not contain answer slots")
	}

	if len(plan.Session.Outputs) != 1 {
		t.Fatalf("len(Session.Outputs) = %d, want 1", len(plan.Session.Outputs))
	}

	if len(plan.Session.Provenance) != 1 {
		t.Fatalf("len(Session.Provenance) = %d, want 1", len(plan.Session.Provenance))
	}
}

func TestPlanStaticQuizGenerationSupportsEveryAuthoringTarget(t *testing.T) {
	now := time.Date(2026, 5, 21, 15, 0, 0, 0, time.UTC)

	for _, target := range staticQuizTemplateTargets() {
		t.Run(string(target), func(t *testing.T) {
			session := validStaticQuizSession(now)
			session.Target = target

			plan, err := PlanStaticQuizGeneration(StaticQuizGenerationOptions{
				ProjectDir: t.TempDir(),
				Session:    session,
				Now:        now,
				DryRun:     true,
			})
			if err != nil {
				t.Fatalf("PlanStaticQuizGeneration returned error: %v", err)
			}

			if len(plan.Questions) == 0 {
				t.Fatal("len(Questions) = 0, want at least one question")
			}
		})
	}
}

func TestPlanStaticQuizGenerationRejectsInvalidOptions(t *testing.T) {
	now := time.Date(2026, 5, 21, 15, 0, 0, 0, time.UTC)

	for _, test := range invalidStaticQuizGenerationTests(now) {
		t.Run(test.name, func(t *testing.T) {
			_, err := PlanStaticQuizGeneration(test.opts)
			if err == nil {
				t.Fatal("PlanStaticQuizGeneration returned nil error")
			}
		})
	}
}

func invalidStaticQuizGenerationTests(now time.Time) []struct {
	name string
	opts StaticQuizGenerationOptions
} {
	session := validStaticQuizSession(now)
	nonStatic := validStaticQuizSession(now)
	nonStatic.Mode = AuthoringSessionModeMarkdownDirect

	return []struct {
		name string
		opts StaticQuizGenerationOptions
	}{
		{
			name: "project dir",
			opts: StaticQuizGenerationOptions{Session: session, Now: now},
		},
		{
			name: "session",
			opts: StaticQuizGenerationOptions{ProjectDir: "/tmp/project", Now: now},
		},
		{
			name: "mode",
			opts: StaticQuizGenerationOptions{ProjectDir: "/tmp/project", Session: nonStatic, Now: now},
		},
		{
			name: "time",
			opts: StaticQuizGenerationOptions{ProjectDir: "/tmp/project", Session: session},
		},
	}
}

func TestWriteStaticQuizWritesArtifact(t *testing.T) {
	projectDir := t.TempDir()
	now := time.Date(2026, 5, 21, 15, 0, 0, 0, time.UTC)
	session := validStaticQuizSession(now)

	plan, err := PlanStaticQuizGeneration(StaticQuizGenerationOptions{
		ProjectDir: projectDir,
		Session:    session,
		Now:        now,
	})
	if err != nil {
		t.Fatalf("PlanStaticQuizGeneration returned error: %v", err)
	}

	err = WriteStaticQuiz(plan)
	if err != nil {
		t.Fatalf("WriteStaticQuiz returned error: %v", err)
	}

	content, err := os.ReadFile(plan.Layout.FullPath)
	if err != nil {
		t.Fatalf("ReadFile returned error: %v", err)
	}

	if !strings.Contains(string(content), session.Topic) {
		t.Fatal("written quiz does not contain session topic")
	}
}

func TestWriteStaticQuizRejectsUnallowedCollision(t *testing.T) {
	projectDir := t.TempDir()
	now := time.Date(2026, 5, 21, 15, 0, 0, 0, time.UTC)
	session := validStaticQuizSession(now)

	initial := mustPlanStaticQuiz(t, projectDir, session, now, false)

	err := os.MkdirAll(filepath.Dir(initial.Layout.FullPath), 0o755)
	if err != nil {
		t.Fatalf("MkdirAll returned error: %v", err)
	}

	err = os.WriteFile(initial.Layout.FullPath, []byte("# Existing\n"), 0o644)
	if err != nil {
		t.Fatalf("WriteFile returned error: %v", err)
	}

	colliding := mustPlanStaticQuiz(t, projectDir, session, now, false)

	err = WriteStaticQuiz(colliding)
	if err == nil {
		t.Fatal("WriteStaticQuiz returned nil error")
	}
}

func mustPlanStaticQuiz(
	t *testing.T,
	projectDir string,
	session AuthoringSession,
	now time.Time,
	allowReplace bool,
) StaticQuizGenerationPlan {
	t.Helper()

	plan, err := PlanStaticQuizGeneration(StaticQuizGenerationOptions{
		ProjectDir:   projectDir,
		Session:      session,
		Now:          now,
		AllowReplace: allowReplace,
	})
	if err != nil {
		t.Fatalf("PlanStaticQuizGeneration returned error: %v", err)
	}

	return plan
}

func validStaticQuizSession(now time.Time) AuthoringSession {
	return AuthoringSession{
		ID:        "auth-20260521-150000",
		ProjectID: "project-1",
		Status:    AuthoringSessionStatusDraft,
		Mode:      AuthoringSessionModeStatic,
		Target:    AuthoringTargetSpec,
		Topic:     "Billing sync behavior",
		CreatedAt: now,
		UpdatedAt: now,
	}
}
