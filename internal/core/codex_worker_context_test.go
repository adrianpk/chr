package core

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestCodexWorkerContextMaterializerWritesContext(t *testing.T) {
	path := filepath.Join(t.TempDir(), ".chc", "user", "runtime", "runs", "CHC-1", "exec-1", "context.md")
	opts := testCodexWorkerContextOptions(path)

	err := NewCodexWorkerContextMaterializer().MaterializeContext(context.Background(), opts)
	if err != nil {
		t.Fatalf("MaterializeContext returned error: %v", err)
	}

	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read context: %v", err)
	}

	for _, want := range []string{
		"# Codex Worker Context",
		"- execution_id: exec-1",
		"- ticket_key: CHC-1",
		"- ticket_path: /workspace/.chc/user/tickets/CHC-1.md",
		"- result_path: /workspace/.chc/user/runtime/runs/CHC-1/exec-1/result.json",
		"- log_path: /workspace/.chc/user/runtime/runs/CHC-1/exec-1/runtime.log",
		"- validation_path: /workspace/.chc/user/runtime/runs/CHC-1/exec-1/validation.log",
		"## Evidence Strategy",
		"Before modifying project files:",
		"Write or update the ticket evidence-strategy checklist.",
		"Before finishing:",
		"Update Validation and Evidence summary.",
		"- `status`: `succeeded` or `failed`.",
		"- `changed_files`: optional list of workspace-relative changed file summaries.",
	} {
		if !strings.Contains(string(content), want) {
			t.Fatalf("content does not contain %q:\n%s", want, string(content))
		}
	}
}

func TestMaterializeCodexWorkerContextForExecutionUsesContainerPaths(t *testing.T) {
	projectDir := t.TempDir()
	runDir := filepath.Join(projectDir, ".chc", "user", "runtime", "runs", "CHC-1", "exec-1")
	plan := RuntimePlan{
		ProjectDir:        projectDir,
		ContainerWorkdir:  "/workspace",
		WorkerContextPath: filepath.Join(runDir, "context.md"),
		ResultPath:        filepath.Join(runDir, "result.json"),
		LogPath:           filepath.Join(runDir, "runtime.log"),
		ValidationPath:    filepath.Join(runDir, "validation.log"),
		ValidationCommand: "go test ./...",
		DockerfilePath:    filepath.Join(projectDir, ".chc", "user", "runtime", "agent", "Dockerfile"),
		RunArtifactDir:    runDir,
		ArtifactDir:       filepath.Join(projectDir, ".chc", "user", "runtime", "runs"),
		Image:             "chc-agent-runtime:local",
	}
	record := ExecutionRecord{
		ID:         "exec-1",
		TicketKey:  "CHC-1",
		TicketPath: filepath.Join(projectDir, ".chc", "user", "tickets", "CHC-1.md"),
	}
	writeCodexWorkerContextTicket(t, record.TicketPath, Ticket{
		Key:       "CHC-1",
		Title:     "Ready work",
		Status:    TicketStatusReadyForDev,
		Priority:  "normal",
		Severity:  "normal",
		Type:      "feature",
		CreatedAt: testCodexWorkerContextTime(),
		UpdatedAt: testCodexWorkerContextTime(),
	})

	err := MaterializeCodexWorkerContextForExecution(context.Background(), record, plan)
	if err != nil {
		t.Fatalf("MaterializeCodexWorkerContextForExecution returned error: %v", err)
	}

	content, err := os.ReadFile(plan.WorkerContextPath)
	if err != nil {
		t.Fatalf("read context: %v", err)
	}

	for _, want := range []string{
		"- ticket_path: /workspace/.chc/user/tickets/CHC-1.md",
		"- result_path: /workspace/.chc/user/runtime/runs/CHC-1/exec-1/result.json",
		"- log_path: /workspace/.chc/user/runtime/runs/CHC-1/exec-1/runtime.log",
		"- validation_path: /workspace/.chc/user/runtime/runs/CHC-1/exec-1/validation.log",
	} {
		if !strings.Contains(string(content), want) {
			t.Fatalf("content does not contain %q:\n%s", want, string(content))
		}
	}
}

