package core

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	workspacecore "forge.adrianpk.com/chr/internal/core/workspace"
)

const executionReportChangedFileLimit = 25

// ExecutionReport describes a package boundary value.
type ExecutionReport struct {
	ExecutionID    string
	ProjectID      string
	TicketKey      string
	TicketPath     string
	Execution      ExecutionReportExecution
	Runtime        ExecutionReportRuntime
	Workspace      ExecutionReportWorkspace
	Artifacts      ExecutionReportArtifacts
	Result         ExecutionReportResult
	Validation     ExecutionReportValidation
	Evidence       ExecutionReportEvidence
	ChangedFiles   []ExecutionReportChangedFile
	ChangedSummary ExecutionReportChangedFilesSummary
	FailureReason  string
	CreatedAt      time.Time
	StartedAt      *time.Time
	FinishedAt     *time.Time
	UpdatedAt      time.Time
}

// ExecutionReportExecution describes a package boundary value.
type ExecutionReportExecution struct {
	Status      ExecutionStatus
	FailureKind ExecutionFailureKind
}

// ExecutionReportRuntime describes a package boundary value.
type ExecutionReportRuntime struct {
	Image             string
	DockerfilePath    string
	ContainerWorkdir  string
	ValidationCommand string
}

// ExecutionReportWorkspace describes execution workspace metadata.
type ExecutionReportWorkspace struct {
	Mode                  workspacecore.WorkWorkspaceMode
	ProjectCheckoutDir    string
	ExecutionWorkspaceDir string
	BaseRef               string
	BaseSHA               string
	TicketBranch          string
}

// ExecutionReportArtifacts describes a package boundary value.
type ExecutionReportArtifacts struct {
	RunDir         string
	ResultPath     string
	LogPath        string
	ValidationPath string
	ReportPath     string
}

// ExecutionReportResult describes a package boundary value.
type ExecutionReportResult struct {
	Available bool
	Status    ExecutionResultStatus
	Summary   string
}

// ExecutionReportValidation describes a package boundary value.
type ExecutionReportValidation struct {
	Command string
	Path    string
	Status  ExecutionReportValidationStatus
}

// ExecutionReportEvidence describes copied strategy and evidence result data.
type ExecutionReportEvidence struct {
	Enabled         bool
	ResultAvailable bool
	Strategy        string
	Produced        string
	Validation      string
	Deviations      []string
	RemainingRisk   []string
	ChecklistStatus string
	ChecklistError  string
	ChecklistItems  []ExecutionReportEvidenceChecklistItem
}

// ExecutionReportEvidenceChecklistItem describes one copied strategy checklist item.
type ExecutionReportEvidenceChecklistItem struct {
	Label   string
	Checked bool
	Status  string
	Detail  string
}

// ExecutionReportValidationStatus defines a package boundary type.
type ExecutionReportValidationStatus string

const (
	// ExecutionReportValidationNotRun defines a package boundary constant.
	ExecutionReportValidationNotRun ExecutionReportValidationStatus = "not_run"
	// ExecutionReportValidationSucceeded defines a package boundary constant.
	ExecutionReportValidationSucceeded ExecutionReportValidationStatus = "succeeded"
	// ExecutionReportValidationFailed defines a package boundary constant.
	ExecutionReportValidationFailed ExecutionReportValidationStatus = "failed"
	// ExecutionReportValidationUnknown defines a package boundary constant.
	ExecutionReportValidationUnknown ExecutionReportValidationStatus = "unknown"
)

// ExecutionReportChangedFile describes a package boundary value.
type ExecutionReportChangedFile struct {
	Path    string
	Summary string
}

// ExecutionReportChangedFilesSummary describes a package boundary value.
type ExecutionReportChangedFilesSummary struct {
	Total     int
	Rendered  int
	Truncated bool
}

// ExecutionReportOptions describes a package boundary value.
type ExecutionReportOptions struct {
	Execution              ExecutionRecord
	Result                 *ExecutionResult
	IncludeEvidence        bool
	EvidenceChecklist      *EvidenceStrategyChecklist
	EvidenceChecklistError string
}

