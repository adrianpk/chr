package core

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	executioncore "forge.adrianpk.com/chr/internal/core/execution"
	harnessruntime "forge.adrianpk.com/chr/internal/core/harness/runtime"
	ticketcore "forge.adrianpk.com/chr/internal/core/ticket"
	ticketmetadata "forge.adrianpk.com/chr/internal/core/ticket/metadata"
	workspacecore "forge.adrianpk.com/chr/internal/core/workspace"
)

func TestWorkerContextMaterializerWritesContext(t *testing.T) {
	path := filepath.Join(t.TempDir(), ".chr", "default", "runtime", "runs", "CHR-1", "exec-1", "context.md")
	opts := testWorkerContextOptions(path)

	err := NewWorkerContextMaterializer().MaterializeContext(context.Background(), opts)
	if err != nil {
		t.Fatalf("MaterializeContext returned error: %v", err)
	}

	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read context: %v", err)
	}

	for _, want := range []string{
		"# Worker Context",
		"- execution_id: exec-1",
		"- ticket_key: CHR-1",
		"- ticket_path: /workspace/.chr/default/tickets/CHR-1.md",
		"## Workspace",
		"- workspace_mode: project",
		"- project_checkout_dir: /project",
		"- execution_workspace_dir: /project",
		"- ticket_branch: chr/CHR-1",
		"- workspace_base_ref: origin/dev",
		"- workspace_base_sha: abc123",
		"- result_path: /workspace/.chr/default/runtime/runs/CHR-1/exec-1/result.json",
		"- log_path: /workspace/.chr/default/runtime/runs/CHR-1/exec-1/runtime.log",
		"- validation_path: /workspace/.chr/default/runtime/runs/CHR-1/exec-1/validation.log",
		"## Evidence Strategy",
		"Before modifying project files:",
		"Write or update the ticket evidence-strategy checklist.",
		"Before finishing:",
		"Update Validation and Evidence summary.",
		"- `status`: `succeeded` or `failed`.",
		"- `evidence_summary`: required object with `strategy`, `produced`, and `validation`.",
		"- `changed_files`: optional list of workspace-relative changed file summaries.",
	} {
		if !strings.Contains(string(content), want) {
			t.Fatalf("content does not contain %q:\n%s", want, string(content))
		}
	}
}

func TestMaterializeWorkerContextForExecutionUsesContainerPaths(t *testing.T) {
	projectDir := t.TempDir()
	runDir := filepath.Join(projectDir, ".chr", "default", "runtime", "runs", "CHR-1", "exec-1")
	plan := harnessruntime.Plan{
		ProjectDir:            projectDir,
		ContainerWorkdir:      "/workspace",
		WorkerContextPath:     filepath.Join(runDir, "context.md"),
		ResultPath:            filepath.Join(runDir, "result.json"),
		LogPath:               filepath.Join(runDir, "runtime.log"),
		ValidationPath:        filepath.Join(runDir, "validation.log"),
		ValidationCommand:     "go test ./...",
		DockerfilePath:        filepath.Join(projectDir, ".chr", "default", "runtime", "agent", "Dockerfile"),
		RunArtifactDir:        runDir,
		ArtifactDir:           filepath.Join(projectDir, ".chr", "default", "runtime", "runs"),
		Image:                 "chr-agent-runtime:local",
		ProjectCheckoutDir:    projectDir,
		ExecutionWorkspaceDir: projectDir,
		WorkspaceMode:         workspacecore.ExecutionProject,
		TicketBranch:          "chr/CHR-1",
		WorkspaceBaseRef:      "origin/dev",
		WorkspaceBaseSHA:      "abc123",
	}
	record := executioncore.ExecutionRecord{
		ID:         "exec-1",
		TicketKey:  "CHR-1",
		TicketPath: filepath.Join(projectDir, ".chr", "default", "tickets", "CHR-1.md"),
	}
	writeWorkerContextTicket(t, record.TicketPath, ticketcore.Ticket{
		Key:       "CHR-1",
		Title:     "Ready work",
		Status:    ticketcore.StatusReadyForDev,
		Priority:  "normal",
		Severity:  "normal",
		Type:      "feature",
		CreatedAt: testWorkerContextTime(),
		UpdatedAt: testWorkerContextTime(),
	})

	err := MaterializeWorkerContextForExecution(context.Background(), record, plan)
	if err != nil {
		t.Fatalf("MaterializeWorkerContextForExecution returned error: %v", err)
	}

	content, err := os.ReadFile(plan.WorkerContextPath)
	if err != nil {
		t.Fatalf("read context: %v", err)
	}

	for _, want := range []string{
		"- ticket_path: /workspace/.chr/default/tickets/CHR-1.md",
		"- result_path: /workspace/.chr/default/runtime/runs/CHR-1/exec-1/result.json",
		"- log_path: /workspace/.chr/default/runtime/runs/CHR-1/exec-1/runtime.log",
		"- validation_path: /workspace/.chr/default/runtime/runs/CHR-1/exec-1/validation.log",
	} {
		if !strings.Contains(string(content), want) {
			t.Fatalf("content does not contain %q:\n%s", want, string(content))
		}
	}
}

