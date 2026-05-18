package chc

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	_ "modernc.org/sqlite"
)

func TestTicketIndexUpsertsAndListsTickets(t *testing.T) {
	index := newTestTicketIndex(t)
	ticket := testIndexedTicket("CHC-1", []string{"uxui", "onboarding"})

	if err := index.UpsertTicket(context.Background(), ticket); err != nil {
		t.Fatalf("UpsertTicket returned error: %v", err)
	}

	tickets, err := index.ListProjectTickets(context.Background(), "project-1")
	if err != nil {
		t.Fatalf("ListProjectTickets returned error: %v", err)
	}

	if len(tickets) != 1 {
		t.Fatalf("len(tickets) = %d, want 1", len(tickets))
	}

	assertIndexedTicket(t, tickets[0], ticket)
}

func TestTicketIndexReplacesTagsOnUpsert(t *testing.T) {
	index := newTestTicketIndex(t)
	ticket := testIndexedTicket("CHC-1", []string{"old"})

	if err := index.UpsertTicket(context.Background(), ticket); err != nil {
		t.Fatalf("first UpsertTicket returned error: %v", err)
	}

	ticket.Tags = []string{"new"}
	ticket.Title = "Updated"

	if err := index.UpsertTicket(context.Background(), ticket); err != nil {
		t.Fatalf("second UpsertTicket returned error: %v", err)
	}

	tickets, err := index.ListProjectTickets(context.Background(), "project-1")
	if err != nil {
		t.Fatalf("ListProjectTickets returned error: %v", err)
	}

	if tickets[0].Title != "Updated" {
		t.Fatalf("Title = %q, want Updated", tickets[0].Title)
	}

	if len(tickets[0].Tags) != 1 || tickets[0].Tags[0] != "new" {
		t.Fatalf("Tags = %#v, want new", tickets[0].Tags)
	}
}

func TestTicketIndexGetsTicketByKey(t *testing.T) {
	index := newTestTicketIndex(t)
	ticket := testIndexedTicket("CHC-1", []string{"uxui"})

	if err := index.UpsertTicket(context.Background(), ticket); err != nil {
		t.Fatalf("UpsertTicket returned error: %v", err)
	}

	got, err := index.GetProjectTicketByKey(context.Background(), "project-1", "CHC-1")
	if err != nil {
		t.Fatalf("GetProjectTicketByKey returned error: %v", err)
	}

	assertIndexedTicket(t, got, ticket)
}

func TestTicketIndexGetTicketByKeyRejectsMissingTicket(t *testing.T) {
	index := newTestTicketIndex(t)

	_, err := index.GetProjectTicketByKey(context.Background(), "project-1", "CHC-1")
	if !errors.Is(err, ErrTicketNotFound) {
		t.Fatalf("err = %v, want ErrTicketNotFound", err)
	}
}

func TestTicketIndexDeletesProjectTickets(t *testing.T) {
	index := newTestTicketIndex(t)

	if err := index.UpsertTicket(context.Background(), testIndexedTicket("CHC-1", nil)); err != nil {
		t.Fatalf("UpsertTicket returned error: %v", err)
	}

	if err := index.DeleteProjectTickets(context.Background(), "project-1"); err != nil {
		t.Fatalf("DeleteProjectTickets returned error: %v", err)
	}

	tickets, err := index.ListProjectTickets(context.Background(), "project-1")
	if err != nil {
		t.Fatalf("ListProjectTickets returned error: %v", err)
	}

	if len(tickets) != 0 {
		t.Fatalf("len(tickets) = %d, want 0", len(tickets))
	}
}

func newTestTicketIndex(t *testing.T) TicketIndex {
	t.Helper()

	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}

	db.SetMaxOpenConns(1)

	registry, err := NewProjectRegistryFromDB(context.Background(), db, testMigrationFS())
	if err != nil {
		t.Fatalf("NewProjectRegistryFromDB returned error: %v", err)
	}

	addTestProject(t, registry, Project{
		ID:           "project-1",
		Name:         "chrysocoord",
		Path:         "/tmp/chrysocoord",
		Enabled:      true,
		RegisteredAt: time.Date(2026, 5, 18, 12, 0, 0, 0, time.UTC),
	})

	index, err := NewTicketIndexFromDB(context.Background(), db, testMigrationFS())
	if err != nil {
		t.Fatalf("NewTicketIndexFromDB returned error: %v", err)
	}

	t.Cleanup(func() {
		if err := index.Close(); err != nil {
			t.Fatalf("close ticket index: %v", err)
		}
	})

	return index
}

func testIndexedTicket(key string, tags []string) IndexedTicket {
	createdAt := time.Date(2026, 5, 18, 12, 30, 0, 0, time.UTC)
	updatedAt := time.Date(2026, 5, 18, 13, 30, 0, 0, time.UTC)

	return IndexedTicket{
		ProjectID:    "project-1",
		Path:         "/tmp/chrysocoord/.chc/tickets/" + key + ".md",
		RelativePath: key + ".md",
		Key:          key,
		Title:        "Do something",
		Status:       "open",
		Priority:     "high",
		Severity:     "critical",
		Type:         "feature",
		Tags:         tags,
		UseCase:      "onboarding",
		ProjectArea:  "frontend",
		CreatedAt:    timePtr(createdAt),
		UpdatedAt:    timePtr(updatedAt),
		FileMTime:    time.Date(2026, 5, 18, 14, 0, 0, 0, time.UTC),
		FileSize:     123,
		ContentHash:  "hash",
		IndexedAt:    time.Date(2026, 5, 18, 15, 0, 0, 0, time.UTC),
	}
}

func assertIndexedTicket(t *testing.T, got, want IndexedTicket) {
	t.Helper()

	if got.ProjectID != want.ProjectID || got.Key != want.Key || got.Title != want.Title {
		t.Fatalf("ticket = %#v, want %#v", got, want)
	}

	if got.ContentHash != want.ContentHash || got.FileSize != want.FileSize {
		t.Fatalf("file state = %#v, want %#v", got, want)
	}

	if len(got.Tags) != len(want.Tags) {
		t.Fatalf("len(Tags) = %d, want %d", len(got.Tags), len(want.Tags))
	}
}
