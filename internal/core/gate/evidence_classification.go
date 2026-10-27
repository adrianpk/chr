package gate

import (
	"context"
	"fmt"
	"path"
	"path/filepath"
	"strings"
)

// EvidenceClassifier describes a package boundary value.
type EvidenceClassifier struct {
	LoadReports func(context.Context, string) ([]ExecutionReport, error)
}

// EvidenceClassificationRequest describes a package boundary value.
type EvidenceClassificationRequest struct {
	ProjectDir    string
	Files         []GitChangedFile
	HighRiskPaths []string
}

// EvidenceClassification describes a package boundary value.
type EvidenceClassification struct {
	AgentManagedWork        []AgentManagedWork
	ManualOrExternalChanges []ManualOrExternalChange
	MissingAgentEvidence    []MissingAgentEvidence
	HighRiskManualChanges   []HighRiskManualChange
}

// AgentManagedWork describes a package boundary value.
type AgentManagedWork struct {
	Path        string
	Status      string
	TicketKey   string
	ExecutionID string
	ReportPath  string
}

// ManualOrExternalChange describes a package boundary value.
type ManualOrExternalChange struct {
	Path   string
	Status string
}

// MissingAgentEvidence describes a package boundary value.
type MissingAgentEvidence struct {
	TicketKey   string
	ExecutionID string
	ReportPath  string
	Reason      string
}

// HighRiskManualChange describes a package boundary value.
type HighRiskManualChange struct {
	Path    string
	Status  string
	Pattern string
}

type executionEvidence struct {
	ticketKey   string
	executionID string
	reportPath  string
}

// NewEvidenceClassifier constructs the corresponding package service or value.
func NewEvidenceClassifier(loadReports func(context.Context, string) ([]ExecutionReport, error)) EvidenceClassifier {
	return EvidenceClassifier{LoadReports: loadReports}
}

// ExecutionReport contains the report identity and the loaded result paths or diagnostic.
type ExecutionReport struct {
	TicketKey, ExecutionID, RelativePath string
	ChangedPaths                         []string
	FailureReason                        string
}

// Classify executes the corresponding package operation.
func (c EvidenceClassifier) Classify(ctx context.Context, request EvidenceClassificationRequest) (EvidenceClassification, error) {
	projectDir := strings.TrimSpace(request.ProjectDir)
	if projectDir == "" {
		return EvidenceClassification{}, fmt.Errorf("project directory is required")
	}

	reports, err := c.LoadReports(ctx, projectDir)
	if err != nil {
		return EvidenceClassification{}, err
	}

	evidenceSet := collectExecutionEvidence(reports)
	evidenceByPath, missingEvidence := evidenceSet.ByPath, evidenceSet.Missing
	classification := EvidenceClassification{
		MissingAgentEvidence: missingEvidence,
	}

	for _, file := range request.Files {
		path := filepath.ToSlash(file.Path)

		evidence, ok := evidenceByPath[path]
		if ok {
			classification.AgentManagedWork = append(classification.AgentManagedWork, AgentManagedWork{
				Path:        path,
				Status:      file.Status,
				TicketKey:   evidence.ticketKey,
				ExecutionID: evidence.executionID,
				ReportPath:  evidence.reportPath,
			})

			continue
		}

		classification.ManualOrExternalChanges = append(classification.ManualOrExternalChanges, ManualOrExternalChange{
			Path:   path,
			Status: file.Status,
		})

		pattern, ok := highRiskPattern(path, request.HighRiskPaths)
		if ok {
			classification.HighRiskManualChanges = append(classification.HighRiskManualChanges, HighRiskManualChange{
				Path:    path,
				Status:  file.Status,
				Pattern: pattern,
			})
		}
	}

	return classification, nil
}

func highRiskPattern(filePath string, patterns []string) (string, bool) {
	for _, pattern := range patterns {
		pattern = filepath.ToSlash(strings.TrimSpace(pattern))
		if pattern == "" {
			continue
		}

		if pathMatchesPattern(filePath, pattern) {
			return pattern, true
		}
	}

	return "", false
}

func pathMatchesPattern(filePath, pattern string) bool {
	if strings.HasSuffix(pattern, "/**") {
		prefix := strings.TrimSuffix(pattern, "/**")

		return filePath == prefix || strings.HasPrefix(filePath, prefix+"/")
	}

	matched, err := path.Match(pattern, filePath)
	if err == nil && matched {
		return true
	}

	return filePath == pattern
}

func collectExecutionEvidence(reports []ExecutionReport) executionEvidenceSet {
	evidenceByPath := map[string]executionEvidence{}
	missingEvidence := []MissingAgentEvidence{}

	for _, report := range reports {
		if report.FailureReason != "" {
			missingEvidence = append(missingEvidence, reportMissingEvidence(report, report.FailureReason))

			continue
		}

		for _, changedPath := range report.ChangedPaths {
			path := filepath.ToSlash(changedPath)
			evidenceByPath[path] = executionEvidence{
				ticketKey:   report.TicketKey,
				executionID: report.ExecutionID,
				reportPath:  report.RelativePath,
			}
		}
	}

	return executionEvidenceSet{ByPath: evidenceByPath, Missing: missingEvidence}
}

func reportMissingEvidence(report ExecutionReport, reason string) MissingAgentEvidence {
	return MissingAgentEvidence{
		TicketKey:   report.TicketKey,
		ExecutionID: report.ExecutionID,
		ReportPath:  report.RelativePath,
		Reason:      reason,
	}
}

type executionEvidenceSet struct {
	ByPath  map[string]executionEvidence
	Missing []MissingAgentEvidence
}