func TestMaterializeWorkerContextForExecutionUsesWorktreeContainerPaths(t *testing.T) {
	projectDir := t.TempDir()
	executionWorkspaceDir := filepath.Join(projectDir, ".chr", "default", "worktrees", "CHR-1")
	plan := testWorkerContextWorktreeRuntimePlan(projectDir, executionWorkspaceDir)
	record := testWorkerContextWorktreeRecord(t, executionWorkspaceDir)

	err := MaterializeWorkerContextForExecution(context.Background(), record, plan)
	if err != nil {
		t.Fatalf("MaterializeWorkerContextForExecution returned error: %v", err)
	}

	content, err := os.ReadFile(plan.WorkerContextPath)
	if err != nil {
		t.Fatalf("read context: %v", err)
	}

	for _, want := range []string{
		"- ticket_path: /workspace/.chr/default/tickets/CHR-1.md",
		"- workspace_mode: worktree",
		"- project_checkout_dir: " + projectDir,
		"- execution_workspace_dir: " + executionWorkspaceDir,
		"- ticket_branch: chr/CHR-1",
		"- workspace_base_ref: origin/dev",
		"- workspace_base_sha: abc123",
		"- result_path: /workspace/.chr/default/runtime/runs/CHR-1/exec-1/result.json",
		"- log_path: /workspace/.chr/default/runtime/runs/CHR-1/exec-1/runtime.log",
		"- validation_path: /workspace/.chr/default/runtime/runs/CHR-1/exec-1/validation.log",
	} {
		if !strings.Contains(string(content), want) {
			t.Fatalf("content does not contain %q:\n%s", want, string(content))
		}
	}
}

func TestMaterializeWorkerContextForExecutionIncludesRepairHandoff(t *testing.T) {
	projectDir := t.TempDir()
	runDir := filepath.Join(projectDir, ".chr", "default", "runtime", "runs", "CHR-1", "exec-1")
	plan := testWorkerContextRuntimePlan(projectDir, runDir)
	record := executioncore.ExecutionRecord{
		ID:         "exec-1",
		TicketKey:  "CHR-1",
		TicketPath: filepath.Join(projectDir, ".chr", "default", "tickets", "CHR-1.md"),
	}
	writeWorkerContextTicket(t, record.TicketPath, ticketcore.Ticket{
		Key:               "CHR-1",
		Title:             "Repair work",
		Status:            ticketcore.StatusReadyForDev,
		Priority:          "normal",
		Severity:          "normal",
		Type:              "feature",
		Branch:            "feat/slice-8-repair-handoff-docs",
		ReviewState:       "changes_requested",
		ReviewRound:       2,
		ReviewResult:      "needs_changes",
		ReviewCommentsURL: "https://forge.example.com/adrianpk/chr/pulls/279",
		BlockingFindings:  []string{"correctness: Missing invalid input handling"},
		ReviewRefs: []ticketmetadata.ReviewRef{{
			Type:     "pull_request",
			Provider: "forgejo",
			URL:      "https://forge.example.com/adrianpk/chr/pulls/279",
			ID:       "279",
			Status:   "open",
		}},
		CreatedAt: testWorkerContextTime(),
		UpdatedAt: testWorkerContextTime(),
	})

	err := MaterializeWorkerContextForExecution(context.Background(), record, plan)
	if err != nil {
		t.Fatalf("MaterializeWorkerContextForExecution returned error: %v", err)
	}

	content, err := os.ReadFile(plan.WorkerContextPath)
	if err != nil {
		t.Fatalf("read context: %v", err)
	}

	for _, want := range []string{
		"## Review Repair Handoff",
		"- branch: feat/slice-8-repair-handoff-docs",
		"- pull_request: https://forge.example.com/adrianpk/chr/pulls/279",
		"- review_comments_url: https://forge.example.com/adrianpk/chr/pulls/279",
		"- blocking_findings:",
		"  - correctness: Missing invalid input handling",
		"Use the local blocking findings and remote review comments as the source of repair work.",
	} {
		if !strings.Contains(string(content), want) {
			t.Fatalf("content does not contain %q:\n%s", want, string(content))
		}
	}
}