// NewExecutionReport constructs the corresponding package service or value.
func NewExecutionReport(opts ExecutionReportOptions) (ExecutionReport, error) {
	err := ValidateExecutionRecord(opts.Execution)
	if err != nil {
		return ExecutionReport{}, err
	}

	report := executionReportFromRecord(opts.Execution)

	if opts.Result == nil {
		if opts.IncludeEvidence {
			report.Evidence = executionReportEvidence(opts)
		}

		return report, nil
	}

	err = validateExecutionReportResult(opts.Execution, *opts.Result)
	if err != nil {
		return ExecutionReport{}, err
	}

	report.Result = ExecutionReportResult{
		Available: true,
		Status:    opts.Result.Status,
		Summary:   opts.Result.Summary,
	}

	report.ChangedFiles, report.ChangedSummary = executionReportChangedFiles(*opts.Result)

	if opts.IncludeEvidence {
		report.Evidence = executionReportEvidence(opts)
	}

	return report, nil
}

func executionReportFromRecord(record ExecutionRecord) ExecutionReport {
	return ExecutionReport{
		ExecutionID: record.ID,
		ProjectID:   record.ProjectID,
		TicketKey:   record.TicketKey,
		TicketPath:  record.TicketPath,
		Execution: ExecutionReportExecution{
			Status:      record.Status,
			FailureKind: ExecutionFailureKindFromReason(record.FailureReason),
		},
		Runtime: ExecutionReportRuntime{
			Image:             record.Runtime.Image,
			DockerfilePath:    record.Runtime.DockerfilePath,
			ContainerWorkdir:  record.Runtime.ContainerWorkdir,
			ValidationCommand: record.Runtime.ValidationCommand,
		},
		Workspace: ExecutionReportWorkspace{
			Mode:                  record.Workspace.Mode,
			ProjectCheckoutDir:    record.Workspace.ProjectCheckoutDir,
			ExecutionWorkspaceDir: record.Workspace.ExecutionWorkspaceDir,
			BaseRef:               record.Workspace.BaseRef,
			BaseSHA:               record.Workspace.BaseSHA,
			TicketBranch:          record.Workspace.TicketBranch,
		},
		Artifacts: ExecutionReportArtifacts{
			RunDir:         record.Artifacts.RunDir,
			ResultPath:     record.Artifacts.ResultPath,
			LogPath:        record.Artifacts.LogPath,
			ValidationPath: record.Artifacts.ValidationPath,
			ReportPath:     ExecutionReportPath(record),
		},
		Validation: ExecutionReportValidation{
			Command: record.Runtime.ValidationCommand,
			Path:    record.Artifacts.ValidationPath,
			Status:  executionReportValidationStatus(record),
		},
		FailureReason: record.FailureReason,
		CreatedAt:     record.CreatedAt,
		StartedAt:     record.StartedAt,
		FinishedAt:    record.FinishedAt,
		UpdatedAt:     record.UpdatedAt,
	}
}

func validateExecutionReportResult(record ExecutionRecord, result ExecutionResult) error {
	err := ValidateExecutionResult(result)
	if err != nil {
		return err
	}

	if record.ID != result.ExecutionID {
		return fmt.Errorf("execution report result execution id mismatch: %s", result.ExecutionID)
	}

	if record.TicketKey != result.TicketKey {
		return fmt.Errorf("execution report result ticket key mismatch: %s", result.TicketKey)
	}

	return nil
}

func executionReportChangedFiles(result ExecutionResult) ([]ExecutionReportChangedFile, ExecutionReportChangedFilesSummary) {
	changedFiles := make([]ExecutionReportChangedFile, 0, len(result.ChangedFiles))

	for _, changedFile := range result.ChangedFiles {
		changedFiles = append(changedFiles, ExecutionReportChangedFile(changedFile))
	}

	sort.SliceStable(changedFiles, func(i, j int) bool {
		return changedFiles[i].Path < changedFiles[j].Path
	})

	summary := ExecutionReportChangedFilesSummary{
		Total:    len(changedFiles),
		Rendered: len(changedFiles),
	}

	if len(changedFiles) > executionReportChangedFileLimit {
		changedFiles = changedFiles[:executionReportChangedFileLimit]
		summary.Rendered = executionReportChangedFileLimit
		summary.Truncated = true
	}

	return changedFiles, summary
}

