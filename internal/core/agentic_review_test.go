package core

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"forge.adrianpk.com/chr/internal/harness"
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

func TestHarnessAgenticReviewRunnerWritesArtifacts(t *testing.T) {
	projectDir := t.TempDir()

	ticketPath := filepath.Join(projectDir, ".chr", "default", "tickets", "CHR-1.md")
	if err := os.MkdirAll(filepath.Dir(ticketPath), 0o755); err != nil {
		t.Fatalf("MkdirAll returned error: %v", err)
	}

	now := time.Date(2026, 6, 20, 12, 0, 0, 0, time.UTC)
	review := testAgenticReviewContext(projectDir, ticketPath, now)

	runner := HarnessAgenticReviewRunner{Runner: fakeAgenticReviewAgentRunner{}}

	artifact, err := runner.Review(context.Background(), review)
	if err != nil {
		t.Fatalf("Review returned error: %v", err)
	}

	if artifact.Result.Decision != IntegrationReviewDecisionPass {
		t.Fatalf("Decision = %q, want pass", artifact.Result.Decision)
	}

	wantRunDir := filepath.Join(projectDir, "ops", "default", "run", "00001-ticket-run", "extension", "chr", "agentic-review")
	if filepath.Dir(artifact.ContextPath) != wantRunDir {
		t.Fatalf("ContextPath dir = %q, want %q", filepath.Dir(artifact.ContextPath), wantRunDir)
	}

	assertAgenticReviewArtifactExists(t, artifact.ContextPath)
	assertAgenticReviewArtifactExists(t, artifact.ResultPath)
	assertAgenticReviewArtifactExists(t, artifact.SummaryPath)
}

func testAgenticReviewContext(projectDir, ticketPath string, now time.Time) AgenticReviewContext {
	return AgenticReviewContext{
		ProjectID:           "project-1",
		ProjectDir:          projectDir,
		Ticket:              testAgenticReviewIndexedTicket(ticketPath),
		Execution:           testAgenticReviewExecution(projectDir, ticketPath, now),
		Application:         testAgenticReviewApplication(),
		ExecutionReportPath: "execution-report.md",
	}
}

func testAgenticReviewIndexedTicket(ticketPath string) IndexedTicket {
	return IndexedTicket{
		IndexedTicketIdentity: IndexedTicketIdentity{
			Key: "CHR-1",

			Path: ticketPath,
		}, IndexedTicketSummary: IndexedTicketSummary{Title: "Implement feature"}, IndexedTicketClassification: IndexedTicketClassification{Status: TicketStatusReadyForReview},
	}
}

func testAgenticReviewExecution(projectDir, ticketPath string, now time.Time) ExecutionRecord {
	return ExecutionRecord{
		ID:         "exec-1",
		ProjectID:  "project-1",
		TicketKey:  "CHR-1",
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
			ArtifactID:     "00001-ticket-run",
			RunDir:         filepath.Join(projectDir, "ops", "default", "run", "00001-ticket-run"),
			ResultPath:     "result.json",
			LogPath:        "runtime.log",
			ValidationPath: "validation.log",
		},
	}
}

func testAgenticReviewApplication() ResultApplication {
	return ResultApplication{
		TicketKey: "CHR-1",
		Evidence: ResultApplicationEvidence{
			ResultPath:        "result.json",
			LogPath:           "runtime.log",
			ValidationPath:    "validation.log",
			ValidationCommand: "go test ./...",
		},
	}
}

func TestApplyAgenticReviewResultPassRequiresMergeHandoff(t *testing.T) {
	ticketPath := writeAgenticReviewTicket(t, Ticket{
		Key:    "CHR-1",
		Title:  "Ready",
		Status: TicketStatusReadyForReview,
		Branch: "feat/chr-1",
	})

	_, err := ApplyAgenticReviewResult(context.Background(), ApplyAgenticReviewResultOptions{
		Ticket: IndexedTicket{
			IndexedTicketIdentity: IndexedTicketIdentity{
				Key:  "CHR-1",
				Path: ticketPath,
			}, IndexedTicketClassification: IndexedTicketClassification{Status: TicketStatusReadyForReview}, IndexedTicketSource: IndexedTicketSource{Branch: "feat/chr-1"},
		},
		Artifact: agenticReviewArtifactForTest(AgenticReviewResult{
			Decision: IntegrationReviewDecisionPass,
			Summary:  "Ready to merge.",
		}),
	})
	if err == nil || !strings.Contains(err.Error(), "requires a pull_request review ref") {
		t.Fatalf("error = %v, want missing pull request review ref", err)
	}
}