func TestMaterializeWorkerContextForExecutionIncludesEvidenceStrategy(t *testing.T) {
	projectDir := t.TempDir()
	runDir := filepath.Join(projectDir, ".chr", "default", "runtime", "runs", "CHR-1", "exec-1")
	plan := testWorkerContextRuntimePlan(projectDir, runDir)
	record := executioncore.ExecutionRecord{
		ID:         "exec-1",
		TicketKey:  "CHR-1",
		TicketPath: filepath.Join(projectDir, ".chr", "default", "tickets", "CHR-1.md"),
	}

	checklist, err := RenderEvidenceStrategyChecklist(testEvidenceStrategyChecklistModel())
	if err != nil {
		t.Fatalf("RenderEvidenceStrategyChecklist returned error: %v", err)
	}

	writeWorkerContextTicket(t, record.TicketPath, ticketcore.Ticket{
		Key:       "CHR-1",
		Title:     "Behavior work",
		Status:    ticketcore.StatusReadyForDev,
		Priority:  "normal",
		Severity:  "normal",
		Type:      "feature",
		Body:      "# Description\n\nImplement behavior.\n\n" + checklist,
		CreatedAt: testWorkerContextTime(),
		UpdatedAt: testWorkerContextTime(),
	})

	err = MaterializeWorkerContextForExecution(context.Background(), record, plan)
	if err != nil {
		t.Fatalf("MaterializeWorkerContextForExecution returned error: %v", err)
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
		"- Gherkin: checked=false; status=required; path=.chr/default/acceptance/bug.feature",
	} {
		if !strings.Contains(string(content), want) {
			t.Fatalf("content does not contain %q:\n%s", want, string(content))
		}
	}
}

func TestMaterializeWorkerContextForExecutionIncludesGherkinPolicyGuidance(t *testing.T) {
	projectDir := t.TempDir()
	runDir := filepath.Join(projectDir, ".chr", "default", "runtime", "runs", "CHR-1", "exec-1")
	plan := testWorkerContextRuntimePlan(projectDir, runDir)
	record := executioncore.ExecutionRecord{
		ID:         "exec-1",
		TicketKey:  "CHR-1",
		TicketPath: filepath.Join(projectDir, ".chr", "default", "tickets", "CHR-1.md"),
	}
	writeEvidencePolicyFeature(t, projectDir, ".chr/default/acceptance/checkout.feature")
	writeWorkerContextTicket(t, record.TicketPath, ticketcore.Ticket{
		Key:         "CHR-1",
		Title:       "Add CLI checkout command",
		Status:      ticketcore.StatusReadyForDev,
		Priority:    "normal",
		Severity:    "normal",
		Type:        "feature",
		UseCase:     "checkout",
		ProjectArea: "cli",
		CreatedAt:   testWorkerContextTime(),
		UpdatedAt:   testWorkerContextTime(),
	})

	err := MaterializeWorkerContextForExecution(context.Background(), record, plan)
	if err != nil {
		t.Fatalf("MaterializeWorkerContextForExecution returned error: %v", err)
	}

	content, err := os.ReadFile(plan.WorkerContextPath)
	if err != nil {
		t.Fatalf("read context: %v", err)
	}

	for _, want := range []string{
		"### Project Gherkin",
		"- active: true",
		"- applies_to_ticket: true",
		"- required_path: .chr/default/acceptance/checkout.feature",
		"### Gherkin Evidence Quality",
		"Do not treat an existing feature path as sufficient evidence by itself.",
		"trivial hardcoded happy path",
		"### Suggested Strategy",
		"- change_kind: cli_behavior",
		"- gherkin: required",
		"### Evidence Decision Matrix",
		"- parsing_or_validation: Prefer table-driven tests and fuzz/property-style checks when the input space warrants it.",
	} {
		if !strings.Contains(string(content), want) {
			t.Fatalf("content does not contain %q:\n%s", want, string(content))
		}
	}
}

