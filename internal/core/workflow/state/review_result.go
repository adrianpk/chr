package state

import (
	"fmt"
	"strings"
)

// IntegrationReviewDecision defines the outcome of an integration review phase.
type IntegrationReviewDecision string

const (
	// IntegrationReviewDecisionPass allows the ticket to proceed toward merge.
	IntegrationReviewDecisionPass IntegrationReviewDecision = "pass"
	// IntegrationReviewDecisionNeedsChanges sends the same ticket back to repair.
	IntegrationReviewDecisionNeedsChanges IntegrationReviewDecision = "needs_changes"
	// IntegrationReviewDecisionInsufficientEvidence blocks until evidence is improved.
	IntegrationReviewDecisionInsufficientEvidence IntegrationReviewDecision = "insufficient_evidence"
	// IntegrationReviewDecisionHumanRequired blocks for human judgement.
	IntegrationReviewDecisionHumanRequired IntegrationReviewDecision = "human_required"
)

// Valid reports whether the review decision is supported.
func (d IntegrationReviewDecision) Valid() bool {
	switch d {
	case IntegrationReviewDecisionPass,
		IntegrationReviewDecisionNeedsChanges,
		IntegrationReviewDecisionInsufficientEvidence,
		IntegrationReviewDecisionHumanRequired:
		return true
	default:
		return false
	}
}

const (
	// AgenticReviewFindingCorrectness defines a package boundary constant.
	AgenticReviewFindingCorrectness = "correctness"
	// AgenticReviewFindingScope defines a package boundary constant.
	AgenticReviewFindingScope = "scope"
	// AgenticReviewFindingEvidence defines a package boundary constant.
	AgenticReviewFindingEvidence = "evidence"
	// AgenticReviewFindingContract defines a package boundary constant.
	AgenticReviewFindingContract = "contract"
	// AgenticReviewFindingSafety defines a package boundary constant.
	AgenticReviewFindingSafety = "safety"
	// AgenticReviewFindingMergeability defines a package boundary constant.
	AgenticReviewFindingMergeability = "mergeability"
	// AgenticReviewFindingCosmetic defines a package boundary constant.
	AgenticReviewFindingCosmetic = "cosmetic"
)

// AgenticReviewResult describes the structured agentic review decision.
type AgenticReviewResult struct {
	Decision          IntegrationReviewDecision `json:"decision"`
	Summary           string                    `json:"summary"`
	BlockingFindings  []ReviewFinding           `json:"blocking_findings,omitempty"`
	NonBlockingNotes  []ReviewFinding           `json:"non_blocking_notes,omitempty"`
	ReviewedArtifacts []string                  `json:"reviewed_artifacts,omitempty"`
}

// ValidateAgenticReviewResult validates one structured review result.
func ValidateAgenticReviewResult(result AgenticReviewResult) error {
	if !result.Decision.Valid() {
		return fmt.Errorf("invalid agentic review decision: %s", result.Decision)
	}

	if strings.TrimSpace(result.Summary) == "" {
		return fmt.Errorf("agentic review summary is required")
	}

	if result.Decision == IntegrationReviewDecisionNeedsChanges && len(result.BlockingFindings) == 0 {
		return fmt.Errorf("agentic review needs_changes requires blocking findings")
	}

	for _, finding := range agenticReviewFindings(result) {
		if err := ValidateReviewFinding(finding); err != nil {
			return err
		}
	}

	return nil
}

func agenticReviewFindings(result AgenticReviewResult) []ReviewFinding {
	findings := make([]ReviewFinding, 0, len(result.BlockingFindings)+len(result.NonBlockingNotes))
	findings = append(findings, result.BlockingFindings...)
	findings = append(findings, result.NonBlockingNotes...)

	return findings
}

// ValidateReviewFinding validates the retained finding value without granting review or publication authority.
func ValidateReviewFinding(finding ReviewFinding) error {
	if !validAgenticReviewFindingCategory(finding.Category) {
		return fmt.Errorf("invalid agentic review finding category: %s", finding.Category)
	}

	if strings.TrimSpace(finding.Title) == "" {
		return fmt.Errorf("agentic review finding title is required")
	}

	if strings.TrimSpace(finding.Detail) == "" {
		return fmt.Errorf("agentic review finding detail is required")
	}

	return nil
}

func validAgenticReviewFindingCategory(category string) bool {
	switch strings.TrimSpace(category) {
	case AgenticReviewFindingCorrectness,
		AgenticReviewFindingScope,
		AgenticReviewFindingEvidence,
		AgenticReviewFindingContract,
		AgenticReviewFindingSafety,
		AgenticReviewFindingMergeability,
		AgenticReviewFindingCosmetic:
		return true
	default:
		return false
	}
}
