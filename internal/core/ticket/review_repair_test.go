package ticket_test

import (
	"testing"

	ticketcore "forge.adrianpk.com/chr/internal/core/ticket"
)

func TestTicketReviewRepairCountUsesReviewRound(t *testing.T) {
	tests := []struct {
		name   string
		ticket ticketcore.Indexed
		want   int
	}{
		{
			name:   "not repair",
			ticket: ticketcore.Indexed{IndexedClassification: ticketcore.IndexedClassification{Status: ticketcore.StatusReadyForDev}, IndexedReview: ticketcore.IndexedReview{ReviewRound: 3}},
			want:   0,
		},
		{
			name: "initial changes request",
			ticket: ticketcore.Indexed{
				IndexedClassification: ticketcore.IndexedClassification{Status: ticketcore.StatusReadyForDev}, IndexedReview: ticketcore.IndexedReview{
					ReviewState:  "changes_requested",
					ReviewResult: "needs_changes",
					ReviewRound:  1,
				},
			},
			want: 0,
		},
		{
			name: "one repair consumed",
			ticket: ticketcore.Indexed{
				IndexedClassification: ticketcore.IndexedClassification{Status: ticketcore.StatusReadyForDev}, IndexedReview: ticketcore.IndexedReview{
					ReviewState:  "changes_requested",
					ReviewResult: "needs_changes",
					ReviewRound:  2,
				},
			},
			want: 1,
		},
		{
			name: "zero review round is defensive",
			ticket: ticketcore.Indexed{
				IndexedClassification: ticketcore.IndexedClassification{Status: ticketcore.StatusReadyForDev}, IndexedReview: ticketcore.IndexedReview{
					ReviewState:  "changes_requested",
					ReviewResult: "needs_changes",
				},
			},
			want: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ticketcore.ReviewRepairCount(tt.ticket); got != tt.want {
				t.Fatalf("TicketReviewRepairCount() = %d, want %d", got, tt.want)
			}
		})
	}
}

func TestTicketReviewRepairAllowedUsesLimit(t *testing.T) {
	ticket := ticketcore.Indexed{
		IndexedClassification: ticketcore.IndexedClassification{Status: ticketcore.StatusReadyForDev}, IndexedReview: ticketcore.IndexedReview{
			ReviewState:  "changes_requested",
			ReviewResult: "needs_changes",
			ReviewRound:  1,
		},
	}

	if !ticketcore.ReviewRepairAllowed(ticket, ticketcore.DefaultReviewRepairLimit) {
		t.Fatal("TicketReviewRepairAllowed() = false, want true before the first repair")
	}

	ticket.ReviewRound = 2
	if ticketcore.ReviewRepairAllowed(ticket, ticketcore.DefaultReviewRepairLimit) {
		t.Fatal("TicketReviewRepairAllowed() = true, want false after the default repair limit")
	}

	if !ticketcore.ReviewRepairLimitExceeded(ticket, ticketcore.DefaultReviewRepairLimit) {
		t.Fatal("TicketReviewRepairLimitExceeded() = false, want true after the default repair limit")
	}
}