func executionReportEvidence(opts ExecutionReportOptions) ExecutionReportEvidence {
	evidence := ExecutionReportEvidence{
		Enabled:         true,
		ChecklistStatus: executionReportEvidenceChecklistStatus(opts.EvidenceChecklist, opts.EvidenceChecklistError),
		ChecklistError:  strings.TrimSpace(opts.EvidenceChecklistError),
		ChecklistItems:  executionReportEvidenceChecklistItems(opts.EvidenceChecklist),
	}

	if opts.Result == nil {
		return evidence
	}

	evidence.ResultAvailable = true
	evidence.Strategy = opts.Result.EvidenceSummary.Strategy
	evidence.Produced = opts.Result.EvidenceSummary.Produced
	evidence.Validation = opts.Result.EvidenceSummary.Validation
	evidence.Deviations = append([]string(nil), opts.Result.EvidenceSummary.Deviations...)
	evidence.RemainingRisk = append([]string(nil), opts.Result.EvidenceSummary.RemainingRisk...)

	return evidence
}

func executionReportEvidenceChecklistStatus(checklist *EvidenceStrategyChecklist, err string) string {
	if strings.TrimSpace(err) != "" {
		return "malformed"
	}

	if checklist == nil || !checklist.Found {
		return "missing"
	}

	return "present"
}

func executionReportEvidenceChecklistItems(checklist *EvidenceStrategyChecklist) []ExecutionReportEvidenceChecklistItem {
	if checklist == nil || !checklist.Found {
		return nil
	}

	items := make([]ExecutionReportEvidenceChecklistItem, 0, len(EvidenceStrategyChecklistLabels()))
	for _, label := range EvidenceStrategyChecklistLabels() {
		item, ok := checklist.Item(label)
		if !ok {
			continue
		}

		items = append(items, ExecutionReportEvidenceChecklistItem(item))
	}

	return items
}

// RenderExecutionReportMarkdown renders the corresponding package data.
func RenderExecutionReportMarkdown(report ExecutionReport) string {
	var output strings.Builder

	output.WriteString("# Execution Report\n\n")
	writeExecutionReportSummary(&output, report)
	writeExecutionReportTimes(&output, report)
	writeExecutionReportRuntime(&output, report.Runtime)
	writeExecutionReportWorkspace(&output, report.Workspace)
	writeExecutionReportArtifacts(&output, report.Artifacts)
	writeExecutionReportValidation(&output, report.Validation)
	writeExecutionReportResult(&output, report.Result)
	writeExecutionReportEvidence(&output, report.Evidence)
	writeExecutionReportChangedFiles(&output, report.ChangedFiles, report.ChangedSummary)
	writeExecutionReportFailure(&output, report)

	return output.String()
}

// ExecutionReportWriter defines a package boundary interface.
type ExecutionReportWriter interface {
	WriteExecutionReport(ctx context.Context, report ExecutionReport) (string, error)
}

type executionReportWriter struct{}

// NewExecutionReportWriter constructs the corresponding package service or value.
func NewExecutionReportWriter() ExecutionReportWriter {
	return &executionReportWriter{}
}

// ExecutionReportPath executes the corresponding package operation.
func ExecutionReportPath(record ExecutionRecord) string {
	if strings.TrimSpace(record.Artifacts.RunDir) == "" {
		return ""
	}

	if strings.TrimSpace(record.Artifacts.ArtifactID) != "" {
		return filepath.Join(record.Artifacts.RunDir, "evidence", "execution-report.md")
	}

	return filepath.Join(record.Artifacts.RunDir, "execution-report.md")
}

// WriteExecutionReport writes the requested output.
func (w *executionReportWriter) WriteExecutionReport(
	ctx context.Context,
	report ExecutionReport,
) (string, error) {
	err := ctx.Err()
	if err != nil {
		return "", err
	}

	path := strings.TrimSpace(report.Artifacts.ReportPath)
	if path == "" {
		return "", fmt.Errorf("execution report path is required")
	}

	err = ensureExecutionReportWriteTarget(path)
	if err != nil {
		return "", err
	}

	content := RenderExecutionReportMarkdown(report)

	err = os.MkdirAll(filepath.Dir(path), 0o755)
	if err != nil {
		return "", fmt.Errorf("create execution report directory: %w", err)
	}

	err = os.WriteFile(path, []byte(content), 0o644)
	if err != nil {
		return "", fmt.Errorf("write execution report: %w", err)
	}

	return path, nil
}

