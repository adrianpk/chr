package chc

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestPlanContextualQuizGenerationUsesProviderAndProjectContext(t *testing.T) {
	projectDir := t.TempDir()
	now := time.Date(2026, 5, 22, 9, 0, 0, 0, time.UTC)
	session := validContextualQuizSession(now)
	session.ID = "author-20260522-090000"
	provider := &FakeAuthoringQuizQuestionProvider{
		Result: validContextualQuizGenerationResult(),
	}

	writeContextualQuizProjectFixture(t, projectDir)

	plan, err := PlanContextualQuizGeneration(context.Background(), ContextualQuizGenerationOptions{
		ProjectDir:    projectDir,
		Session:       session,
		Now:           now,
		DryRun:        true,
		QuestionCount: 8,
		Provider:      provider,
	})
	if err != nil {
		t.Fatalf("PlanContextualQuizGeneration returned error: %v", err)
	}

	assertContextualQuizPlan(t, plan, provider)
}

func assertContextualQuizPlan(
	t *testing.T,
	plan ContextualQuizGenerationPlan,
	provider *FakeAuthoringQuizQuestionProvider,
) {
	t.Helper()

	if plan.WouldWrite {
		t.Fatal("WouldWrite = true, want false for dry run")
	}

	if len(provider.Requests) != 1 {
		t.Fatalf("len(provider.Requests) = %d, want 1", len(provider.Requests))
	}

	assertContextualQuizRequest(t, provider.Requests[0])

	if plan.RequestedQuestionCount != 8 {
		t.Fatalf("RequestedQuestionCount = %d, want 8", plan.RequestedQuestionCount)
	}

	if plan.GeneratedQuestionCount != 3 {
		t.Fatalf("GeneratedQuestionCount = %d, want 3", plan.GeneratedQuestionCount)
	}

	if !strings.Contains(plan.Content, "# Context-Aware Quiz") {
		t.Fatal("Content does not contain Context-Aware Quiz heading")
	}

	if !strings.Contains(plan.Content, "generated_questions: 3") {
		t.Fatal("Content does not record generated question count")
	}
}

func assertContextualQuizRequest(t *testing.T, request AuthoringQuizGenerationRequest) {
	t.Helper()

	if request.Action != AuthoringQuizActionGeneration {
		t.Fatalf("Action = %q, want %q", request.Action, AuthoringQuizActionGeneration)
	}

	if request.Strategy != AuthoringQuizGenerationStrategyDeep {
		t.Fatalf("Strategy = %q, want %q", request.Strategy, AuthoringQuizGenerationStrategyDeep)
	}

	if request.QuestionCount != 8 {
		t.Fatalf("QuestionCount = %d, want 8", request.QuestionCount)
	}

	for _, name := range []string{
		"project",
		"session",
		"agent_protocol",
		"workspace_graph_context",
		"docs",
		"tree",
		"go_inventory",
	} {
		assertAuthoringQuizContextSection(t, request.ContextSections, name)
	}

	workspaceContext := authoringQuizContextSection(t, request.ContextSections, "workspace_graph_context")
	if !strings.Contains(workspaceContext.Content, "Project Context Plan") {
		t.Fatalf("workspace graph context missing project-wide plan:\n%s", workspaceContext.Content)
	}
}

func writeContextualQuizProjectFixture(t *testing.T, projectDir string) {
	t.Helper()

	err := os.WriteFile(filepath.Join(projectDir, "AGENTS.md"), []byte("# Agent Rules\n"), 0o644)
	if err != nil {
		t.Fatalf("WriteFile returned error: %v", err)
	}

	err = os.MkdirAll(filepath.Join(projectDir, "docs"), 0o755)
	if err != nil {
		t.Fatalf("MkdirAll returned error: %v", err)
	}

	err = os.WriteFile(filepath.Join(projectDir, "docs", "overview.md"), []byte("# Overview\n"), 0o644)
	if err != nil {
		t.Fatalf("WriteFile returned error: %v", err)
	}

	writeTestFile(t, ProjectMetadataPath(projectDir), RenderProjectMetadata(ProjectMetadata{
		SchemaVersion: ProjectMetadataSchemaVersion,
		Name:          "Contextual Quiz Project",
		Description:   "Project description.",
		Context:       "Project body.",
	}))
	writeTestFile(t, filepath.Join(projectDir, ".chc", "plans", "project-context-plan.md"), `---
title: Project Context Plan
---

# Project Context Plan
`)
}

