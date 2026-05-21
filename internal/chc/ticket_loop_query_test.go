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

	gotKeys := indexedTicketKeys(got)
	if !slices.Equal(gotKeys, []string{"CHC-2", "CHC-3"}) {
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

	gotKeys := indexedTicketKeys(got)
	if !slices.Equal(gotKeys, []string{"CHC-1", "CHC-2"}) {
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

func TestTicketLoopQueryFiltersByMetadata(t *testing.T) {
	index := newTestTicketIndex(t)
	upsertLoopQueryTicketWithMetadata(t, index, IndexedTicket{
		Key:         "CHC-1",
		Status:      "ready",
		Priority:    "high",
		Severity:    "critical",
		Type:        "feature",
		UseCase:     "onboarding",
		ProjectArea: "frontend",
		Tags:        []string{"uxui", "quick"},
	})
	upsertLoopQueryTicketWithMetadata(t, index, IndexedTicket{
		Key:         "CHC-2",
		Status:      "ready",
		Priority:    "normal",
		Severity:    "normal",
		Type:        "bug",
		UseCase:     "maintenance",
		ProjectArea: "backend",
		Tags:        []string{"api"},
	})
	upsertLoopQueryTicketWithMetadata(t, index, IndexedTicket{
		Key:         "CHC-3",
		Status:      "ready",
		Priority:    "high",
		Severity:    "critical",
		Type:        "feature",
		UseCase:     "onboarding",
		ProjectArea: "frontend",
		Tags:        []string{"uxui"},
	})

	got, err := NewTicketLoopQuery(index).EligibleTickets(context.Background(), TicketLoopQueryOptions{
		ProjectID: "project-1",
		Filters: TicketLoopFilters{
			Priorities:   []string{"high"},
			Severities:   []string{"critical"},
			ProjectAreas: []string{"frontend"},
			Tags:         []string{"uxui", "quick"},
			Types:        []string{"feature"},
			UseCases:     []string{"onboarding"},
		},
	})
	if err != nil {
		t.Fatalf("EligibleTickets returned error: %v", err)
	}

	gotKeys := indexedTicketKeys(got)
	if !slices.Equal(gotKeys, []string{"CHC-1"}) {
		t.Fatalf("keys = %#v, want filtered ticket", gotKeys)
	}
}

func TestTicketLoopQueryAllowsAnyValueForEmptyFilterDimensions(t *testing.T) {
	index := newTestTicketIndex(t)
	upsertLoopQueryTicketWithMetadata(t, index, IndexedTicket{
		Key:      "CHC-1",
		Status:   "ready",
		Priority: "high",
		Severity: "critical",
		Type:     "feature",
	})
	upsertLoopQueryTicketWithMetadata(t, index, IndexedTicket{
		Key:      "CHC-2",
		Status:   "ready",
		Priority: "normal",
		Severity: "normal",
		Type:     "bug",
	})

	got, err := NewTicketLoopQuery(index).EligibleTickets(context.Background(), TicketLoopQueryOptions{
		ProjectID: "project-1",
		Filters: TicketLoopFilters{
			Priorities: []string{"normal"},
		},
	})
	if err != nil {
		t.Fatalf("EligibleTickets returned error: %v", err)
	}

	gotKeys := indexedTicketKeys(got)
	if !slices.Equal(gotKeys, []string{"CHC-2"}) {
		t.Fatalf("keys = %#v, want priority-filtered ticket", gotKeys)
	}
}

func TestTicketLoopQueryOrdersByPriority(t *testing.T) {
	index := newTestTicketIndex(t)
	upsertLoopQueryTicketWithMetadata(t, index, IndexedTicket{
		Key:      "CHC-1",
		Status:   "ready",
		Priority: "low",
	})
	upsertLoopQueryTicketWithMetadata(t, index, IndexedTicket{
		Key:      "CHC-2",
		Status:   "ready",
		Priority: "critical",
	})
	upsertLoopQueryTicketWithMetadata(t, index, IndexedTicket{
		Key:      "CHC-3",
		Status:   "ready",
		Priority: "high",
	})

	got, err := NewTicketLoopQuery(index).EligibleTickets(context.Background(), TicketLoopQueryOptions{
		ProjectID: "project-1",
		Order:     TicketLoopOrderPriority,
	})
	if err != nil {
		t.Fatalf("EligibleTickets returned error: %v", err)
	}

	gotKeys := indexedTicketKeys(got)
	if !slices.Equal(gotKeys, []string{"CHC-2", "CHC-3", "CHC-1"}) {
		t.Fatalf("keys = %#v, want priority order", gotKeys)
	}
}

func TestTicketLoopQueryOrdersBySeverity(t *testing.T) {
	index := newTestTicketIndex(t)
	upsertLoopQueryTicketWithMetadata(t, index, IndexedTicket{
		Key:      "CHC-1",
		Status:   "ready",
		Severity: "normal",
	})
	upsertLoopQueryTicketWithMetadata(t, index, IndexedTicket{
		Key:      "CHC-2",
		Status:   "ready",
		Severity: "critical",
	})
	upsertLoopQueryTicketWithMetadata(t, index, IndexedTicket{
		Key:      "CHC-3",
		Status:   "ready",
		Severity: "high",
	})

	got, err := NewTicketLoopQuery(index).EligibleTickets(context.Background(), TicketLoopQueryOptions{
		ProjectID: "project-1",
		Order:     TicketLoopOrderSeverity,
	})
	if err != nil {
		t.Fatalf("EligibleTickets returned error: %v", err)
	}

	gotKeys := indexedTicketKeys(got)
	if !slices.Equal(gotKeys, []string{"CHC-2", "CHC-3", "CHC-1"}) {
		t.Fatalf("keys = %#v, want severity order", gotKeys)
	}
}

func TestTicketLoopQueryOrdersByPrioritySeverityAndKey(t *testing.T) {
	index := newTestTicketIndex(t)
	upsertLoopQueryTicketWithMetadata(t, index, IndexedTicket{
		Key:      "CHC-3",
		Status:   "ready",
		Priority: "high",
		Severity: "normal",
	})
	upsertLoopQueryTicketWithMetadata(t, index, IndexedTicket{
		Key:      "CHC-1",
		Status:   "ready",
		Priority: "high",
		Severity: "critical",
	})
	upsertLoopQueryTicketWithMetadata(t, index, IndexedTicket{
		Key:      "CHC-2",
		Status:   "ready",
		Priority: "high",
		Severity: "critical",
	})

	got, err := NewTicketLoopQuery(index).EligibleTickets(context.Background(), TicketLoopQueryOptions{
		ProjectID: "project-1",
		Order:     TicketLoopOrderPrioritySeverity,
	})
	if err != nil {
		t.Fatalf("EligibleTickets returned error: %v", err)
	}

	gotKeys := indexedTicketKeys(got)
	if !slices.Equal(gotKeys, []string{"CHC-1", "CHC-2", "CHC-3"}) {
		t.Fatalf("keys = %#v, want priority, severity, and key order", gotKeys)
	}
}

func TestTicketLoopQueryOrdersByCreatedAt(t *testing.T) {
	index := newTestTicketIndex(t)
	older := testLoopQueryTicketWithCreatedAt("CHC-1", "2026-05-18T12:00:00Z")
	newer := testLoopQueryTicketWithCreatedAt("CHC-2", "2026-05-18T14:00:00Z")
	middle := testLoopQueryTicketWithCreatedAt("CHC-3", "2026-05-18T13:00:00Z")

	upsertLoopQueryTicketWithMetadata(t, index, newer)
	upsertLoopQueryTicketWithMetadata(t, index, middle)
	upsertLoopQueryTicketWithMetadata(t, index, older)

	got, err := NewTicketLoopQuery(index).EligibleTickets(context.Background(), TicketLoopQueryOptions{
		ProjectID: "project-1",
		Order:     TicketLoopOrderCreatedAt,
	})
	if err != nil {
		t.Fatalf("EligibleTickets returned error: %v", err)
	}

	gotKeys := indexedTicketKeys(got)
	if !slices.Equal(gotKeys, []string{"CHC-1", "CHC-3", "CHC-2"}) {
		t.Fatalf("keys = %#v, want created_at order", gotKeys)
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
		{
			name: "invalid priority",
			opts: TicketLoopQueryOptions{
				ProjectID: "project-1",
				Filters:   TicketLoopFilters{Priorities: []string{"urgent"}},
			},
			want: "invalid ticket priority: urgent",
		},
		{
			name: "invalid severity",
			opts: TicketLoopQueryOptions{
				ProjectID: "project-1",
				Filters:   TicketLoopFilters{Severities: []string{"maximum"}},
			},
			want: "invalid ticket severity: maximum",
		},
		{
			name: "invalid type",
			opts: TicketLoopQueryOptions{
				ProjectID: "project-1",
				Filters:   TicketLoopFilters{Types: []string{"unknown"}},
			},
			want: "invalid ticket type: unknown",
		},
		{
			name: "invalid order",
			opts: TicketLoopQueryOptions{
				ProjectID: "project-1",
				Order:     TicketLoopOrder("unknown"),
			},
			want: "invalid ticket loop order: unknown",
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

	upsertLoopQueryTicketWithMetadata(t, index, ticket)
}

func upsertLoopQueryTicketWithMetadata(t *testing.T, index TicketIndex, metadata IndexedTicket) {
	t.Helper()

	ticket := testIndexedTicket(metadata.Key, metadata.Tags)
	ticket.Status = metadata.Status
	ticket.Priority = metadata.Priority
	ticket.Severity = metadata.Severity
	ticket.Type = metadata.Type
	ticket.UseCase = metadata.UseCase
	ticket.ProjectArea = metadata.ProjectArea
	ticket.CreatedAt = metadata.CreatedAt

	err := index.UpsertTicket(context.Background(), ticket)
	if err != nil {
		t.Fatalf("UpsertTicket returned error: %v", err)
	}
}

func testLoopQueryTicketWithCreatedAt(key, createdAt string) IndexedTicket {
	parsed, err := parseProjectTime(createdAt)
	if err != nil {
		panic(err)
	}

	return IndexedTicket{
		Key:       key,
		Status:    "ready",
		Priority:  "normal",
		Severity:  "normal",
		Type:      "task",
		CreatedAt: &parsed,
	}
}

func indexedTicketKeys(tickets []IndexedTicket) []string {
	keys := make([]string, 0, len(tickets))
	for _, ticket := range tickets {
		keys = append(keys, ticket.Key)
	}

	return keys
}