func ensureExecutionReportWriteTarget(path string) error {
	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}

		return fmt.Errorf("stat execution report: %w", err)
	}

	if info.IsDir() {
		return fmt.Errorf("execution report path is a directory: %s", path)
	}

	content, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("read existing execution report: %w", err)
	}

	if !strings.HasPrefix(string(content), "# Execution Report\n") {
		return fmt.Errorf("refuse to overwrite non-report file: %s", path)
	}

	return nil
}

func writeExecutionReportSummary(output *strings.Builder, report ExecutionReport) {
	output.WriteString("## Summary\n\n")
	writeExecutionReportField(output, "execution", report.ExecutionID)
	writeExecutionReportField(output, "project", report.ProjectID)
	writeExecutionReportField(output, "ticket", report.TicketKey)
	writeExecutionReportField(output, "ticket_path", report.TicketPath)
	writeExecutionReportField(output, "status", string(report.Execution.Status))
	writeExecutionReportField(output, "failure_kind", string(report.Execution.FailureKind))
	output.WriteString("\n")
}

func writeExecutionReportTimes(output *strings.Builder, report ExecutionReport) {
	output.WriteString("## Times\n\n")
	writeExecutionReportField(output, "created_at", formatExecutionReportTime(report.CreatedAt))
	writeExecutionReportField(output, "started_at", formatExecutionReportOptionalTime(report.StartedAt))
	writeExecutionReportField(output, "finished_at", formatExecutionReportOptionalTime(report.FinishedAt))
	writeExecutionReportField(output, "updated_at", formatExecutionReportTime(report.UpdatedAt))
	output.WriteString("\n")
}

func writeExecutionReportRuntime(output *strings.Builder, runtime ExecutionReportRuntime) {
	output.WriteString("## Runtime\n\n")
	writeExecutionReportField(output, "image", runtime.Image)
	writeExecutionReportField(output, "dockerfile", runtime.DockerfilePath)
	writeExecutionReportField(output, "container_workdir", runtime.ContainerWorkdir)
	writeExecutionReportField(output, "validation_command", runtime.ValidationCommand)
	output.WriteString("\n")
}

func writeExecutionReportWorkspace(output *strings.Builder, workspace ExecutionReportWorkspace) {
	output.WriteString("## Workspace\n\n")
	writeExecutionReportField(output, "mode", string(workspace.Mode))
	writeExecutionReportField(output, "project_checkout_dir", workspace.ProjectCheckoutDir)
	writeExecutionReportField(output, "execution_workspace_dir", workspace.ExecutionWorkspaceDir)
	writeExecutionReportField(output, "base_ref", workspace.BaseRef)
	writeExecutionReportField(output, "base_sha", workspace.BaseSHA)
	writeExecutionReportField(output, "ticket_branch", workspace.TicketBranch)
	output.WriteString("\n")
}

func writeExecutionReportArtifacts(output *strings.Builder, artifacts ExecutionReportArtifacts) {
	output.WriteString("## Artifacts\n\n")
	writeExecutionReportField(output, "run_dir", artifacts.RunDir)
	writeExecutionReportField(output, "result", artifacts.ResultPath)
	writeExecutionReportField(output, "log", artifacts.LogPath)
	writeExecutionReportField(output, "validation", artifacts.ValidationPath)
	writeExecutionReportField(output, "report", artifacts.ReportPath)
	output.WriteString("\n")
}

func writeExecutionReportValidation(output *strings.Builder, validation ExecutionReportValidation) {
	output.WriteString("## Validation\n\n")
	writeExecutionReportField(output, "status", string(validation.Status))
	writeExecutionReportField(output, "command", validation.Command)
	writeExecutionReportField(output, "path", validation.Path)
	output.WriteString("\n")
}

func writeExecutionReportResult(output *strings.Builder, result ExecutionReportResult) {
	output.WriteString("## Result\n\n")

	if !result.Available {
		output.WriteString("No execution result was available.\n\n")

		return
	}

	writeExecutionReportField(output, "status", string(result.Status))
	writeExecutionReportField(output, "summary", result.Summary)
	output.WriteString("\n")
}

