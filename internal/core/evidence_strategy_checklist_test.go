package core

import (
	"strings"
	"testing"
)

func TestParseEvidenceStrategyChecklistParsesAnchoredBlock(t *testing.T) {
	checklist, err := ParseEvidenceStrategyChecklist(testEvidenceStrategyChecklistMarkdown())
	if err != nil {
		t.Fatalf("ParseEvidenceStrategyChecklist returned error: %v", err)
	}

	if !checklist.Found {
		t.Fatal("Found = false, want true")
	}

	strategy := assertEvidenceStrategyChecklistItem(t, checklist, EvidenceStrategyChecklistLabelStrategy)
	if !strategy.Checked {
		t.Fatal("strategy Checked = false, want true")
	}
	if strategy.Status != "planned" {
		t.Fatalf("strategy Status = %q, want planned", strategy.Status)
	}
	if strategy.Detail != "change_kind=bugfix; risk=medium; test_timing=before; note=regression" {
		t.Fatalf("strategy Detail = %q", strategy.Detail)
	}

	gherkin := assertEvidenceStrategyChecklistItem(t, checklist, EvidenceStrategyChecklistLabelGherkin)
	if gherkin.Checked {
		t.Fatal("gherkin Checked = true, want false")
	}
	if gherkin.Status != "required" {
		t.Fatalf("gherkin Status = %q, want required", gherkin.Status)
	}
}

func TestParseEvidenceStrategyChecklistReportsMissingBlock(t *testing.T) {
	checklist, err := ParseEvidenceStrategyChecklist("# Ticket\n\nNo strategy yet.\n")
	if err != nil {
		t.Fatalf("ParseEvidenceStrategyChecklist returned error: %v", err)
	}

	if checklist.Found {
		t.Fatal("Found = true, want false")
	}
}

func TestParseEvidenceStrategyChecklistRejectsMalformedBlocks(t *testing.T) {
	for _, tt := range []struct {
		name    string
		content string
		wantErr string
	}{
		{
			name:    "end without start",
			content: "# Ticket\n" + EvidenceStrategyChecklistEndAnchor + "\n",
			wantErr: "end anchor without start anchor",
		},
		{
			name: "start without end",
			content: "# Ticket\n" +
				EvidenceStrategyChecklistStartAnchor + "\n" +
				"- [ ] Evidence strategy: planned\n",
			wantErr: "start anchor without end anchor",
		},
		{
			name: "duplicate label",
			content: strings.Replace(
				testEvidenceStrategyChecklistMarkdown(),
				"- [ ] Evidence summary: pending",
				"- [ ] Evidence summary: pending\n- [ ] Evidence summary: pending",
				1,
			),
			wantErr: "duplicate evidence strategy checklist label: Evidence summary",
		},
		{
			name: "missing label",
			content: strings.Replace(
				testEvidenceStrategyChecklistMarkdown(),
				"- [ ] Tests: mixed; reason=unit and fuzz coverage\n",
				"",
				1,
			),
			wantErr: "missing evidence strategy checklist label: Tests",
		},
		{
			name: "invalid checked state",
			content: strings.Replace(
				testEvidenceStrategyChecklistMarkdown(),
				"- [ ] Tests: mixed",
				"- [?] Tests: mixed",
				1,
			),
			wantErr: "invalid checklist checked state",
		},
		{
			name: "unknown label",
			content: strings.Replace(
				testEvidenceStrategyChecklistMarkdown(),
				"- [ ] Tests: mixed",
				"- [ ] Test plan: mixed",
				1,
			),
			wantErr: "unknown evidence strategy checklist label: Test plan",
		},
		{
			name: "invalid status",
			content: strings.Replace(
				testEvidenceStrategyChecklistMarkdown(),
				"- [ ] Validation: pending",
				"- [ ] Validation: maybe",
				1,
			),
			wantErr: "invalid Validation status: maybe",
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			_, err := ParseEvidenceStrategyChecklist(tt.content)
			if err == nil {
				t.Fatal("ParseEvidenceStrategyChecklist returned nil error")
			}

			if !strings.Contains(err.Error(), tt.wantErr) {
				t.Fatalf("error = %q, want %q", err.Error(), tt.wantErr)
			}
		})
	}
}

func assertEvidenceStrategyChecklistItem(
	t *testing.T,
	checklist EvidenceStrategyChecklist,
	label string,
) EvidenceStrategyChecklistItem {
	t.Helper()

	item, ok := checklist.Item(label)
	if !ok {
		t.Fatalf("missing item %s", label)
	}

	return item
}

func testEvidenceStrategyChecklistMarkdown() string {
	return `---
key: "CHC-1"
---

# Description

<!-- chc:evidence-strategy:v1 -->
  - [x] Evidence strategy: planned; change_kind=bugfix; risk=medium; test_timing=before; note=regression
- [ ] Gherkin: required; path=.chc/user/acceptance/bug.feature; reason=user-visible
- [ ] Tests: mixed; reason=unit and fuzz coverage
- [ ] Validation: pending; commands=make verify; evidence=none
- [ ] Evidence summary: pending
<!-- /chc:evidence-strategy -->
`
}
