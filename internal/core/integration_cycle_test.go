package core

import (
	"strings"
	"testing"

	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

func TestIntegrationCyclePolicyFromReviewConfigMapsModes(t *testing.T) {
	for _, testCase := range []struct {
		name   string
		mode   workflowstate.ReviewWorkflowMode
		policy IntegrationCyclePolicy
	}{
		{
			name: "human review",
			mode: workflowstate.ReviewWorkflowModeHumanReview,
			policy: IntegrationCyclePolicy{
				Mode:                IntegrationCycleModeHumanReview,
				RequiresReview:      true,
				RequiresHumanReview: true,
				RequiresMerge:       false,
			},
		},
		{
			name: "agentic review",
			mode: workflowstate.ReviewWorkflowModeAgenticReview,
			policy: IntegrationCyclePolicy{
				Mode:           IntegrationCycleModeAgenticReview,
				RequiresReview: true,
				RequiresMerge:  true,
			},
		},
		{
			name: "trusted auto merge",
			mode: workflowstate.ReviewWorkflowModeTrustedAutoMerge,
			policy: IntegrationCyclePolicy{
				Mode:          IntegrationCycleModeTrustedAutoMerge,
				RequiresMerge: true,
			},
		},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			got, err := IntegrationCyclePolicyFromReviewConfig(ReviewWorkflowConfig{
				Mode: testCase.mode,
			})
			if err != nil {
				t.Fatalf("IntegrationCyclePolicyFromReviewConfig returned error: %v", err)
			}

			if got != testCase.policy {
				t.Fatalf("policy = %#v, want %#v", got, testCase.policy)
			}
		})
	}
}

func TestIntegrationCyclePolicyFromReviewConfigRejectsInvalidConfig(t *testing.T) {
	_, err := IntegrationCyclePolicyFromReviewConfig(ReviewWorkflowConfig{
		Mode: "unknown",
	})
	if err == nil || !strings.Contains(err.Error(), "invalid review workflow mode: unknown") {
		t.Fatalf("error = %v, want invalid review workflow mode", err)
	}
}

func TestPlanIntegrationCycleBuildsSelectedPendingPlan(t *testing.T) {
	ticket := IndexedTicket{IndexedTicketIdentity: IndexedTicketIdentity{Key: "CHR-1"}, IndexedTicketSummary: IndexedTicketSummary{Title: "Add serialized loop"}, IndexedTicketClassification: IndexedTicketClassification{Status: TicketStatusReadyForDev}}

	got, err := PlanIntegrationCycle(IntegrationCyclePlanOptions{
		ProjectID: "project-1",
		Ticket:    ticket,
		ReviewConfig: ReviewWorkflowConfig{
			Mode: workflowstate.ReviewWorkflowModeAgenticReview,
		},
	})
	if err != nil {
		t.Fatalf("PlanIntegrationCycle returned error: %v", err)
	}

	if got.ProjectID != "project-1" || got.Ticket.Key != "CHR-1" {
		t.Fatalf("plan identity = %#v, want project and ticket identity", got)
	}

	if got.Policy.Mode != IntegrationCycleModeAgenticReview || !got.Policy.RequiresReview || !got.Policy.RequiresMerge {
		t.Fatalf("policy = %#v, want agentic review policy", got.Policy)
	}

	if got.Phase != IntegrationCyclePhaseSelected || got.Status != IntegrationCycleStatusPending || got.Blocker != IntegrationCycleBlockerNone {
		t.Fatalf("plan state = phase %q status %q blocker %q, want selected pending none", got.Phase, got.Status, got.Blocker)
	}
}

func TestPlanIntegrationCycleRejectsInvalidInput(t *testing.T) {
	for _, testCase := range []struct {
		name string
		opts IntegrationCyclePlanOptions
		want string
	}{
		{
			name: "missing project id",
			opts: IntegrationCyclePlanOptions{
				Ticket:       IndexedTicket{IndexedTicketIdentity: IndexedTicketIdentity{Key: "CHR-1"}},
				ReviewConfig: ReviewWorkflowConfig{Mode: workflowstate.ReviewWorkflowModeHumanReview},
			},
			want: "integration cycle project id is required",
		},
		{
			name: "missing ticket key",
			opts: IntegrationCyclePlanOptions{
				ProjectID:    "project-1",
				ReviewConfig: ReviewWorkflowConfig{Mode: workflowstate.ReviewWorkflowModeHumanReview},
			},
			want: "integration cycle ticket key is required",
		},
		{
			name: "invalid review config",
			opts: IntegrationCyclePlanOptions{
				ProjectID:    "project-1",
				Ticket:       IndexedTicket{IndexedTicketIdentity: IndexedTicketIdentity{Key: "CHR-1"}},
				ReviewConfig: ReviewWorkflowConfig{Mode: "unknown"},
			},
			want: "invalid review workflow mode: unknown",
		},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			_, err := PlanIntegrationCycle(testCase.opts)
			if err == nil || err.Error() != testCase.want {
				t.Fatalf("error = %v, want %q", err, testCase.want)
			}
		})
	}
}

func TestIntegrationCycleValuesValidateSupportedConstants(t *testing.T) {
	t.Run("mode", func(t *testing.T) {
		assertIntegrationCycleValidity(t, IntegrationCycleModeAgenticReview.Valid(), IntegrationCycleMode("unknown").Valid())
	})

	t.Run("phase", func(t *testing.T) {
		assertIntegrationCycleValidity(t, IntegrationCyclePhaseReviewing.Valid(), IntegrationCyclePhase("unknown").Valid())
	})

	t.Run("status", func(t *testing.T) {
		assertIntegrationCycleValidity(t, IntegrationCycleStatusBlocked.Valid(), IntegrationCycleStatus("unknown").Valid())
	})

	t.Run("blocker", func(t *testing.T) {
		assertIntegrationCycleValidity(t, IntegrationCycleBlockerMergeBlocked.Valid(), IntegrationCycleBlocker("unknown").Valid())
	})

	t.Run("review decision", func(t *testing.T) {
		assertIntegrationCycleValidity(t, IntegrationReviewDecisionNeedsChanges.Valid(), IntegrationReviewDecision("unknown").Valid())
	})
}

func assertIntegrationCycleValidity(t *testing.T, valid, invalid bool) {
	t.Helper()

	if !valid || invalid {
		t.Fatalf("valid = %t invalid = %t, want true false", valid, invalid)
	}
}