func TestRenderWorkerContextIncludesDocsOnlyEvidenceStrategy(t *testing.T) {
	checklist := EvidenceStrategyChecklist{
		Found: true,
		Items: map[string]EvidenceStrategyChecklistItem{
			EvidenceStrategyChecklistLabelStrategy: {
				Label:  EvidenceStrategyChecklistLabelStrategy,
				Status: "planned",
				Detail: "change_kind=docs_only; risk=low; test_timing=none",
			},
			EvidenceStrategyChecklistLabelGherkin: {
				Label:  EvidenceStrategyChecklistLabelGherkin,
				Status: "not_applicable",
				Detail: "reason=documentation-only change",
			},
			EvidenceStrategyChecklistLabelTests: {
				Label:  EvidenceStrategyChecklistLabelTests,
				Status: "none",
				Detail: "reason=no runtime behavior",
			},
			EvidenceStrategyChecklistLabelValidation: {
				Label:  EvidenceStrategyChecklistLabelValidation,
				Status: "pending",
				Detail: "commands=docs review",
			},
			EvidenceStrategyChecklistLabelSummary: {
				Label:  EvidenceStrategyChecklistLabelSummary,
				Status: "pending",
			},
		},
	}

	opts := testWorkerContextOptions("/tmp/context.md")
	opts.Evidence = &WorkerEvidenceContext{
		Config:    DefaultEvidenceStrategyConfig(),
		Checklist: &checklist,
	}

	content := RenderWorkerContext(opts)

	for _, want := range []string{
		"- Evidence strategy: checked=false; status=planned; change_kind=docs_only",
		"- Gherkin: checked=false; status=not_applicable; reason=documentation-only change",
		"- Tests: checked=false; status=none; reason=no runtime behavior",
		"- `evidence_summary`: required object with `strategy`, `produced`, and `validation`.",
	} {
		if !strings.Contains(content, want) {
			t.Fatalf("content does not contain %q:\n%s", want, content)
		}
	}
}

func TestRenderWorkerContextIncludesBehaviorChangeEvidenceStrategy(t *testing.T) {
	checklist := EvidenceStrategyChecklist{
		Found: true,
		Items: map[string]EvidenceStrategyChecklistItem{
			EvidenceStrategyChecklistLabelStrategy: {
				Label:  EvidenceStrategyChecklistLabelStrategy,
				Status: "planned",
				Detail: "change_kind=behavior; risk=high; test_timing=before",
			},
			EvidenceStrategyChecklistLabelGherkin: {
				Label:  EvidenceStrategyChecklistLabelGherkin,
				Status: "required",
				Detail: "path=.chr/default/acceptance/checkout.feature",
			},
			EvidenceStrategyChecklistLabelTests: {
				Label:  EvidenceStrategyChecklistLabelTests,
				Status: "mixed",
				Detail: "reason=unit table and fuzz coverage",
			},
			EvidenceStrategyChecklistLabelValidation: {
				Label:  EvidenceStrategyChecklistLabelValidation,
				Status: "pending",
				Detail: "commands=make verify",
			},
			EvidenceStrategyChecklistLabelSummary: {
				Label:  EvidenceStrategyChecklistLabelSummary,
				Status: "pending",
			},
		},
	}

	opts := testWorkerContextOptions("/tmp/context.md")
	opts.Evidence = &WorkerEvidenceContext{
		Config:    DefaultEvidenceStrategyConfig(),
		Checklist: &checklist,
	}

	content := RenderWorkerContext(opts)

	for _, want := range []string{
		"- Evidence strategy: checked=false; status=planned; change_kind=behavior",
		"- Gherkin: checked=false; status=required; path=.chr/default/acceptance/checkout.feature",
		"- Tests: checked=false; status=mixed; reason=unit table and fuzz coverage",
		"- `evidence_summary.deviations`: optional list of deviations from the planned strategy.",
	} {
		if !strings.Contains(content, want) {
			t.Fatalf("content does not contain %q:\n%s", want, content)
		}
	}
}

