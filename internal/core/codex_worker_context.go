package core

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

// CodexWorkerContextMaterializer defines a package boundary interface.
type CodexWorkerContextMaterializer interface {
	MaterializeContext(ctx context.Context, opts MaterializeCodexWorkerContextOptions) error
}

// MaterializeCodexWorkerContextOptions describes a package boundary value.
type MaterializeCodexWorkerContextOptions struct {
	Path           string
	ExecutionID    string
	TicketKey      string
	TicketPath     string
	ResultPath     string
	LogPath        string
	ValidationPath string
	RepairHandoff  *CodexWorkerRepairHandoff
	Evidence       *WorkerEvidenceContext
}

// CodexWorkerRepairHandoff describes a package boundary value.
type CodexWorkerRepairHandoff struct {
	Branch            string
	PullRequestURL    string
	ReviewCommentsURL string
	ReviewState       string
	ReviewResult      string
	ReviewRound       int
}

// WorkerEvidenceContext describes ticket evidence strategy input.
type WorkerEvidenceContext struct {
	Config         EvidenceStrategyConfig
	Checklist      *EvidenceStrategyChecklist
	ChecklistError string
	Policy         EvidencePolicyContext
}

type codexWorkerContextMaterializer struct{}

// NewCodexWorkerContextMaterializer constructs the corresponding package service or value.
func NewCodexWorkerContextMaterializer() CodexWorkerContextMaterializer {
	return &codexWorkerContextMaterializer{}
}

// MaterializeCodexWorkerContextForExecution executes the corresponding package operation.
func MaterializeCodexWorkerContextForExecution(
	ctx context.Context,
	record ExecutionRecord,
	plan RuntimePlan,
) error {
	repairHandoff, err := codexWorkerRepairHandoff(ctx, plan.ProjectDir, record.TicketPath)
	if err != nil {
		return err
	}

	evidence, err := loadWorkerEvidence(ctx, plan.ProjectDir, record.TicketPath)
	if err != nil {
		return err
	}

	return NewCodexWorkerContextMaterializer().MaterializeContext(ctx, MaterializeCodexWorkerContextOptions{
		Path:           plan.WorkerContextPath,
		ExecutionID:    record.ID,
		TicketKey:      record.TicketKey,
		TicketPath:     containerProjectPath(plan, record.TicketPath),
		ResultPath:     containerProjectPath(plan, plan.ResultPath),
		LogPath:        containerProjectPath(plan, plan.LogPath),
		ValidationPath: containerProjectPath(plan, plan.ValidationPath),
		RepairHandoff:  repairHandoff,
		Evidence:       evidence,
	})
}

// MaterializeContext executes the corresponding package operation.
func (m *codexWorkerContextMaterializer) MaterializeContext(
	ctx context.Context,
	opts MaterializeCodexWorkerContextOptions,
) error {
	err := ctx.Err()
	if err != nil {
		return err
	}

	err = validateCodexWorkerContextOptions(opts)
	if err != nil {
		return err
	}

	err = os.MkdirAll(filepath.Dir(opts.Path), 0o755)
	if err != nil {
		return fmt.Errorf("create codex worker context directory: %w", err)
	}

	content := RenderCodexWorkerContext(opts)

	err = os.WriteFile(opts.Path, []byte(content), 0o644)
	if err != nil {
		return fmt.Errorf("write codex worker context: %w", err)
	}

	return nil
}

// RenderCodexWorkerContext renders the corresponding package data.
func RenderCodexWorkerContext(opts MaterializeCodexWorkerContextOptions) string {
	var output strings.Builder

	output.WriteString("# Codex Worker Context\n\n")
	output.WriteString("## Ticket\n\n")
	output.WriteString("- execution_id: ")
	output.WriteString(opts.ExecutionID)
	output.WriteString("\n")
	output.WriteString("- ticket_key: ")
	output.WriteString(opts.TicketKey)
	output.WriteString("\n")
	output.WriteString("- ticket_path: ")
	output.WriteString(opts.TicketPath)
	output.WriteString("\n\n")
	output.WriteString("## Artifacts\n\n")
	output.WriteString("- result_path: ")
	output.WriteString(opts.ResultPath)
	output.WriteString("\n")
	output.WriteString("- log_path: ")
	output.WriteString(opts.LogPath)
	output.WriteString("\n")
	output.WriteString("- validation_path: ")
	output.WriteString(opts.ValidationPath)
	output.WriteString("\n\n")
	writeCodexWorkerRepairHandoff(&output, opts)
	writeCodexWorkerEvidenceStrategy(&output, opts)
	output.WriteString("## Result Contract\n\n")
	output.WriteString("Write a JSON object to `result_path` with these fields:\n\n")
	output.WriteString("- `execution_id`: must match the execution ID above.\n")
	output.WriteString("- `ticket_key`: must match the ticket key above.\n")
	output.WriteString("- `status`: `succeeded` or `failed`.\n")
	output.WriteString("- `summary`: concise human-readable result summary.\n")
	output.WriteString("- `evidence_summary`: required object with `strategy`, `produced`, and `validation`.\n")
	output.WriteString("- `evidence_summary.deviations`: optional list of deviations from the planned strategy.\n")
	output.WriteString("- `evidence_summary.remaining_risk`: optional list of remaining risks.\n")
	output.WriteString("- `changed_files`: optional list of workspace-relative changed file summaries.\n")

	return output.String()
}

