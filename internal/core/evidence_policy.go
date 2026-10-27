package core

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"

	acceptancecore "forge.adrianpk.com/chr/internal/core/authoring/acceptance"
)

// EvidenceGherkinContext describes project and ticket-level Gherkin policy.
type EvidenceGherkinContext struct {
	Policy       EvidenceGherkinPolicy
	Active       bool
	Applies      bool
	FeaturePaths []string
	RequiredPath string
	Reason       string
}

// EvidenceStrategyGuidance describes the default evidence policy for a ticket.
type EvidenceStrategyGuidance struct {
	ChangeKind string
	Risk       string
	TestTiming string
	Tests      string
	Gherkin    string
	Reason     string
}

// EvidenceDecisionMatrixRow describes one default evidence matrix row.
type EvidenceDecisionMatrixRow struct {
	ChangeKind      string
	DefaultEvidence string
}

type evidenceKeywordRule struct {
	Kind  string
	Terms []string
}

// EvidencePolicyContext describes ticket policy guidance for a worker.
type EvidencePolicyContext struct {
	Gherkin  EvidenceGherkinContext
	Guidance EvidenceStrategyGuidance
}

// DetectEvidencePolicyContext builds worker-facing evidence policy context.
func DetectEvidencePolicyContext(projectDir string, config EvidenceStrategyConfig, ticket IndexedTicket) (EvidencePolicyContext, error) {
	gherkin, err := DetectEvidenceGherkinContext(projectDir, config, ticket)
	if err != nil {
		return EvidencePolicyContext{}, err
	}

	return EvidencePolicyContext{
		Gherkin:  gherkin,
		Guidance: EvidenceGuidanceForTicket(ticket, gherkin),
	}, nil
}

// DetectEvidenceGherkinContext detects whether Gherkin is active and applicable.
func DetectEvidenceGherkinContext(projectDir string, config EvidenceStrategyConfig, ticket IndexedTicket) (EvidenceGherkinContext, error) {
	featurePaths, err := acceptanceFeaturePaths(projectDir)
	if err != nil {
		return EvidenceGherkinContext{}, err
	}

	applies := TicketNeedsGherkin(ticket)
	active := config.GherkinPolicy == EvidenceGherkinPolicyActiveWhenApplicable && len(featurePaths) > 0

	context := EvidenceGherkinContext{
		Policy:       config.GherkinPolicy,
		Active:       active,
		Applies:      applies,
		FeaturePaths: featurePaths,
		Reason:       evidenceGherkinReason(config, featurePaths, applies),
	}
	if active && applies {
		context.RequiredPath = evidenceRequiredGherkinPath(featurePaths, ticket)
	}

	return context, nil
}

// TicketNeedsGherkin reports whether a ticket affects observable behavior.
func TicketNeedsGherkin(ticket IndexedTicket) bool {
	kind := evidenceChangeKind(ticket)
	switch kind {
	case "behavior", "cli_behavior", "tui_behavior", "api_behavior":
		return true
	case "bugfix":
		return hasUserVisibleSignal(ticket)
	default:
		return false
	}
}

// EvidenceGuidanceForTicket returns default evidence guidance for a ticket.
func EvidenceGuidanceForTicket(ticket IndexedTicket, gherkin EvidenceGherkinContext) EvidenceStrategyGuidance {
	kind := evidenceChangeKind(ticket)
	guidance := EvidenceStrategyGuidance{
		ChangeKind: kind,
		Risk:       evidenceRisk(ticket, kind),
		TestTiming: evidenceTestTiming(kind),
		Tests:      evidenceTestGuidance(kind),
		Gherkin:    evidenceGherkinGuidance(gherkin),
		Reason:     evidenceGuidanceReason(kind),
	}

	return guidance
}

