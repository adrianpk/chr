package workspace

import (
	"context"
	"fmt"
	"path"
	"path/filepath"
	"strings"
)

// GateEvidenceClassifier describes a package boundary value.
type GateEvidenceClassifier struct {
	ReportLister ReportLister
}

// GateEvidenceClassificationRequest describes a package boundary value.
type GateEvidenceClassificationRequest struct {
	ProjectDir    string
	Files         []GateGitChangedFile
	HighRiskPaths []string
}

// GateEvidenceClassification describes a package boundary value.
type GateEvidenceClassification struct {
	AgentManagedWork        []GateAgentManagedWork
	ManualOrExternalChanges []GateManualOrExternalChange
	MissingAgentEvidence    []GateMissingAgentEvidence
	HighRiskManualChanges   []GateHighRiskManualChange
}

// GateAgentManagedWork describes a package boundary value.
type GateAgentManagedWork struct {
	Path        string
	Status      string
	TicketKey   string
	ExecutionID string
	ReportPath  string
}

// GateManualOrExternalChange describes a package boundary value.
type GateManualOrExternalChange struct {
	Path   string
	Status string
}

// GateMissingAgentEvidence describes a package boundary value.
type GateMissingAgentEvidence struct {
	TicketKey   string
	ExecutionID string
	ReportPath  string
	Reason      string
}

// GateHighRiskManualChange describes a package boundary value.
type GateHighRiskManualChange struct {
	Path    string
	Status  string
	Pattern string
}

type gateExecutionEvidence struct {
	ticketKey   string
	executionID string
	reportPath  string
	changedFile ExecutionResultChangedFile
}

// NewGateEvidenceClassifier constructs the corresponding package service or value.
func NewGateEvidenceClassifier() GateEvidenceClassifier {
	return GateEvidenceClassifier{
		ReportLister: NewReportLister(),
	}
}

// Classify executes the corresponding package operation.
func (c GateEvidenceClassifier) Classify(
	ctx context.Context,
	request GateEvidenceClassificationRequest,
) (GateEvidenceClassification, error) {
	projectDir := strings.TrimSpace(request.ProjectDir)
	if projectDir == "" {
		return GateEvidenceClassification{}, fmt.Errorf("project directory is required")
	}

	reports, err := c.ReportLister.ListProjectReports(ctx, projectDir, ReportListFilter{
		Kind: ReportKindExecution,
	})
	if err != nil {
		return GateEvidenceClassification{}, err
	}

	evidenceByPath, missingEvidence := collectGateExecutionEvidence(reports)
	classification := GateEvidenceClassification{
		MissingAgentEvidence: missingEvidence,
	}

	for _, file := range request.Files {
		path := filepath.ToSlash(file.Path)

		evidence, ok := evidenceByPath[path]
		if ok {
			classification.AgentManagedWork = append(classification.AgentManagedWork, GateAgentManagedWork{
				Path:        path,
				Status:      file.Status,
				TicketKey:   evidence.ticketKey,
				ExecutionID: evidence.executionID,
				ReportPath:  evidence.reportPath,
			})

			continue
		}

		classification.ManualOrExternalChanges = append(classification.ManualOrExternalChanges, GateManualOrExternalChange{
			Path:   path,
			Status: file.Status,
		})

		pattern, ok := gateHighRiskPattern(path, request.HighRiskPaths)
		if ok {
			classification.HighRiskManualChanges = append(classification.HighRiskManualChanges, GateHighRiskManualChange{
				Path:    path,
				Status:  file.Status,
				Pattern: pattern,
			})
		}
	}

	return classification, nil
}

func gateHighRiskPattern(filePath string, patterns []string) (string, bool) {
	for _, pattern := range patterns {
		pattern = filepath.ToSlash(strings.TrimSpace(pattern))
		if pattern == "" {
			continue
		}

		if gatePathMatchesPattern(filePath, pattern) {
			return pattern, true
		}
	}

	return "", false
}

func gatePathMatchesPattern(filePath, pattern string) bool {
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

func collectGateExecutionEvidence(
	reports []ReportListEntry,
) (map[string]gateExecutionEvidence, []GateMissingAgentEvidence) {
	evidenceByPath := map[string]gateExecutionEvidence{}
	missingEvidence := []GateMissingAgentEvidence{}

	for _, report := range reports {
		resultPath := strings.TrimSpace(report.ResultPath)
		if resultPath == "" {
			missingEvidence = append(missingEvidence, gateMissingEvidence(report, "execution result path is missing"))

			continue
		}

		result, err := LoadExecutionResult(resultPath)
		if err != nil {
			missingEvidence = append(missingEvidence, gateMissingEvidence(report, err.Error()))

			continue
		}

		for _, changedFile := range result.ChangedFiles {
			path := filepath.ToSlash(changedFile.Path)
			evidenceByPath[path] = gateExecutionEvidence{
				ticketKey:   report.TicketKey,
				executionID: report.ExecutionID,
				reportPath:  report.RelativePath,
				changedFile: changedFile,
			}
		}
	}

	return evidenceByPath, missingEvidence
}

func gateMissingEvidence(report ReportListEntry, reason string) GateMissingAgentEvidence {
	return GateMissingAgentEvidence{
		TicketKey:   report.TicketKey,
		ExecutionID: report.ExecutionID,
		ReportPath:  report.RelativePath,
		Reason:      reason,
	}
}
