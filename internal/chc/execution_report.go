package chc

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

const executionReportChangedFileLimit = 25

type ExecutionReport struct {
	ExecutionID    string
	ProjectID      string
	TicketKey      string
	TicketPath     string
	Execution      ExecutionReportExecution
	Runtime        ExecutionReportRuntime
	Artifacts      ExecutionReportArtifacts
	Result         ExecutionReportResult
	Validation     ExecutionReportValidation
	ChangedFiles   []ExecutionReportChangedFile
	ChangedSummary ExecutionReportChangedFilesSummary
	FailureReason  string
	CreatedAt      time.Time
	StartedAt      *time.Time
	FinishedAt     *time.Time
	UpdatedAt      time.Time
}

type ExecutionReportExecution struct {
	Status      ExecutionStatus
	FailureKind ExecutionFailureKind
}

type ExecutionReportRuntime struct {
	Image             string
	DockerfilePath    string
	ContainerWorkdir  string
	ValidationCommand string
}

type ExecutionReportArtifacts struct {
	RunDir         string
	ResultPath     string
	LogPath        string
	ValidationPath string
	ReportPath     string
}

type ExecutionReportResult struct {
	Available bool
	Status    ExecutionResultStatus
	Summary   string
}

type ExecutionReportValidation struct {
	Command string
	Path    string
	Status  ExecutionReportValidationStatus
}

type ExecutionReportValidationStatus string

const (
	ExecutionReportValidationNotRun    ExecutionReportValidationStatus = "not_run"
	ExecutionReportValidationSucceeded ExecutionReportValidationStatus = "succeeded"
	ExecutionReportValidationFailed    ExecutionReportValidationStatus = "failed"
	ExecutionReportValidationUnknown   ExecutionReportValidationStatus = "unknown"
)

type ExecutionReportChangedFile struct {
	Path    string
	Summary string
}

type ExecutionReportChangedFilesSummary struct {
	Total     int
	Rendered  int
	Truncated bool
}

type ExecutionReportOptions struct {
	Execution ExecutionRecord
	Result    *ExecutionResult
}

func NewExecutionReport(opts ExecutionReportOptions) (ExecutionReport, error) {
	err := ValidateExecutionRecord(opts.Execution)
	if err != nil {
		return ExecutionReport{}, err
	}

	report := executionReportFromRecord(opts.Execution)

	if opts.Result == nil {
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

func RenderExecutionReportMarkdown(report ExecutionReport) string {
	var output strings.Builder

	output.WriteString("# Execution Report\n\n")
	writeExecutionReportSummary(&output, report)
	writeExecutionReportTimes(&output, report)
	writeExecutionReportRuntime(&output, report.Runtime)
	writeExecutionReportArtifacts(&output, report.Artifacts)
	writeExecutionReportValidation(&output, report.Validation)
	writeExecutionReportResult(&output, report.Result)
	writeExecutionReportChangedFiles(&output, report.ChangedFiles, report.ChangedSummary)
	writeExecutionReportFailure(&output, report)

	return output.String()
}

type ExecutionReportWriter interface {
	WriteExecutionReport(ctx context.Context, report ExecutionReport) (string, error)
}

type executionReportWriter struct{}

func NewExecutionReportWriter() ExecutionReportWriter {
	return &executionReportWriter{}
}

func ExecutionReportPath(record ExecutionRecord) string {
	if strings.TrimSpace(record.Artifacts.RunDir) == "" {
		return ""
	}

	return filepath.Join(record.Artifacts.RunDir, "execution-report.md")
}

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
