package core

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	ticketcore "forge.adrianpk.com/chr/internal/core/ticket"
)

func TestTicketStatusUpdaterUpdatesStatus(t *testing.T) {
	path := filepath.Join(t.TempDir(), "ticket.md")
	createdAt := time.Date(2026, 5, 18, 12, 0, 0, 0, time.UTC)
	updatedAt := time.Date(2026, 5, 18, 13, 0, 0, 0, time.UTC)

	content, err := ticketcore.NewRenderer().Render(ticketcore.Ticket{
		Key:       "CHR-1",
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
		ticketcore.StatusBacklog,
		ticketcore.StatusReadyForDev,
		ticketcore.StatusInProgress,
		ticketcore.StatusReadyForReview,
		ticketcore.StatusReadyForMerge,
		ticketcore.StatusDone,
		ticketcore.StatusOpenLegacy,
		ticketcore.StatusReadyLegacy,
		ticketcore.StatusReviewingLegacy,
		ticketcore.StatusBlockedLegacy,
		ticketcore.StatusSolvedLegacy,
		ticketcore.StatusMergedLegacy,
		ticketcore.StatusClosedLegacy,
	} {
		t.Run(status, func(t *testing.T) {
			if !ticketcore.ValidStatus(status) {
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
		{name: "normal", status: ticketcore.StatusReadyForDev, want: ticketcore.StatusReadyForDev, wantOK: true},
		{name: "open legacy", status: ticketcore.StatusOpenLegacy, want: ticketcore.StatusBacklog, wantOK: true},
		{name: "ready legacy", status: ticketcore.StatusReadyLegacy, want: ticketcore.StatusReadyForDev, wantOK: true},
		{name: "reviewing legacy", status: ticketcore.StatusReviewingLegacy, want: ticketcore.StatusReadyForReview, wantOK: true},
		{name: "solved legacy", status: ticketcore.StatusSolvedLegacy, want: ticketcore.StatusReadyForMerge, wantOK: true},
		{name: "merged legacy", status: ticketcore.StatusMergedLegacy, want: ticketcore.StatusDone, wantOK: true},
		{name: "closed legacy", status: ticketcore.StatusClosedLegacy, want: ticketcore.StatusDone, wantOK: true},
		{name: "blocked legacy", status: ticketcore.StatusBlockedLegacy, wantOK: false},
		{name: "unknown", status: "unknown", wantOK: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := ticketcore.BoardStatus(tt.status)
			if ok != tt.wantOK {
				t.Fatalf("ok = %v, want %v", ok, tt.wantOK)
			}

			if got != tt.want {
				t.Fatalf("status = %q, want %q", got, tt.want)
			}
		})
	}
}