func TestMaterializeCodexWorkerContextForExecutionIncludesRepairHandoff(t *testing.T) {
	projectDir := t.TempDir()
	runDir := filepath.Join(projectDir, ".chc", "user", "runtime", "runs", "CHC-1", "exec-1")
	plan := testCodexWorkerContextRuntimePlan(projectDir, runDir)
	record := ExecutionRecord{
		ID:         "exec-1",
		TicketKey:  "CHC-1",
		TicketPath: filepath.Join(projectDir, ".chc", "user", "tickets", "CHC-1.md"),
	}
	writeCodexWorkerContextTicket(t, record.TicketPath, Ticket{
		Key:               "CHC-1",
		Title:             "Repair work",
		Status:            TicketStatusReadyForDev,
		Priority:          "normal",
		Severity:          "normal",
		Type:              "feature",
		Branch:            "feat/slice-8-repair-handoff-docs",
		ReviewState:       "changes_requested",
		ReviewRound:       2,
		ReviewResult:      "needs_changes",
		ReviewCommentsURL: "https://forge.example.com/adrianpk/chc/pulls/279",
		ReviewRefs: []TicketReviewRef{{
			Type:     "pull_request",
			Provider: "forgejo",
			URL:      "https://forge.example.com/adrianpk/chc/pulls/279",
			ID:       "279",
			Status:   "open",
		}},
		CreatedAt: testCodexWorkerContextTime(),
		UpdatedAt: testCodexWorkerContextTime(),
	})

	err := MaterializeCodexWorkerContextForExecution(context.Background(), record, plan)
	if err != nil {
		t.Fatalf("MaterializeCodexWorkerContextForExecution returned error: %v", err)
	}

	content, err := os.ReadFile(plan.WorkerContextPath)
	if err != nil {
		t.Fatalf("read context: %v", err)
	}

	for _, want := range []string{
		"## Review Repair Handoff",
		"- branch: feat/slice-8-repair-handoff-docs",
		"- pull_request: https://forge.example.com/adrianpk/chc/pulls/279",
		"- review_comments_url: https://forge.example.com/adrianpk/chc/pulls/279",
		"Use the remote review comments as the source of detailed findings.",
	} {
		if !strings.Contains(string(content), want) {
			t.Fatalf("content does not contain %q:\n%s", want, string(content))
		}
	}
}

func TestMaterializeCodexWorkerContextForExecutionIncludesEvidenceStrategy(t *testing.T) {
	projectDir := t.TempDir()
	runDir := filepath.Join(projectDir, ".chc", "user", "runtime", "runs", "CHC-1", "exec-1")
	plan := testCodexWorkerContextRuntimePlan(projectDir, runDir)
	record := ExecutionRecord{
		ID:         "exec-1",
		TicketKey:  "CHC-1",
		TicketPath: filepath.Join(projectDir, ".chc", "user", "tickets", "CHC-1.md"),
	}

	checklist, err := RenderEvidenceStrategyChecklist(testEvidenceStrategyChecklistModel())
	if err != nil {
		t.Fatalf("RenderEvidenceStrategyChecklist returned error: %v", err)
	}

	writeCodexWorkerContextTicket(t, record.TicketPath, Ticket{
		Key:       "CHC-1",
		Title:     "Behavior work",
		Status:    TicketStatusReadyForDev,
		Priority:  "normal",
		Severity:  "normal",
		Type:      "feature",
		Body:      "# Description\n\nImplement behavior.\n\n" + checklist,
		CreatedAt: testCodexWorkerContextTime(),
		UpdatedAt: testCodexWorkerContextTime(),
	})

	err = MaterializeCodexWorkerContextForExecution(context.Background(), record, plan)
	if err != nil {
		t.Fatalf("MaterializeCodexWorkerContextForExecution returned error: %v", err)
	}

	content, err := os.ReadFile(plan.WorkerContextPath)
	if err != nil {
		t.Fatalf("read context: %v", err)
	}

	for _, want := range []string{
		"### Effective Config",
		"- enabled: true",
		"- mode: inline",
		"- ticket_checklist: required",
		"- gherkin_policy: active_when_applicable",
		"### Existing Checklist",
		"- status: present",
		"- Evidence strategy: checked=false; status=planned; change_kind=bugfix",
		"- Gherkin: checked=false; status=required; path=.chc/user/acceptance/bug.feature",
	} {
		if !strings.Contains(string(content), want) {
			t.Fatalf("content does not contain %q:\n%s", want, string(content))
		}
	}
}

