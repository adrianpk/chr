package chc

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadExecutionResultParsesEnvelope(t *testing.T) {
	path := writeExecutionResultFile(t, `{
		"execution_id": "exec-1",
		"ticket_key": "CHC-1",
		"status": "succeeded",
		"summary": "done",
		"changed_files": [
			{
				"path": "internal/chc/execution_result.go",
				"summary": "add changed files"
			}
		]
	}`)

	result, err := LoadExecutionResult(path)
	if err != nil {
		t.Fatalf("LoadExecutionResult returned error: %v", err)
	}

	if result.ExecutionID != "exec-1" {
		t.Fatalf("ExecutionID = %q, want exec-1", result.ExecutionID)
	}

	if result.Status != ExecutionResultStatusSucceeded {
		t.Fatalf("Status = %q, want succeeded", result.Status)
	}

	if len(result.ChangedFiles) != 1 {
		t.Fatalf("ChangedFiles len = %d, want 1", len(result.ChangedFiles))
	}

	if result.ChangedFiles[0].Path != "internal/chc/execution_result.go" {
		t.Fatalf("ChangedFiles[0].Path = %q, want result path", result.ChangedFiles[0].Path)
	}
}

func TestLoadExecutionResultRejectsMalformedJSON(t *testing.T) {
	path := writeExecutionResultFile(t, `{`)

	_, err := LoadExecutionResult(path)
	if err == nil {
		t.Fatal("LoadExecutionResult returned nil error")
	}
}

func TestLoadExecutionResultRejectsInvalidEnvelope(t *testing.T) {
	path := writeExecutionResultFile(t, `{
		"execution_id": "exec-1",
		"ticket_key": "CHC-1",
		"status": "unknown"
	}`)

	_, err := LoadExecutionResult(path)
	if err == nil {
		t.Fatal("LoadExecutionResult returned nil error")
	}
}

func TestLoadExecutionResultRejectsInvalidChangedFilePath(t *testing.T) {
	path := writeExecutionResultFile(t, `{
		"execution_id": "exec-1",
		"ticket_key": "CHC-1",
		"status": "succeeded",
		"changed_files": [
			{
				"path": "../outside.go"
			}
		]
	}`)

	_, err := LoadExecutionResult(path)
	if err == nil {
		t.Fatal("LoadExecutionResult returned nil error")
	}
}

func writeExecutionResultFile(t *testing.T, content string) string {
	t.Helper()

	path := filepath.Join(t.TempDir(), "result.json")
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write result: %v", err)
	}

	return path
}
