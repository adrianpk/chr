package core

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
	writeTestFile(t, filepath.Join(projectDir, ".chr", "default", "plans", "project-context-plan.md"), `---
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

func TestCodexAuthoringQuizQuestionProviderRecordsTraceEvents(t *testing.T) {
	projectDir := t.TempDir()
	t.Setenv(ActiveOperatorIDEnv, "adrian")
	t.Setenv(LLMTraceEnabledEnv, "1")

	program := writeAuthoringQuizProviderScript(t, projectDir, `#!/bin/sh
cat >/dev/null
printf '%s\n' '{"questions":[{"id":"Q1","question":"Which workflow should this spec cover?","tags":["workflow"],"follow_up_depth":1,"completion_recommended":false}]}'
`)
	request := validAuthoringQuizGenerationRequest(projectDir)

	result, err := (CodexAuthoringQuizQuestionProvider{Program: program}).GenerateAuthoringQuizQuestions(context.Background(), request)
	if err != nil {
		t.Fatalf("GenerateAuthoringQuizQuestions returned error: %v", err)
	}

	if len(result.Questions) != 1 {
		t.Fatalf("question count = %d, want 1", len(result.Questions))
	}

	events := readLLMTraceEvents(t, projectDir)

	assertCodexTraceEventsWithoutPayloads(t, events)
}

func TestCodexAuthoringQuizQuestionProviderRecordsTracePayloadsWhenEnabled(t *testing.T) {
	projectDir := t.TempDir()
	t.Setenv(ActiveOperatorIDEnv, "adrian")
	t.Setenv(LLMTraceEnabledEnv, "1")
	t.Setenv(LLMTracePayloadsEnv, "1")

	program := writeAuthoringQuizProviderScript(t, projectDir, `#!/bin/sh
cat >/dev/null
printf '%s\n' '{"questions":[{"id":"Q1","question":"Which workflow should this spec cover?","tags":["workflow"],"follow_up_depth":1,"completion_recommended":false}]}'
`)
	request := validAuthoringQuizGenerationRequest(projectDir)

	_, err := (CodexAuthoringQuizQuestionProvider{Program: program}).GenerateAuthoringQuizQuestions(context.Background(), request)
	if err != nil {
		t.Fatalf("GenerateAuthoringQuizQuestions returned error: %v", err)
	}

	events := readLLMTraceEvents(t, projectDir)
	if _, ok := events[0]["request_payload"]; !ok {
		t.Fatalf("request payload missing with payload opt-in: %#v", events[0])
	}

	if _, ok := events[1]["response_payload"]; !ok {
		t.Fatalf("response payload missing with payload opt-in: %#v", events[1])
	}
}

func TestCodexAuthoringQuizQuestionProviderRecordsParseFailureTrace(t *testing.T) {
	projectDir := t.TempDir()
	t.Setenv(ActiveOperatorIDEnv, "adrian")
	t.Setenv(LLMTraceEnabledEnv, "1")

	program := writeAuthoringQuizProviderScript(t, projectDir, `#!/bin/sh
cat >/dev/null
printf '%s\n' 'not-json'
`)
	request := validAuthoringQuizGenerationRequest(projectDir)

	_, err := (CodexAuthoringQuizQuestionProvider{Program: program}).GenerateAuthoringQuizQuestions(context.Background(), request)
	if err == nil {
		t.Fatal("GenerateAuthoringQuizQuestions returned nil error")
	}

	events := readLLMTraceEvents(t, projectDir)
	if len(events) != 2 {
		t.Fatalf("trace event count = %d, want 2: %#v", len(events), events)
	}

	if events[1]["status"] != "parse_error" {
		t.Fatalf("response status = %#v, want parse_error", events[1]["status"])
	}

	if events[1]["error"] == "" {
		t.Fatalf("missing parse error in response event: %#v", events[1])
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

func validAuthoringQuizGenerationRequest(projectDir string) AuthoringQuizGenerationRequest {
	now := time.Date(2026, 5, 22, 9, 0, 0, 0, time.UTC)

	return AuthoringQuizGenerationRequest{
		ID:            "authoring-quiz-request-1",
		Action:        AuthoringQuizActionGeneration,
		Strategy:      AuthoringQuizGenerationStrategyDeep,
		ProjectDir:    projectDir,
		SessionID:     "author-20260522-090000",
		Target:        AuthoringTargetSpec,
		Topic:         "Billing sync behavior",
		QuestionCount: 1,
		ContextSections: []AuthoringQuizContextSection{
			{Name: "project", Layer: "project", Authority: "project_metadata", Content: "Project context."},
		},
		OutputContract: AuthoringQuizOutputContract{
			RequiredFields: []string{"id", "question", "tags", "follow_up_depth", "completion_recommended"},
		},
		CreatedAt: now,
	}
}

func writeAuthoringQuizProviderScript(t *testing.T, projectDir, content string) string {
	t.Helper()

	path := filepath.Join(projectDir, "provider.sh")

	err := os.WriteFile(path, []byte(content), 0o755)
	if err != nil {
		t.Fatalf("WriteFile returned error: %v", err)
	}

	return path
}

func assertCodexTraceEventsWithoutPayloads(t *testing.T, events []map[string]any) {
	t.Helper()

	if len(events) != 2 {
		t.Fatalf("trace event count = %d, want 2: %#v", len(events), events)
	}

	assertTraceField(t, events[0], "event", "llm_request")
	assertTraceField(t, events[1], "event", "llm_response")
	assertTraceField(t, events[0], "provider", "codex")
	assertTraceField(t, events[1], "provider", "codex")
	assertTraceField(t, events[1], "status", "ready")
	assertMatchingTraceCallID(t, events)
	assertTraceOmitsField(t, events[0], "request_payload")
	assertTraceOmitsField(t, events[1], "response_payload")
}

func assertTraceField(t *testing.T, event map[string]any, key string, want any) {
	t.Helper()

	if event[key] != want {
		t.Fatalf("trace %s = %#v, want %#v in %#v", key, event[key], want, event)
	}
}

func assertTraceOmitsField(t *testing.T, event map[string]any, key string) {
	t.Helper()

	if _, ok := event[key]; ok {
		t.Fatalf("trace event captured %s: %#v", key, event)
	}
}

func assertMatchingTraceCallID(t *testing.T, events []map[string]any) {
	t.Helper()

	if events[0]["llm_call_id"] != events[1]["llm_call_id"] {
		t.Fatalf("trace call ids differ: %#v / %#v", events[0]["llm_call_id"], events[1]["llm_call_id"])
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
