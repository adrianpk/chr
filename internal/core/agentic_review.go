package core

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	executioncore "forge.adrianpk.com/chr/internal/core/execution"
	harnesshost "forge.adrianpk.com/chr/internal/core/harness/host"
	ticketcore "forge.adrianpk.com/chr/internal/core/ticket"
	ticketmetadata "forge.adrianpk.com/chr/internal/core/ticket/metadata"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
	workspacecore "forge.adrianpk.com/chr/internal/core/workspace"
	"forge.adrianpk.com/chr/internal/harness"
)

const (
	// AgenticReviewFindingCorrectness defines a package boundary constant.
	AgenticReviewFindingCorrectness = "correctness"
	// AgenticReviewFindingScope defines a package boundary constant.
	AgenticReviewFindingScope = "scope"
	// AgenticReviewFindingEvidence defines a package boundary constant.
	AgenticReviewFindingEvidence = "evidence"
	// AgenticReviewFindingContract defines a package boundary constant.
	AgenticReviewFindingContract = "contract"
	// AgenticReviewFindingSafety defines a package boundary constant.
	AgenticReviewFindingSafety = "safety"
	// AgenticReviewFindingMergeability defines a package boundary constant.
	AgenticReviewFindingMergeability = "mergeability"
	// AgenticReviewFindingCosmetic defines a package boundary constant.
	AgenticReviewFindingCosmetic = "cosmetic"
)

// AgenticReviewContext describes the local evidence given to the reviewer.
type AgenticReviewContext struct {
	ProjectID           string
	ProjectDir          string
	Ticket              ticketcore.Indexed
	Execution           executioncore.ExecutionRecord
	Application         ResultApplication
	ExecutionReportPath string
}

// AgenticReviewResult describes the structured agentic review decision.
type AgenticReviewResult struct {
	Decision          IntegrationReviewDecision     `json:"decision"`
	Summary           string                        `json:"summary"`
	BlockingFindings  []workflowstate.ReviewFinding `json:"blocking_findings,omitempty"`
	NonBlockingNotes  []workflowstate.ReviewFinding `json:"non_blocking_notes,omitempty"`
	ReviewedArtifacts []string                      `json:"reviewed_artifacts,omitempty"`
}

// AgenticReviewArtifact describes files produced by one local review.
type AgenticReviewArtifact struct {
	ContextPath string
	ResultPath  string
	SummaryPath string
	Run         harness.AgentRunResult
	Result      AgenticReviewResult
}

// ApplyAgenticReviewResultOptions describes local ticket metadata application.
type ApplyAgenticReviewResultOptions struct {
	Ticket     ticketcore.Indexed
	Artifact   AgenticReviewArtifact
	ReviewedAt time.Time
}

// AgenticReviewRunner executes a local agentic review.
type AgenticReviewRunner interface {
	Review(ctx context.Context, review AgenticReviewContext) (AgenticReviewArtifact, error)
}

// HarnessAgenticReviewRunner runs a high-level review through the injected harness.
type HarnessAgenticReviewRunner struct {
	Runner harness.AgentRunner
}

