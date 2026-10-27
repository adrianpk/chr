package gate

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// ReportStatus defines a package boundary type.
type ReportStatus string

const (
	// ReportStatusPassed defines a package boundary constant.
	ReportStatusPassed ReportStatus = "passed"
	// ReportStatusFailed defines a package boundary constant.
	ReportStatusFailed ReportStatus = "failed"
)

// Report describes a package boundary value.
type Report struct {
	CreatedAt               time.Time                      `json:"created_at"`
	Status                  ReportStatus                   `json:"status"`
	Base                    ReportRef                      `json:"base"`
	Head                    ReportRef                      `json:"head"`
	MergeBase               string                         `json:"merge_base"`
	Commits                 []ReportCommit                 `json:"commits"`
	ChangedFiles            []ReportChangedFile            `json:"changed_files"`
	DiffStat                string                         `json:"diff_stat"`
	AgentManagedWork        []ReportAgentManagedWork       `json:"agent_managed_work"`
	ManualOrExternalChanges []ReportManualOrExternalChange `json:"manual_or_external_changes"`
	MissingAgentEvidence    []ReportMissingAgentEvidence   `json:"missing_agent_evidence"`
	HighRiskManualChanges   []ReportHighRiskManualChange   `json:"high_risk_manual_changes"`
	Validation              ReportValidation               `json:"validation"`
	Artifacts               ReportArtifacts                `json:"artifacts"`
	Failures                []string                       `json:"failures"`
	Warnings                []string                       `json:"warnings"`
	Info                    []string                       `json:"info"`
	Metadata                map[string]string              `json:"metadata,omitempty"`
}

// ReportRef describes a package boundary value.
type ReportRef struct {
	Ref string `json:"ref"`
	SHA string `json:"sha"`
}

// ReportCommit describes a package boundary value.
type ReportCommit struct {
	SHA     string `json:"sha"`
	Subject string `json:"subject"`
}

// ReportChangedFile describes a package boundary value.
type ReportChangedFile struct {
	Status string `json:"status"`
	Path   string `json:"path"`
	Old    string `json:"old,omitempty"`
}

// ReportAgentManagedWork describes a package boundary value.
type ReportAgentManagedWork struct {
	Path        string `json:"path"`
	Status      string `json:"status"`
	TicketKey   string `json:"ticket_key"`
	ExecutionID string `json:"execution_id"`
	ReportPath  string `json:"report_path"`
}

// ReportManualOrExternalChange describes a package boundary value.
type ReportManualOrExternalChange struct {
	Path   string `json:"path"`
	Status string `json:"status"`
}

// ReportMissingAgentEvidence describes a package boundary value.
type ReportMissingAgentEvidence struct {
	TicketKey   string `json:"ticket_key"`
	ExecutionID string `json:"execution_id"`
	ReportPath  string `json:"report_path"`
	Reason      string `json:"reason"`
}

// ReportHighRiskManualChange describes a package boundary value.
type ReportHighRiskManualChange struct {
	Path    string `json:"path"`
	Status  string `json:"status"`
	Pattern string `json:"pattern"`
}

// ReportValidation describes a package boundary value.
type ReportValidation struct {
	Command  string `json:"command"`
	ExitCode int    `json:"exit_code"`
	Output   string `json:"output"`
	Passed   bool   `json:"passed"`
}

// ReportArtifacts describes a package boundary value.
type ReportArtifacts struct {
	JSONPath     string `json:"json_path"`
	MarkdownPath string `json:"markdown_path"`
}

// ReportWriter describes a package boundary value.
type ReportWriter struct{}

// NewReportWriter constructs the corresponding package service or value.
func NewReportWriter() ReportWriter {
	return ReportWriter{}
}

