package core

import (
	"strings"
	"testing"
	"time"
)

func TestEvaluateEvidenceReadinessAllowsCompleteEvidence(t *testing.T) {
	result := EvaluateEvidenceReadiness(EvidenceReadinessRequest{
		Config:    DefaultEvidenceStrategyConfig(),
		Checklist: testCompleteEvidenceReadinessChecklist(),
		Execution: testEvidenceReadinessExecution(func(record *ExecutionRecord) {
			record.Status = ExecutionStatusSucceeded
		}),
		Result: testEvidenceReadinessResult(func(result *ExecutionResult) {
			result.Status = ExecutionResultStatusSucceeded
		}),
	})

	if !result.Allowed {
		t.Fatalf("Allowed = false, diagnostics = %#v", result.Diagnostics)
	}
}

func TestEvaluateEvidenceReadinessClassifiesBlockers(t *testing.T) {
	for _, tt := range evidenceReadinessBlockerCases() {
		t.Run(tt.name, func(t *testing.T) {
			result := EvaluateEvidenceReadiness(tt.request())
			if result.Allowed {
				t.Fatalf("Allowed = true, want false")
			}

			if !evidenceDiagnosticsContain(result.Diagnostics, tt.want) {
				t.Fatalf("diagnostics = %#v, want %s", result.Diagnostics, tt.want)
			}

			if !strings.Contains(result.BlockerReason(), string(tt.want)) {
				t.Fatalf("BlockerReason = %q, want code %s", result.BlockerReason(), tt.want)
			}
		})
	}
}

type evidenceReadinessBlockerCase struct {
	name    string
	request func() EvidenceReadinessRequest
	want    EvidenceStrategyDiagnosticCode
}

func evidenceReadinessBlockerCases() []evidenceReadinessBlockerCase {
	return []evidenceReadinessBlockerCase{
		{name: "missing checklist", request: missingEvidenceChecklistRequest, want: EvidenceStrategyDiagnosticMissing},
		{name: "malformed checklist", request: malformedEvidenceChecklistRequest, want: EvidenceStrategyDiagnosticMalformed},
		{name: "blocked checklist item", request: blockedEvidenceChecklistRequest, want: EvidenceStrategyDiagnosticBlocked},
		{name: "validation failed", request: failedEvidenceValidationRequest, want: EvidenceStrategyDiagnosticValidationFailed},
		{name: "runtime failed", request: failedEvidenceRuntimeRequest, want: EvidenceStrategyDiagnosticRuntimeUnavailable},
		{name: "summary missing", request: missingEvidenceSummaryRequest, want: EvidenceStrategyDiagnosticSummaryMissing},
	}
}

func missingEvidenceChecklistRequest() EvidenceReadinessRequest {
	return EvidenceReadinessRequest{Config: DefaultEvidenceStrategyConfig()}
}

func malformedEvidenceChecklistRequest() EvidenceReadinessRequest {
	return EvidenceReadinessRequest{
		Config:         DefaultEvidenceStrategyConfig(),
		ChecklistError: "invalid Tests status: maybe",
	}
}

func blockedEvidenceChecklistRequest() EvidenceReadinessRequest {
	return EvidenceReadinessRequest{
		Config: DefaultEvidenceStrategyConfig(),
		Checklist: testCompleteEvidenceReadinessChecklistWith(func(checklist *EvidenceStrategyChecklist) {
			item, _ := checklist.Item(EvidenceStrategyChecklistLabelTests)
			item.Status = "blocked"
			checklist.Items[item.Label] = item
		}),
	}
}

func failedEvidenceValidationRequest() EvidenceReadinessRequest {
	return failedEvidenceExecutionRequest("validation: make verify failed")
}

func failedEvidenceRuntimeRequest() EvidenceReadinessRequest {
	return failedEvidenceExecutionRequest("runtime: docker build failed")
}

func failedEvidenceExecutionRequest(reason string) EvidenceReadinessRequest {
	return EvidenceReadinessRequest{
		Config:    DefaultEvidenceStrategyConfig(),
		Checklist: testCompleteEvidenceReadinessChecklist(),
		Execution: testEvidenceReadinessExecution(func(record *ExecutionRecord) {
			record.Status = ExecutionStatusFailed
			record.FailureReason = reason
		}),
	}
}

func missingEvidenceSummaryRequest() EvidenceReadinessRequest {
	return EvidenceReadinessRequest{
		Config:    DefaultEvidenceStrategyConfig(),
		Checklist: testCompleteEvidenceReadinessChecklist(),
		Result: testEvidenceReadinessResult(func(result *ExecutionResult) {
			result.EvidenceSummary.Validation = ""
		}),
	}
}

