package app

import (
	"testing"

	generationcore "forge.adrianpk.com/chr/internal/core/authoring/generation"
	opscore "forge.adrianpk.com/chr/internal/core/ops"
	quizcore "forge.adrianpk.com/chr/internal/core/quiz"
	quizmodel "forge.adrianpk.com/chr/internal/core/quiz/model"
	"forge.adrianpk.com/chr/internal/tui"
)

func TestTUIDumpAdapterUsesDumpServices(t *testing.T) {
	projectDir := t.TempDir()
	writeServiceWorkspaceQueryFixture(t, projectDir)

	operations := newTestService(t, projectDir)
	executor := newTUIDumpActionExecutor(operations)

	dumpPath := createDumpThroughTUIAdapter(t, executor)
	editDumpThroughTUIAdapter(t, executor, dumpPath)
	saveDumpThroughTUIAdapter(t, executor, dumpPath)

	listResult, err := operations.ListDumps()
	if err != nil {
		t.Fatalf("ListDumps returned error: %v", err)
	}

	saved := findDumpViewByTitle(t, listResult.Dumps, "Adapter Notes")
	if saved.Scope.CaseSlug != "blog" {
		t.Fatalf("saved scope = %#v, want blog", saved.Scope)
	}
}

func createDumpThroughTUIAdapter(t *testing.T, executor tuiDumpActionExecutor) string {
	t.Helper()

	createResult, err := executor.ExecuteDumpAction(tui.DumpActionRequest{
		Action:  tui.DumpActionCreate,
		Title:   "Adapter Notes",
		Content: "Created through TUI adapter.",
		Scope:   tui.DumpScope{CaseSlug: "blog"},
	})
	if err != nil {
		t.Fatalf("ExecuteDumpAction create returned error: %v", err)
	}

	if createResult.Message != "Captured dump ops/default/dump/00000-adapter-notes/artifact.md." {
		t.Fatalf("create message = %q, want created dump message", createResult.Message)
	}

	return "ops/default/dump/00000-adapter-notes/artifact.md"
}

func editDumpThroughTUIAdapter(t *testing.T, executor tuiDumpActionExecutor, dumpPath string) {
	t.Helper()

	editResult, err := executor.ExecuteDumpAction(tui.DumpActionRequest{
		Action: tui.DumpActionEdit,
		Path:   dumpPath,
	})
	if err != nil {
		t.Fatalf("ExecuteDumpAction edit returned error: %v", err)
	}

	if editResult.SelectedPath != dumpPath {
		t.Fatalf("edit selected path = %q, want %q", editResult.SelectedPath, dumpPath)
	}
}

func saveDumpThroughTUIAdapter(t *testing.T, executor tuiDumpActionExecutor, dumpPath string) {
	t.Helper()

	observed, err := executor.operations.ReadDump(DumpReadRequest{RelativePath: dumpPath})
	if err != nil {
		t.Fatal(err)
	}

	saveResult, err := executor.ExecuteDumpAction(tui.DumpActionRequest{
		Action: tui.DumpActionSave, OperationID: "tui-save", Revision: observed.Revision,
		Path: dumpPath,
		Content: `---
title: "Adapter Notes"
status: "raw"
case: "blog"
---

# Adapter Notes

Saved through TUI adapter.`,
	})
	if err != nil {
		t.Fatalf("ExecuteDumpAction save returned error: %v", err)
	}

	if saveResult.SelectedPath != dumpPath {
		t.Fatalf("save selected path = %q, want %q", saveResult.SelectedPath, dumpPath)
	}
}