func writeExecutionReportEvidence(output *strings.Builder, evidence ExecutionReportEvidence) {
	if !evidence.Enabled {
		return
	}

	output.WriteString("## Evidence Strategy\n\n")
	output.WriteString("### Result Summary\n\n")

	if !evidence.ResultAvailable {
		output.WriteString("No execution evidence summary was available.\n\n")
	} else {
		writeExecutionReportField(output, "strategy", evidence.Strategy)
		writeExecutionReportField(output, "produced", evidence.Produced)
		writeExecutionReportField(output, "validation", evidence.Validation)
		writeExecutionReportList(output, "deviations", evidence.Deviations)
		writeExecutionReportList(output, "remaining_risk", evidence.RemainingRisk)
		output.WriteString("\n")
	}

	output.WriteString("### Ticket Checklist\n\n")
	writeExecutionReportField(output, "status", evidence.ChecklistStatus)

	if strings.TrimSpace(evidence.ChecklistError) != "" {
		writeExecutionReportField(output, "error", evidence.ChecklistError)
	}

	if len(evidence.ChecklistItems) == 0 {
		output.WriteString("\n")

		return
	}

	output.WriteString("\n")

	for _, item := range evidence.ChecklistItems {
		writeExecutionReportEvidenceChecklistItem(output, item)
	}

	output.WriteString("\n")
}

func writeExecutionReportEvidenceChecklistItem(output *strings.Builder, item ExecutionReportEvidenceChecklistItem) {
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

	if strings.TrimSpace(item.Detail) != "" {
		output.WriteString("; ")
		output.WriteString(item.Detail)
	}

	output.WriteString("\n")
}

func writeExecutionReportChangedFiles(
	output *strings.Builder,
	changedFiles []ExecutionReportChangedFile,
	summary ExecutionReportChangedFilesSummary,
) {
	output.WriteString("## Changed Files\n\n")
	writeExecutionReportField(output, "total", fmt.Sprintf("%d", summary.Total))
	writeExecutionReportField(output, "rendered", fmt.Sprintf("%d", summary.Rendered))
	writeExecutionReportField(output, "truncated", fmt.Sprintf("%t", summary.Truncated))
	output.WriteString("\n")

	if len(changedFiles) == 0 {
		output.WriteString("No changed files were reported.\n\n")

		return
	}

	for _, changedFile := range changedFiles {
		output.WriteString("- ")
		output.WriteString(emptyExecutionReportValue(changedFile.Path))

		summary := strings.TrimSpace(changedFile.Summary)
		if summary != "" {
			output.WriteString(": ")
			output.WriteString(summary)
		}

		output.WriteString("\n")
	}

	output.WriteString("\n")
}

func executionReportValidationStatus(record ExecutionRecord) ExecutionReportValidationStatus {
	if strings.TrimSpace(record.Artifacts.ValidationPath) == "" {
		return ExecutionReportValidationNotRun
	}

	switch record.Status {
	case ExecutionStatusSucceeded:
		return ExecutionReportValidationSucceeded
	case ExecutionStatusFailed,
		ExecutionStatusBlocked:
		return ExecutionReportValidationFailed
	case ExecutionStatusPending,
		ExecutionStatusRunning:
		return ExecutionReportValidationUnknown
	default:
		return ExecutionReportValidationUnknown
	}
}

func writeExecutionReportFailure(output *strings.Builder, report ExecutionReport) {
	output.WriteString("## Failure\n\n")

	if strings.TrimSpace(report.FailureReason) == "" {
		output.WriteString("No failure was recorded.\n")

		return
	}

	writeExecutionReportField(output, "reason", report.FailureReason)
}

func writeExecutionReportList(output *strings.Builder, name string, values []string) {
	if len(values) == 0 {
		writeExecutionReportField(output, name, "none")

		return
	}

	output.WriteString("- ")
	output.WriteString(name)
	output.WriteString(":\n")

	for _, value := range values {
		output.WriteString("  - ")
		output.WriteString(emptyExecutionReportValue(value))
		output.WriteString("\n")
	}
}

func writeExecutionReportField(output *strings.Builder, name, value string) {
	output.WriteString("- ")
	output.WriteString(name)
	output.WriteString(": ")
	output.WriteString(emptyExecutionReportValue(value))
	output.WriteString("\n")
}

func formatExecutionReportTime(value time.Time) string {
	if value.IsZero() {
		return ""
	}

	return value.UTC().Format(time.RFC3339)
}

func formatExecutionReportOptionalTime(value *time.Time) string {
	if value == nil {
		return ""
	}

	return formatExecutionReportTime(*value)
}

func emptyExecutionReportValue(value string) string {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return "none"
	}

	return trimmed
}