// EvidenceDecisionMatrixRows returns the default evidence decision matrix.
func EvidenceDecisionMatrixRows() []EvidenceDecisionMatrixRow {
	return []EvidenceDecisionMatrixRow{
		{"docs_only, godoc_only, metadata_only", "No new tests. Validate links, formatting, or repository docs checks when available."},
		{"test_only", "Run the relevant test command and repository validation. No Gherkin unless observable behavior is being specified."},
		{"refactor_no_behavior", "Run existing tests and validation. Add tests only when the refactor exposes an obvious missing safety net."},
		{"bugfix", "Prefer a failing regression test before or during implementation. Add Gherkin when user-visible behavior changes and Gherkin is active."},
		{"behavior, cli_behavior, tui_behavior, api_behavior", "Add or update tests for the behavior. Add or update Gherkin when active and applicable."},
		{"parsing_or_validation", "Prefer table-driven tests and fuzz/property-style checks when the input space warrants it."},
		{"state_machine", "Prefer table-driven transition tests, invariants, and invalid-state checks."},
		{"persistence_or_filesystem", "Prefer integration tests with temp directories, fixtures, and failure cases."},
		{"network_or_external_io", "Prefer fakes, timeout/failure cases, and integration tests when practical."},
		{"concurrency_or_timeout", "Prefer boundedness checks, timeout tests, race-oriented validation, and failure-path tests."},
		{"configuration", "Test defaulting, invalid config, and override behavior."},
		{"performance", "Add benchmarks or boundedness evidence only when the ticket changes performance-sensitive behavior."},
		{"security_or_safety", "Require explicit tests or validation evidence unless impossible; missing evidence blocks review."},
		{"unknown", "Default to conservative evidence and do not move to review without a clear strategy."},
	}
}

func acceptanceFeaturePaths(projectDir string) ([]string, error) {
	acceptanceDir := filepath.Join(projectDir, ".chr", "default", acceptancecore.AcceptanceFeatureDirectory)
	entries := make([]string, 0)

	err := filepath.WalkDir(acceptanceDir, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}

		if entry.IsDir() || filepath.Ext(path) != ".feature" {
			return nil
		}

		relative, err := filepath.Rel(projectDir, path)
		if err != nil {
			return err
		}

		entries = append(entries, filepath.ToSlash(relative))

		return nil
	})
	if os.IsNotExist(err) {
		return nil, nil
	}

	if err != nil {
		return nil, fmt.Errorf("scan acceptance feature artifacts: %w", err)
	}

	sort.Strings(entries)

	return entries, nil
}

func evidenceGherkinReason(config EvidenceStrategyConfig, featurePaths []string, applies bool) string {
	switch {
	case config.GherkinPolicy == EvidenceGherkinPolicyDisabled:
		return "gherkin_policy=disabled"
	case config.GherkinPolicy == EvidenceGherkinPolicyManualOnly:
		return "gherkin_policy=manual_only"
	case len(featurePaths) == 0:
		return "no acceptance feature artifacts found"
	case applies:
		return "active acceptance feature artifacts and observable behavior"
	default:
		return "active acceptance feature artifacts, but ticket is not observable behavior"
	}
}

func evidenceRequiredGherkinPath(featurePaths []string, ticket IndexedTicket) string {
	if len(featurePaths) == 1 {
		return featurePaths[0]
	}

	useCase := normalizeEvidenceToken(ticket.UseCase)
	for _, path := range featurePaths {
		if useCase != "" && strings.Contains(normalizeEvidenceToken(path), useCase) {
			return path
		}
	}

	return featurePaths[0]
}

func evidenceChangeKind(ticket IndexedTicket) string {
	text := strings.ToLower(strings.Join(evidenceTicketTerms(ticket), " "))

	if ticket.Type == "docs" {
		return "docs_only"
	}

	if strings.Contains(text, "godoc") {
		return "godoc_only"
	}

	if ticket.Type == "bug" {
		return "bugfix"
	}

	kind := evidenceKeywordChangeKind(text)
	if kind != "" {
		return kind
	}

	if ticket.Type == "feature" || strings.TrimSpace(ticket.UseCase) != "" {
		return "behavior"
	}

	if ticket.Type == "cleanup" {
		return "refactor_no_behavior"
	}

	return "unknown"
}

func evidenceTicketTerms(ticket IndexedTicket) []string {
	terms := []string{ticket.Title, ticket.Synopsis, ticket.Type, ticket.UseCase, ticket.ProjectArea}
	terms = append(terms, ticket.Tags...)

	return terms
}

func evidenceKeywordChangeKind(text string) string {
	for _, rule := range evidenceKeywordRules() {
		if evidenceTextContainsAny(text, rule.Terms) {
			return rule.Kind
		}
	}

	return ""
}