// Review executes the corresponding operation.
func (r HarnessAgenticReviewRunner) Review(ctx context.Context, review AgenticReviewContext) (AgenticReviewArtifact, error) {
	if err := ValidateAgenticReviewContext(review); err != nil {
		return AgenticReviewArtifact{}, err
	}

	runDir := AgenticReviewRunDirForContext(review)
	if err := os.MkdirAll(runDir, 0o755); err != nil {
		return AgenticReviewArtifact{}, fmt.Errorf("create agentic review directory: %w", err)
	}

	contextPath := filepath.Join(runDir, "context.md")

	contextBody := RenderAgenticReviewContext(review)
	if err := os.WriteFile(contextPath, []byte(contextBody), 0o644); err != nil {
		return AgenticReviewArtifact{}, fmt.Errorf("write agentic review context: %w", err)
	}

	agentRun, err := harnesshost.RunAgent(ctx, r.Runner, harness.AgentRunRequest{
		ID:            "review-" + review.Ticket.Key + "-" + review.Execution.ID,
		Purpose:       "agentic_review",
		ProjectDir:    review.ProjectDir,
		RunDir:        runDir,
		Prompt:        AgenticReviewPrompt(contextBody),
		OutputSchema:  AgenticReviewOutputSchema(),
		OutputName:    "result.json",
		ReadOnlyFiles: true,
		ReadOnlyRefs:  []string{review.ExecutionReportPath, review.Application.Evidence.ResultPath, review.Application.Evidence.ValidationPath, review.Application.Evidence.LogPath},
	})
	if err != nil {
		return AgenticReviewArtifact{}, err
	}

	result, err := ParseAgenticReviewResult(agentRun.Output)

	err = harnesshost.ValidateGenerationOutput(ctx, harness.AgentRunRequest{ID: "review-" + review.Ticket.Key + "-" + review.Execution.ID, Purpose: "agentic_review", ProjectDir: review.ProjectDir, RunDir: runDir}, err)
	if err != nil {
		return AgenticReviewArtifact{}, err
	}

	artifact := AgenticReviewArtifact{
		ContextPath: contextPath,
		ResultPath:  agentRun.OutputPath,
		SummaryPath: filepath.Join(runDir, "summary.md"),
		Run:         agentRun,
		Result:      result,
	}

	if err := WriteAgenticReviewSummary(artifact); err != nil {
		return AgenticReviewArtifact{}, err
	}

	return artifact, nil
}

// AgenticReviewRunDirForContext returns the artifact directory for one review.
func AgenticReviewRunDirForContext(review AgenticReviewContext) string {
	if review.Execution.Artifacts.ArtifactID != "" && review.Execution.Artifacts.RunDir != "" {
		return filepath.Join(review.Execution.Artifacts.RunDir, "extension", "chr", "agentic-review")
	}

	return AgenticReviewRunDir(review.ProjectDir, review.Ticket.Key, review.Execution.ID)
}

// AgenticReviewRunDir returns the compatibility local artifact directory for one review.
func AgenticReviewRunDir(projectDir, ticketKey, executionID string) string {
	return workspacecore.UserPath(projectDir, "runtime", "reviews", safeAgenticReviewPathPart(ticketKey), safeAgenticReviewPathPart(executionID))
}

// ValidateAgenticReviewContext validates the corresponding package boundary value.
func ValidateAgenticReviewContext(review AgenticReviewContext) error {
	required := []struct {
		name  string
		value string
	}{
		{name: "project id", value: review.ProjectID},
		{name: "project dir", value: review.ProjectDir},
		{name: "ticket key", value: review.Ticket.Key},
		{name: "ticket path", value: review.Ticket.Path},
		{name: "execution id", value: review.Execution.ID},
		{name: "execution report path", value: review.ExecutionReportPath},
	}

	for _, field := range required {
		if strings.TrimSpace(field.value) == "" {
			return fmt.Errorf("agentic review %s is required", field.name)
		}
	}

	if review.Execution.ProjectID != review.ProjectID {
		return fmt.Errorf("agentic review execution project mismatch: %s", review.Execution.ProjectID)
	}

	if review.Execution.TicketKey != review.Ticket.Key {
		return fmt.Errorf("agentic review execution ticket mismatch: %s", review.Execution.TicketKey)
	}

	return nil
}

