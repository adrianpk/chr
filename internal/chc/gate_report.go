package chc

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// GateReportStatus defines a package boundary type.
type GateReportStatus string

const (
	// GateReportStatusPassed defines a package boundary constant.
	GateReportStatusPassed GateReportStatus = "passed"
	// GateReportStatusFailed defines a package boundary constant.
	GateReportStatusFailed GateReportStatus = "failed"
)

// GateReport describes a package boundary value.
type GateReport struct {
	CreatedAt               time.Time                          `json:"created_at"`
	Status                  GateReportStatus                   `json:"status"`
	Base                    GateReportRef                      `json:"base"`
	Head                    GateReportRef                      `json:"head"`
	MergeBase               string                             `json:"merge_base"`
	Commits                 []GateReportCommit                 `json:"commits"`
	ChangedFiles            []GateReportChangedFile            `json:"changed_files"`
	DiffStat                string                             `json:"diff_stat"`
	AgentManagedWork        []GateReportAgentManagedWork       `json:"agent_managed_work"`
	ManualOrExternalChanges []GateReportManualOrExternalChange `json:"manual_or_external_changes"`
	MissingAgentEvidence    []GateReportMissingAgentEvidence   `json:"missing_agent_evidence"`
	HighRiskManualChanges   []GateReportHighRiskManualChange   `json:"high_risk_manual_changes"`
	Validation              GateReportValidation               `json:"validation"`
	Artifacts               GateReportArtifacts                `json:"artifacts"`
	Failures                []string                           `json:"failures"`
	Warnings                []string                           `json:"warnings"`
	Info                    []string                           `json:"info"`
	Metadata                map[string]string                  `json:"metadata,omitempty"`
}

// GateReportRef describes a package boundary value.
type GateReportRef struct {
	Ref string `json:"ref"`
	SHA string `json:"sha"`
}

// GateReportCommit describes a package boundary value.
type GateReportCommit struct {
	SHA     string `json:"sha"`
	Subject string `json:"subject"`
}

// GateReportChangedFile describes a package boundary value.
type GateReportChangedFile struct {
	Status string `json:"status"`
	Path   string `json:"path"`
	Old    string `json:"old,omitempty"`
}

// GateReportAgentManagedWork describes a package boundary value.
type GateReportAgentManagedWork struct {
	Path        string `json:"path"`
	Status      string `json:"status"`
	TicketKey   string `json:"ticket_key"`
	ExecutionID string `json:"execution_id"`
	ReportPath  string `json:"report_path"`
}

// GateReportManualOrExternalChange describes a package boundary value.
type GateReportManualOrExternalChange struct {
	Path   string `json:"path"`
	Status string `json:"status"`
}

// GateReportMissingAgentEvidence describes a package boundary value.
type GateReportMissingAgentEvidence struct {
	TicketKey   string `json:"ticket_key"`
	ExecutionID string `json:"execution_id"`
	ReportPath  string `json:"report_path"`
	Reason      string `json:"reason"`
}

// GateReportHighRiskManualChange describes a package boundary value.
type GateReportHighRiskManualChange struct {
	Path    string `json:"path"`
	Status  string `json:"status"`
	Pattern string `json:"pattern"`
}

// GateReportValidation describes a package boundary value.
type GateReportValidation struct {
	Command  string `json:"command"`
	ExitCode int    `json:"exit_code"`
	Output   string `json:"output"`
	Passed   bool   `json:"passed"`
}

// GateReportArtifacts describes a package boundary value.
type GateReportArtifacts struct {
	JSONPath     string `json:"json_path"`
	MarkdownPath string `json:"markdown_path"`
}

// GateReportWriter describes a package boundary value.
type GateReportWriter struct{}

// NewGateReportWriter constructs the corresponding package service or value.
func NewGateReportWriter() GateReportWriter {
	return GateReportWriter{}
}

// NewGateReport constructs the corresponding package service or value.
func NewGateReport(
	createdAt time.Time,
	evidence GateGitEvidence,
	validation GateValidationResult,
	classification GateEvidenceClassification,
	artifacts GateReportArtifacts,
) GateReport {
	status := GateReportStatusFailed
	if validation.Passed {
		status = GateReportStatusPassed
	}

	return GateReport{
		CreatedAt:               createdAt.UTC(),
		Status:                  status,
		Base:                    GateReportRef(evidence.Base),
		Head:                    GateReportRef(evidence.Head),
		MergeBase:               evidence.MergeBase,
		Commits:                 gateReportCommits(evidence.Commits),
		ChangedFiles:            gateReportChangedFiles(evidence.Files),
		DiffStat:                evidence.DiffStat,
		AgentManagedWork:        gateReportAgentManagedWork(classification.AgentManagedWork),
		ManualOrExternalChanges: gateReportManualOrExternalChanges(classification.ManualOrExternalChanges),
		MissingAgentEvidence:    gateReportMissingAgentEvidence(classification.MissingAgentEvidence),
		HighRiskManualChanges:   gateReportHighRiskManualChanges(classification.HighRiskManualChanges),
		Validation:              GateReportValidation(validation),
		Artifacts:               artifacts,
	}
}

