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

type GateReportStatus string

const (
	GateReportStatusPassed GateReportStatus = "passed"
	GateReportStatusFailed GateReportStatus = "failed"
)

type GateReport struct {
	CreatedAt    time.Time               `json:"created_at"`
	Status       GateReportStatus        `json:"status"`
	Base         GateReportRef           `json:"base"`
	Head         GateReportRef           `json:"head"`
	MergeBase    string                  `json:"merge_base"`
	Commits      []GateReportCommit      `json:"commits"`
	ChangedFiles []GateReportChangedFile `json:"changed_files"`
	DiffStat     string                  `json:"diff_stat"`
	Validation   GateReportValidation    `json:"validation"`
	Artifacts    GateReportArtifacts     `json:"artifacts"`
	Failures     []string                `json:"failures"`
	Warnings     []string                `json:"warnings"`
	Info         []string                `json:"info"`
	Metadata     map[string]string       `json:"metadata,omitempty"`
}

type GateReportRef struct {
	Ref string `json:"ref"`
	SHA string `json:"sha"`
}

type GateReportCommit struct {
	SHA     string `json:"sha"`
	Subject string `json:"subject"`
}

type GateReportChangedFile struct {
	Status string `json:"status"`
	Path   string `json:"path"`
	Old    string `json:"old,omitempty"`
}

type GateReportValidation struct {
	Command  string `json:"command"`
	ExitCode int    `json:"exit_code"`
	Output   string `json:"output"`
	Passed   bool   `json:"passed"`
}

type GateReportArtifacts struct {
	JSONPath     string `json:"json_path"`
	MarkdownPath string `json:"markdown_path"`
}

type GateReportWriter struct{}

func NewGateReportWriter() GateReportWriter {
	return GateReportWriter{}
}

func NewGateReport(
	createdAt time.Time,
	evidence GateGitEvidence,
	validation GateValidationResult,
	artifacts GateReportArtifacts,
) GateReport {
	status := GateReportStatusFailed
	if validation.Passed {
		status = GateReportStatusPassed
	}

	return GateReport{
		CreatedAt:    createdAt.UTC(),
		Status:       status,
		Base:         GateReportRef(evidence.Base),
		Head:         GateReportRef(evidence.Head),
		MergeBase:    evidence.MergeBase,
		Commits:      gateReportCommits(evidence.Commits),
		ChangedFiles: gateReportChangedFiles(evidence.Files),
		DiffStat:     evidence.DiffStat,
		Validation:   GateReportValidation(validation),
		Artifacts:    artifacts,
	}
}

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

func NewGateReportArtifacts(projectDir string, createdAt time.Time) GateReportArtifacts {
	fileBase := "dev-main-" + createdAt.UTC().Format("20060102T150405Z")

	return GateReportArtifacts{
		JSONPath:     UserPath(projectDir, "reports", "gates", fileBase+".json"),
		MarkdownPath: UserPath(projectDir, "reports", "gates", fileBase+".md"),
	}
}

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