func evidenceKeywordRules() []evidenceKeywordRule {
	return []evidenceKeywordRule{
		{Kind: "parsing_or_validation", Terms: []string{"parser", "parse", "validation", "validator"}},
		{Kind: "state_machine", Terms: []string{"state machine", "transition"}},
		{Kind: "persistence_or_filesystem", Terms: []string{"filesystem", "file ", "sqlite", "persistence"}},
		{Kind: "network_or_external_io", Terms: []string{"network", "http", "api client"}},
		{Kind: "concurrency_or_timeout", Terms: []string{"concurrency", "timeout", "worker", "queue"}},
		{Kind: "configuration", Terms: []string{"config", "default", "override"}},
		{Kind: "cli_behavior", Terms: []string{"cli"}},
		{Kind: "tui_behavior", Terms: []string{"tui", "board", "dash"}},
		{Kind: "api_behavior", Terms: []string{"api"}},
	}
}

func evidenceTextContainsAny(text string, terms []string) bool {
	for _, term := range terms {
		if strings.Contains(text, term) {
			return true
		}
	}

	return false
}

func evidenceRisk(ticket IndexedTicket, kind string) string {
	if ticket.Severity == "critical" || ticket.Priority == "high" {
		return "high"
	}

	switch kind {
	case "docs_only", "godoc_only", "metadata_only":
		return "low"
	case "unknown", "security_or_safety", "concurrency_or_timeout", "state_machine":
		return "high"
	case "bugfix", "behavior", "cli_behavior", "tui_behavior", "api_behavior", "parsing_or_validation":
		return "medium"
	default:
		return "medium"
	}
}

func evidenceTestTiming(kind string) string {
	switch kind {
	case "docs_only", "godoc_only", "metadata_only":
		return "none"
	case "bugfix", "parsing_or_validation", "state_machine":
		return "before"
	case "behavior", "cli_behavior", "tui_behavior", "api_behavior":
		return "during"
	default:
		return "during"
	}
}

func evidenceTestGuidance(kind string) string {
	switch kind {
	case "docs_only", "godoc_only", "metadata_only":
		return "No new tests; run available docs or formatting validation."
	case "bugfix":
		return "Prefer a regression test before or during implementation."
	case "parsing_or_validation":
		return "Prefer table-driven tests and fuzz/property-style checks when input space warrants it."
	case "state_machine":
		return "Prefer table-driven transition tests, invariants, and invalid-state checks."
	case "persistence_or_filesystem":
		return "Prefer integration tests with temp directories, fixtures, and failure cases."
	case "network_or_external_io":
		return "Prefer fakes, timeout/failure cases, and integration tests when practical."
	case "concurrency_or_timeout":
		return "Prefer boundedness checks, timeout tests, race-oriented validation, and failure-path tests."
	case "configuration":
		return "Test defaults, invalid config, and override behavior."
	default:
		return "Add or update tests matching the behavior and risk."
	}
}

func evidenceGherkinGuidance(gherkin EvidenceGherkinContext) string {
	switch {
	case gherkin.Policy == EvidenceGherkinPolicyDisabled:
		return "disabled"
	case gherkin.Policy == EvidenceGherkinPolicyManualOnly:
		return "manual_only"
	case gherkin.Active && gherkin.Applies:
		return "required"
	case gherkin.Active:
		return "not_applicable"
	default:
		return "inactive"
	}
}

func evidenceGuidanceReason(kind string) string {
	switch kind {
	case "docs_only", "godoc_only", "metadata_only":
		return "No runtime behavior changed."
	case "unknown":
		return "Classify conservatively before review."
	default:
		return "Use the evidence decision matrix defaults unless ticket instructions override them."
	}
}

func hasUserVisibleSignal(ticket IndexedTicket) bool {
	text := strings.ToLower(strings.Join(evidenceTicketTerms(ticket), " "))

	return strings.Contains(text, "user") ||
		strings.Contains(text, "cli") ||
		strings.Contains(text, "tui") ||
		strings.Contains(text, "api") ||
		strings.TrimSpace(ticket.UseCase) != ""
}

func normalizeEvidenceToken(value string) string {
	return strings.NewReplacer("-", "", "_", "", " ", "", "/", "").Replace(strings.ToLower(value))
}
