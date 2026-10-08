package core

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestApplyTrustedAutoMergeReviewMovesTicketToReadyForMerge(t *testing.T) {
	ref := TicketReviewRef{
		Type:     "pull_request",
		Provider: "forgejo",
		URL:      "https://forge.example.com/owner/repo/pulls/300",
		ID:       "300",
		Status:   "open",
	}
	path := writeTrustedAutoMergeTicket(t, Ticket{
		Key:        "CHR-1",
		Title:      "Trusted ticket",
		Status:     TicketResultStatusReviewing,
		Branch:     "feat/chr-1",
		ReviewRefs: []TicketReviewRef{ref},
		CreatedAt:  testTrustedAutoMergeTime(),
		UpdatedAt:  testTrustedAutoMergeTime(),
	})

	result, err := ApplyTrustedAutoMergeReview(context.Background(), ApplyTrustedAutoMergeReviewOptions{
		Ticket: IndexedTicket{
			Key:        "CHR-1",
			Path:       path,
			Status:     TicketResultStatusReviewing,
			Branch:     "feat/chr-1",
			ReviewRefs: []TicketReviewRef{ref},
		},
		ReviewedAt: testTrustedAutoMergeTime(),
	})
	if err != nil {
		t.Fatalf("ApplyTrustedAutoMergeReview returned error: %v", err)
	}

	if result.Status != TicketStatusReadyForMerge || result.ReviewResult != TicketReviewCommandResultPass {
		t.Fatalf("result = %#v, want trusted pass ready_for_merge", result)
	}

	content := readTrustedAutoMergeTicket(t, path)
	for _, want := range []string{
		`status: "ready_for_merge"`,
		`review_state: "pass"`,
		`review_result: "pass"`,
		`review_comments_url: "https://forge.example.com/owner/repo/pulls/300"`,
	} {
		if !strings.Contains(content, want) {
			t.Fatalf("ticket content missing %q:\n%s", want, content)
		}
	}
}

func TestApplyTrustedAutoMergeReviewRequiresPullRequestRef(t *testing.T) {
	path := writeTrustedAutoMergeTicket(t, Ticket{
		Key:       "CHR-1",
		Title:     "Trusted ticket",
		Status:    TicketResultStatusReviewing,
		Branch:    "feat/chr-1",
		CreatedAt: testTrustedAutoMergeTime(),
		UpdatedAt: testTrustedAutoMergeTime(),
	})

	_, err := ApplyTrustedAutoMergeReview(context.Background(), ApplyTrustedAutoMergeReviewOptions{
		Ticket: IndexedTicket{
			Key:    "CHR-1",
			Path:   path,
			Status: TicketResultStatusReviewing,
			Branch: "feat/chr-1",
		},
		ReviewedAt: testTrustedAutoMergeTime(),
	})
	if err == nil || !strings.Contains(err.Error(), "requires a pull_request review ref") {
		t.Fatalf("error = %v, want missing pull_request ref", err)
	}
}

func writeTrustedAutoMergeTicket(t *testing.T, ticket Ticket) string {
	t.Helper()

	content, err := NewTicketRenderer().RenderTicket(ticket)
	if err != nil {
		t.Fatalf("RenderTicket returned error: %v", err)
	}

	path := filepath.Join(t.TempDir(), "CHR-1.md")
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("WriteFile returned error: %v", err)
	}

	return path
}

func readTrustedAutoMergeTicket(t *testing.T, path string) string {
	t.Helper()

	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile returned error: %v", err)
	}

	return string(content)
}

func testTrustedAutoMergeTime() time.Time {
	return time.Date(2026, 6, 20, 15, 0, 0, 0, time.UTC)
}
