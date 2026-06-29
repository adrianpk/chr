package core

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestTicketReviewMetadataUpdaterAddsBranchAndReviewRef(t *testing.T) {
	path := writeReviewMetadataTicket(t, Ticket{
		Key:       "CHR-1",
		Title:     "Do something",
		Status:    TicketStatusInProgress,
		CreatedAt: testReviewMetadataTime(12),
		UpdatedAt: testReviewMetadataTime(12),
	})

	err := NewTicketReviewMetadataUpdater().UpdateReviewMetadata(context.Background(), UpdateTicketReviewMetadataOptions{
		Path:   path,
		Branch: "feat/slice-2-pr-reference-workflow",
		ReviewRef: &TicketReviewRef{
			Type:     "pull_request",
			Provider: "forgejo",
			URL:      "https://forge.example.com/adrianpk/chr/pulls/274",
			ID:       "274",
			Status:   "open",
		},
		UpdatedAt: testReviewMetadataTime(13),
	})
	if err != nil {
		t.Fatalf("UpdateReviewMetadata returned error: %v", err)
	}

	content := readReviewMetadataTicket(t, path)
	for _, want := range []string{
		`branch: "feat/slice-2-pr-reference-workflow"`,
		`review_refs:`,
		`  - type: "pull_request"`,
		`    provider: "forgejo"`,
		`    url: "https://forge.example.com/adrianpk/chr/pulls/274"`,
		`    id: "274"`,
		`    status: "open"`,
		`updated_at: "2026-05-18T13:00:00Z"`,
	} {
		if !strings.Contains(content, want) {
			t.Fatalf("ticket content does not contain %q:\n%s", want, content)
		}
	}
}

func TestTicketReviewMetadataUpdaterReplacesExistingReviewRef(t *testing.T) {
	path := writeReviewMetadataTicket(t, Ticket{
		Key:    "CHR-1",
		Title:  "Do something",
		Status: TicketStatusInProgress,
		ReviewRefs: []TicketReviewRef{
			{
				Type:     "pull_request",
				Provider: "forgejo",
				URL:      "https://forge.example.com/adrianpk/chr/pulls/274",
				ID:       "274",
				Status:   "open",
			},
		},
		CreatedAt: testReviewMetadataTime(12),
		UpdatedAt: testReviewMetadataTime(12),
	})

	err := NewTicketReviewMetadataUpdater().UpdateReviewMetadata(context.Background(), UpdateTicketReviewMetadataOptions{
		Path: path,
		ReviewRef: &TicketReviewRef{
			Type:     "pull_request",
			Provider: "forgejo",
			URL:      "https://forge.example.com/adrianpk/chr/pulls/274",
			ID:       "274",
			Status:   "merged",
		},
		UpdatedAt: testReviewMetadataTime(13),
	})
	if err != nil {
		t.Fatalf("UpdateReviewMetadata returned error: %v", err)
	}

	content := readReviewMetadataTicket(t, path)
	if strings.Count(content, `url: "https://forge.example.com/adrianpk/chr/pulls/274"`) != 1 {
		t.Fatalf("ticket content should contain one PR URL:\n%s", content)
	}

	if !strings.Contains(content, `status: "merged"`) {
		t.Fatalf("ticket content does not contain merged ref status:\n%s", content)
	}
}

func TestTicketStatusUpdaterRequiresReviewMetadataForReadyForReview(t *testing.T) {
	path := writeReviewMetadataTicket(t, Ticket{
		Key:       "CHR-1",
		Title:     "Do something",
		Status:    TicketStatusInProgress,
		CreatedAt: testReviewMetadataTime(12),
		UpdatedAt: testReviewMetadataTime(12),
	})

	err := NewTicketStatusUpdater().UpdateStatus(context.Background(), UpdateTicketStatusOptions{
		Path:      path,
		Status:    TicketStatusReadyForReview,
		UpdatedAt: testReviewMetadataTime(13),
	})
	if err == nil {
		t.Fatal("UpdateStatus returned nil error")
	}

	if !strings.Contains(err.Error(), "requires branch metadata") {
		t.Fatalf("err = %v, want branch metadata error", err)
	}
}