func TestPlanContextualQuizGenerationRejectsMissingProvider(t *testing.T) {
	now := time.Date(2026, 5, 22, 9, 0, 0, 0, time.UTC)
	session := validContextualQuizSession(now)

	_, err := PlanContextualQuizGeneration(context.Background(), ContextualQuizGenerationOptions{
		ProjectDir: t.TempDir(),
		Session:    session,
		Now:        now,
	})
	if err == nil {
		t.Fatal("PlanContextualQuizGeneration returned nil error")
	}
}

func TestValidateAuthoringQuizGenerationResultRejectsGenericUnsafeOutput(t *testing.T) {
	result := validContextualQuizGenerationResult()
	result.Questions[0].Text = "Which API token should this project use?"

	err := ValidateAuthoringQuizGenerationResult(result)
	if err == nil {
		t.Fatal("ValidateAuthoringQuizGenerationResult returned nil error")
	}
}

func TestWriteContextualQuizWritesArtifact(t *testing.T) {
	projectDir := t.TempDir()
	now := time.Date(2026, 5, 22, 9, 0, 0, 0, time.UTC)
	session := validContextualQuizSession(now)
	provider := &FakeAuthoringQuizQuestionProvider{
		Result: validContextualQuizGenerationResult(),
	}

	plan, err := PlanContextualQuizGeneration(context.Background(), ContextualQuizGenerationOptions{
		ProjectDir: projectDir,
		Session:    session,
		Now:        now,
		Provider:   provider,
	})
	if err != nil {
		t.Fatalf("PlanContextualQuizGeneration returned error: %v", err)
	}

	err = WriteContextualQuiz(plan)
	if err != nil {
		t.Fatalf("WriteContextualQuiz returned error: %v", err)
	}

	content, err := os.ReadFile(plan.Layout.FullPath)
	if err != nil {
		t.Fatalf("ReadFile returned error: %v", err)
	}

	if !strings.Contains(string(content), "Which user workflow should the specification pin down first?") {
		t.Fatal("written quiz does not contain generated question")
	}
}

func assertAuthoringQuizContextSection(
	t *testing.T,
	sections []AuthoringQuizContextSection,
	name string,
) {
	t.Helper()

	_ = authoringQuizContextSection(t, sections, name)
}

func authoringQuizContextSection(
	t *testing.T,
	sections []AuthoringQuizContextSection,
	name string,
) AuthoringQuizContextSection {
	t.Helper()

	for _, section := range sections {
		if section.Name == name {
			return section
		}
	}

	t.Fatalf("context sections do not contain %q: %#v", name, sections)

	return AuthoringQuizContextSection{}
}

func validContextualQuizGenerationResult() AuthoringQuizGenerationResult {
	return AuthoringQuizGenerationResult{
		Provider: "fake",
		Strategy: AuthoringQuizGenerationStrategyDeep,
		Questions: []AuthoringQuizQuestion{
			{
				ID:                    "Q1",
				Text:                  "Which user workflow should the specification pin down first?",
				Tags:                  []string{"workflow", "specification"},
				FollowUpDepth:         1,
				CompletionRecommended: false,
			},
			{
				ID:                    "Q2",
				Text:                  "What project files or docs should be treated as authoritative context?",
				Tags:                  []string{"context", "documents"},
				FollowUpDepth:         1,
				CompletionRecommended: false,
			},
			{
				ID:                    "Q3",
				Text:                  "What validation evidence should prove the generated spec is useful?",
				Tags:                  []string{"validation", "evidence"},
				FollowUpDepth:         0,
				CompletionRecommended: true,
			},
		},
		RawText: "{}",
	}
}

func validContextualQuizSession(now time.Time) AuthoringSession {
	return AuthoringSession{
		ID:        "author-20260522-090000",
		ProjectID: "project-1",
		Status:    AuthoringSessionStatusDraft,
		Mode:      AuthoringSessionModeContextual,
		Target:    AuthoringTargetSpec,
		Topic:     "Billing sync behavior",
		CreatedAt: now,
		UpdatedAt: now,
	}
}
