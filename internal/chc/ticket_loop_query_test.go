package chc

import (
	"context"
	"slices"
	"testing"
)

func TestTicketLoopQuerySelectsDefaultReadyTickets(t *testing.T) {
	index := newTestTicketIndex(t)
	upsertLoopQueryTicket(t, index, "CHC-3", "ready")
	upsertLoopQueryTicket(t, index, "CHC-1", "open")
	upsertLoopQueryTicket(t, index, "CHC-2", "ready")
	upsertLoopQueryTicket(t, index, "CHC-4", "blocked")

	got, err := NewTicketLoopQuery(index).EligibleTickets(context.Background(), TicketLoopQueryOptions{
		ProjectID: "project-1",
	})
	if err != nil {
		t.Fatalf("EligibleTickets returned error: %v", err)
	}

	if gotKeys := indexedTicketKeys(got); !slices.Equal(gotKeys, []string{"CHC-2", "CHC-3"}) {
		t.Fatalf("keys = %#v, want ready tickets sorted by key", gotKeys)
	}
}

func TestTicketLoopQuerySelectsExplicitStatuses(t *testing.T) {
	index := newTestTicketIndex(t)
	upsertLoopQueryTicket(t, index, "CHC-1", "open")
	upsertLoopQueryTicket(t, index, "CHC-2", "ready")
	upsertLoopQueryTicket(t, index, "CHC-3", "reviewing")
	upsertLoopQueryTicket(t, index, "CHC-4", "blocked")

	got, err := NewTicketLoopQuery(index).EligibleTickets(context.Background(), TicketLoopQueryOptions{
		ProjectID: "project-1",
		Statuses:  []string{"open", "ready"},
	})
	if err != nil {
		t.Fatalf("EligibleTickets returned error: %v", err)
	}

	if gotKeys := indexedTicketKeys(got); !slices.Equal(gotKeys, []string{"CHC-1", "CHC-2"}) {
		t.Fatalf("keys = %#v, want open and ready tickets", gotKeys)
	}
}

func TestTicketLoopQueryScopesToProject(t *testing.T) {
	index := newTestTicketIndex(t)
	upsertLoopQueryTicket(t, index, "CHC-1", "ready")

	got, err := NewTicketLoopQuery(index).EligibleTickets(context.Background(), TicketLoopQueryOptions{
		ProjectID: "other-project",
	})
	if err != nil {
		t.Fatalf("EligibleTickets returned error: %v", err)
	}

	if len(got) != 0 {
		t.Fatalf("len(tickets) = %d, want 0", len(got))
	}
}

func TestTicketLoopQueryRejectsInvalidOptions(t *testing.T) {
	index := newTestTicketIndex(t)

	for _, testCase := range []struct {
		name string
		opts TicketLoopQueryOptions
		want string
	}{
		{
			name: "missing project id",
			opts: TicketLoopQueryOptions{},
			want: "project id is required",
		},
		{
			name: "invalid status",
			opts: TicketLoopQueryOptions{ProjectID: "project-1", Statuses: []string{"unknown"}},
			want: "invalid ticket status: unknown",
		},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			_, err := NewTicketLoopQuery(index).EligibleTickets(context.Background(), testCase.opts)
			if err == nil || err.Error() != testCase.want {
				t.Fatalf("err = %v, want %q", err, testCase.want)
			}
		})
	}
}

func upsertLoopQueryTicket(t *testing.T, index TicketIndex, key, status string) {
	t.Helper()

	ticket := testIndexedTicket(key, nil)
	ticket.Status = status

	if err := index.UpsertTicket(context.Background(), ticket); err != nil {
		t.Fatalf("UpsertTicket returned error: %v", err)
	}
}

func indexedTicketKeys(tickets []IndexedTicket) []string {
	keys := make([]string, 0, len(tickets))
	for _, ticket := range tickets {
		keys = append(keys, ticket.Key)
	}

	return keys
}
