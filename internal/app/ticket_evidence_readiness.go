package app

import (
	"context"
	"fmt"
	"time"

	"forge.adrianpk.com/chr/internal/core"
	executioncore "forge.adrianpk.com/chr/internal/core/execution"
	ticketcore "forge.adrianpk.com/chr/internal/core/ticket"
)

type ticketEvidenceRequest struct {
	ProjectDir string
	TicketPath string
	Execution  *executioncore.ExecutionRecord
	Result     *executioncore.ExecutionResult
}

func ticketEvidenceReadiness(ctx context.Context, request ticketEvidenceRequest) (core.EvidenceReadinessResult, error) {
	config, err := core.LoadWorkspaceConfig(request.ProjectDir)
	if err != nil {
		return core.EvidenceReadinessResult{}, fmt.Errorf("load evidence strategy config: %w", err)
	}

	checklistResult, err := ticketExecutionReportEvidenceChecklist(request.TicketPath)
	checklist, checklistError := checklistResult.Checklist, checklistResult.Diagnostic

	if err != nil {
		return core.EvidenceReadinessResult{}, err
	}

	ticket, err := ticketcore.NewParser().Parse(ctx, ticketcore.ParseOptions{
		ProjectDir: request.ProjectDir,
		Path:       request.TicketPath,
		IndexedAt:  time.Now().UTC(),
	})
	if err != nil {
		return core.EvidenceReadinessResult{}, fmt.Errorf("read evidence strategy ticket metadata: %w", err)
	}

	policy, err := core.DetectEvidencePolicyContext(request.ProjectDir, config.Evidence.Strategy, ticket)
	if err != nil {
		return core.EvidenceReadinessResult{}, err
	}

	return core.EvaluateEvidenceReadiness(core.EvidenceReadinessRequest{
		Config:         config.Evidence.Strategy,
		Checklist:      checklist,
		ChecklistError: checklistError,
		Execution:      request.Execution,
		Result:         request.Result,
		Gherkin:        policy.Gherkin,
	}), nil
}
