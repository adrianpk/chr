package core

import (
	"fmt"
	"strings"

	executioncore "forge.adrianpk.com/chr/internal/core/execution"
	ticketmetadata "forge.adrianpk.com/chr/internal/core/ticket/metadata"
	workspaceconfig "forge.adrianpk.com/chr/internal/core/workspace/config"
)

// EvidenceReadinessRequest describes inputs for ready-for-review evidence checks.
type EvidenceReadinessRequest struct {
	Config         workspaceconfig.EvidenceStrategyConfig
	Checklist      *EvidenceStrategyChecklist
	ChecklistError string
	Execution      *executioncore.Record
	Result         *executioncore.Result
	Gherkin        EvidenceGherkinContext
	ReviewRefs     []ticketmetadata.ReviewRef
}

// EvidenceReadinessResult describes the ready-for-review evidence decision.
type EvidenceReadinessResult struct {
	Allowed     bool
	Diagnostics []EvidenceReadinessDiagnostic
}

// EvidenceReadinessDiagnostic describes one evidence readiness blocker.
type EvidenceReadinessDiagnostic struct {
	Code   EvidenceStrategyDiagnosticCode
	Reason string
}

type evidenceFinalDiagnosticRequest struct {
	Item   EvidenceStrategyChecklistItem
	Code   EvidenceStrategyDiagnosticCode
	Reason string
}

// EvaluateEvidenceReadiness evaluates whether evidence permits review.
func EvaluateEvidenceReadiness(request EvidenceReadinessRequest) EvidenceReadinessResult {
	if evidenceGateDisabled(request.Config) {
		return EvidenceReadinessResult{Allowed: true}
	}

	diagnostics := evidenceReadinessDiagnostics(request)

	return EvidenceReadinessResult{
		Allowed:     len(diagnostics) == 0,
		Diagnostics: diagnostics,
	}
}

// BlockerReason renders diagnostics into a stable ticket blocker reason.
func (r EvidenceReadinessResult) BlockerReason() string {
	if r.Allowed || len(r.Diagnostics) == 0 {
		return ""
	}

	reasons := make([]string, 0, len(r.Diagnostics))
	for _, diagnostic := range r.Diagnostics {
		reasons = append(reasons, string(diagnostic.Code)+": "+diagnostic.Reason)
	}

	return "evidence readiness failed: " + strings.Join(reasons, "; ")
}

func evidenceGateDisabled(config workspaceconfig.EvidenceStrategyConfig) bool {
	return !config.Enabled ||
		!config.StrictReadyForReview ||
		config.TicketChecklist == workspaceconfig.EvidenceTicketChecklistPolicyOptional
}

func evidenceReadinessDiagnostics(request EvidenceReadinessRequest) []EvidenceReadinessDiagnostic {
	diagnostics := evidenceRuntimeDiagnostics(request)
	diagnostics = append(diagnostics, evidenceChecklistDiagnostics(request)...)
	diagnostics = append(diagnostics, evidencePolicyDiagnostics(request)...)
	diagnostics = append(diagnostics, evidenceSummaryDiagnostics(request)...)

	return diagnostics
}

func evidenceRuntimeDiagnostics(request EvidenceReadinessRequest) []EvidenceReadinessDiagnostic {
	if request.Execution == nil && request.Result == nil {
		return nil
	}

	if request.Execution != nil && request.Execution.Status != executioncore.StatusSucceeded {
		return []EvidenceReadinessDiagnostic{
			evidenceRuntimeDiagnostic(*request.Execution),
		}
	}

	if request.Result != nil && request.Result.Status != executioncore.ResultStatusSucceeded {
		return []EvidenceReadinessDiagnostic{{
			Code:   EvidenceStrategyDiagnosticValidationFailed,
			Reason: "worker result status is " + string(request.Result.Status),
		}}
	}

	return nil
}

func evidenceRuntimeDiagnostic(record executioncore.Record) EvidenceReadinessDiagnostic {
	kind := executioncore.FailureKindFromReason(record.FailureReason)
	switch kind {
	case executioncore.FailureValidation:
		return EvidenceReadinessDiagnostic{
			Code:   EvidenceStrategyDiagnosticValidationFailed,
			Reason: nonEmptyEvidenceReason(record.FailureReason, "execution validation failed"),
		}
	case executioncore.FailureRuntime,
		executioncore.FailureHarness,
		executioncore.FailurePreflight:
		return EvidenceReadinessDiagnostic{
			Code:   EvidenceStrategyDiagnosticRuntimeUnavailable,
			Reason: nonEmptyEvidenceReason(record.FailureReason, "execution runtime failed"),
		}
	case executioncore.FailureWorker,
		executioncore.FailureResultCapture:
		return EvidenceReadinessDiagnostic{
			Code:   EvidenceStrategyDiagnosticCodexRuntimeUnavailable,
			Reason: nonEmptyEvidenceReason(record.FailureReason, "worker result is unavailable"),
		}
	default:
		return EvidenceReadinessDiagnostic{
			Code:   EvidenceStrategyDiagnosticRuntimeUnavailable,
			Reason: fmt.Sprintf("execution status is %s", record.Status),
		}
	}
}

func evidenceChecklistDiagnostics(request EvidenceReadinessRequest) []EvidenceReadinessDiagnostic {
	if strings.TrimSpace(request.ChecklistError) != "" {
		return []EvidenceReadinessDiagnostic{{
			Code:   EvidenceStrategyDiagnosticMalformed,
			Reason: request.ChecklistError,
		}}
	}

	if request.Checklist == nil || !request.Checklist.Found {
		return []EvidenceReadinessDiagnostic{{
			Code:   EvidenceStrategyDiagnosticMissing,
			Reason: "ticket evidence strategy checklist is missing",
		}}
	}

	return evidenceChecklistItemDiagnostics(*request.Checklist)
}

