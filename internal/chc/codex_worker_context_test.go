package chc

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
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