// RenderAgenticReviewContext renders the reviewer context artifact.
func RenderAgenticReviewContext(review AgenticReviewContext) string {
	var output strings.Builder

	output.WriteString("# Agentic Review Context\n\n")
	writeAgenticReviewField(&output, "project_id", review.ProjectID)
	writeAgenticReviewField(&output, "ticket_key", review.Ticket.Key)
	writeAgenticReviewField(&output, "ticket_title", review.Ticket.Title)
	writeAgenticReviewField(&output, "ticket_status", review.Ticket.Status)
	writeAgenticReviewField(&output, "ticket_path", review.Ticket.Path)
	writeAgenticReviewField(&output, "branch", review.Ticket.Branch)
	writeAgenticReviewField(&output, "execution_id", review.Execution.ID)
	writeAgenticReviewField(&output, "execution_status", string(review.Execution.Status))
	writeAgenticReviewField(&output, "execution_report_path", review.ExecutionReportPath)
	writeAgenticReviewField(&output, "result_path", review.Application.Evidence.ResultPath)
	writeAgenticReviewField(&output, "validation_command", review.Application.Evidence.ValidationCommand)
	writeAgenticReviewField(&output, "validation_log", review.Application.Evidence.ValidationPath)
	writeAgenticReviewField(&output, "runtime_log", review.Application.Evidence.LogPath)

	output.WriteString("\n## Review References\n\n")

	if len(review.Ticket.ReviewRefs) == 0 {
		output.WriteString("- none\n")
	} else {
		for _, ref := range review.Ticket.ReviewRefs {
			output.WriteString("- type: ")
			output.WriteString(ref.Type)
			output.WriteString("; provider: ")
			output.WriteString(ref.Provider)
			output.WriteString("; url: ")
			output.WriteString(ref.URL)
			output.WriteString("\n")
		}
	}

	output.WriteString("\n## Changed Files\n\n")

	if len(review.Application.ChangedFiles) == 0 {
		output.WriteString("- none reported\n")
	} else {
		for _, changedFile := range review.Application.ChangedFiles {
			output.WriteString("- ")
			output.WriteString(changedFile.Path)

			if strings.TrimSpace(changedFile.Summary) != "" {
				output.WriteString(": ")
				output.WriteString(changedFile.Summary)
			}

			output.WriteString("\n")
		}
	}

	return output.String()
}

// AgenticReviewPrompt renders the harness reviewer prompt.
func AgenticReviewPrompt(contextBody string) string {
	return strings.TrimSpace(`You are reviewing one completed implementation ticket.

Use the provided context and inspect referenced local files as needed.
This is a high-level correctness review, not a style preference pass.

Block only for objective issues:
- correctness defects
- scope mismatch
- missing or weak evidence
- broken contracts
- safety or data-loss risk
- mergeability problems

Do not block for cosmetic style, optional refactors, naming preferences, or
non-critical cleanup when configured validation passed.

Return only JSON matching the provided schema.

`+contextBody) + "\n"
}

// AgenticReviewOutputSchema returns the JSON schema for harness review output.
func AgenticReviewOutputSchema() []byte {
	return []byte(`{
  "type": "object",
  "additionalProperties": false,
  "required": ["decision", "summary", "blocking_findings", "non_blocking_notes"],
  "properties": {
    "decision": {
      "type": "string",
      "enum": ["pass", "needs_changes", "insufficient_evidence", "human_required"]
    },
    "summary": {"type": "string"},
    "blocking_findings": {
      "type": "array",
      "items": {"$ref": "#/$defs/finding"}
    },
    "non_blocking_notes": {
      "type": "array",
      "items": {"$ref": "#/$defs/finding"}
    },
    "reviewed_artifacts": {
      "type": "array",
      "items": {"type": "string"}
    }
  },
  "$defs": {
    "finding": {
      "type": "object",
      "additionalProperties": false,
      "required": ["category", "title", "detail"],
      "properties": {
        "category": {
          "type": "string",
          "enum": ["correctness", "scope", "evidence", "contract", "safety", "mergeability", "cosmetic"]
        },
        "title": {"type": "string"},
        "detail": {"type": "string"}
      }
    }
  }
}`)
}

// ParseAgenticReviewResult parses and validates JSON review output.
func ParseAgenticReviewResult(payload []byte) (AgenticReviewResult, error) {
	var result AgenticReviewResult

	decoder := json.NewDecoder(strings.NewReader(string(payload)))
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(&result); err != nil {
		return AgenticReviewResult{}, fmt.Errorf("decode agentic review result: %w", err)
	}

	if err := ValidateAgenticReviewResult(result); err != nil {
		return AgenticReviewResult{}, err
	}

	return result, nil
}

