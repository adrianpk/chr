package workspace

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestTicketStatusUpdaterUpdatesStatus(t *testing.T) {
	path := filepath.Join(t.TempDir(), "ticket.md")
	createdAt := time.Date(2026, 5, 18, 12, 0, 0, 0, time.UTC)
	updatedAt := time.Date(2026, 5, 18, 13, 0, 0, 0, time.UTC)

	content, err := NewTicketRenderer().RenderTicket(Ticket{
		Key:       "CHC-1",
		Title:     "Do something",
		Status:    "open",
		CreatedAt: createdAt,
		UpdatedAt: createdAt,
	})
	if err != nil {
		t.Fatalf("RenderTicket returned error: %v", err)
	}

	err = os.WriteFile(path, []byte(content), 0o644)
	if err != nil {
		t.Fatalf("write ticket: %v", err)
	}

	err = NewTicketStatusUpdater().UpdateStatus(context.Background(), UpdateTicketStatusOptions{
		Path:      path,
		Status:    "ready",
		UpdatedAt: updatedAt,
	})
	if err != nil {
		t.Fatalf("UpdateStatus returned error: %v", err)
	}

	updatedContent, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read ticket: %v", err)
	}

	for _, want := range []string{
		`status: "ready"`,
		`updated_at: "2026-05-18T13:00:00Z"`,
	} {
		if !strings.Contains(string(updatedContent), want) {
			t.Fatalf("updated content does not contain %q:\n%s", want, string(updatedContent))
		}
	}
}

func TestTicketStatusUpdaterRejectsInvalidStatus(t *testing.T) {
	err := NewTicketStatusUpdater().UpdateStatus(context.Background(), UpdateTicketStatusOptions{
		Path:   filepath.Join(t.TempDir(), "ticket.md"),
		Status: "unknown",
	})
	if err == nil {
		t.Fatal("UpdateStatus returned nil error")
	}
}

func TestTicketStatusValidationAcceptsBoardAndLegacyStatuses(t *testing.T) {
	for _, status := range []string{
		TicketStatusBacklog,
		TicketStatusReadyForDev,
		TicketStatusInProgress,
		TicketStatusReadyForReview,
		TicketStatusReadyForMerge,
		TicketStatusDone,
		TicketStatusOpenLegacy,
		TicketStatusReadyLegacy,
		TicketStatusReviewingLegacy,
		TicketStatusBlockedLegacy,
		TicketStatusSolvedLegacy,
		TicketStatusMergedLegacy,
		TicketStatusClosedLegacy,
	} {
		t.Run(status, func(t *testing.T) {
			if !IsValidTicketStatus(status) {
				t.Fatalf("IsValidTicketStatus(%q) = false, want true", status)
			}
		})
	}
}

func TestBoardTicketStatusForMapsNormalAndLegacyStatuses(t *testing.T) {
	tests := []struct {
		name   string
		status string
		want   string
		wantOK bool
	}{
		{name: "normal", status: TicketStatusReadyForDev, want: TicketStatusReadyForDev, wantOK: true},
		{name: "open legacy", status: TicketStatusOpenLegacy, want: TicketStatusBacklog, wantOK: true},
		{name: "ready legacy", status: TicketStatusReadyLegacy, want: TicketStatusReadyForDev, wantOK: true},
		{name: "reviewing legacy", status: TicketStatusReviewingLegacy, want: TicketStatusReadyForReview, wantOK: true},
		{name: "solved legacy", status: TicketStatusSolvedLegacy, want: TicketStatusReadyForMerge, wantOK: true},
		{name: "merged legacy", status: TicketStatusMergedLegacy, want: TicketStatusDone, wantOK: true},
		{name: "closed legacy", status: TicketStatusClosedLegacy, want: TicketStatusDone, wantOK: true},
		{name: "blocked legacy", status: TicketStatusBlockedLegacy, wantOK: false},
		{name: "unknown", status: "unknown", wantOK: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := BoardTicketStatusFor(tt.status)
			if ok != tt.wantOK {
				t.Fatalf("ok = %v, want %v", ok, tt.wantOK)
			}

			if got != tt.want {
				t.Fatalf("status = %q, want %q", got, tt.want)
			}
		})
	}
}
