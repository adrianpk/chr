package ticket_test

import (
	"strings"
	"testing"
	"time"

	ticketcore "forge.adrianpk.com/chr/internal/core/ticket"
	ticketmetadata "forge.adrianpk.com/chr/internal/core/ticket/metadata"
)

func TestValidateIndexedTicketAcceptsBoardMetadata(t *testing.T) {
	now := time.Date(2026, 5, 26, 10, 0, 0, 0, time.UTC)
	ticket := validIndexedTicket(now)
	ticket.Status = ticketcore.StatusBacklog
	ticket.Synopsis = "Implement the Board ticket creation path."
	ticket.SourceKind = "manual"

	err := ticketcore.ValidateIndexed(ticket)
	if err != nil {
		t.Fatalf("ValidateIndexedTicket returned error: %v", err)
	}
}

func TestValidateIndexedTicketAcceptsLegacyTicketWithoutSynopsis(t *testing.T) {
	now := time.Date(2026, 5, 26, 10, 0, 0, 0, time.UTC)
	ticket := validIndexedTicket(now)
	ticket.Status = ticketcore.StatusOpenLegacy
	ticket.Synopsis = ""

	err := ticketcore.ValidateIndexed(ticket)
	if err != nil {
		t.Fatalf("ValidateIndexedTicket returned error: %v", err)
	}
}

func TestValidateIndexedTicketRejectsUnknownStatus(t *testing.T) {
	now := time.Date(2026, 5, 26, 10, 0, 0, 0, time.UTC)
	ticket := validIndexedTicket(now)
	ticket.Status = "unknown"

	err := ticketcore.ValidateIndexed(ticket)
	if err == nil {
		t.Fatal("ValidateIndexedTicket returned nil error")
	}

	if !strings.Contains(err.Error(), "invalid ticket status: unknown") {
		t.Fatalf("error = %q, want invalid status", err.Error())
	}
}

func TestValidateIndexedTicketAcceptsReviewMetadata(t *testing.T) {
	now := time.Date(2026, 5, 26, 10, 0, 0, 0, time.UTC)
	ticket := validIndexedTicket(now)
	ticket.Branch = "feat/slice-1-ticket-review-metadata-contract"
	ticket.ReviewRefs = []ticketmetadata.ReviewRef{{
		Type:     "pull_request",
		Provider: "forgejo",
		URL:      "https://forge.example.com/owner/repo/pulls/123",
		ID:       "123",
		Status:   "open",
	}}
	ticket.ReviewState = "changes_requested"
	ticket.ReviewRound = 1
	ticket.ReviewResult = "needs_changes"
	ticket.ReviewCommentsURL = "https://forge.example.com/owner/repo/pulls/123"
	ticket.BlockingFindings = []string{"RF-1"}

	err := ticketcore.ValidateIndexed(ticket)
	if err != nil {
		t.Fatalf("ValidateIndexedTicket returned error: %v", err)
	}
}

func TestValidateIndexedTicketAcceptsMergeMetadata(t *testing.T) {
	now := time.Date(2026, 5, 26, 10, 0, 0, 0, time.UTC)
	ticket := validIndexedTicket(now)
	ticket.MergeState = ticketcore.MergeStateBlocked
	ticket.MergeResult = ticketcore.MergeResultNotMergeable
	ticket.MergeActor = ticketcore.MergeActorMerger
	ticket.MergeCommit = "abc123"
	ticket.MergedAt = &now
	ticket.MergeBlocker = "merge conflict"

	err := ticketcore.ValidateIndexed(ticket)
	if err != nil {
		t.Fatalf("ValidateIndexedTicket returned error: %v", err)
	}
}

func TestValidateIndexedTicketRejectsInvalidReviewMetadata(t *testing.T) {
	now := time.Date(2026, 5, 26, 10, 0, 0, 0, time.UTC)

	tests := []struct {
		name    string
		mutate  func(*ticketcore.Indexed)
		wantErr string
	}{
		{
			name: "invalid review state",
			mutate: func(ticket *ticketcore.Indexed) {
				ticket.ReviewState = "pending"
			},
			wantErr: "invalid ticket review_state: pending",
		},
		{
			name: "invalid review result",
			mutate: func(ticket *ticketcore.Indexed) {
				ticket.ReviewResult = "pass_with_notes"
			},
			wantErr: "invalid ticket review_result: pass_with_notes",
		},
		{
			name: "negative review round",
			mutate: func(ticket *ticketcore.Indexed) {
				ticket.ReviewRound = -1
			},
			wantErr: "ticket review_round cannot be negative",
		},
		{
			name: "review ref missing url",
			mutate: func(ticket *ticketcore.Indexed) {
				ticket.ReviewRefs = []ticketmetadata.ReviewRef{{Type: "pull_request", Provider: "forgejo"}}
			},
			wantErr: "ticket review ref url is required",
		},
		{
			name: "empty blocking finding",
			mutate: func(ticket *ticketcore.Indexed) {
				ticket.BlockingFindings = []string{""}
			},
			wantErr: "ticket blocking_findings cannot contain empty values",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ticket := validIndexedTicket(now)
			tt.mutate(&ticket)

			err := ticketcore.ValidateIndexed(ticket)
			if err == nil {
				t.Fatal("ValidateIndexedTicket returned nil error")
			}

			if !strings.Contains(err.Error(), tt.wantErr) {
				t.Fatalf("error = %q, want %q", err.Error(), tt.wantErr)
			}
		})
	}
}

func TestValidateIndexedTicketRejectsInvalidMergeMetadata(t *testing.T) {
	now := time.Date(2026, 5, 26, 10, 0, 0, 0, time.UTC)

	tests := []struct {
		name    string
		mutate  func(*ticketcore.Indexed)
		wantErr string
	}{
		{
			name: "invalid merge state",
			mutate: func(ticket *ticketcore.Indexed) {
				ticket.MergeState = "waiting"
			},
			wantErr: "invalid ticket merge_state: waiting",
		},
		{
			name: "invalid merge result",
			mutate: func(ticket *ticketcore.Indexed) {
				ticket.MergeResult = "approved"
			},
			wantErr: "invalid ticket merge_result: approved",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ticket := validIndexedTicket(now)
			tt.mutate(&ticket)

			err := ticketcore.ValidateIndexed(ticket)
			if err == nil {
				t.Fatal("ValidateIndexedTicket returned nil error")
			}

			if !strings.Contains(err.Error(), tt.wantErr) {
				t.Fatalf("error = %q, want %q", err.Error(), tt.wantErr)
			}
		})
	}
}

func validIndexedTicket(now time.Time) ticketcore.Indexed {
	return ticketcore.Indexed{
		IndexedIdentity: ticketcore.IndexedIdentity{
			ProjectID:    "project-1",
			Path:         "/tmp/project/.chr/default/tickets/CHR-1-Test.md",
			RelativePath: "CHR-1-Test.md",
			Key:          "CHR-1",
		}, IndexedSummary: ticketcore.IndexedSummary{Title: "Test ticket"}, IndexedClassification: ticketcore.IndexedClassification{
			Status:   ticketcore.StatusBacklog,
			Priority: "normal",
			Severity: "normal",
			Type:     "task",
		}, IndexedTimes: ticketcore.IndexedTimes{
			CreatedAt: &now,
			UpdatedAt: &now,
		},
	}
}