func TestApplyAgenticReviewResultPassMovesTicketToReadyForMerge(t *testing.T) {
	ref := TicketReviewRef{
		Type:     "pull_request",
		Provider: "forgejo",
		URL:      "https://forge.example.com/owner/repo/pulls/1",
		ID:       "1",
		Status:   "open",
	}
	ticketPath := writeAgenticReviewTicket(t, Ticket{
		Key:        "CHR-1",
		Title:      "Ready",
		Status:     TicketStatusReadyForReview,
		Branch:     "feat/chr-1",
		ReviewRefs: []TicketReviewRef{ref},
	})

	_, err := ApplyAgenticReviewResult(context.Background(), ApplyAgenticReviewResultOptions{
		Ticket: IndexedTicket{
			IndexedTicketIdentity: IndexedTicketIdentity{
				Key:  "CHR-1",
				Path: ticketPath,
			}, IndexedTicketClassification: IndexedTicketClassification{Status: TicketStatusReadyForReview}, IndexedTicketSource: IndexedTicketSource{Branch: "feat/chr-1"}, IndexedTicketReview: IndexedTicketReview{ReviewRefs: []TicketReviewRef{ref}},
		},
		Artifact: agenticReviewArtifactForTest(AgenticReviewResult{
			Decision: IntegrationReviewDecisionPass,
			Summary:  "Ready to merge.",
		}),
	})
	if err != nil {
		t.Fatalf("ApplyAgenticReviewResult returned error: %v", err)
	}

	content := readAgenticReviewTicket(t, ticketPath)
	for _, want := range []string{
		`status: "ready_for_merge"`,
		`review_state: "pass"`,
		`review_result: "pass"`,
		`review_comments_url: "https://forge.example.com/owner/repo/pulls/1"`,
	} {
		if !strings.Contains(content, want) {
			t.Fatalf("ticket missing %q:\n%s", want, content)
		}
	}
}

func TestApplyAgenticReviewResultNeedsChangesMovesTicketToReadyForDev(t *testing.T) {
	ticketPath := writeAgenticReviewTicket(t, Ticket{
		Key:    "CHR-1",
		Title:  "Ready",
		Status: TicketStatusReadyForReview,
		Branch: "feat/chr-1",
	})

	_, err := ApplyAgenticReviewResult(context.Background(), ApplyAgenticReviewResultOptions{
		Ticket: IndexedTicket{
			IndexedTicketIdentity: IndexedTicketIdentity{
				Key:  "CHR-1",
				Path: ticketPath,
			}, IndexedTicketClassification: IndexedTicketClassification{Status: TicketStatusReadyForReview}, IndexedTicketSource: IndexedTicketSource{Branch: "feat/chr-1"},
		},
		Artifact: agenticReviewArtifactForTest(AgenticReviewResult{
			Decision: IntegrationReviewDecisionNeedsChanges,
			Summary:  "Needs repair.",
			BlockingFindings: []AgenticReviewFinding{{
				Category: AgenticReviewFindingCorrectness,
				Title:    "Broken behavior",
				Detail:   "The main flow fails.",
			}},
		}),
	})
	if err != nil {
		t.Fatalf("ApplyAgenticReviewResult returned error: %v", err)
	}

	content := readAgenticReviewTicket(t, ticketPath)
	for _, want := range []string{
		`status: "ready_for_dev"`,
		`review_state: "changes_requested"`,
		`review_result: "needs_changes"`,
		`correctness: Broken behavior - The main flow fails.`,
	} {
		if !strings.Contains(content, want) {
			t.Fatalf("ticket missing %q:\n%s", want, content)
		}
	}
}

type fakeAgenticReviewAgentRunner struct{}

func (fakeAgenticReviewAgentRunner) RunAgent(_ context.Context, request harness.AgentRunRequest) (harness.AgentRunResult, error) {
	output := []byte(`{
		"decision": "pass",
		"summary": "Looks correct.",
		"blocking_findings": [],
		"non_blocking_notes": []
	}`)

	if err := os.WriteFile(filepath.Join(request.RunDir, request.OutputName), output, 0o644); err != nil {
		return harness.AgentRunResult{}, err
	}

	return harness.AgentRunResult{
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

func writeAgenticReviewTicket(t *testing.T, ticket Ticket) string {
	t.Helper()

	dir := t.TempDir()
	path := filepath.Join(dir, "CHR-1.md")

	content, err := NewTicketRenderer().RenderTicket(ticket)
	if err != nil {
		t.Fatalf("RenderTicket returned error: %v", err)
	}

	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("WriteFile returned error: %v", err)
	}

	return path
}

func readAgenticReviewTicket(t *testing.T, path string) string {
	t.Helper()

	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile returned error: %v", err)
	}

	return string(content)
}

func agenticReviewArtifactForTest(result AgenticReviewResult) AgenticReviewArtifact {
	return AgenticReviewArtifact{
		SummaryPath: "/tmp/chr-agentic-review-summary.md",
		Run: harness.AgentRunResult{
			ID: "review-1",
		},
		Result: result,
	}
}
