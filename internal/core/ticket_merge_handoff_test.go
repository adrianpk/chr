package core

import (
	"strings"
	"testing"
	"time"

	ticketmetadata "forge.adrianpk.com/chr/internal/core/ticket/metadata"
)

func TestValidateTicketMergeHandoffAcceptsPassedReview(t *testing.T) {
	now := time.Date(2026, 6, 2, 10, 0, 0, 0, time.UTC)
	ticket := validIndexedTicket(now)
	ticket.Status = TicketStatusReadyForMerge
	ticket.Branch = "feat/slice-1-merger-handoff-contract"
	ticket.ReviewState = TicketReviewCommandResultPass
	ticket.ReviewResult = TicketReviewCommandResultPass
	ticket.ReviewRefs = []ticketmetadata.ReviewRef{{
		Type:     "pull_request",
		Provider: "forgejo",
		URL:      "https://forge.example.com/owner/repo/pulls/281",
		ID:       "281",
		Status:   "open",
	}}

	handoff, err := ValidateTicketMergeHandoff(ticket)
	if err != nil {
		t.Fatalf("ValidateTicketMergeHandoff returned error: %v", err)
	}

	if handoff.TicketKey != "CHR-1" || handoff.Branch != "feat/slice-1-merger-handoff-contract" {
		t.Fatalf("handoff = %#v, want ticket key and branch", handoff)
	}

	if handoff.ReviewRef.URL != "https://forge.example.com/owner/repo/pulls/281" {
		t.Fatalf("ReviewRef = %#v, want PR ref", handoff.ReviewRef)
	}
}

func TestValidateTicketMergeHandoffRejectsMissingPassedReview(t *testing.T) {
	now := time.Date(2026, 6, 2, 10, 0, 0, 0, time.UTC)
	ticket := validIndexedTicket(now)
	ticket.Status = TicketStatusReadyForMerge
	ticket.Branch = "feat/slice-1-merger-handoff-contract"
	ticket.ReviewState = "changes_requested"
	ticket.ReviewResult = TicketReviewCommandResultNeedsChanges
	ticket.ReviewRefs = []ticketmetadata.ReviewRef{{
		Type:     "pull_request",
		Provider: "forgejo",
		URL:      "https://forge.example.com/owner/repo/pulls/281",
	}}

	_, err := ValidateTicketMergeHandoff(ticket)
	if err == nil {
		t.Fatal("ValidateTicketMergeHandoff returned nil error")
	}

	if !strings.Contains(err.Error(), "requires passed review metadata") {
		t.Fatalf("error = %q, want passed review metadata", err.Error())
	}
}

func TestValidateTicketMergeHandoffRejectsMissingPullRequestRef(t *testing.T) {
	now := time.Date(2026, 6, 2, 10, 0, 0, 0, time.UTC)
	ticket := validIndexedTicket(now)
	ticket.Status = TicketStatusReadyForMerge
	ticket.Branch = "feat/slice-1-merger-handoff-contract"
	ticket.ReviewState = TicketReviewCommandResultPass
	ticket.ReviewResult = TicketReviewCommandResultPass

	_, err := ValidateTicketMergeHandoff(ticket)
	if err == nil {
		t.Fatal("ValidateTicketMergeHandoff returned nil error")
	}

	if !strings.Contains(err.Error(), "requires a pull_request review ref") {
		t.Fatalf("error = %q, want missing pull request ref", err.Error())
	}
}
