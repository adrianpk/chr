package chc

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestQuizArtifactLayoutResolvesProjectAndUseCasePaths(t *testing.T) {
	projectDir := t.TempDir()

	projectPaths, err := QuizArtifactLayout(projectDir, QuizScope{ProjectID: "project-1"})
	if err != nil {
		t.Fatalf("QuizArtifactLayout project returned error: %v", err)
	}

	if projectPaths.RawRelativePath != ".chc/user/quizzes/project-quiz-raw.md" {
		t.Fatalf("project raw path = %q", projectPaths.RawRelativePath)
	}

	if projectPaths.CuratedRelativePath != ".chc/user/quizzes/project-quiz.md" {
		t.Fatalf("project curated path = %q", projectPaths.CuratedRelativePath)
	}

	if projectPaths.RuntimeRelativePath != ".chc/user/tmp/quiz-runtime/project.json" {
		t.Fatalf("project runtime path = %q", projectPaths.RuntimeRelativePath)
	}

	casePaths, err := QuizArtifactLayout(projectDir, QuizScope{ProjectID: "project-1", CaseSlug: "article-authoring"})
	if err != nil {
		t.Fatalf("QuizArtifactLayout case returned error: %v", err)
	}

	if casePaths.RawRelativePath != ".chc/user/quizzes/article-authoring-quiz-raw.md" {
		t.Fatalf("case raw path = %q", casePaths.RawRelativePath)
	}

	if casePaths.CuratedRelativePath != ".chc/user/quizzes/article-authoring-quiz.md" {
		t.Fatalf("case curated path = %q", casePaths.CuratedRelativePath)
	}

	if casePaths.RuntimeRelativePath != ".chc/user/tmp/quiz-runtime/article-authoring.json" {
		t.Fatalf("case runtime path = %q", casePaths.RuntimeRelativePath)
	}
}

func TestQuizArtifactLayoutRejectsNestedCaseSlug(t *testing.T) {
	_, err := QuizArtifactLayout(t.TempDir(), QuizScope{ProjectID: "project-1", CaseSlug: "../bad"})
	if err == nil {
		t.Fatal("QuizArtifactLayout returned nil error")
	}
}

func TestQuizArtifactRepositoryAppendsRawCapture(t *testing.T) {
	dir := t.TempDir()
	repository := NewQuizArtifactRepository(dir)
	state := validQuizRuntimeState()
	entry := validQuizEntry()

	paths, err := repository.AppendRawCapture(context.Background(), state, entry)
	if err != nil {
		t.Fatalf("AppendRawCapture returned error: %v", err)
	}

	content := readQuizArtifactTestFile(t, paths.RawPath)
	for _, want := range []string{
		`status: "raw"`,
		`runtime_id: "quiz-project"`,
		"question_id: question-1",
		"Prompt intent:",
		`"follow_up_depth":0`,
		"Question:",
		"What should this workflow preserve?",
		"Answer:",
		"It should preserve one visible question at a time.",
	} {
		if !strings.Contains(content, want) {
			t.Fatalf("raw capture missing %q in:\n%s", want, content)
		}
	}
}

func TestQuizArtifactRepositoryWritesCuratedNotes(t *testing.T) {
	dir := t.TempDir()
	repository := NewQuizArtifactRepository(dir)

	paths, err := repository.WriteCuratedNotes(context.Background(), QuizCuratedNotes{
		Scope:       QuizScope{ProjectID: "project-1", CaseSlug: "article-authoring"},
		Decisions:   []string{"Use one visible question at a time."},
		Constraints: []string{"Do not expose fast candidate internals."},
		Entries:     []QuizEntry{validQuizEntry()},
		UpdatedAt:   quizRuntimeTestTime().Format(time.RFC3339),
	})
	if err != nil {
		t.Fatalf("WriteCuratedNotes returned error: %v", err)
	}

	content := readQuizArtifactTestFile(t, paths.CuratedPath)
	for _, want := range []string{
		`status: "curated"`,
		`case_slug: "article-authoring"`,
		`raw_path: ".chc/user/quizzes/article-authoring-quiz-raw.md"`,
		"## Decisions",
		"- Use one visible question at a time.",
		"## Constraints",
		"- Do not expose fast candidate internals.",
		"## Provenance",
	} {
		if !strings.Contains(content, want) {
			t.Fatalf("curated notes missing %q in:\n%s", want, content)
		}
	}
}

func TestQuizArtifactRepositorySavesAndLoadsRuntimeState(t *testing.T) {
	dir := t.TempDir()
	repository := NewQuizArtifactRepository(dir)
	state := validQuizRuntimeState()

	paths, err := repository.SaveRuntimeState(context.Background(), state)
	if err != nil {
		t.Fatalf("SaveRuntimeState returned error: %v", err)
	}

	content := readQuizArtifactTestFile(t, paths.RuntimePath)
	if !strings.Contains(content, "candidate-1") {
		t.Fatal("candidate id missing from runtime state")
	}

	loaded, err := repository.LoadRuntimeState(context.Background(), QuizScope{ProjectID: "project-1"})
	if err != nil {
		t.Fatalf("LoadRuntimeState returned error: %v", err)
	}

	if loaded.ID != state.ID {
		t.Fatalf("loaded runtime id = %q, want %q", loaded.ID, state.ID)
	}

	if len(loaded.Candidates) != 1 {
		t.Fatalf("loaded candidates = %#v, want one candidate", loaded.Candidates)
	}
}