func TestEvaluateEvidenceReadinessAllowsDisabledOrOptionalStrategy(t *testing.T) {
	for _, config := range []EvidenceStrategyConfig{
		testEvidenceReadinessConfig(func(config *EvidenceStrategyConfig) {
			config.Enabled = false
		}),
		testEvidenceReadinessConfig(func(config *EvidenceStrategyConfig) {
			config.TicketChecklist = EvidenceTicketChecklistPolicyOptional
		}),
		testEvidenceReadinessConfig(func(config *EvidenceStrategyConfig) {
			config.StrictReadyForReview = false
		}),
	} {
		result := EvaluateEvidenceReadiness(EvidenceReadinessRequest{Config: config})
		if !result.Allowed {
			t.Fatalf("Allowed = false for config %#v", config)
		}
	}
}

func TestEvaluateEvidenceReadinessRequiresFinalChecklistEvidence(t *testing.T) {
	result := EvaluateEvidenceReadiness(EvidenceReadinessRequest{
		Config: DefaultEvidenceStrategyConfig(),
		Checklist: testCompleteEvidenceReadinessChecklistWith(func(checklist *EvidenceStrategyChecklist) {
			item, _ := checklist.Item(EvidenceStrategyChecklistLabelGherkin)
			item.Checked = false
			checklist.Items[item.Label] = item
		}),
	})

	if result.Allowed {
		t.Fatal("Allowed = true, want false")
	}

	if !evidenceDiagnosticsContain(result.Diagnostics, EvidenceStrategyDiagnosticGherkinRequiredMissing) {
		t.Fatalf("diagnostics = %#v, want gherkin blocker", result.Diagnostics)
	}
}

func testCompleteEvidenceReadinessChecklist() *EvidenceStrategyChecklist {
	return testCompleteEvidenceReadinessChecklistWith(func(_ *EvidenceStrategyChecklist) {})
}

func testCompleteEvidenceReadinessChecklistWith(edit func(*EvidenceStrategyChecklist)) *EvidenceStrategyChecklist {
	checklist := EvidenceStrategyChecklist{
		Found: true,
		Items: map[string]EvidenceStrategyChecklistItem{
			EvidenceStrategyChecklistLabelStrategy: {
				Label:   EvidenceStrategyChecklistLabelStrategy,
				Checked: true,
				Status:  "complete",
			},
			EvidenceStrategyChecklistLabelGherkin: {
				Label:   EvidenceStrategyChecklistLabelGherkin,
				Checked: true,
				Status:  "required",
				Detail:  "path=.chc/user/acceptance/flow.feature",
			},
			EvidenceStrategyChecklistLabelTests: {
				Label:   EvidenceStrategyChecklistLabelTests,
				Checked: true,
				Status:  "mixed",
				Detail:  "reason=unit and fuzz coverage",
			},
			EvidenceStrategyChecklistLabelValidation: {
				Label:   EvidenceStrategyChecklistLabelValidation,
				Checked: true,
				Status:  "passed",
			},
			EvidenceStrategyChecklistLabelSummary: {
				Label:   EvidenceStrategyChecklistLabelSummary,
				Checked: true,
				Status:  "complete",
			},
		},
	}

	edit(&checklist)

	return &checklist
}

func testEvidenceReadinessConfig(edit func(*EvidenceStrategyConfig)) EvidenceStrategyConfig {
	config := DefaultEvidenceStrategyConfig()
	edit(&config)

	return config
}

func testEvidenceReadinessExecution(edit func(*ExecutionRecord)) *ExecutionRecord {
	now := time.Date(2026, 6, 19, 10, 0, 0, 0, time.UTC)
	record := ExecutionRecord{
		ID:        "exec-1",
		ProjectID: "project-1",
		TicketKey: "CHC-1",
		Status:    ExecutionStatusSucceeded,
		CreatedAt: now,
		UpdatedAt: now,
	}

	edit(&record)

	return &record
}

func testEvidenceReadinessResult(edit func(*ExecutionResult)) *ExecutionResult {
	result := ExecutionResult{
		ExecutionID: "exec-1",
		TicketKey:   "CHC-1",
		Status:      ExecutionResultStatusSucceeded,
		Summary:     "done",
		EvidenceSummary: ExecutionResultEvidenceSummary{
			Strategy:   "complete strategy",
			Produced:   "unit and fuzz evidence",
			Validation: "make verify passed",
		},
	}

	edit(&result)

	return &result
}

func evidenceDiagnosticsContain(diagnostics []EvidenceReadinessDiagnostic, code EvidenceStrategyDiagnosticCode) bool {
	for _, diagnostic := range diagnostics {
		if diagnostic.Code == code {
			return true
		}
	}

	return false
}
