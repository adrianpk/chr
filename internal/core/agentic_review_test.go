package core

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestParseAgenticReviewResultAcceptsCleanPassWithCosmeticNotes(t *testing.T) {
	payload := []byte(`{
		"decision": "pass",
		"summary": "Implementation is correct and validated.",
		"blocking_findings": [],
		"non_blocking_notes": [
			{"category": "cosmetic", "title": "Name could be shorter", "detail": "This is preference-only and does not block."}
		],
		"reviewed_artifacts": ["result.json", "execution-report.md"]
	}`)

	got, err := ParseAgenticReviewResult(payload)
	if err != nil {
		t.Fatalf("ParseAgenticReviewResult returned error: %v", err)
	}

	if got.Decision != IntegrationReviewDecisionPass {
		t.Fatalf("Decision = %q, want pass", got.Decision)
	}

	if len(got.BlockingFindings) != 0 || len(got.NonBlockingNotes) != 1 {
		t.Fatalf("findings = blocking %#v notes %#v", got.BlockingFindings, got.NonBlockingNotes)
	}
}

func TestParseAgenticReviewResultRequiresBlockingFindingForNeedsChanges(t *testing.T) {
	payload := []byte(`{
		"decision": "needs_changes",
		"summary": "Needs changes.",
		"blocking_findings": [],
		"non_blocking_notes": []
	}`)

	_, err := ParseAgenticReviewResult(payload)
	if err == nil || !strings.Contains(err.Error(), "needs_changes requires blocking findings") {
		t.Fatalf("error = %v, want blocking findings requirement", err)
	}
}

func TestParseAgenticReviewResultRejectsUnknownFindingCategory(t *testing.T) {
	payload := []byte(`{
		"decision": "insufficient_evidence",
		"summary": "Cannot prove validation.",
		"blocking_findings": [
			{"category": "preference", "title": "Preference", "detail": "Not an allowed blocking category."}
		],
		"non_blocking_notes": []
	}`)

	_, err := ParseAgenticReviewResult(payload)
	if err == nil || !strings.Contains(err.Error(), "invalid agentic review finding category") {
		t.Fatalf("error = %v, want invalid category", err)
	}
}

func TestCodexAgenticReviewRunnerWritesArtifacts(t *testing.T) {
	projectDir := t.TempDir()
	ticketPath := filepath.Join(projectDir, ".chc", "user", "tickets", "CHC-1.md")
	if err := os.MkdirAll(filepath.Dir(ticketPath), 0o755); err != nil {
		t.Fatalf("MkdirAll returned error: %v", err)
	}

	now := time.Date(2026, 6, 20, 12, 0, 0, 0, time.UTC)

	review := AgenticReviewContext{
		ProjectID:  "project-1",
		ProjectDir: projectDir,
		Ticket: IndexedTicket{
			Key:    "CHC-1",
			Title:  "Implement feature",
			Status: TicketStatusReadyForReview,
			Path:   ticketPath,
		},
		Execution: ExecutionRecord{
			ID:         "exec-1",
			ProjectID:  "project-1",
			TicketKey:  "CHC-1",
			TicketPath: ticketPath,
			Status:     ExecutionStatusSucceeded,
			CreatedAt:  now,
			UpdatedAt:  now,
			Runtime: ExecutionRuntimeRef{
				Image:             "runtime",
				DockerfilePath:    "Dockerfile",
				ContainerWorkdir:  "/workspace",
				ValidationCommand: "go test ./...",
			},
			Artifacts: ExecutionArtifacts{
				RunDir:         filepath.Join(projectDir, ".chc", "user", "runtime", "runs", "CHC-1", "exec-1"),
				ResultPath:     "result.json",
				LogPath:        "runtime.log",
				ValidationPath: "validation.log",
			},
		},
		Application: ResultApplication{
			TicketKey: "CHC-1",
			Evidence: ResultApplicationEvidence{
				ResultPath:        "result.json",
				LogPath:           "runtime.log",
				ValidationPath:    "validation.log",
				ValidationCommand: "go test ./...",
			},
		},
		ExecutionReportPath: "execution-report.md",
	}

	runner := CodexAgenticReviewRunner{Runner: fakeAgenticReviewAgentRunner{}}

	artifact, err := runner.Review(context.Background(), review)
	if err != nil {
		t.Fatalf("Review returned error: %v", err)
	}

	if artifact.Result.Decision != IntegrationReviewDecisionPass {
		t.Fatalf("Decision = %q, want pass", artifact.Result.Decision)
	}

	assertAgenticReviewArtifactExists(t, artifact.ContextPath)
	assertAgenticReviewArtifactExists(t, artifact.ResultPath)
	assertAgenticReviewArtifactExists(t, artifact.SummaryPath)
}

type fakeAgenticReviewAgentRunner struct{}

func (fakeAgenticReviewAgentRunner) RunAgent(_ context.Context, request AgentRunRequest) (AgentRunResult, error) {
	output := []byte(`{
		"decision": "pass",
		"summary": "Looks correct.",
		"blocking_findings": [],
		"non_blocking_notes": []
	}`)

	if err := os.WriteFile(filepath.Join(request.RunDir, request.OutputName), output, 0o644); err != nil {
		return AgentRunResult{}, err
	}

	return AgentRunResult{
		ID:         request.ID,
		RunDir:     request.RunDir,
		OutputPath: filepath.Join(request.RunDir, request.OutputName),
		Output:     output,
	}, nil
}

func assertAgenticReviewArtifactExists(t *testing.T, path string) {
	t.Helper()

	if _, err := os.Stat(path); err != nil {
		t.Fatalf("artifact %s stat error: %v", path, err)
	}
}
