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

func TestRenderEvidenceStrategyChecklistRendersCanonicalBlock(t *testing.T) {
	checklist := EvidenceStrategyChecklist{
		Found: true,
		Items: map[string]EvidenceStrategyChecklistItem{
			EvidenceStrategyChecklistLabelStrategy: {
				Label:   EvidenceStrategyChecklistLabelStrategy,
				Checked: true,
				Status:  "planned",
				Detail:  "change_kind=bugfix; risk=medium",
			},
			EvidenceStrategyChecklistLabelGherkin: {
				Label:  EvidenceStrategyChecklistLabelGherkin,
				Status: "required",
				Detail: "path=.chc/user/acceptance/bug.feature; reason=user-visible",
			},
			EvidenceStrategyChecklistLabelTests: {
				Label:  EvidenceStrategyChecklistLabelTests,
				Status: "mixed",
				Detail: "reason=unit and fuzz coverage",
			},
			EvidenceStrategyChecklistLabelValidation: {
				Label:  EvidenceStrategyChecklistLabelValidation,
				Status: "pending",
				Detail: "commands=make verify; evidence=none",
			},
			EvidenceStrategyChecklistLabelSummary: {
				Label:  EvidenceStrategyChecklistLabelSummary,
				Status: "pending",
			},
		},
	}

	got, err := RenderEvidenceStrategyChecklist(checklist)
	if err != nil {
		t.Fatalf("RenderEvidenceStrategyChecklist returned error: %v", err)
	}

	want := `<!-- chc:evidence-strategy:v1 -->
- [x] Evidence strategy: planned; change_kind=bugfix; risk=medium
- [ ] Gherkin: required; path=.chc/user/acceptance/bug.feature; reason=user-visible
- [ ] Tests: mixed; reason=unit and fuzz coverage
- [ ] Validation: pending; commands=make verify; evidence=none
- [ ] Evidence summary: pending
<!-- /chc:evidence-strategy -->
`
	if got != want {
		t.Fatalf("rendered checklist:\n%s\nwant:\n%s", got, want)
	}
}

func TestRenderEvidenceStrategyChecklistRejectsMissingItems(t *testing.T) {
	_, err := RenderEvidenceStrategyChecklist(EvidenceStrategyChecklist{
		Items: map[string]EvidenceStrategyChecklistItem{},
	})
	if err == nil {
		t.Fatal("RenderEvidenceStrategyChecklist returned nil error")
	}

	if !strings.Contains(err.Error(), "missing evidence strategy checklist label: Evidence strategy") {
		t.Fatalf("error = %q", err.Error())
	}
}

func TestUpsertEvidenceStrategyChecklistInsertsMissingBlock(t *testing.T) {
	content := `---
key: "CHC-1"
title: "Ticket"
updated_at: "2026-06-19T10:00:00Z"
---

# Description

Keep this body.

# Comments

Keep comments.

# Blocker

Keep blockers.
`

	updated, err := UpsertEvidenceStrategyChecklist(content, testEvidenceStrategyChecklistModel())
	if err != nil {
		t.Fatalf("UpsertEvidenceStrategyChecklist returned error: %v", err)
	}

	assertTextOrder(t, updated, `updated_at: "2026-06-19T10:00:00Z"`, EvidenceStrategyChecklistStartAnchor)
	assertTextOrder(t, updated, EvidenceStrategyChecklistEndAnchor, "# Description")

	for _, want := range []string{
		"Keep this body.",
		"# Comments\n\nKeep comments.",
		"# Blocker\n\nKeep blockers.",
	} {
		if !strings.Contains(updated, want) {
			t.Fatalf("updated content missing %q:\n%s", want, updated)
		}
	}
}

func TestUpsertEvidenceStrategyChecklistUpdatesOnlyProvidedItems(t *testing.T) {
	updates := EvidenceStrategyChecklist{
		Items: map[string]EvidenceStrategyChecklistItem{
			EvidenceStrategyChecklistLabelValidation: {
				Label:   EvidenceStrategyChecklistLabelValidation,
				Checked: true,
				Status:  "passed",
				Detail:  "commands=make verify; evidence=.chc/user/runtime/runs/1/report.md",
			},
		},
	}

	updated, err := UpsertEvidenceStrategyChecklist(testEvidenceStrategyChecklistMarkdown(), updates)
	if err != nil {
		t.Fatalf("UpsertEvidenceStrategyChecklist returned error: %v", err)
	}

	checklist, err := ParseEvidenceStrategyChecklist(updated)
	if err != nil {
		t.Fatalf("ParseEvidenceStrategyChecklist returned error: %v", err)
	}

	validation := assertEvidenceStrategyChecklistItem(t, checklist, EvidenceStrategyChecklistLabelValidation)
	if !validation.Checked || validation.Status != "passed" {
		t.Fatalf("validation item = %#v, want checked passed", validation)
	}

	tests := assertEvidenceStrategyChecklistItem(t, checklist, EvidenceStrategyChecklistLabelTests)
	if tests.Status != "mixed" || tests.Detail != "reason=unit and fuzz coverage" {
		t.Fatalf("tests item = %#v, want preserved mixed detail", tests)
	}
}

func TestUpsertEvidenceStrategyChecklistRejectsMalformedExistingBlock(t *testing.T) {
	content := strings.Replace(
		testEvidenceStrategyChecklistMarkdown(),
		"- [ ] Validation: pending",
		"- [ ] Validation: maybe",
		1,
	)

	_, err := UpsertEvidenceStrategyChecklist(content, testEvidenceStrategyChecklistModel())
	if err == nil {
		t.Fatal("UpsertEvidenceStrategyChecklist returned nil error")
	}

	if !strings.Contains(err.Error(), "invalid Validation status: maybe") {
		t.Fatalf("error = %q", err.Error())
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

func assertTextOrder(t *testing.T, content, before, after string) {
	t.Helper()

	beforeIndex := strings.Index(content, before)
	if beforeIndex < 0 {
		t.Fatalf("content missing %q:\n%s", before, content)
	}

	afterIndex := strings.Index(content, after)
	if afterIndex < 0 {
		t.Fatalf("content missing %q:\n%s", after, content)
	}

	if beforeIndex >= afterIndex {
		t.Fatalf("%q appears after %q in:\n%s", before, after, content)
	}
}

func testEvidenceStrategyChecklistModel() EvidenceStrategyChecklist {
	return EvidenceStrategyChecklist{
		Found: true,
		Items: map[string]EvidenceStrategyChecklistItem{
			EvidenceStrategyChecklistLabelStrategy: {
				Label:  EvidenceStrategyChecklistLabelStrategy,
				Status: "planned",
				Detail: "change_kind=bugfix; risk=medium; test_timing=before; note=regression",
			},
			EvidenceStrategyChecklistLabelGherkin: {
				Label:  EvidenceStrategyChecklistLabelGherkin,
				Status: "required",
				Detail: "path=.chc/user/acceptance/bug.feature; reason=user-visible",
			},
			EvidenceStrategyChecklistLabelTests: {
				Label:  EvidenceStrategyChecklistLabelTests,
				Status: "mixed",
				Detail: "reason=unit and fuzz coverage",
			},
			EvidenceStrategyChecklistLabelValidation: {
				Label:  EvidenceStrategyChecklistLabelValidation,
				Status: "pending",
				Detail: "commands=make verify; evidence=none",
			},
			EvidenceStrategyChecklistLabelSummary: {
				Label:  EvidenceStrategyChecklistLabelSummary,
				Status: "pending",
			},
		},
	}
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
