package workspace

import (
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestLoadPreviousAttemptContextLoadsValidationAndResult(t *testing.T) {
	dir := t.TempDir()
	runDir := filepath.Join(dir, ".chc", "user", "runtime", "runs", "CHC-1", "exec-1")
	validationPath := filepath.Join(runDir, "validation.log")
	resultPath := filepath.Join(runDir, "result.json")

	writeTestFile(t, validationPath, "go test ./...\nFAIL package/store\nexpected complete task\ngot incomplete task\nFAIL package/store\n")
	writeTestFile(t, resultPath, `{
		"execution_id": "exec-1",
		"ticket_key": "CHC-1",
		"status": "failed",
		"summary": "tests still fail",
		"changed_files": [
			{"path": "internal/todo/store.go", "summary": "implemented completion"}
		]
	}`)

	ctx, err := LoadPreviousAttemptContext(PreviousAttemptContextOptions{
		Execution:                  testPreviousAttemptExecutionRecord(validationPath, resultPath),
		PreviousAttemptContextPath: filepath.Join(runDir, "context", "previous-attempt.md"),
		MaxValidationLines:         3,
	})
	if err != nil {
		t.Fatalf("LoadPreviousAttemptContext returned error: %v", err)
	}

	if ctx.Result == nil {
		t.Fatal("Result was nil")
	}

	if len(ctx.ChangedFiles) != 1 {
		t.Fatalf("ChangedFiles = %d, want 1", len(ctx.ChangedFiles))
	}

	if len(ctx.ValidationExcerpt) != 3 {
		t.Fatalf("ValidationExcerpt = %d, want 3", len(ctx.ValidationExcerpt))
	}

	if !containsString(ctx.RepeatedFailureSignatures, "FAIL package/store") {
		t.Fatalf("RepeatedFailureSignatures = %#v, want repeated FAIL", ctx.RepeatedFailureSignatures)
	}

	rendered := RenderPreviousAttemptContext(ctx)
	for _, want := range []string{
		"Previous Attempt Context",
		"tests still fail",
		"internal/todo/store.go",
		"Implicit conversation history is not included",
	} {
		if !strings.Contains(rendered, want) {
			t.Fatalf("rendered previous attempt missing %q:\n%s", want, rendered)
		}
	}
}

func TestLoadPreviousAttemptContextDegradesWhenArtifactsAreMissing(t *testing.T) {
	dir := t.TempDir()
	record := testPreviousAttemptExecutionRecord(
		filepath.Join(dir, "missing-validation.log"),
		filepath.Join(dir, "missing-result.json"),
	)

	ctx, err := LoadPreviousAttemptContext(PreviousAttemptContextOptions{Execution: record})
	if err != nil {
		t.Fatalf("LoadPreviousAttemptContext returned error: %v", err)
	}

	if ctx.Result != nil {
		t.Fatal("missing result loaded a result")
	}

	for _, source := range ctx.Sources {
		if source.Included {
			t.Fatalf("source %s was marked included", source.ID)
		}
	}
}

func TestRenderPreviousAttemptContextLink(t *testing.T) {
	rendered := RenderPreviousAttemptContextLink(".chc/user/runtime/runs/CHC-1/exec-2/context/previous-attempt.md")

	if !strings.Contains(rendered, "Previous Attempt") {
		t.Fatalf("link missing heading:\n%s", rendered)
	}

	if !strings.Contains(rendered, "previous-attempt.md") {
		t.Fatalf("link missing path:\n%s", rendered)
	}
}

func testPreviousAttemptExecutionRecord(validationPath, resultPath string) ExecutionRecord {
	now := time.Date(2026, 5, 20, 12, 0, 0, 0, time.UTC)

	return ExecutionRecord{
		ID:            "exec-1",
		ProjectID:     "project-1",
		TicketKey:     "CHC-1",
		TicketPath:    "/tmp/project/.chc/user/tickets/CHC-1.md",
		Status:        ExecutionStatusFailed,
		FailureReason: string(ExecutionFailureValidation) + ": make verify failed",
		Runtime: ExecutionRuntimeRef{
			ValidationCommand: "make verify",
		},
		Artifacts: ExecutionArtifacts{
			ResultPath:     resultPath,
			ValidationPath: validationPath,
		},
		CreatedAt: now,
		UpdatedAt: now,
	}
}