func TestMaterializeWorkerContextForExecutionBlocksRepairWithoutReviewComments(t *testing.T) {
	projectDir := t.TempDir()
	runDir := filepath.Join(projectDir, ".chr", "default", "runtime", "runs", "CHR-1", "exec-1")
	plan := testWorkerContextRuntimePlan(projectDir, runDir)
	record := executioncore.ExecutionRecord{
		ID:         "exec-1",
		TicketKey:  "CHR-1",
		TicketPath: filepath.Join(projectDir, ".chr", "default", "tickets", "CHR-1.md"),
	}
	writeWorkerContextTicket(t, record.TicketPath, ticketcore.Ticket{
		Key:          "CHR-1",
		Title:        "Repair work",
		Status:       ticketcore.StatusReadyForDev,
		Priority:     "normal",
		Severity:     "normal",
		Type:         "feature",
		Branch:       "feat/slice-8-repair-handoff-docs",
		ReviewState:  "changes_requested",
		ReviewResult: "needs_changes",
		ReviewRefs: []ticketmetadata.ReviewRef{{
			Type:     "pull_request",
			Provider: "forgejo",
			URL:      "https://forge.example.com/adrianpk/chr/pulls/279",
			ID:       "279",
			Status:   "open",
		}},
		CreatedAt: testWorkerContextTime(),
		UpdatedAt: testWorkerContextTime(),
	})

	err := MaterializeWorkerContextForExecution(context.Background(), record, plan)
	if err == nil || !strings.Contains(err.Error(), "repair handoff requires review_comments_url") {
		t.Fatalf("MaterializeWorkerContextForExecution error = %v, want review_comments_url blocker", err)
	}
}

func TestRenderWorkerContextIsDeterministic(t *testing.T) {
	opts := testWorkerContextOptions("/tmp/context.md")

	first := RenderWorkerContext(opts)
	second := RenderWorkerContext(opts)

	if first != second {
		t.Fatalf("context render is not deterministic:\nfirst:\n%s\nsecond:\n%s", first, second)
	}
}

func TestWorkerContextMaterializerRejectsMissingRequiredFields(t *testing.T) {
	opts := testWorkerContextOptions(filepath.Join(t.TempDir(), "context.md"))
	opts.TicketPath = ""

	err := NewWorkerContextMaterializer().MaterializeContext(context.Background(), opts)
	if err == nil {
		t.Fatal("MaterializeContext returned nil error")
	}
}

func testWorkerContextOptions(path string) MaterializeWorkerContextOptions {
	return MaterializeWorkerContextOptions{
		Path:                  path,
		ExecutionID:           "exec-1",
		TicketKey:             "CHR-1",
		TicketPath:            "/workspace/.chr/default/tickets/CHR-1.md",
		ResultPath:            "/workspace/.chr/default/runtime/runs/CHR-1/exec-1/result.json",
		LogPath:               "/workspace/.chr/default/runtime/runs/CHR-1/exec-1/runtime.log",
		ValidationPath:        "/workspace/.chr/default/runtime/runs/CHR-1/exec-1/validation.log",
		WorkspaceMode:         workspacecore.ExecutionProject,
		ProjectCheckoutDir:    "/project",
		ExecutionWorkspaceDir: "/project",
		TicketBranch:          "chr/CHR-1",
		WorkspaceBaseRef:      "origin/dev",
		WorkspaceBaseSHA:      "abc123",
	}
}

