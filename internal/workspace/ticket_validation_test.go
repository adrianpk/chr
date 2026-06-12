package workspace

import (
	"strings"
	"testing"
	"time"
)

func TestValidateIndexedTicketAcceptsBoardMetadata(t *testing.T) {
	now := time.Date(2026, 5, 26, 10, 0, 0, 0, time.UTC)
	ticket := validIndexedTicket(now)
	ticket.Status = TicketStatusBacklog
	ticket.Synopsis = "Implement the Board ticket creation path."
	ticket.SourceKind = "manual"

	err := ValidateIndexedTicket(ticket)
	if err != nil {
		t.Fatalf("ValidateIndexedTicket returned error: %v", err)
	}
}

func TestValidateIndexedTicketAcceptsLegacyTicketWithoutSynopsis(t *testing.T) {
	now := time.Date(2026, 5, 26, 10, 0, 0, 0, time.UTC)
	ticket := validIndexedTicket(now)
	ticket.Status = TicketStatusOpenLegacy
	ticket.Synopsis = ""

	err := ValidateIndexedTicket(ticket)
	if err != nil {
		t.Fatalf("ValidateIndexedTicket returned error: %v", err)
	}
}

func TestValidateIndexedTicketRejectsUnknownStatus(t *testing.T) {
	now := time.Date(2026, 5, 26, 10, 0, 0, 0, time.UTC)
	ticket := validIndexedTicket(now)
	ticket.Status = "unknown"

	err := ValidateIndexedTicket(ticket)
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
	ticket.ReviewRefs = []TicketReviewRef{{
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

	err := ValidateIndexedTicket(ticket)
	if err != nil {
		t.Fatalf("ValidateIndexedTicket returned error: %v", err)
	}
}

func TestValidateIndexedTicketAcceptsMergeMetadata(t *testing.T) {
	now := time.Date(2026, 5, 26, 10, 0, 0, 0, time.UTC)
	ticket := validIndexedTicket(now)
	ticket.MergeState = TicketMergeStateBlocked
	ticket.MergeResult = TicketMergeResultNotMergeable
	ticket.MergeActor = TicketMergeActorMerger
	ticket.MergeCommit = "abc123"
	ticket.MergedAt = &now
	ticket.MergeBlocker = "merge conflict"

	err := ValidateIndexedTicket(ticket)
	if err != nil {
		t.Fatalf("ValidateIndexedTicket returned error: %v", err)
	}
}

func TestValidateIndexedTicketRejectsInvalidReviewMetadata(t *testing.T) {
	now := time.Date(2026, 5, 26, 10, 0, 0, 0, time.UTC)

	tests := []struct {
		name    string
		mutate  func(*IndexedTicket)
		wantErr string
	}{
		{
			name: "invalid review state",
			mutate: func(ticket *IndexedTicket) {
				ticket.ReviewState = "pending"
			},
			wantErr: "invalid ticket review_state: pending",
		},
		{
			name: "invalid review result",
			mutate: func(ticket *IndexedTicket) {
				ticket.ReviewResult = "pass_with_notes"
			},
			wantErr: "invalid ticket review_result: pass_with_notes",
		},
		{
			name: "negative review round",
			mutate: func(ticket *IndexedTicket) {
				ticket.ReviewRound = -1
			},
			wantErr: "ticket review_round cannot be negative",
		},
		{
			name: "review ref missing url",
			mutate: func(ticket *IndexedTicket) {
				ticket.ReviewRefs = []TicketReviewRef{{Type: "pull_request", Provider: "forgejo"}}
			},
			wantErr: "ticket review ref url is required",
		},
		{
			name: "empty blocking finding",
			mutate: func(ticket *IndexedTicket) {
				ticket.BlockingFindings = []string{""}
			},
			wantErr: "ticket blocking_findings cannot contain empty values",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ticket := validIndexedTicket(now)
			tt.mutate(&ticket)

			err := ValidateIndexedTicket(ticket)
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
		mutate  func(*IndexedTicket)
		wantErr string
	}{
		{
			name: "invalid merge state",
			mutate: func(ticket *IndexedTicket) {
				ticket.MergeState = "waiting"
			},
			wantErr: "invalid ticket merge_state: waiting",
		},
		{
			name: "invalid merge result",
			mutate: func(ticket *IndexedTicket) {
				ticket.MergeResult = "approved"
			},
			wantErr: "invalid ticket merge_result: approved",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ticket := validIndexedTicket(now)
			tt.mutate(&ticket)

			err := ValidateIndexedTicket(ticket)
			if err == nil {
				t.Fatal("ValidateIndexedTicket returned nil error")
			}

			if !strings.Contains(err.Error(), tt.wantErr) {
				t.Fatalf("error = %q, want %q", err.Error(), tt.wantErr)
			}
		})
	}
}

func validIndexedTicket(now time.Time) IndexedTicket {
	return IndexedTicket{
		ProjectID:    "project-1",
		Path:         "/tmp/project/.chc/user/tickets/CHC-1-Test.md",
		RelativePath: "CHC-1-Test.md",
		Key:          "CHC-1",
		Title:        "Test ticket",
		Status:       TicketStatusBacklog,
		Priority:     "normal",
		Severity:     "normal",
		Type:         "task",
		CreatedAt:    &now,
		UpdatedAt:    &now,
	}
}
