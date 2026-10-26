package core

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDetectEvidenceGherkinContextUsesAcceptanceFeatureArtifacts(t *testing.T) {
	projectDir := t.TempDir()
	writeEvidencePolicyFeature(t, projectDir, ".chr/default/acceptance/checkout.feature")

	ticket := testEvidencePolicyTicket(func(ticket *IndexedTicket) {
		ticket.Type = "feature"
		ticket.UseCase = "checkout"
		ticket.ProjectArea = "cli"
	})

	context, err := DetectEvidenceGherkinContext(projectDir, DefaultEvidenceStrategyConfig(), ticket)
	if err != nil {
		t.Fatalf("DetectEvidenceGherkinContext returned error: %v", err)
	}

	if !context.Active || !context.Applies {
		t.Fatalf("context = %#v, want active applicable Gherkin", context)
	}

	if context.RequiredPath != ".chr/default/acceptance/checkout.feature" {
		t.Fatalf("RequiredPath = %q, want checkout feature", context.RequiredPath)
	}
}

func TestDetectEvidenceGherkinContextHonorsDisabledPolicy(t *testing.T) {
	projectDir := t.TempDir()
	writeEvidencePolicyFeature(t, projectDir, ".chr/default/acceptance/checkout.feature")

	config := DefaultEvidenceStrategyConfig()
	config.GherkinPolicy = EvidenceGherkinPolicyDisabled

	context, err := DetectEvidenceGherkinContext(projectDir, config, testEvidencePolicyTicket(func(ticket *IndexedTicket) {
		ticket.Type = "feature"
		ticket.UseCase = "checkout"
	}))
	if err != nil {
		t.Fatalf("DetectEvidenceGherkinContext returned error: %v", err)
	}

	if context.Active || context.RequiredPath != "" {
		t.Fatalf("context = %#v, want inactive disabled Gherkin", context)
	}
}

func TestEvidenceGuidanceForTicketClassifiesDocsOnlyNoTests(t *testing.T) {
	ticket := testEvidencePolicyTicket(func(ticket *IndexedTicket) {
		ticket.Type = "docs"
		ticket.Title = "Update user guide"
	})

	guidance := EvidenceGuidanceForTicket(ticket, EvidenceGherkinContext{})

	assertEvidenceGuidance(t, guidance, "docs_only", "low", "none", "inactive")
}

func TestEvidenceGuidanceForTicketRequiresBehaviorGherkinWhenActive(t *testing.T) {
	ticket := testEvidencePolicyTicket(func(ticket *IndexedTicket) {
		ticket.Type = "feature"
		ticket.Title = "Add CLI checkout command"
		ticket.UseCase = "checkout"
	})
	gherkin := EvidenceGherkinContext{
		Policy:       EvidenceGherkinPolicyActiveWhenApplicable,
		Active:       true,
		Applies:      true,
		RequiredPath: ".chr/default/acceptance/checkout.feature",
	}

	guidance := EvidenceGuidanceForTicket(ticket, gherkin)

	assertEvidenceGuidance(t, guidance, "cli_behavior", "medium", "during", "required")
}

func TestEvidenceGuidanceForTicketPrefersParserTableFuzzCoverage(t *testing.T) {
	ticket := testEvidencePolicyTicket(func(ticket *IndexedTicket) {
		ticket.Type = "bug"
		ticket.Title = "Fix parser validation for quoted input"
	})

	guidance := EvidenceGuidanceForTicket(ticket, EvidenceGherkinContext{})

	if guidance.ChangeKind != "bugfix" {
		t.Fatalf("ChangeKind = %q, want bugfix", guidance.ChangeKind)
	}

	if guidance.TestTiming != "before" {
		t.Fatalf("TestTiming = %q, want before", guidance.TestTiming)
	}

	if guidance.Tests != "Prefer a regression test before or during implementation." {
		t.Fatalf("Tests = %q, want regression guidance", guidance.Tests)
	}
}

func TestEvidenceGuidanceForTicketClassifiesParserValidation(t *testing.T) {
	ticket := testEvidencePolicyTicket(func(ticket *IndexedTicket) {
		ticket.Type = "task"
		ticket.Title = "Add parser validation table"
	})

	guidance := EvidenceGuidanceForTicket(ticket, EvidenceGherkinContext{})

	if guidance.ChangeKind != "parsing_or_validation" {
		t.Fatalf("ChangeKind = %q, want parsing_or_validation", guidance.ChangeKind)
	}

	if guidance.Tests != "Prefer table-driven tests and fuzz/property-style checks when input space warrants it." {
		t.Fatalf("Tests = %q, want parser test guidance", guidance.Tests)
	}
}

func assertEvidenceGuidance(t *testing.T, guidance EvidenceStrategyGuidance, kind, risk, timing, gherkin string) {
	t.Helper()

	if guidance.ChangeKind != kind {
		t.Fatalf("ChangeKind = %q, want %q", guidance.ChangeKind, kind)
	}

	if guidance.Risk != risk {
		t.Fatalf("Risk = %q, want %q", guidance.Risk, risk)
	}

	if guidance.TestTiming != timing {
		t.Fatalf("TestTiming = %q, want %q", guidance.TestTiming, timing)
	}

	if guidance.Gherkin != gherkin {
		t.Fatalf("Gherkin = %q, want %q", guidance.Gherkin, gherkin)
	}
}

func testEvidencePolicyTicket(edit func(*IndexedTicket)) IndexedTicket {
	ticket := IndexedTicket{
		IndexedTicketIdentity: IndexedTicketIdentity{Key: "CHR-1"}, IndexedTicketSummary: IndexedTicketSummary{Title: "Do work"}, IndexedTicketClassification: IndexedTicketClassification{
			Type:     "task",
			Priority: "normal",
			Severity: "normal",
		},
	}
	edit(&ticket)

	return ticket
}

func writeEvidencePolicyFeature(t *testing.T, projectDir, relativePath string) {
	t.Helper()

	path := filepath.Join(projectDir, relativePath)

	err := os.MkdirAll(filepath.Dir(path), 0o755)
	if err != nil {
		t.Fatalf("create feature dir: %v", err)
	}

	err = os.WriteFile(path, []byte("Feature: Checkout\n"), 0o644)
	if err != nil {
		t.Fatalf("write feature: %v", err)
	}
}