func TestQuizArtifactRepositoryKeepsCandidatesOutOfRawAndCuratedArtifacts(t *testing.T) {
	dir := t.TempDir()
	repository := NewQuizArtifactRepository(dir)
	state := validQuizRuntimeState()
	entry := validQuizEntry()

	rawPaths, err := repository.AppendRawCapture(context.Background(), state, entry)
	if err != nil {
		t.Fatalf("AppendRawCapture returned error: %v", err)
	}

	curatedPaths, err := repository.WriteCuratedNotes(context.Background(), QuizCuratedNotes{
		Scope:     QuizScope{ProjectID: "project-1"},
		Decisions: []string{"Use service-owned quiz state."},
		Entries:   []QuizEntry{entry},
		UpdatedAt: quizRuntimeTestTime().Format(time.RFC3339),
	})
	if err != nil {
		t.Fatalf("WriteCuratedNotes returned error: %v", err)
	}

	raw := readQuizArtifactTestFile(t, rawPaths.RawPath)
	curated := readQuizArtifactTestFile(t, curatedPaths.CuratedPath)

	for _, content := range []string{raw, curated} {
		if strings.Contains(content, "candidate-1") {
			t.Fatalf("candidate leaked into artifact:\n%s", content)
		}
	}
}

func TestQuizArtifactRepositoryLoadsRawCaptureEntries(t *testing.T) {
	dir := t.TempDir()
	repository := NewQuizArtifactRepository(dir)
	state := validQuizRuntimeState()

	_, err := repository.AppendRawCapture(context.Background(), state, validQuizEntry())
	if err != nil {
		t.Fatalf("AppendRawCapture returned error: %v", err)
	}

	entries, err := repository.LoadRawCaptureEntries(context.Background(), QuizScope{ProjectID: "project-1"})
	if err != nil {
		t.Fatalf("LoadRawCaptureEntries returned error: %v", err)
	}

	if len(entries) != 1 {
		t.Fatalf("entries = %#v, want one entry", entries)
	}

	if entries[0].QuestionID != "question-1" {
		t.Fatalf("QuestionID = %q, want question-1", entries[0].QuestionID)
	}

	if !strings.Contains(entries[0].PromptIntent, `"follow_up_depth":0`) {
		t.Fatalf("PromptIntent = %q, want persisted follow-up depth", entries[0].PromptIntent)
	}
}

func TestCurateQuizEntriesDistillsAnsweredTurns(t *testing.T) {
	entry := validQuizEntry()
	entry.Answer = "The TUI must preserve one visible question at a time. Extra raw detail."

	notes := CurateQuizEntries(
		QuizScope{ProjectID: "project-1"},
		[]QuizEntry{entry},
		quizRuntimeTestTime(),
	)

	if len(notes.Decisions) != 1 {
		t.Fatalf("Decisions = %#v, want one distilled decision", notes.Decisions)
	}

	if strings.Contains(notes.Decisions[0], entry.Question) {
		t.Fatalf("decision mirrors raw question:\n%s", notes.Decisions[0])
	}

	if !strings.Contains(notes.Decisions[0], "workflow") {
		t.Fatalf("decision missing topic:\n%s", notes.Decisions[0])
	}
}

func TestWriteQuizDraftSpecWritesSpecArtifactWithDraftStatus(t *testing.T) {
	dir := t.TempDir()
	spec := QuizDraftSpec{
		Scope:           QuizScope{ProjectID: "project-1", CaseSlug: "article-authoring"},
		Title:           "Article Authoring Spec",
		Body:            "Generated behavior.",
		SourceQuizPath:  ".chc/user/quizzes/article-authoring-quiz.md",
		SourceRawPath:   ".chc/user/quizzes/article-authoring-quiz-raw.md",
		SourceDumpPaths: []string{".chc/user/dumps/article-authoring-context.md"},
		CreatedAt:       quizRuntimeTestTime(),
	}

	paths, err := WriteQuizDraftSpec(context.Background(), dir, spec)
	if err != nil {
		t.Fatalf("WriteQuizDraftSpec returned error: %v", err)
	}

	if paths.RelativePath != ".chc/user/specs/drafts/article-authoring-quiz-spec.md" {
		t.Fatalf("relative path = %q", paths.RelativePath)
	}

	content := readQuizArtifactTestFile(t, paths.Path)
	for _, want := range []string{
		`status: "draft"`,
		`case_slug: "article-authoring"`,
		"# Article Authoring Spec",
		"Generated behavior.",
		"- curated quiz: .chc/user/quizzes/article-authoring-quiz.md",
		"- dump: .chc/user/dumps/article-authoring-context.md",
	} {
		if !strings.Contains(content, want) {
			t.Fatalf("draft spec missing %q in:\n%s", want, content)
		}
	}
}

func readQuizArtifactTestFile(t *testing.T, path string) string {
	t.Helper()

	content, err := os.ReadFile(filepath.Clean(path))
	if err != nil {
		t.Fatalf("ReadFile returned error: %v", err)
	}

	return string(content)
}