func writeCodexWorkerEvidenceStrategy(output *strings.Builder, opts MaterializeCodexWorkerContextOptions) {
	output.WriteString("## Evidence Strategy\n\n")
	writeCodexWorkerEvidenceStrategyConfig(output, opts.Evidence)
	writeCodexWorkerEvidencePolicy(output, opts.Evidence)
	writeCodexWorkerEvidenceStrategyChecklist(output, opts.Evidence)
	output.WriteString("Before modifying project files:\n\n")
	output.WriteString("1. Read the ticket and any existing evidence-strategy checklist.\n")
	output.WriteString("2. Classify the change kind and risk.\n")
	output.WriteString("3. Decide whether Gherkin applies.\n")
	output.WriteString("4. Decide whether new tests are required.\n")
	output.WriteString("5. Decide test timing: before, during, after, or none.\n")
	output.WriteString("6. Write or update the ticket evidence-strategy checklist.\n")
	output.WriteString("7. Treat the checklist as an operational contract during implementation.\n\n")
	output.WriteString("Before finishing:\n\n")
	output.WriteString("1. Re-read the strategy checklist.\n")
	output.WriteString("2. Reconcile planned evidence with produced evidence.\n")
	output.WriteString("3. Update Validation and Evidence summary.\n")
	output.WriteString("4. Do not move the ticket to ready_for_review if required evidence is missing, failing, or blocked.\n\n")
	output.WriteString("The ticket checklist is the operational source of truth.\n\n")
}

func writeCodexWorkerEvidenceStrategyConfig(output *strings.Builder, evidence *WorkerEvidenceContext) {
	if evidence == nil {
		return
	}

	config := evidence.Config

	output.WriteString("### Effective Config\n\n")
	output.WriteString("- enabled: ")
	output.WriteString(strconv.FormatBool(config.Enabled))
	output.WriteString("\n")
	output.WriteString("- mode: ")
	output.WriteString(string(config.Mode))
	output.WriteString("\n")
	output.WriteString("- ticket_checklist: ")
	output.WriteString(string(config.TicketChecklist))
	output.WriteString("\n")
	output.WriteString("- gherkin_policy: ")
	output.WriteString(string(config.GherkinPolicy))
	output.WriteString("\n")
	output.WriteString("- strict_ready_for_review: ")
	output.WriteString(strconv.FormatBool(config.StrictReadyForReview))
	output.WriteString("\n")
	output.WriteString("- allow_user_edits: ")
	output.WriteString(strconv.FormatBool(config.AllowUserEdits))
	output.WriteString("\n")
	output.WriteString("- execution_report_copy: ")
	output.WriteString(strconv.FormatBool(config.ExecutionReportCopy))
	output.WriteString("\n\n")
}

func writeCodexWorkerEvidencePolicy(output *strings.Builder, evidence *WorkerEvidenceContext) {
	if evidence == nil {
		return
	}

	if evidence.Policy.Gherkin.Policy == "" && evidence.Policy.Guidance.ChangeKind == "" {
		return
	}

	writeCodexWorkerGherkinPolicy(output, evidence.Policy.Gherkin)
	writeCodexWorkerSuggestedStrategy(output, evidence.Policy.Guidance)
	writeCodexWorkerDecisionMatrix(output)
}

