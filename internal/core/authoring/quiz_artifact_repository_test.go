package authoring

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	generationcore "forge.adrianpk.com/chr/internal/core/authoring/generation"
	quizcore "forge.adrianpk.com/chr/internal/core/quiz"
)

func TestQuizArtifactLayoutResolvesProjectAndUseCasePaths(t *testing.T) {
	projectDir := t.TempDir()

	projectPaths, err := QuizArtifactLayout(projectDir, quizcore.Scope{ProjectID: "project-1"})
	if err != nil {
		t.Fatalf("QuizArtifactLayout project returned error: %v", err)
	}

	if projectPaths.RawRelativePath != "ops/default/quiz/00000-project-quiz-raw/artifact.md" {
		t.Fatalf("project raw path = %q", projectPaths.RawRelativePath)
	}

	if projectPaths.CuratedRelativePath != "ops/default/quiz/00000-project-quiz/artifact.md" {
		t.Fatalf("project curated path = %q", projectPaths.CuratedRelativePath)
	}

	if projectPaths.RuntimeRelativePath != ".chr/default/tmp/quiz-runtime/project.json" {
		t.Fatalf("project runtime path = %q", projectPaths.RuntimeRelativePath)
	}

	casePaths, err := QuizArtifactLayout(projectDir, quizcore.Scope{ProjectID: "project-1", CaseSlug: "article-authoring"})
	if err != nil {
		t.Fatalf("QuizArtifactLayout case returned error: %v", err)
	}

	if casePaths.RawRelativePath != "ops/default/quiz/00000-article-authoring-quiz-raw/artifact.md" {
		t.Fatalf("case raw path = %q", casePaths.RawRelativePath)
	}

	if casePaths.CuratedRelativePath != "ops/default/quiz/00000-article-authoring-quiz/artifact.md" {
		t.Fatalf("case curated path = %q", casePaths.CuratedRelativePath)
	}

	if casePaths.RuntimeRelativePath != ".chr/default/tmp/quiz-runtime/article-authoring.json" {
		t.Fatalf("case runtime path = %q", casePaths.RuntimeRelativePath)
	}
}

func TestQuizArtifactLayoutRejectsNestedCaseSlug(t *testing.T) {
	_, err := QuizArtifactLayout(t.TempDir(), quizcore.Scope{ProjectID: "project-1", CaseSlug: "../bad"})
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

	paths, err := repository.WriteCuratedNotes(context.Background(), quizcore.CuratedNotes{
		Scope:               quizcore.Scope{ProjectID: "project-1", CaseSlug: "article-authoring"},
		Decisions:           []string{"Use one visible question at a time."},
		AcceptanceReadiness: []string{"Decision: needs_more_info; missing: actor"},
		Constraints:         []string{"Do not expose fast candidate internals."},
		Entries:             []quizcore.Entry{validQuizEntry()},
		UpdatedAt:           quizRuntimeTestTime().Format(time.RFC3339),
	})
	if err != nil {
		t.Fatalf("WriteCuratedNotes returned error: %v", err)
	}

	content := readQuizArtifactTestFile(t, paths.CuratedPath)
	for _, want := range []string{
		`status: "curated"`,
		`case_slug: "article-authoring"`,
		`raw_path: "ops/default/quiz/00000-article-authoring-quiz-raw/artifact.md"`,
		"## Decisions",
		"- Use one visible question at a time.",
		"## Acceptance Readiness",
		"- Decision: needs_more_info; missing: actor",
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

	loaded, err := repository.LoadRuntimeState(context.Background(), quizcore.Scope{ProjectID: "project-1"})
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

	curatedPaths, err := repository.WriteCuratedNotes(context.Background(), quizcore.CuratedNotes{
		Scope:     quizcore.Scope{ProjectID: "project-1"},
		Decisions: []string{"Use service-owned quiz state."},
		Entries:   []quizcore.Entry{entry},
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

	entries, err := repository.LoadRawCaptureEntries(context.Background(), quizcore.Scope{ProjectID: "project-1"})
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

	notes := quizcore.CurateEntries(
		quizcore.Scope{ProjectID: "project-1"},
		[]quizcore.Entry{entry},
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

func TestCurateQuizEntriesPreservesAcceptanceReadiness(t *testing.T) {
	entry := quizEntryWith(func(entry *quizcore.Entry) {
		entry.PromptIntent = `{"kind":"quiz_question","metadata":{"routing_state":{"acceptance_readiness":{"decision":"needs_more_info","missing":["actor","trigger"],"inferred_elements":{"observable_outcome":["Report is written."]}}}}}`
	})

	notes := quizcore.CurateEntries(
		quizcore.Scope{ProjectID: "project-1"},
		[]quizcore.Entry{entry},
		quizRuntimeTestTime(),
	)

	if !containsString(notes.AcceptanceReadiness, "Decision: needs_more_info; missing: actor, trigger; inferred: observable_outcome") {
		t.Fatalf("acceptance readiness = %#v", notes.AcceptanceReadiness)
	}
}

func TestWriteQuizDraftSpecWritesSpecArtifactWithDraftStatus(t *testing.T) {
	dir := t.TempDir()
	spec := generationcore.QuizDraftSpec{
		Scope:           quizcore.Scope{ProjectID: "project-1", CaseSlug: "article-authoring"},
		Title:           "Article Authoring Spec",
		Body:            "Generated behavior.",
		SourceQuizPath:  "ops/default/quiz/00000-article-authoring-quiz/artifact.md",
		SourceRawPath:   "ops/default/quiz/00000-article-authoring-quiz-raw/artifact.md",
		SourceDumpPaths: []string{".chr/default/dumps/article-authoring-context.md"},
		CreatedAt:       quizRuntimeTestTime(),
	}

	paths, err := WriteQuizDraftSpec(context.Background(), dir, spec)
	if err != nil {
		t.Fatalf("WriteQuizDraftSpec returned error: %v", err)
	}

	if paths.RelativePath != "ops/default/spec/00000-article-authoring-quiz-spec/artifact.md" {
		t.Fatalf("relative path = %q", paths.RelativePath)
	}

	content := readQuizArtifactTestFile(t, paths.Path)
	for _, want := range []string{
		`status: "draft"`,
		`case_slug: "article-authoring"`,
		"# Article Authoring Spec",
		"Generated behavior.",
		"- curated quiz: ops/default/quiz/00000-article-authoring-quiz/artifact.md",
		"- dump: .chr/default/dumps/article-authoring-context.md",
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