func TestMaterializeCodexWorkerContextForExecutionBlocksRepairWithoutReviewComments(t *testing.T) {
	projectDir := t.TempDir()
	runDir := filepath.Join(projectDir, ".chc", "user", "runtime", "runs", "CHC-1", "exec-1")
	plan := testCodexWorkerContextRuntimePlan(projectDir, runDir)
	record := ExecutionRecord{
		ID:         "exec-1",
		TicketKey:  "CHC-1",
		TicketPath: filepath.Join(projectDir, ".chc", "user", "tickets", "CHC-1.md"),
	}
	writeCodexWorkerContextTicket(t, record.TicketPath, Ticket{
		Key:          "CHC-1",
		Title:        "Repair work",
		Status:       TicketStatusReadyForDev,
		Priority:     "normal",
		Severity:     "normal",
		Type:         "feature",
		Branch:       "feat/slice-8-repair-handoff-docs",
		ReviewState:  "changes_requested",
		ReviewResult: "needs_changes",
		ReviewRefs: []TicketReviewRef{{
			Type:     "pull_request",
			Provider: "forgejo",
			URL:      "https://forge.example.com/adrianpk/chc/pulls/279",
			ID:       "279",
			Status:   "open",
		}},
		CreatedAt: testCodexWorkerContextTime(),
		UpdatedAt: testCodexWorkerContextTime(),
	})

	err := MaterializeCodexWorkerContextForExecution(context.Background(), record, plan)
	if err == nil || !strings.Contains(err.Error(), "repair handoff requires review_comments_url") {
		t.Fatalf("MaterializeCodexWorkerContextForExecution error = %v, want review_comments_url blocker", err)
	}
}

func TestRenderCodexWorkerContextIsDeterministic(t *testing.T) {
	opts := testCodexWorkerContextOptions("/tmp/context.md")

	first := RenderCodexWorkerContext(opts)
	second := RenderCodexWorkerContext(opts)

	if first != second {
		t.Fatalf("context render is not deterministic:\nfirst:\n%s\nsecond:\n%s", first, second)
	}
}

func TestCodexWorkerContextMaterializerRejectsMissingRequiredFields(t *testing.T) {
	opts := testCodexWorkerContextOptions(filepath.Join(t.TempDir(), "context.md"))
	opts.TicketPath = ""

	err := NewCodexWorkerContextMaterializer().MaterializeContext(context.Background(), opts)
	if err == nil {
		t.Fatal("MaterializeContext returned nil error")
	}
}

func testCodexWorkerContextOptions(path string) MaterializeCodexWorkerContextOptions {
	return MaterializeCodexWorkerContextOptions{
		Path:           path,
		ExecutionID:    "exec-1",
		TicketKey:      "CHC-1",
		TicketPath:     "/workspace/.chc/user/tickets/CHC-1.md",
		ResultPath:     "/workspace/.chc/user/runtime/runs/CHC-1/exec-1/result.json",
		LogPath:        "/workspace/.chc/user/runtime/runs/CHC-1/exec-1/runtime.log",
		ValidationPath: "/workspace/.chc/user/runtime/runs/CHC-1/exec-1/validation.log",
	}
}

func testCodexWorkerContextRuntimePlan(projectDir, runDir string) RuntimePlan {
	return RuntimePlan{
		ProjectDir:        projectDir,
		ContainerWorkdir:  "/workspace",
		WorkerContextPath: filepath.Join(runDir, "context.md"),
		ResultPath:        filepath.Join(runDir, "result.json"),
		LogPath:           filepath.Join(runDir, "runtime.log"),
		ValidationPath:    filepath.Join(runDir, "validation.log"),
		ValidationCommand: "go test ./...",
		DockerfilePath:    filepath.Join(projectDir, ".chc", "user", "runtime", "agent", "Dockerfile"),
		RunArtifactDir:    runDir,
		ArtifactDir:       filepath.Join(projectDir, ".chc", "user", "runtime", "runs"),
		Image:             "chc-agent-runtime:local",
	}
}

func writeCodexWorkerContextTicket(t *testing.T, path string, ticket Ticket) {
	t.Helper()

	content, err := NewTicketRenderer().RenderTicket(ticket)
	if err != nil {
		t.Fatalf("RenderTicket returned error: %v", err)
	}

	err = os.MkdirAll(filepath.Dir(path), 0o755)
	if err != nil {
		t.Fatalf("create ticket dir: %v", err)
	}

	err = os.WriteFile(path, []byte(content), 0o644)
	if err != nil {
		t.Fatalf("write ticket: %v", err)
	}
}

func testCodexWorkerContextTime() time.Time {
	return time.Date(2026, 6, 1, 12, 0, 0, 0, time.UTC)
}
