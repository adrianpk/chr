package core

import (
	"strings"
	"testing"
)

func TestSummarizeEvidenceStrategyChecklistClassifiesStatus(t *testing.T) {
	for _, tt := range evidenceStrategySummaryCases() {
		t.Run(tt.name, func(t *testing.T) {
			summary := SummarizeEvidenceStrategyChecklist(tt.content)
			if summary.Status != tt.want {
				t.Fatalf("Status = %q, want %q", summary.Status, tt.want)
			}
		})
	}
}

func TestSummarizeEvidenceStrategyChecklistReturnsCanonicalItems(t *testing.T) {
	summary := SummarizeEvidenceStrategyChecklist(completeEvidenceStrategySummaryMarkdown())
	if len(summary.Items) != len(EvidenceStrategyChecklistLabels()) {
		t.Fatalf("Items len = %d, want %d", len(summary.Items), len(EvidenceStrategyChecklistLabels()))
	}

	if summary.Items[0].Label != EvidenceStrategyChecklistLabelStrategy {
		t.Fatalf("first item = %q, want strategy", summary.Items[0].Label)
	}

	if !summary.Items[0].Checked || summary.Items[0].Status != "complete" {
		t.Fatalf("strategy item = %#v, want checked complete", summary.Items[0])
	}
}

type evidenceStrategySummaryCase struct {
	name    string
	content string
	want    EvidenceStrategySummaryStatus
}

func evidenceStrategySummaryCases() []evidenceStrategySummaryCase {
	return []evidenceStrategySummaryCase{
		{
			name:    "complete",
			content: completeEvidenceStrategySummaryMarkdown(),
			want:    EvidenceStrategySummaryStatusComplete,
		},
		{
			name:    "missing",
			content: "# Ticket\n\nNo strategy yet.\n",
			want:    EvidenceStrategySummaryStatusMissing,
		},
		{
			name: "malformed",
			content: strings.Replace(
				completeEvidenceStrategySummaryMarkdown(),
				EvidenceStrategyChecklistEndAnchor,
				"",
				1,
			),
			want: EvidenceStrategySummaryStatusMalformed,
		},
		{
			name: "blocked",
			content: strings.Replace(
				completeEvidenceStrategySummaryMarkdown(),
				"- [x] Tests: mixed",
				"- [ ] Tests: blocked",
				1,
			),
			want: EvidenceStrategySummaryStatusBlocked,
		},
		{
			name: "incomplete",
			content: strings.Replace(
				completeEvidenceStrategySummaryMarkdown(),
				"- [x] Validation: passed",
				"- [ ] Validation: pending",
				1,
			),
			want: EvidenceStrategySummaryStatusIncomplete,
		},
	}
}

func completeEvidenceStrategySummaryMarkdown() string {
	return `# Ticket

<!-- chr:evidence-strategy:v1 -->
- [x] Evidence strategy: complete; change_kind=behavior; risk=medium
- [x] Gherkin: required; path=.chr/user/acceptance/flow.feature
- [x] Tests: mixed; unit and CLI coverage
- [x] Validation: passed; make verify
- [x] Evidence summary: complete; ready for review
<!-- /chr:evidence-strategy -->
`
}
