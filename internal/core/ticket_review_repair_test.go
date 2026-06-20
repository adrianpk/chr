package core

import "testing"

func TestTicketReviewRepairCountUsesReviewRound(t *testing.T) {
	tests := []struct {
		name   string
		ticket IndexedTicket
		want   int
	}{
		{
			name: "not repair",
			ticket: IndexedTicket{
				Status:      TicketStatusReadyForDev,
				ReviewRound: 3,
			},
			want: 0,
		},
		{
			name: "initial changes request",
			ticket: IndexedTicket{
				Status:       TicketStatusReadyForDev,
				ReviewState:  "changes_requested",
				ReviewResult: "needs_changes",
				ReviewRound:  1,
			},
			want: 0,
		},
		{
			name: "one repair consumed",
			ticket: IndexedTicket{
				Status:       TicketStatusReadyForDev,
				ReviewState:  "changes_requested",
				ReviewResult: "needs_changes",
				ReviewRound:  2,
			},
			want: 1,
		},
		{
			name: "zero review round is defensive",
			ticket: IndexedTicket{
				Status:       TicketStatusReadyForDev,
				ReviewState:  "changes_requested",
				ReviewResult: "needs_changes",
			},
			want: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := TicketReviewRepairCount(tt.ticket); got != tt.want {
				t.Fatalf("TicketReviewRepairCount() = %d, want %d", got, tt.want)
			}
		})
	}
}

func TestTicketReviewRepairAllowedUsesLimit(t *testing.T) {
	ticket := IndexedTicket{
		Status:       TicketStatusReadyForDev,
		ReviewState:  "changes_requested",
		ReviewResult: "needs_changes",
		ReviewRound:  1,
	}

	if !TicketReviewRepairAllowed(ticket, DefaultTicketReviewRepairLimit) {
		t.Fatal("TicketReviewRepairAllowed() = false, want true before the first repair")
	}

	ticket.ReviewRound = 2
	if TicketReviewRepairAllowed(ticket, DefaultTicketReviewRepairLimit) {
		t.Fatal("TicketReviewRepairAllowed() = true, want false after the default repair limit")
	}

	if !TicketReviewRepairLimitExceeded(ticket, DefaultTicketReviewRepairLimit) {
		t.Fatal("TicketReviewRepairLimitExceeded() = false, want true after the default repair limit")
	}
}
