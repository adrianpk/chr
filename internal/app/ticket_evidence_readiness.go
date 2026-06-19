package app

import (
	"fmt"

	"forge.adrianpk.com/chc/internal/core"
)

type ticketEvidenceRequest struct {
	ProjectDir string
	TicketPath string
	Execution  *core.ExecutionRecord
	Result     *core.ExecutionResult
}

func ticketEvidenceReadiness(request ticketEvidenceRequest) (core.EvidenceReadinessResult, error) {
	config, err := core.LoadWorkspaceConfig(request.ProjectDir)
	if err != nil {
		return core.EvidenceReadinessResult{}, fmt.Errorf("load evidence strategy config: %w", err)
	}

	checklist, checklistError, err := ticketExecutionReportEvidenceChecklist(request.TicketPath)
	if err != nil {
		return core.EvidenceReadinessResult{}, err
	}

	return core.EvaluateEvidenceReadiness(core.EvidenceReadinessRequest{
		Config:         config.Evidence.Strategy,
		Checklist:      checklist,
		ChecklistError: checklistError,
		Execution:      request.Execution,
		Result:         request.Result,
	}), nil
}