func evidenceChecklistItemDiagnostics(checklist EvidenceStrategyChecklist) []EvidenceReadinessDiagnostic {
	diagnostics := make([]EvidenceReadinessDiagnostic, 0)

	for _, label := range EvidenceStrategyChecklistLabels() {
		item, ok := checklist.Item(label)
		if !ok {
			diagnostics = append(diagnostics, evidenceDiagnostic(EvidenceStrategyDiagnosticMissing, label+" checklist item is missing"))

			continue
		}

		diagnostics = append(diagnostics, evidenceChecklistItemDiagnostic(item)...)
	}

	return diagnostics
}

func evidenceChecklistItemDiagnostic(item EvidenceStrategyChecklistItem) []EvidenceReadinessDiagnostic {
	if item.Status == "blocked" {
		return []EvidenceReadinessDiagnostic{
			evidenceDiagnostic(EvidenceStrategyDiagnosticBlocked, item.Label+" evidence is blocked"),
		}
	}

	switch item.Label {
	case EvidenceStrategyChecklistLabelStrategy:
		return evidenceFinalDiagnostics(evidenceFinalDiagnosticRequest{
			Item:   item,
			Code:   EvidenceStrategyDiagnosticMissing,
			Reason: "evidence strategy is not complete",
		})
	case EvidenceStrategyChecklistLabelGherkin:
		return evidenceGherkinDiagnostics(item)
	case EvidenceStrategyChecklistLabelTests:
		return evidenceTestsDiagnostics(item)
	case EvidenceStrategyChecklistLabelValidation:
		return evidenceValidationDiagnostics(item)
	case EvidenceStrategyChecklistLabelSummary:
		return evidenceFinalDiagnostics(evidenceFinalDiagnosticRequest{
			Item:   item,
			Code:   EvidenceStrategyDiagnosticSummaryMissing,
			Reason: "evidence summary is not complete",
		})
	default:
		return nil
	}
}

func evidenceGherkinDiagnostics(item EvidenceStrategyChecklistItem) []EvidenceReadinessDiagnostic {
	if item.Status != "required" || item.Checked {
		return nil
	}

	return []EvidenceReadinessDiagnostic{
		evidenceDiagnostic(EvidenceStrategyDiagnosticGherkinRequiredMissing, "required Gherkin evidence is missing"),
	}
}

func evidenceTestsDiagnostics(item EvidenceStrategyChecklistItem) []EvidenceReadinessDiagnostic {
	if item.Status == "none" || item.Checked {
		return nil
	}

	return []EvidenceReadinessDiagnostic{
		evidenceDiagnostic(EvidenceStrategyDiagnosticTestsRequiredMissing, "required test evidence is missing"),
	}
}

func evidenceValidationDiagnostics(item EvidenceStrategyChecklistItem) []EvidenceReadinessDiagnostic {
	switch {
	case item.Status == "failed":
		return []EvidenceReadinessDiagnostic{
			evidenceDiagnostic(EvidenceStrategyDiagnosticValidationFailed, "validation evidence failed"),
		}
	case item.Status != "passed" || !item.Checked:
		return []EvidenceReadinessDiagnostic{
			evidenceDiagnostic(EvidenceStrategyDiagnosticValidationMissing, "validation evidence is not passed"),
		}
	default:
		return nil
	}
}

func evidenceFinalDiagnostics(request evidenceFinalDiagnosticRequest) []EvidenceReadinessDiagnostic {
	if request.Item.Status == "complete" && request.Item.Checked {
		return nil
	}

	return []EvidenceReadinessDiagnostic{
		evidenceDiagnostic(request.Code, request.Reason),
	}
}

func evidenceSummaryDiagnostics(request EvidenceReadinessRequest) []EvidenceReadinessDiagnostic {
	if request.Result == nil {
		return nil
	}

	summary := request.Result.EvidenceSummary

	summaryComplete := strings.TrimSpace(summary.Strategy) != "" &&
		strings.TrimSpace(summary.Produced) != "" &&
		strings.TrimSpace(summary.Validation) != ""
	if summaryComplete {
		return nil
	}

	return []EvidenceReadinessDiagnostic{
		evidenceDiagnostic(EvidenceStrategyDiagnosticSummaryMissing, "execution result evidence_summary is incomplete"),
	}
}

func evidencePolicyDiagnostics(request EvidenceReadinessRequest) []EvidenceReadinessDiagnostic {
	if !request.Gherkin.Active || !request.Gherkin.Applies {
		return nil
	}

	if request.Checklist == nil || request.ChecklistError != "" {
		return nil
	}

	item, ok := request.Checklist.Item(EvidenceStrategyChecklistLabelGherkin)
	if !ok {
		return nil
	}

	if item.Status == "required" {
		return nil
	}

	reason := "active Gherkin applies to this observable behavior"
	if strings.TrimSpace(request.Gherkin.RequiredPath) != "" {
		reason += "; required_path=" + request.Gherkin.RequiredPath
	}

	return []EvidenceReadinessDiagnostic{
		evidenceDiagnostic(EvidenceStrategyDiagnosticGherkinRequiredMissing, reason),
	}
}

func evidenceDiagnostic(code EvidenceStrategyDiagnosticCode, reason string) EvidenceReadinessDiagnostic {
	return EvidenceReadinessDiagnostic{Code: code, Reason: reason}
}

func nonEmptyEvidenceReason(value, fallback string) string {
	if strings.TrimSpace(value) == "" {
		return fallback
	}

	return value
}