func writeCodexWorkerGherkinPolicy(output *strings.Builder, gherkin EvidenceGherkinContext) {
	output.WriteString("### Project Gherkin\n\n")
	output.WriteString("- policy: ")
	output.WriteString(string(gherkin.Policy))
	output.WriteString("\n")
	output.WriteString("- active: ")
	output.WriteString(strconv.FormatBool(gherkin.Active))
	output.WriteString("\n")
	output.WriteString("- applies_to_ticket: ")
	output.WriteString(strconv.FormatBool(gherkin.Applies))
	output.WriteString("\n")
	output.WriteString("- required_path: ")
	output.WriteString(nonEmptyWorkerContextValue(gherkin.RequiredPath, "none"))
	output.WriteString("\n")
	output.WriteString("- reason: ")
	output.WriteString(nonEmptyWorkerContextValue(gherkin.Reason, "not evaluated"))
	output.WriteString("\n")

	if len(gherkin.FeaturePaths) > 0 {
		output.WriteString("- feature_paths:\n")

		for _, path := range gherkin.FeaturePaths {
			output.WriteString("  - ")
			output.WriteString(path)
			output.WriteString("\n")
		}
	}

	output.WriteString("\n")
}

func writeCodexWorkerSuggestedStrategy(output *strings.Builder, guidance EvidenceStrategyGuidance) {
	output.WriteString("### Suggested Strategy\n\n")
	output.WriteString("- change_kind: ")
	output.WriteString(nonEmptyWorkerContextValue(guidance.ChangeKind, "unknown"))
	output.WriteString("\n")
	output.WriteString("- risk: ")
	output.WriteString(nonEmptyWorkerContextValue(guidance.Risk, "unknown"))
	output.WriteString("\n")
	output.WriteString("- test_timing: ")
	output.WriteString(nonEmptyWorkerContextValue(guidance.TestTiming, "during"))
	output.WriteString("\n")
	output.WriteString("- gherkin: ")
	output.WriteString(nonEmptyWorkerContextValue(guidance.Gherkin, "inactive"))
	output.WriteString("\n")
	output.WriteString("- tests: ")
	output.WriteString(nonEmptyWorkerContextValue(guidance.Tests, "Classify evidence before implementation."))
	output.WriteString("\n")
	output.WriteString("- reason: ")
	output.WriteString(nonEmptyWorkerContextValue(guidance.Reason, "Use ticket context."))
	output.WriteString("\n\n")
}

func writeCodexWorkerDecisionMatrix(output *strings.Builder) {
	output.WriteString("### Evidence Decision Matrix\n\n")

	for _, row := range EvidenceDecisionMatrixRows() {
		output.WriteString("- ")
		output.WriteString(row.ChangeKind)
		output.WriteString(": ")
		output.WriteString(row.DefaultEvidence)
		output.WriteString("\n")
	}

	output.WriteString("\n")
}

func writeCodexWorkerEvidenceStrategyChecklist(output *strings.Builder, evidence *WorkerEvidenceContext) {
	if evidence == nil {
		return
	}

	output.WriteString("### Existing Checklist\n\n")

	if evidence.ChecklistError != "" {
		output.WriteString("- status: malformed\n")
		output.WriteString("- error: ")
		output.WriteString(evidence.ChecklistError)
		output.WriteString("\n\n")

		return
	}

	if evidence.Checklist == nil || !evidence.Checklist.Found {
		output.WriteString("- status: missing\n\n")

		return
	}

	output.WriteString("- status: present\n\n")

	for _, label := range EvidenceStrategyChecklistLabels() {
		item, ok := evidence.Checklist.Item(label)
		if !ok {
			continue
		}

		writeCodexWorkerChecklistItem(output, item)
	}

	output.WriteString("\n")
}

func writeCodexWorkerChecklistItem(output *strings.Builder, item EvidenceStrategyChecklistItem) {
	checked := "false"
	if item.Checked {
		checked = "true"
	}

	output.WriteString("- ")
	output.WriteString(item.Label)
	output.WriteString(": checked=")
	output.WriteString(checked)
	output.WriteString("; status=")
	output.WriteString(item.Status)

	if item.Detail != "" {
		output.WriteString("; ")
		output.WriteString(item.Detail)
	}

	output.WriteString("\n")
}

func nonEmptyWorkerContextValue(value, fallback string) string {
	if strings.TrimSpace(value) == "" {
		return fallback
	}

	return value
}