func TestTicketStatusUpdaterAcceptsReadyForReviewWithReviewMetadata(t *testing.T) {
	path := writeReviewMetadataTicket(t, Ticket{
		Key:    "CHR-1",
		Title:  "Do something",
		Status: TicketStatusInProgress,
		Branch: "feat/slice-2-pr-reference-workflow",
		ReviewRefs: []TicketReviewRef{
			{
				Type:     "pull_request",
				Provider: "forgejo",
				URL:      "https://forge.example.com/adrianpk/chr/pulls/274",
				ID:       "274",
				Status:   "open",
			},
		},
		CreatedAt: testReviewMetadataTime(12),
		UpdatedAt: testReviewMetadataTime(12),
	})

	err := NewTicketStatusUpdater().UpdateStatus(context.Background(), UpdateTicketStatusOptions{
		Path:      path,
		Status:    TicketStatusReadyForReview,
		UpdatedAt: testReviewMetadataTime(13),
	})
	if err != nil {
		t.Fatalf("UpdateStatus returned error: %v", err)
	}

	content := readReviewMetadataTicket(t, path)
	if !strings.Contains(content, `status: "ready_for_review"`) {
		t.Fatalf("ticket content does not contain ready_for_review:\n%s", content)
	}
}

func TestTicketStatusUpdaterRequiresPassedReviewForReadyForMerge(t *testing.T) {
	path := writeReviewMetadataTicket(t, Ticket{
		Key:         "CHR-1",
		Title:       "Do something",
		Status:      TicketStatusReadyForReview,
		Branch:      "feat/slice-1-merger-handoff-contract",
		ReviewState: "changes_requested",
		ReviewRefs: []TicketReviewRef{
			{
				Type:     "pull_request",
				Provider: "forgejo",
				URL:      "https://forge.example.com/adrianpk/chr/pulls/281",
				ID:       "281",
				Status:   "open",
			},
		},
		CreatedAt: testReviewMetadataTime(12),
		UpdatedAt: testReviewMetadataTime(12),
	})

	err := NewTicketStatusUpdater().UpdateStatus(context.Background(), UpdateTicketStatusOptions{
		Path:      path,
		Status:    TicketStatusReadyForMerge,
		UpdatedAt: testReviewMetadataTime(13),
	})
	if err == nil {
		t.Fatal("UpdateStatus returned nil error")
	}

	if !strings.Contains(err.Error(), "requires passed review metadata") {
		t.Fatalf("err = %v, want passed review metadata error", err)
	}
}

func TestTicketStatusUpdaterAcceptsReadyForMergeWithPassedReview(t *testing.T) {
	path := writeReviewMetadataTicket(t, Ticket{
		Key:          "CHR-1",
		Title:        "Do something",
		Status:       TicketStatusReadyForReview,
		Branch:       "feat/slice-1-merger-handoff-contract",
		ReviewState:  TicketReviewCommandResultPass,
		ReviewResult: TicketReviewCommandResultPass,
		ReviewRefs: []TicketReviewRef{
			{
				Type:     "pull_request",
				Provider: "forgejo",
				URL:      "https://forge.example.com/adrianpk/chr/pulls/281",
				ID:       "281",
				Status:   "open",
			},
		},
		CreatedAt: testReviewMetadataTime(12),
		UpdatedAt: testReviewMetadataTime(12),
	})

	err := NewTicketStatusUpdater().UpdateStatus(context.Background(), UpdateTicketStatusOptions{
		Path:      path,
		Status:    TicketStatusReadyForMerge,
		UpdatedAt: testReviewMetadataTime(13),
	})
	if err != nil {
		t.Fatalf("UpdateStatus returned error: %v", err)
	}

	content := readReviewMetadataTicket(t, path)
	if !strings.Contains(content, `status: "ready_for_merge"`) {
		t.Fatalf("ticket content does not contain ready_for_merge:\n%s", content)
	}
}

func writeReviewMetadataTicket(t *testing.T, ticket Ticket) string {
	t.Helper()

	content, err := NewTicketRenderer().RenderTicket(ticket)
	if err != nil {
		t.Fatalf("RenderTicket returned error: %v", err)
	}

	path := filepath.Join(t.TempDir(), "ticket.md")

	err = os.WriteFile(path, []byte(content), 0o644)
	if err != nil {
		t.Fatalf("write ticket: %v", err)
	}

	return path
}

func readReviewMetadataTicket(t *testing.T, path string) string {
	t.Helper()

	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read ticket: %v", err)
	}

	return string(content)
}

func testReviewMetadataTime(hour int) time.Time {
	return time.Date(2026, 5, 18, hour, 0, 0, 0, time.UTC)
}