// NewReport constructs the corresponding package service or value.
func NewReport(createdAt time.Time, evidence GitEvidence, validation ValidationResult, classification EvidenceClassification, artifacts ReportArtifacts) Report {
	status := ReportStatusFailed
	if validation.Passed {
		status = ReportStatusPassed
	}

	return Report{
		CreatedAt:               createdAt.UTC(),
		Status:                  status,
		Base:                    ReportRef(evidence.Base),
		Head:                    ReportRef(evidence.Head),
		MergeBase:               evidence.MergeBase,
		Commits:                 reportCommits(evidence.Commits),
		ChangedFiles:            reportChangedFiles(evidence.Files),
		DiffStat:                evidence.DiffStat,
		AgentManagedWork:        reportAgentManagedWork(classification.AgentManagedWork),
		ManualOrExternalChanges: reportManualOrExternalChanges(classification.ManualOrExternalChanges),
		MissingAgentEvidence:    reportMissingAgentEvidence(classification.MissingAgentEvidence),
		HighRiskManualChanges:   reportHighRiskManualChanges(classification.HighRiskManualChanges),
		Validation:              ReportValidation(validation),
		Artifacts:               artifacts,
	}
}

// NewFailedReport constructs the corresponding package service or value.
func NewFailedReport(createdAt time.Time, baseRef string, headRef string, artifacts ReportArtifacts, failures []string) Report {
	return Report{
		CreatedAt: createdAt.UTC(),
		Status:    ReportStatusFailed,
		Base:      ReportRef{Ref: strings.TrimSpace(baseRef)},
		Head:      ReportRef{Ref: strings.TrimSpace(headRef)},
		Artifacts: artifacts,
		Failures:  failures,
	}
}

// WriteJSON writes the requested output.
func (w ReportWriter) WriteJSON(ctx context.Context, report Report) (string, error) {
	err := ctx.Err()
	if err != nil {
		return "", err
	}

	path := strings.TrimSpace(report.Artifacts.JSONPath)
	if path == "" {
		return "", fmt.Errorf("gate JSON report path is required")
	}

	err = ensureReportWriteTarget(path)
	if err != nil {
		return "", err
	}

	content, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return "", fmt.Errorf("encode gate JSON report: %w", err)
	}

	err = os.MkdirAll(filepath.Dir(path), 0o755)
	if err != nil {
		return "", fmt.Errorf("create gate report directory: %w", err)
	}

	err = os.WriteFile(path, append(content, '\n'), 0o644)
	if err != nil {
		return "", fmt.Errorf("write gate JSON report: %w", err)
	}

	return path, nil
}

// WriteMarkdown writes the requested output.
func (w ReportWriter) WriteMarkdown(ctx context.Context, report Report) (string, error) {
	err := ctx.Err()
	if err != nil {
		return "", err
	}

	path := strings.TrimSpace(report.Artifacts.MarkdownPath)
	if path == "" {
		return "", fmt.Errorf("gate Markdown report path is required")
	}

	err = ensureReportWriteTarget(path)
	if err != nil {
		return "", err
	}

	content := RenderReportMarkdown(report)

	err = os.MkdirAll(filepath.Dir(path), 0o755)
	if err != nil {
		return "", fmt.Errorf("create gate report directory: %w", err)
	}

	err = os.WriteFile(path, []byte(content), 0o644)
	if err != nil {
		return "", fmt.Errorf("write gate Markdown report: %w", err)
	}

	return path, nil
}

// RenderReportMarkdown renders a human-readable gate report.
func RenderReportMarkdown(report Report) string {
	var output strings.Builder

	output.WriteString("# Dev to Main Gate Report\n\n")
	writeReportSummary(&output, report)
	writeReportGitEvidence(&output, report)
	writeReportAgentManagedWork(&output, report.AgentManagedWork)
	writeReportManualOrExternalChanges(&output, report.ManualOrExternalChanges)
	writeReportHighRiskManualChanges(&output, report.HighRiskManualChanges)
	writeReportMissingAgentEvidence(&output, report.MissingAgentEvidence)
	writeReportValidation(&output, report.Validation)
	writeReportMessages(&output, "Failures", report.Failures)
	writeReportMessages(&output, "Warnings", report.Warnings)
	writeReportMessages(&output, "Info", report.Info)
	writeReportArtifacts(&output, report.Artifacts)

	return output.String()
}