// ValidateAgenticReviewResult validates one structured review result.
func ValidateAgenticReviewResult(result AgenticReviewResult) error {
	if !result.Decision.Valid() {
		return fmt.Errorf("invalid agentic review decision: %s", result.Decision)
	}

	if strings.TrimSpace(result.Summary) == "" {
		return fmt.Errorf("agentic review summary is required")
	}

	if result.Decision == IntegrationReviewDecisionNeedsChanges && len(result.BlockingFindings) == 0 {
		return fmt.Errorf("agentic review needs_changes requires blocking findings")
	}

	for _, finding := range agenticReviewFindings(result) {
		if err := validateAgenticReviewFinding(finding); err != nil {
			return err
		}
	}

	return nil
}

// ApplyAgenticReviewResult writes local review metadata to the ticket.
func ApplyAgenticReviewResult(ctx context.Context, opts ApplyAgenticReviewResultOptions) (TicketReviewCommandResult, error) {
	if strings.TrimSpace(opts.Ticket.Path) == "" {
		return TicketReviewCommandResult{}, fmt.Errorf("agentic review ticket path is required")
	}

	reviewResult := ticketReviewResultForAgenticDecision(opts.Artifact.Result.Decision)
	ref := agenticReviewMetadataRef(opts.Ticket, opts.Artifact)
	result := TicketReviewCommandResult{
		TicketKey:        opts.Ticket.Key,
		ReviewRef:        ref,
		ReviewState:      ticketReviewStateForResult(reviewResult),
		ReviewResult:     reviewResult,
		ReviewRound:      opts.Ticket.ReviewRound + 1,
		Status:           ticketReviewStatusForResult(reviewResult),
		BlockingFindings: agenticReviewBlockingFindingBodies(opts.Artifact.Result),
	}

	if result.Status == ticketcore.StatusReadyForMerge {
		ticket := opts.Ticket
		ticket.Status = result.Status
		ticket.ReviewState = result.ReviewState
		ticket.ReviewResult = result.ReviewResult

		if _, err := ValidateTicketMergeHandoff(ticket); err != nil {
			return TicketReviewCommandResult{}, err
		}
	}

	if err := applyTicketReviewCommandResult(ctx, opts.Ticket.Path, result, opts.ReviewedAt); err != nil {
		return TicketReviewCommandResult{}, err
	}

	return result, nil
}

// WriteAgenticReviewSummary writes a human-readable review summary.
func WriteAgenticReviewSummary(artifact AgenticReviewArtifact) error {
	var output strings.Builder

	output.WriteString("# Agentic Review\n\n")
	writeAgenticReviewField(&output, "decision", string(artifact.Result.Decision))
	writeAgenticReviewField(&output, "summary", artifact.Result.Summary)
	writeAgenticReviewField(&output, "result", artifact.ResultPath)

	output.WriteString("\n## Blocking Findings\n\n")
	writeAgenticReviewFindings(&output, artifact.Result.BlockingFindings)

	output.WriteString("\n## Non-Blocking Notes\n\n")
	writeAgenticReviewFindings(&output, artifact.Result.NonBlockingNotes)

	if err := os.WriteFile(artifact.SummaryPath, []byte(output.String()), 0o644); err != nil {
		return fmt.Errorf("write agentic review summary: %w", err)
	}

	return nil
}

func agenticReviewFindings(result AgenticReviewResult) []workflowstate.ReviewFinding {
	findings := make([]workflowstate.ReviewFinding, 0, len(result.BlockingFindings)+len(result.NonBlockingNotes))
	findings = append(findings, result.BlockingFindings...)
	findings = append(findings, result.NonBlockingNotes...)

	return findings
}

