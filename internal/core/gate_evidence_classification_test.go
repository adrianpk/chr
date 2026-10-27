package core

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	executioncore "forge.adrianpk.com/chr/internal/core/execution"
	"forge.adrianpk.com/chr/internal/core/gate"
	opscore "forge.adrianpk.com/chr/internal/core/ops"
	opsnamespace "forge.adrianpk.com/chr/internal/core/ops/namespace"
)

func TestGateEvidenceClassifierClassifiesAgentManagedAndManualChanges(t *testing.T) {
	projectDir := t.TempDir()
	writeGateEvidenceExecutionArtifacts(t, projectDir, "CHR-1", "exec-1", executioncore.Result{
		ExecutionID: "exec-1",
		TicketKey:   "CHR-1",
		Status:      executioncore.ResultStatusSucceeded,
		Summary:     "implemented",
		EvidenceSummary: executioncore.ResultEvidenceSummary{
			Strategy:   "planned",
			Produced:   "updated gate",
			Validation: "validation passed",
		},
		ChangedFiles: []executioncore.ResultChangedFile{
			{Path: "internal/core/gate.go", Summary: "updated gate"},
		},
	})

	classification, err := gate.NewEvidenceClassifier(NewReportLister().ListGateExecutionReports).Classify(context.Background(), gate.EvidenceClassificationRequest{
		ProjectDir: projectDir,
		Files: []gate.GitChangedFile{
			{Status: "M", Path: "internal/core/gate.go"},
			{Status: "M", Path: "README.md"},
		},
	})
	if err != nil {
		t.Fatalf("Classify returned error: %v", err)
	}

	if len(classification.AgentManagedWork) != 1 {
		t.Fatalf("AgentManagedWork = %#v, want one entry", classification.AgentManagedWork)
	}

	managed := classification.AgentManagedWork[0]
	if managed.Path != "internal/core/gate.go" || managed.TicketKey != "CHR-1" || managed.ExecutionID != "exec-1" {
		t.Fatalf("managed = %#v, want CHR-1 exec-1 gate.go", managed)
	}

	if len(classification.ManualOrExternalChanges) != 1 {
		t.Fatalf("ManualOrExternalChanges = %#v, want one entry", classification.ManualOrExternalChanges)
	}

	if classification.ManualOrExternalChanges[0].Path != "README.md" {
		t.Fatalf("manual path = %q, want README.md", classification.ManualOrExternalChanges[0].Path)
	}
}

func TestGateEvidenceClassifierClassifiesHighRiskManualChanges(t *testing.T) {
	projectDir := t.TempDir()

	classification, err := gate.NewEvidenceClassifier(NewReportLister().ListGateExecutionReports).Classify(context.Background(), gate.EvidenceClassificationRequest{
		ProjectDir: projectDir,
		Files: []gate.GitChangedFile{
			{Status: "M", Path: "internal/runtime/runner.go"},
			{Status: "M", Path: "README.md"},
		},
		HighRiskPaths: []string{"internal/runtime/**"},
	})
	if err != nil {
		t.Fatalf("Classify returned error: %v", err)
	}

	if len(classification.HighRiskManualChanges) != 1 {
		t.Fatalf("HighRiskManualChanges = %#v, want one entry", classification.HighRiskManualChanges)
	}

	if classification.HighRiskManualChanges[0].Pattern != "internal/runtime/**" {
		t.Fatalf("Pattern = %q, want internal/runtime/**", classification.HighRiskManualChanges[0].Pattern)
	}
}

func TestGateEvidenceClassifierRecordsMissingExecutionResultEvidence(t *testing.T) {
	projectDir := t.TempDir()
	reportPath := filepath.Join(
		projectDir,
		opscore.DirectoryName,
		opsnamespace.Default,
		"run",
		"00001-ticket-run",
		"evidence",
		"execution-report.md",
	)
	writeGateEvidenceFile(t, reportPath, `# Execution Report

## Summary

- execution: exec-1
- ticket: CHR-1
- status: succeeded

## Artifacts

- result: ops/default/run/00001-ticket-run/evidence/result.json
`)

	classification, err := gate.NewEvidenceClassifier(NewReportLister().ListGateExecutionReports).Classify(context.Background(), gate.EvidenceClassificationRequest{
		ProjectDir: projectDir,
		Files: []gate.GitChangedFile{
			{Status: "M", Path: "internal/core/gate.go"},
		},
	})
	if err != nil {
		t.Fatalf("Classify returned error: %v", err)
	}

	if len(classification.MissingAgentEvidence) != 1 {
		t.Fatalf("MissingAgentEvidence = %#v, want one entry", classification.MissingAgentEvidence)
	}

	if !strings.Contains(classification.MissingAgentEvidence[0].Reason, "open execution result") {
		t.Fatalf("Reason = %q, want result load context", classification.MissingAgentEvidence[0].Reason)
	}
}

func writeGateEvidenceExecutionArtifacts(
	t *testing.T,
	projectDir string,
	ticketKey string,
	executionID string,
	result executioncore.Result,
) {
	t.Helper()

	runDir := filepath.Join(projectDir, opscore.DirectoryName, opsnamespace.Default, "run", "00001-ticket-run")
	resultPath := filepath.Join(runDir, "evidence", "result.json")
	reportPath := filepath.Join(runDir, "evidence", "execution-report.md")
	resultContent := `{"execution_id":"` + result.ExecutionID +
		`","ticket_key":"` + result.TicketKey +
		`","status":"` + string(result.Status) +
		`","summary":"` + result.Summary +
		`","evidence_summary":{"strategy":"` + result.EvidenceSummary.Strategy +
		`","produced":"` + result.EvidenceSummary.Produced +
		`","validation":"` + result.EvidenceSummary.Validation +
		`"},"changed_files":[{"path":"` + result.ChangedFiles[0].Path +
		`","summary":"` + result.ChangedFiles[0].Summary + `"}]}`

	writeGateEvidenceFile(t, resultPath, resultContent)
	writeGateEvidenceFile(t, reportPath, `# Execution Report

## Summary

- execution: `+executionID+`
- ticket: `+ticketKey+`
- status: succeeded

## Artifacts

- result: ops/default/run/00001-ticket-run/evidence/result.json
`)
}

func writeGateEvidenceFile(t *testing.T, path, content string) {
	t.Helper()

	err := os.MkdirAll(filepath.Dir(path), 0o755)
	if err != nil {
		t.Fatalf("create gate evidence file dir: %v", err)
	}

	err = os.WriteFile(path, []byte(content), 0o644)
	if err != nil {
		t.Fatalf("write gate evidence file: %v", err)
	}
}