func writeCodexWorkerRepairHandoff(output *strings.Builder, opts MaterializeCodexWorkerContextOptions) {
	if opts.RepairHandoff == nil {
		return
	}

	handoff := opts.RepairHandoff

	output.WriteString("## Review Repair Handoff\n\n")
	output.WriteString("- branch: ")
	output.WriteString(handoff.Branch)
	output.WriteString("\n")
	output.WriteString("- pull_request: ")
	output.WriteString(handoff.PullRequestURL)
	output.WriteString("\n")
	output.WriteString("- review_comments_url: ")
	output.WriteString(handoff.ReviewCommentsURL)
	output.WriteString("\n")
	output.WriteString("- review_state: ")
	output.WriteString(handoff.ReviewState)
	output.WriteString("\n")
	output.WriteString("- review_result: ")
	output.WriteString(handoff.ReviewResult)
	output.WriteString("\n")
	output.WriteString("- review_round: ")
	output.WriteString(strconv.Itoa(handoff.ReviewRound))
	output.WriteString("\n\n")
	output.WriteString("Use the remote review comments as the source of detailed findings. Do not duplicate remote comments into the ticket.\n\n")
}

func loadWorkerEvidence(ctx context.Context, projectDir, ticketPath string) (*WorkerEvidenceContext, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	config, err := LoadWorkspaceConfig(projectDir)
	if err != nil {
		return nil, fmt.Errorf("load evidence strategy config: %w", err)
	}

	content, err := os.ReadFile(ticketPath)
	if err != nil {
		return nil, fmt.Errorf("read evidence strategy ticket: %w", err)
	}

	checklist, checklistError := parseWorkerEvidenceChecklist(string(content))

	ticket, err := NewTicketParser().ParseTicket(ctx, ParseTicketOptions{
		ProjectDir: projectDir,
		Path:       ticketPath,
		IndexedAt:  time.Now().UTC(),
	})
	if err != nil {
		return nil, fmt.Errorf("read evidence strategy ticket metadata: %w", err)
	}

	policy, err := DetectEvidencePolicyContext(projectDir, config.Evidence.Strategy, ticket)
	if err != nil {
		return nil, err
	}

	return &WorkerEvidenceContext{
		Config:         config.Evidence.Strategy,
		Checklist:      checklist,
		ChecklistError: checklistError,
		Policy:         policy,
	}, nil
}

func parseWorkerEvidenceChecklist(content string) (*EvidenceStrategyChecklist, string) {
	checklist, err := ParseEvidenceStrategyChecklist(content)
	if err != nil {
		return nil, err.Error()
	}

	return &checklist, ""
}

func codexWorkerRepairHandoff(ctx context.Context, projectDir, ticketPath string) (*CodexWorkerRepairHandoff, error) {
	ticket, err := NewTicketParser().ParseTicket(ctx, ParseTicketOptions{
		ProjectDir: projectDir,
		Path:       ticketPath,
		IndexedAt:  time.Now().UTC(),
	})
	if err != nil {
		return nil, fmt.Errorf("read repair handoff ticket metadata: %w", err)
	}

	if !isReviewRepairTicket(ticket) {
		return nil, nil
	}

	ref, err := ticketReviewCommandPullRequestRef(ticket)
	if err != nil {
		return nil, fmt.Errorf("repair handoff requires pull request ref: %w", err)
	}

	if strings.TrimSpace(ticket.Branch) == "" {
		return nil, fmt.Errorf("repair handoff requires branch metadata")
	}

	if strings.TrimSpace(ticket.ReviewCommentsURL) == "" {
		return nil, fmt.Errorf("repair handoff requires review_comments_url")
	}

	return &CodexWorkerRepairHandoff{
		Branch:            ticket.Branch,
		PullRequestURL:    ref.URL,
		ReviewCommentsURL: ticket.ReviewCommentsURL,
		ReviewState:       ticket.ReviewState,
		ReviewResult:      ticket.ReviewResult,
		ReviewRound:       ticket.ReviewRound,
	}, nil
}

func validateCodexWorkerContextOptions(opts MaterializeCodexWorkerContextOptions) error {
	required := []struct {
		name  string
		value string
	}{
		{name: "context path", value: opts.Path},
		{name: "execution id", value: opts.ExecutionID},
		{name: "ticket key", value: opts.TicketKey},
		{name: "ticket path", value: opts.TicketPath},
		{name: "result path", value: opts.ResultPath},
		{name: "log path", value: opts.LogPath},
		{name: "validation path", value: opts.ValidationPath},
	}

	for _, field := range required {
		if strings.TrimSpace(field.value) == "" {
			return fmt.Errorf("codex worker %s is required", field.name)
		}
	}

	return nil
}