func ensureReportWriteTarget(path string) error {
	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}

		return fmt.Errorf("stat gate report: %w", err)
	}

	if info.IsDir() {
		return fmt.Errorf("gate report path is a directory: %s", path)
	}

	return nil
}

func writeReportSummary(output *strings.Builder, report Report) {
	output.WriteString("## Summary\n\n")
	writeReportField(output, "status", string(report.Status))
	writeReportField(output, "created_at", report.CreatedAt.UTC().Format(time.RFC3339))
	output.WriteString("\n")
}

func writeReportGitEvidence(output *strings.Builder, report Report) {
	output.WriteString("## Git Evidence\n\n")
	writeReportField(output, "base", report.Base.Ref+" "+report.Base.SHA)
	writeReportField(output, "head", report.Head.Ref+" "+report.Head.SHA)
	writeReportField(output, "merge_base", report.MergeBase)
	writeReportField(output, "diff_stat", report.DiffStat)
	writeReportField(output, "commits", fmt.Sprintf("%d", len(report.Commits)))
	writeReportField(output, "changed_files", fmt.Sprintf("%d", len(report.ChangedFiles)))
	output.WriteString("\n")

	if len(report.Commits) > 0 {
		output.WriteString("### Commits\n\n")

		for _, commit := range report.Commits {
			output.WriteString("- `" + emptyReportValue(commit.SHA) + "` " + emptyReportValue(commit.Subject) + "\n")
		}

		output.WriteString("\n")
	}

	if len(report.ChangedFiles) > 0 {
		output.WriteString("### Changed Files\n\n")

		for _, file := range report.ChangedFiles {
			if file.Old != "" {
				output.WriteString("- " + file.Status + " " + file.Old + " -> " + file.Path + "\n")

				continue
			}

			output.WriteString("- " + file.Status + " " + file.Path + "\n")
		}

		output.WriteString("\n")
	}
}

func writeReportValidation(output *strings.Builder, validation ReportValidation) {
	output.WriteString("## Validation\n\n")
	writeReportField(output, "command", validation.Command)
	writeReportField(output, "passed", fmt.Sprintf("%t", validation.Passed))
	writeReportField(output, "exit_code", fmt.Sprintf("%d", validation.ExitCode))

	if validation.Output != "" {
		output.WriteString("\n```text\n")
		output.WriteString(validation.Output)
		output.WriteString("\n```\n")
	}

	output.WriteString("\n")
}

func writeReportAgentManagedWork(output *strings.Builder, work []ReportAgentManagedWork) {
	output.WriteString("## Agent-Managed Work\n\n")

	if len(work) == 0 {
		output.WriteString("- none\n\n")

		return
	}

	for _, item := range work {
		output.WriteString("- ")
		output.WriteString(item.Status)
		output.WriteString(" ")
		output.WriteString(item.Path)
		output.WriteString(" ticket=")
		output.WriteString(emptyReportValue(item.TicketKey))
		output.WriteString(" execution=")
		output.WriteString(emptyReportValue(item.ExecutionID))
		output.WriteString(" report=")
		output.WriteString(emptyReportValue(item.ReportPath))
		output.WriteString("\n")
	}

	output.WriteString("\n")
}

func writeReportManualOrExternalChanges(output *strings.Builder, changes []ReportManualOrExternalChange) {
	output.WriteString("## Manual or External Changes\n\n")

	if len(changes) == 0 {
		output.WriteString("- none\n\n")

		return
	}

	for _, item := range changes {
		output.WriteString("- ")
		output.WriteString(item.Status)
		output.WriteString(" ")
		output.WriteString(item.Path)
		output.WriteString("\n")
	}

	output.WriteString("\n")
}