func TestTUIQuizAdapterDispatchesRuntimeQuizActions(t *testing.T) {
	projectDir := t.TempDir()
	writeServiceQuizFixture(t, projectDir)

	clock := serviceQuizClock()
	model := &fakeServiceQuizModel{
		responses: []quizmodel.Response{
			serviceQuizModelResponse("response-1", "question-1", clock()),
			serviceQuizModelResponse("response-2", "question-2", clock()),
		},
	}
	operations := newTestService(t, projectDir)
	operations.quizModel = model
	operations.now = clock

	executor := newTUIQuizActionExecutor(operations)

	startResult, err := executor.ExecuteQuizAction(tui.QuizActionRequest{
		Binding: observedTUIQuizBinding(t, operations, "blog", "adapter-start"),
		Action:  tui.QuizActionStart,
		Scope:   tui.QuizScope{CaseSlug: "blog"},
	})
	if err != nil {
		t.Fatalf("ExecuteQuizAction start returned error: %v", err)
	}

	if startResult.Message != "Quiz started." {
		t.Fatalf("start message = %q, want Quiz started.", startResult.Message)
	}

	answerResult, err := executor.ExecuteQuizAction(tui.QuizActionRequest{
		Binding: observedTUIQuizBinding(t, operations, "blog", "adapter-answer"),
		Action:  tui.QuizActionAnswer,
		Scope:   tui.QuizScope{CaseSlug: "blog"},
		Answer:  "Adapter answer.",
	})
	if err != nil {
		t.Fatalf("ExecuteQuizAction answer returned error: %v", err)
	}

	if answerResult.Message != "Quiz answer recorded." {
		t.Fatalf("answer message = %q, want answer recorded", answerResult.Message)
	}

	modeResult, err := executor.ExecuteQuizAction(tui.QuizActionRequest{
		Binding: observedTUIQuizBinding(t, operations, "blog", "adapter-mode"),
		Action:  tui.QuizActionFastMode,
		Scope:   tui.QuizScope{CaseSlug: "blog"},
		Mode:    string(quizcore.ModeFast),
	})
	if err != nil {
		t.Fatalf("ExecuteQuizAction mode returned error: %v", err)
	}

	if modeResult.Message != "Quiz mode switched." {
		t.Fatalf("mode message = %q, want mode switched", modeResult.Message)
	}
}

func TestTUISpecAdapterGeneratesAcceptanceFeature(t *testing.T) {
	projectDir := t.TempDir()
	writeAcceptanceFeatureServiceFixture(t, projectDir)
	// The concrete catalog reads the canonical body and metadata together.
	for _, reference := range []string{"ops/default/spec/00000-blog-spec/artifact.md", "ops/default/spec/00000-blog-prior-spec/artifact.md"} {
		if err := opscore.EnsureArtifactBodyMetadata(projectDir, reference); err != nil {
			t.Fatal(err)
		}
	}

	clock := serviceQuizClock()
	model := &fakeServiceQuizModel{
		responses: []quizmodel.Response{
			serviceQuizAcceptanceFeatureResponse("response-feature", clock(), generationcore.AcceptanceFeatureGeneration{
				Status: generationcore.AcceptanceFeatureStatusGenerated,
				Body: `@spec:blog-spec @use-case:blog
Feature: Blog publishing

  Scenario: Editorial review is preserved
    Given an article draft has editorial review status
    When the article is published
    Then the published article keeps the editorial review status

  Scenario: Publishing dependency fails
    Given an article draft has editorial review status
    When the publishing dependency is unavailable
    Then the article remains unpublished and the editor sees the failure
`,
			}),
		},
	}
	operations := newTestService(t, projectDir)
	operations.quizModel = model
	operations.now = clock

	executor := newTUISpecActionExecutor(operations)

	result, err := executor.ExecuteSpecAction(tui.SpecActionRequest{
		Action: tui.SpecActionGenerateFeature,
		Scope:  tui.QuizScope{CaseSlug: "blog"},
	})
	if err != nil {
		t.Fatalf("ExecuteSpecAction returned error: %v", err)
	}

	if result.AcceptanceFeatureStatus != generationcore.AcceptanceFeatureStatusGenerated {
		t.Fatalf("AcceptanceFeatureStatus = %q, want generated", result.AcceptanceFeatureStatus)
	}

	if result.AcceptanceFeaturePath != "ops/default/spec/00000-blog-acceptance-feature/artifact.md" {
		t.Fatalf("AcceptanceFeaturePath = %q, want feature path", result.AcceptanceFeaturePath)
	}

	if len(result.Specs) == 0 {
		t.Fatal("Specs is empty, want refreshed specs")
	}
}