// NewFailedGateReport constructs the corresponding package service or value.
func NewFailedGateReport(
	createdAt time.Time,
	baseRef string,
	headRef string,
	artifacts GateReportArtifacts,
	failures []string,
) GateReport {
	return GateReport{
		CreatedAt: createdAt.UTC(),
		Status:    GateReportStatusFailed,
		Base:      GateReportRef{Ref: strings.TrimSpace(baseRef)},
		Head:      GateReportRef{Ref: strings.TrimSpace(headRef)},
		Artifacts: artifacts,
		Failures:  failures,
	}
}

// NewGateReportArtifacts constructs the corresponding package service or value.
func NewGateReportArtifacts(projectDir string, createdAt time.Time) GateReportArtifacts {
	fileBase := "dev-main-" + createdAt.UTC().Format("20060102T150405Z")

	return GateReportArtifacts{
		JSONPath:     UserPath(projectDir, "reports", "gates", fileBase+".json"),
		MarkdownPath: UserPath(projectDir, "reports", "gates", fileBase+".md"),
	}
}

// WriteJSON writes the requested output.
func (w GateReportWriter) WriteJSON(ctx context.Context, report GateReport) (string, error) {
	err := ctx.Err()
	if err != nil {
		return "", err
	}

	path := strings.TrimSpace(report.Artifacts.JSONPath)
	if path == "" {
		return "", fmt.Errorf("gate JSON report path is required")
	}

	err = ensureGateReportWriteTarget(path)
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
func (w GateReportWriter) WriteMarkdown(ctx context.Context, report GateReport) (string, error) {
	err := ctx.Err()
	if err != nil {
		return "", err
	}

	path := strings.TrimSpace(report.Artifacts.MarkdownPath)
	if path == "" {
		return "", fmt.Errorf("gate Markdown report path is required")
	}

	err = ensureGateReportWriteTarget(path)
	if err != nil {
		return "", err
	}

	content := RenderGateReportMarkdown(report)

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

// RenderGateReportMarkdown renders the corresponding package data.
func RenderGateReportMarkdown(report GateReport) string {
	var output strings.Builder

	output.WriteString("# Dev to Main Gate Report\n\n")
	writeGateReportSummary(&output, report)
	writeGateReportGitEvidence(&output, report)
	writeGateReportAgentManagedWork(&output, report.AgentManagedWork)
	writeGateReportManualOrExternalChanges(&output, report.ManualOrExternalChanges)
	writeGateReportHighRiskManualChanges(&output, report.HighRiskManualChanges)
	writeGateReportMissingAgentEvidence(&output, report.MissingAgentEvidence)
	writeGateReportValidation(&output, report.Validation)
	writeGateReportMessages(&output, "Failures", report.Failures)
	writeGateReportMessages(&output, "Warnings", report.Warnings)
	writeGateReportMessages(&output, "Info", report.Info)
	writeGateReportArtifacts(&output, report.Artifacts)

	return output.String()
}

func ensureGateReportWriteTarget(path string) error {
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

func writeGateReportSummary(output *strings.Builder, report GateReport) {
	output.WriteString("## Summary\n\n")
	writeGateReportField(output, "status", string(report.Status))
	writeGateReportField(output, "created_at", report.CreatedAt.UTC().Format(time.RFC3339))
	output.WriteString("\n")
}

func writeGateReportGitEvidence(output *strings.Builder, report GateReport) {
	output.WriteString("## Git Evidence\n\n")
	writeGateReportField(output, "base", report.Base.Ref+" "+report.Base.SHA)
	writeGateReportField(output, "head", report.Head.Ref+" "+report.Head.SHA)
	writeGateReportField(output, "merge_base", report.MergeBase)
	writeGateReportField(output, "diff_stat", report.DiffStat)
	writeGateReportField(output, "commits", fmt.Sprintf("%d", len(report.Commits)))
	writeGateReportField(output, "changed_files", fmt.Sprintf("%d", len(report.ChangedFiles)))
	output.WriteString("\n")

	if len(report.Commits) > 0 {
		output.WriteString("### Commits\n\n")

		for _, commit := range report.Commits {
			output.WriteString("- `" + emptyGateReportValue(commit.SHA) + "` " + emptyGateReportValue(commit.Subject) + "\n")
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

func writeGateReportValidation(output *strings.Builder, validation GateReportValidation) {
	output.WriteString("## Validation\n\n")
	writeGateReportField(output, "command", validation.Command)
	writeGateReportField(output, "passed", fmt.Sprintf("%t", validation.Passed))
	writeGateReportField(output, "exit_code", fmt.Sprintf("%d", validation.ExitCode))

	if validation.Output != "" {
		output.WriteString("\n```text\n")
		output.WriteString(validation.Output)
		output.WriteString("\n```\n")
	}

	output.WriteString("\n")
}

func writeGateReportAgentManagedWork(output *strings.Builder, work []GateReportAgentManagedWork) {
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
		output.WriteString(emptyGateReportValue(item.TicketKey))
		output.WriteString(" execution=")
		output.WriteString(emptyGateReportValue(item.ExecutionID))
		output.WriteString(" report=")
		output.WriteString(emptyGateReportValue(item.ReportPath))
		output.WriteString("\n")
	}

	output.WriteString("\n")
}

func writeGateReportManualOrExternalChanges(output *strings.Builder, changes []GateReportManualOrExternalChange) {
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

func writeGateReportMissingAgentEvidence(output *strings.Builder, missing []GateReportMissingAgentEvidence) {
	output.WriteString("## Missing Agent Evidence\n\n")

	if len(missing) == 0 {
		output.WriteString("- none\n\n")

		return
	}

	for _, item := range missing {
		output.WriteString("- ticket=")
		output.WriteString(emptyGateReportValue(item.TicketKey))
		output.WriteString(" execution=")
		output.WriteString(emptyGateReportValue(item.ExecutionID))
		output.WriteString(" report=")
		output.WriteString(emptyGateReportValue(item.ReportPath))
		output.WriteString(" reason=")
		output.WriteString(emptyGateReportValue(item.Reason))
		output.WriteString("\n")
	}

	output.WriteString("\n")
}

func writeGateReportHighRiskManualChanges(output *strings.Builder, changes []GateReportHighRiskManualChange) {
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
		output.WriteString(emptyGateReportValue(item.Pattern))
		output.WriteString("\n")
	}

	output.WriteString("\n")
}

func writeGateReportMessages(output *strings.Builder, title string, messages []string) {
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

func writeGateReportArtifacts(output *strings.Builder, artifacts GateReportArtifacts) {
	output.WriteString("## Artifacts\n\n")
	writeGateReportField(output, "json", artifacts.JSONPath)
	writeGateReportField(output, "markdown", artifacts.MarkdownPath)
	output.WriteString("\n")
}

func writeGateReportField(output *strings.Builder, name, value string) {
	output.WriteString("- ")
	output.WriteString(name)
	output.WriteString(": ")
	output.WriteString(emptyGateReportValue(value))
	output.WriteString("\n")
}

func emptyGateReportValue(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return "none"
	}

	return value
}

func gateReportCommits(commits []GateGitCommit) []GateReportCommit {
	reportCommits := make([]GateReportCommit, 0, len(commits))

	for _, commit := range commits {
		reportCommits = append(reportCommits, GateReportCommit(commit))
	}

	return reportCommits
}

func gateReportChangedFiles(files []GateGitChangedFile) []GateReportChangedFile {
	reportFiles := make([]GateReportChangedFile, 0, len(files))

	for _, file := range files {
		reportFiles = append(reportFiles, GateReportChangedFile(file))
	}

	return reportFiles
}

func gateReportAgentManagedWork(work []GateAgentManagedWork) []GateReportAgentManagedWork {
	reportWork := make([]GateReportAgentManagedWork, 0, len(work))

	for _, item := range work {
		reportWork = append(reportWork, GateReportAgentManagedWork(item))
	}

	return reportWork
}

func gateReportManualOrExternalChanges(changes []GateManualOrExternalChange) []GateReportManualOrExternalChange {
	reportChanges := make([]GateReportManualOrExternalChange, 0, len(changes))

	for _, item := range changes {
		reportChanges = append(reportChanges, GateReportManualOrExternalChange(item))
	}

	return reportChanges
}

func gateReportMissingAgentEvidence(missing []GateMissingAgentEvidence) []GateReportMissingAgentEvidence {
	reportMissing := make([]GateReportMissingAgentEvidence, 0, len(missing))

	for _, item := range missing {
		reportMissing = append(reportMissing, GateReportMissingAgentEvidence(item))
	}

	return reportMissing
}

func gateReportHighRiskManualChanges(changes []GateHighRiskManualChange) []GateReportHighRiskManualChange {
	reportChanges := make([]GateReportHighRiskManualChange, 0, len(changes))

	for _, item := range changes {
		reportChanges = append(reportChanges, GateReportHighRiskManualChange(item))
	}

	return reportChanges
}