func writeReportMissingAgentEvidence(output *strings.Builder, missing []ReportMissingAgentEvidence) {
	output.WriteString("## Missing Agent Evidence\n\n")

	if len(missing) == 0 {
		output.WriteString("- none\n\n")

		return
	}

	for _, item := range missing {
		output.WriteString("- ticket=")
		output.WriteString(emptyReportValue(item.TicketKey))
		output.WriteString(" execution=")
		output.WriteString(emptyReportValue(item.ExecutionID))
		output.WriteString(" report=")
		output.WriteString(emptyReportValue(item.ReportPath))
		output.WriteString(" reason=")
		output.WriteString(emptyReportValue(item.Reason))
		output.WriteString("\n")
	}

	output.WriteString("\n")
}

func writeReportHighRiskManualChanges(output *strings.Builder, changes []ReportHighRiskManualChange) {
	output.WriteString("## High-Risk Manual Changes\n\n")

	if len(changes) == 0 {
		output.WriteString("- none\n\n")

		return
	}

	for _, item := range changes {
		output.WriteString("- ")
		output.WriteString(item.Status)
		output.WriteString(" ")
		output.WriteString(item.Path)
		output.WriteString(" pattern=")
		output.WriteString(emptyReportValue(item.Pattern))
		output.WriteString("\n")
	}

	output.WriteString("\n")
}

func writeReportMessages(output *strings.Builder, title string, messages []string) {
	output.WriteString("## " + title + "\n\n")

	if len(messages) == 0 {
		output.WriteString("- none\n\n")

		return
	}

	for _, message := range messages {
		output.WriteString("- " + message + "\n")
	}

	output.WriteString("\n")
}

func writeReportArtifacts(output *strings.Builder, artifacts ReportArtifacts) {
	output.WriteString("## Artifacts\n\n")
	writeReportField(output, "json", artifacts.JSONPath)
	writeReportField(output, "markdown", artifacts.MarkdownPath)
	output.WriteString("\n")
}

func writeReportField(output *strings.Builder, name, value string) {
	output.WriteString("- ")
	output.WriteString(name)
	output.WriteString(": ")
	output.WriteString(emptyReportValue(value))
	output.WriteString("\n")
}

func emptyReportValue(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return "none"
	}

	return value
}

func reportCommits(commits []GitCommit) []ReportCommit {
	reportCommits := make([]ReportCommit, 0, len(commits))

	for _, commit := range commits {
		reportCommits = append(reportCommits, ReportCommit(commit))
	}

	return reportCommits
}

func reportChangedFiles(files []GitChangedFile) []ReportChangedFile {
	reportFiles := make([]ReportChangedFile, 0, len(files))

	for _, file := range files {
		reportFiles = append(reportFiles, ReportChangedFile(file))
	}

	return reportFiles
}

func reportAgentManagedWork(work []AgentManagedWork) []ReportAgentManagedWork {
	reportWork := make([]ReportAgentManagedWork, 0, len(work))

	for _, item := range work {
		reportWork = append(reportWork, ReportAgentManagedWork(item))
	}

	return reportWork
}

func reportManualOrExternalChanges(changes []ManualOrExternalChange) []ReportManualOrExternalChange {
	reportChanges := make([]ReportManualOrExternalChange, 0, len(changes))

	for _, item := range changes {
		reportChanges = append(reportChanges, ReportManualOrExternalChange(item))
	}

	return reportChanges
}

func reportMissingAgentEvidence(missing []MissingAgentEvidence) []ReportMissingAgentEvidence {
	reportMissing := make([]ReportMissingAgentEvidence, 0, len(missing))

	for _, item := range missing {
		reportMissing = append(reportMissing, ReportMissingAgentEvidence(item))
	}

	return reportMissing
}

func reportHighRiskManualChanges(changes []HighRiskManualChange) []ReportHighRiskManualChange {
	reportChanges := make([]ReportHighRiskManualChange, 0, len(changes))

	for _, item := range changes {
		reportChanges = append(reportChanges, ReportHighRiskManualChange(item))
	}

	return reportChanges
}
