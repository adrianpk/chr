package core

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
		"evidence_summary": {
			"strategy": "planned mixed tests",
			"produced": "unit test added",
			"validation": "go test ./internal/core passed",
			"deviations": ["none"],
			"remaining_risk": ["none"]
		},
		"changed_files": [
			{
				"path": "internal/core/execution_result.go",
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

	if result.EvidenceSummary.Strategy != "planned mixed tests" {
		t.Fatalf("EvidenceSummary.Strategy = %q, want planned mixed tests", result.EvidenceSummary.Strategy)
	}

	if len(result.ChangedFiles) != 1 {
		t.Fatalf("ChangedFiles len = %d, want 1", len(result.ChangedFiles))
	}

	if result.ChangedFiles[0].Path != "internal/core/execution_result.go" {
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
		"status": "unknown",
		"evidence_summary": {
			"strategy": "planned",
			"produced": "none",
			"validation": "not run"
		}
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
		"evidence_summary": {
			"strategy": "planned",
			"produced": "none",
			"validation": "not run"
		},
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

func TestLoadExecutionResultRejectsMissingEvidenceSummary(t *testing.T) {
	path := writeExecutionResultFile(t, `{
		"execution_id": "exec-1",
		"ticket_key": "CHC-1",
		"status": "succeeded",
		"summary": "done"
	}`)

	_, err := LoadExecutionResult(path)
	if err == nil {
		t.Fatal("LoadExecutionResult returned nil error")
	}
}

func TestLoadExecutionResultAcceptsEvidenceSummaryForChangeKinds(t *testing.T) {
	for _, tt := range []struct {
		name    string
		content string
		want    string
	}{
		{
			name: "docs only",
			content: `{
				"execution_id": "exec-1",
				"ticket_key": "CHC-1",
				"status": "succeeded",
				"summary": "updated docs",
				"evidence_summary": {
					"strategy": "docs-only strategy; tests not required",
					"produced": "documentation updated",
					"validation": "manual documentation review"
				}
			}`,
			want: "docs-only strategy; tests not required",
		},
		{
			name: "behavior change",
			content: `{
				"execution_id": "exec-1",
				"ticket_key": "CHC-1",
				"status": "succeeded",
				"summary": "updated behavior",
				"evidence_summary": {
					"strategy": "behavior strategy; table tests and fuzzing",
					"produced": "unit table tests, fuzz target, and Gherkin scenario",
					"validation": "make verify passed",
					"deviations": ["mutation testing not configured"],
					"remaining_risk": ["manual CLI smoke deferred"]
				}
			}`,
			want: "behavior strategy; table tests and fuzzing",
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			path := writeExecutionResultFile(t, tt.content)

			result, err := LoadExecutionResult(path)
			if err != nil {
				t.Fatalf("LoadExecutionResult returned error: %v", err)
			}

			if result.EvidenceSummary.Strategy != tt.want {
				t.Fatalf("EvidenceSummary.Strategy = %q, want %q", result.EvidenceSummary.Strategy, tt.want)
			}
		})
	}
}

func testExecutionResultEvidenceSummary() ExecutionResultEvidenceSummary {
	return ExecutionResultEvidenceSummary{
		Strategy:   "planned",
		Produced:   "evidence produced",
		Validation: "validation passed",
	}
}

func writeExecutionResultFile(t *testing.T, content string) string {
	t.Helper()

	path := filepath.Join(t.TempDir(), "result.json")

	err := os.WriteFile(path, []byte(content), 0o644)
	if err != nil {
		t.Fatalf("write result: %v", err)
	}

	return path
}