func ticketReviewResultForAgenticDecision(decision IntegrationReviewDecision) string {
	switch decision {
	case IntegrationReviewDecisionPass:
		return TicketReviewCommandResultPass
	case IntegrationReviewDecisionNeedsChanges:
		return TicketReviewCommandResultNeedsChanges
	case IntegrationReviewDecisionInsufficientEvidence:
		return TicketReviewCommandResultInsufficientEvidence
	case IntegrationReviewDecisionHumanRequired:
		return TicketReviewCommandResultNeedsHumanDecision
	default:
		return TicketReviewCommandResultNeedsHumanDecision
	}
}

func ticketReviewStateForResult(result string) string {
	state, _ := ticketReviewCommandStateAndStatus(result)

	return state
}

func ticketReviewStatusForResult(result string) string {
	_, status := ticketReviewCommandStateAndStatus(result)

	return status
}

func agenticReviewMetadataRef(ticket ticketcore.Indexed, artifact AgenticReviewArtifact) ticketmetadata.ReviewRef {
	if ref, ok := firstTicketPullRequestReviewRef(ticket.ReviewRefs); ok {
		return ref
	}

	return ticketmetadata.ReviewRef{
		Type:     "local_review",
		Provider: "chr",
		URL:      artifact.SummaryPath,
		ID:       artifact.Run.ID,
		Status:   string(artifact.Result.Decision),
	}
}

func firstTicketPullRequestReviewRef(refs []ticketmetadata.ReviewRef) (ticketmetadata.ReviewRef, bool) {
	for _, ref := range refs {
		if ref.Type == "pull_request" && strings.TrimSpace(ref.URL) != "" {
			return ref, true
		}
	}

	return ticketmetadata.ReviewRef{}, false
}

func agenticReviewBlockingFindingBodies(result AgenticReviewResult) []string {
	if len(result.BlockingFindings) == 0 && result.Decision != IntegrationReviewDecisionPass {
		return []string{result.Summary}
	}

	findings := make([]string, 0, len(result.BlockingFindings))
	for _, finding := range result.BlockingFindings {
		findings = append(findings, finding.Category+": "+finding.Title+" - "+finding.Detail)
	}

	return findings
}

func validateAgenticReviewFinding(finding workflowstate.ReviewFinding) error {
	if !validAgenticReviewFindingCategory(finding.Category) {
		return fmt.Errorf("invalid agentic review finding category: %s", finding.Category)
	}

	if strings.TrimSpace(finding.Title) == "" {
		return fmt.Errorf("agentic review finding title is required")
	}

	if strings.TrimSpace(finding.Detail) == "" {
		return fmt.Errorf("agentic review finding detail is required")
	}

	return nil
}

func validAgenticReviewFindingCategory(category string) bool {
	switch strings.TrimSpace(category) {
	case AgenticReviewFindingCorrectness,
		AgenticReviewFindingScope,
		AgenticReviewFindingEvidence,
		AgenticReviewFindingContract,
		AgenticReviewFindingSafety,
		AgenticReviewFindingMergeability,
		AgenticReviewFindingCosmetic:
		return true
	default:
		return false
	}
}

func writeAgenticReviewFindings(output *strings.Builder, findings []workflowstate.ReviewFinding) {
	if len(findings) == 0 {
		output.WriteString("- none\n")

		return
	}

	for _, finding := range findings {
		output.WriteString("- ")
		output.WriteString(finding.Category)
		output.WriteString(": ")
		output.WriteString(finding.Title)
		output.WriteString(" - ")
		output.WriteString(finding.Detail)
		output.WriteString("\n")
	}
}

func writeAgenticReviewField(output *strings.Builder, key, value string) {
	output.WriteString("- ")
	output.WriteString(key)
	output.WriteString(": ")

	if strings.TrimSpace(value) == "" {
		output.WriteString("(none)")
	} else {
		output.WriteString(value)
	}

	output.WriteString("\n")
}

func safeAgenticReviewPathPart(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	value = strings.ReplaceAll(value, string(os.PathSeparator), "-")
	value = strings.ReplaceAll(value, "/", "-")
	value = strings.ReplaceAll(value, "\\", "-")

	value = strings.ReplaceAll(value, " ", "-")

	if value == "" {
		return time.Now().UTC().Format("20060102150405")
	}

	return value
}
