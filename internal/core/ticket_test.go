package core

import (
	"strings"
	"testing"
	"time"
)

func TestTicketRendererRendersFrontmatterAndBody(t *testing.T) {
	renderer := NewTicketRenderer()

	content, err := renderer.RenderTicket(testRenderedTicket())
	if err != nil {
		t.Fatalf("RenderTicket returned error: %v", err)
	}

	if content != wantRenderedTicket {
		t.Fatalf("content =\n%s\nwant =\n%s", content, wantRenderedTicket)
	}
}

func testRenderedTicket() Ticket {
	createdAt := time.Date(2026, 5, 18, 12, 30, 0, 0, time.UTC)
	updatedAt := time.Date(2026, 5, 18, 13, 45, 0, 0, time.UTC)

	return Ticket{
		Key:         "CHR-1",
		Title:       "Do something",
		Synopsis:    "Implement the onboarding flow.",
		Priority:    "high",
		Severity:    "critical",
		Type:        "feature",
		Tags:        []string{"uxui", "onboarding"},
		UseCase:     "onboarding",
		ProjectArea: "frontend",
		SourceKind:  "plan_tracking",
		SourceRef:   ".chr/user/plans/onboarding-plan.md",
		Slice:       "Slice 1",
		Branch:      "feat/slice-1-onboarding",
		ReviewRefs: []TicketReviewRef{
			{
				Type:     "pull_request",
				Provider: "forgejo",
				URL:      "https://forge.example.com/owner/repo/pulls/123",
				ID:       "123",
				Status:   "open",
			},
		},
		ReviewState:       "changes_requested",
		ReviewRound:       1,
		ReviewResult:      "needs_changes",
		ReviewCommentsURL: "https://forge.example.com/owner/repo/pulls/123",
		BlockingFindings:  []string{"RF-1"},
		MergeState:        "blocked",
		MergeResult:       "not_mergeable",
		MergeActor:        "merger",
		MergeCommit:       "abc123",
		MergedAt:          time.Date(2026, 5, 18, 14, 0, 0, 0, time.UTC),
		MergeBlocker:      "merge conflict",
		CreatedAt:         createdAt,
		UpdatedAt:         updatedAt,
	}
}

const wantRenderedTicket = `---
key: "CHR-1"
title: "Do something"
synopsis: "Implement the onboarding flow."
status: "open"
priority: "high"
severity: "critical"
type: "feature"
tags: ["uxui", "onboarding"]
use_case: "onboarding"
project_area: "frontend"
source_kind: "plan_tracking"
source_ref: ".chr/user/plans/onboarding-plan.md"
slice: "Slice 1"
branch: "feat/slice-1-onboarding"
review_refs:
  - type: "pull_request"
    provider: "forgejo"
    url: "https://forge.example.com/owner/repo/pulls/123"
    id: "123"
    status: "open"
review_state: "changes_requested"
review_result: "needs_changes"
review_comments_url: "https://forge.example.com/owner/repo/pulls/123"
review_round: 1
blocking_findings: ["RF-1"]
merge_state: "blocked"
merge_result: "not_mergeable"
merge_actor: "merger"
merge_commit: "abc123"
merge_blocker: "merge conflict"
merged_at: "2026-05-18T14:00:00Z"
created_at: "2026-05-18T12:30:00Z"
updated_at: "2026-05-18T13:45:00Z"
---

# Description

# Checklist

# Acceptance

# Comments

# Blocker

# Validation
`

func TestTicketRendererKeepsOptionalFieldsEmpty(t *testing.T) {
	renderer := NewTicketRenderer()

	content, err := renderer.RenderTicket(Ticket{Key: "CHR-1", Title: "Do something"})
	if err != nil {
		t.Fatalf("RenderTicket returned error: %v", err)
	}

	for _, want := range []string{
		`synopsis: ""`,
		`priority: ""`,
		`severity: ""`,
		`type: ""`,
		`tags: []`,
		`use_case: ""`,
		`project_area: ""`,
		`source_kind: ""`,
		`source_ref: ""`,
		`slice: ""`,
		`branch: ""`,
		`review_refs: []`,
		`review_state: ""`,
		`review_result: ""`,
		`review_comments_url: ""`,
		`review_round: ""`,
		`blocking_findings: []`,
		`merge_state: ""`,
		`merge_result: ""`,
		`merge_actor: ""`,
		`merge_commit: ""`,
		`merge_blocker: ""`,
		`merged_at: ""`,
		`created_at: ""`,
		`updated_at: ""`,
	} {
		if !strings.Contains(content, want) {
			t.Fatalf("content does not contain %q:\n%s", want, content)
		}
	}
}
