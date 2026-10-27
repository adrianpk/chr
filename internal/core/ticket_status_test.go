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

	content, err := ticketcore.NewTicketRenderer().RenderTicket(ticketcore.Ticket{
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
		ticketcore.TicketStatusBacklog,
		ticketcore.TicketStatusReadyForDev,
		ticketcore.TicketStatusInProgress,
		ticketcore.TicketStatusReadyForReview,
		ticketcore.TicketStatusReadyForMerge,
		ticketcore.TicketStatusDone,
		ticketcore.TicketStatusOpenLegacy,
		ticketcore.TicketStatusReadyLegacy,
		ticketcore.TicketStatusReviewingLegacy,
		ticketcore.TicketStatusBlockedLegacy,
		ticketcore.TicketStatusSolvedLegacy,
		ticketcore.TicketStatusMergedLegacy,
		ticketcore.TicketStatusClosedLegacy,
	} {
		t.Run(status, func(t *testing.T) {
			if !ticketcore.IsValidTicketStatus(status) {
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
		{name: "normal", status: ticketcore.TicketStatusReadyForDev, want: ticketcore.TicketStatusReadyForDev, wantOK: true},
		{name: "open legacy", status: ticketcore.TicketStatusOpenLegacy, want: ticketcore.TicketStatusBacklog, wantOK: true},
		{name: "ready legacy", status: ticketcore.TicketStatusReadyLegacy, want: ticketcore.TicketStatusReadyForDev, wantOK: true},
		{name: "reviewing legacy", status: ticketcore.TicketStatusReviewingLegacy, want: ticketcore.TicketStatusReadyForReview, wantOK: true},
		{name: "solved legacy", status: ticketcore.TicketStatusSolvedLegacy, want: ticketcore.TicketStatusReadyForMerge, wantOK: true},
		{name: "merged legacy", status: ticketcore.TicketStatusMergedLegacy, want: ticketcore.TicketStatusDone, wantOK: true},
		{name: "closed legacy", status: ticketcore.TicketStatusClosedLegacy, want: ticketcore.TicketStatusDone, wantOK: true},
		{name: "blocked legacy", status: ticketcore.TicketStatusBlockedLegacy, wantOK: false},
		{name: "unknown", status: "unknown", wantOK: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := ticketcore.BoardTicketStatusFor(tt.status)
			if ok != tt.wantOK {
				t.Fatalf("ok = %v, want %v", ok, tt.wantOK)
			}

			if got != tt.want {
				t.Fatalf("status = %q, want %q", got, tt.want)
			}
		})
	}
}
