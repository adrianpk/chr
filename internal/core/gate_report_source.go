package core

import (
	"context"
	"strings"

	executioncore "forge.adrianpk.com/chr/internal/core/execution"
	"forge.adrianpk.com/chr/internal/core/gate"
)

// ListGateExecutionReports adapts report storage to gate classification inputs.
func (l ReportLister) ListGateExecutionReports(ctx context.Context, projectDir string) ([]gate.ExecutionReport, error) {
	reports, err := l.ListProjectReports(ctx, projectDir, ReportListFilter{Kind: ReportKindExecution})
	if err != nil {
		return nil, err
	}

	evidence := make([]gate.ExecutionReport, 0, len(reports))
	for _, report := range reports {
		item := loadGateExecutionReport(report)

		evidence = append(evidence, item)
	}

	return evidence, nil
}

func loadGateExecutionReport(report ReportListEntry) gate.ExecutionReport {
	item := gate.ExecutionReport{TicketKey: report.TicketKey, ExecutionID: report.ExecutionID, RelativePath: report.RelativePath}

	resultPath := strings.TrimSpace(report.ResultPath)
	if resultPath == "" {
		item.FailureReason = "execution result path is missing"

		return item
	}

	result, err := executioncore.LoadExecutionResult(resultPath)
	if err != nil {
		item.FailureReason = err.Error()

		return item
	}

	for _, file := range result.ChangedFiles {
		item.ChangedPaths = append(item.ChangedPaths, file.Path)
	}

	return item
}