func testWorkerContextRuntimePlan(projectDir, runDir string) harnessruntime.Plan {
	return harnessruntime.Plan{
		ProjectDir:            projectDir,
		ProjectCheckoutDir:    projectDir,
		ExecutionWorkspaceDir: projectDir,
		WorkspaceMode:         workspacecore.ExecutionProject,
		TicketBranch:          "chr/CHR-1",
		WorkspaceBaseRef:      "origin/dev",
		WorkspaceBaseSHA:      "abc123",
		ContainerWorkdir:      "/workspace",
		WorkerContextPath:     filepath.Join(runDir, "context.md"),
		ResultPath:            filepath.Join(runDir, "result.json"),
		LogPath:               filepath.Join(runDir, "runtime.log"),
		ValidationPath:        filepath.Join(runDir, "validation.log"),
		ValidationCommand:     "go test ./...",
		DockerfilePath:        filepath.Join(projectDir, ".chr", "default", "runtime", "agent", "Dockerfile"),
		RunArtifactDir:        runDir,
		ArtifactDir:           filepath.Join(projectDir, ".chr", "default", "runtime", "runs"),
		Image:                 "chr-agent-runtime:local",
	}
}

func testWorkerContextWorktreeRuntimePlan(projectDir, executionWorkspaceDir string) harnessruntime.Plan {
	runDir := filepath.Join(executionWorkspaceDir, ".chr", "default", "runtime", "runs", "CHR-1", "exec-1")

	return harnessruntime.Plan{
		ProjectDir:            projectDir,
		ProjectCheckoutDir:    projectDir,
		ExecutionWorkspaceDir: executionWorkspaceDir,
		WorkspaceMode:         workspacecore.ExecutionWorktree,
		TicketBranch:          "chr/CHR-1",
		WorkspaceBaseRef:      "origin/dev",
		WorkspaceBaseSHA:      "abc123",
		ContainerWorkdir:      "/workspace",
		WorkerContextPath:     filepath.Join(runDir, "context.md"),
		ResultPath:            filepath.Join(runDir, "result.json"),
		LogPath:               filepath.Join(runDir, "runtime.log"),
		ValidationPath:        filepath.Join(runDir, "validation.log"),
		ValidationCommand:     "go test ./...",
		DockerfilePath:        filepath.Join(projectDir, ".chr", "default", "runtime", "agent", "Dockerfile"),
		RunArtifactDir:        runDir,
		ArtifactDir:           filepath.Join(executionWorkspaceDir, ".chr", "default", "runtime", "runs"),
		Image:                 "chr-agent-runtime:local",
	}
}

func testWorkerContextWorktreeRecord(t *testing.T, executionWorkspaceDir string) executioncore.ExecutionRecord {
	t.Helper()

	record := executioncore.ExecutionRecord{
		ID:         "exec-1",
		TicketKey:  "CHR-1",
		TicketPath: filepath.Join(executionWorkspaceDir, ".chr", "default", "tickets", "CHR-1.md"),
	}
	writeWorkerContextTicket(t, record.TicketPath, ticketcore.Ticket{
		Key:       "CHR-1",
		Title:     "Ready work",
		Status:    ticketcore.StatusReadyForDev,
		Priority:  "normal",
		Severity:  "normal",
		Type:      "feature",
		CreatedAt: testWorkerContextTime(),
		UpdatedAt: testWorkerContextTime(),
	})

	return record
}

func writeWorkerContextTicket(t *testing.T, path string, ticket ticketcore.Ticket) {
	t.Helper()

	content, err := ticketcore.NewRenderer().Render(ticket)
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

func testWorkerContextTime() time.Time {
	return time.Date(2026, 6, 1, 12, 0, 0, 0, time.UTC)
}
