package core

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	_ "modernc.org/sqlite"

	projectcore "forge.adrianpk.com/chr/internal/core/project"
	ticketcore "forge.adrianpk.com/chr/internal/core/ticket"
	ticketmetadata "forge.adrianpk.com/chr/internal/core/ticket/metadata"
)

func TestTicketIndexUpsertsAndListsTickets(t *testing.T) {
	index := newTestTicketIndex(t)
	ticket := testIndexedTicket("CHR-1", []string{"uxui", "onboarding"})

	err := index.UpsertTicket(context.Background(), ticket)
	if err != nil {
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
	ticket := testIndexedTicket("CHR-1", []string{"old"})

	err := index.UpsertTicket(context.Background(), ticket)
	if err != nil {
		t.Fatalf("first UpsertTicket returned error: %v", err)
	}

	ticket.Tags = []string{"new"}
	ticket.Title = "Updated"

	err = index.UpsertTicket(context.Background(), ticket)
	if err != nil {
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
	ticket := testIndexedTicket("CHR-1", []string{"uxui"})

	err := index.UpsertTicket(context.Background(), ticket)
	if err != nil {
		t.Fatalf("UpsertTicket returned error: %v", err)
	}

	got, err := index.GetProjectTicketByKey(context.Background(), "project-1", "CHR-1")
	if err != nil {
		t.Fatalf("GetProjectTicketByKey returned error: %v", err)
	}

	assertIndexedTicket(t, got, ticket)
}

func TestTicketIndexRoundTripsReviewMetadata(t *testing.T) {
	index := newTestTicketIndex(t)
	ticket := testIndexedTicket("CHR-1", []string{"uxui"})
	ticket.Branch = "feat/slice-4-indexed-ticket-enrichment"
	ticket.ReviewRefs = []ticketmetadata.ReviewRef{
		{
			Type:     "pull_request",
			Provider: "forgejo",
			URL:      "https://forge.example.com/owner/repo/pulls/44",
			ID:       "44",
			Status:   "open",
		},
	}
	ticket.ReviewState = "changes_requested"
	ticket.ReviewRound = 1
	ticket.ReviewResult = "needs_changes"
	ticket.ReviewCommentsURL = "https://forge.example.com/owner/repo/pulls/44"
	ticket.BlockingFindings = []string{"RF-1"}

	err := index.UpsertTicket(context.Background(), ticket)
	if err != nil {
		t.Fatalf("UpsertTicket returned error: %v", err)
	}

	got, err := index.GetProjectTicketByKey(context.Background(), "project-1", "CHR-1")
	if err != nil {
		t.Fatalf("GetProjectTicketByKey returned error: %v", err)
	}

	assertIndexedTicket(t, got, ticket)
}

func TestTicketIndexGetTicketByKeyRejectsMissingTicket(t *testing.T) {
	index := newTestTicketIndex(t)

	_, err := index.GetProjectTicketByKey(context.Background(), "project-1", "CHR-1")
	if !errors.Is(err, ErrTicketNotFound) {
		t.Fatalf("err = %v, want ErrTicketNotFound", err)
	}
}

func TestTicketIndexDeletesProjectTickets(t *testing.T) {
	index := newTestTicketIndex(t)

	err := index.UpsertTicket(context.Background(), testIndexedTicket("CHR-1", nil))
	if err != nil {
		t.Fatalf("UpsertTicket returned error: %v", err)
	}

	err = index.DeleteProjectTickets(context.Background(), "project-1")
	if err != nil {
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

	registry, err := projectcore.NewRegistryFromDB(context.Background(), db, testMigrationFS())
	if err != nil {
		t.Fatalf("NewProjectRegistryFromDB returned error: %v", err)
	}

	addTestProject(t, registry, projectcore.Project{
		ID:           "project-1",
		Name:         "chr",
		Path:         "/tmp/chr",
		Enabled:      true,
		RegisteredAt: time.Date(2026, 5, 18, 12, 0, 0, 0, time.UTC),
	})

	index, err := NewTicketIndexFromDB(context.Background(), db, testMigrationFS())
	if err != nil {
		t.Fatalf("NewTicketIndexFromDB returned error: %v", err)
	}

	t.Cleanup(func() {
		err := index.Close()
		if err != nil {
			t.Fatalf("close ticket index: %v", err)
		}
	})

	return index
}

func testIndexedTicket(key string, tags []string) ticketcore.Indexed {
	createdAt := time.Date(2026, 5, 18, 12, 30, 0, 0, time.UTC)
	updatedAt := time.Date(2026, 5, 18, 13, 30, 0, 0, time.UTC)

	return ticketcore.Indexed{
		IndexedIdentity: ticketcore.IndexedIdentity{
			ProjectID:    "project-1",
			Path:         "/tmp/chr/.chr/default/tickets/" + key + ".md",
			RelativePath: key + ".md",
			Key:          key,
		}, IndexedSummary: ticketcore.IndexedSummary{
			Title:    "Do something",
			Synopsis: "Implement indexed ticket enrichment.",
		}, IndexedClassification: ticketcore.IndexedClassification{
			Status:      ticketcore.StatusReadyForDev,
			BoardColumn: ticketcore.StatusReadyForDev,
			Priority:    "high",
			Severity:    "critical",
			Type:        "feature",
			Tags:        tags,
			UseCase:     "onboarding",
			ProjectArea: "frontend",
		}, IndexedSource: ticketcore.IndexedSource{
			SourceKind: "plan_tracking",
			SourceRef:  ".chr/default/plans/onboarding-plan.md",
			Slice:      "Slice 4",
		}, IndexedTimes: ticketcore.IndexedTimes{
			CreatedAt: timePtr(createdAt),
			UpdatedAt: timePtr(updatedAt),
		}, IndexedProjection: ticketcore.IndexedProjection{
			FileMTime:   time.Date(2026, 5, 18, 14, 0, 0, 0, time.UTC),
			FileSize:    123,
			ContentHash: "hash",
			IndexedAt:   time.Date(2026, 5, 18, 15, 0, 0, 0, time.UTC),
		},
	}
}

func assertIndexedTicket(t *testing.T, got, want ticketcore.Indexed) {
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

	assertIndexedTicketEnrichedMetadata(t, got, want)
}

func assertIndexedTicketEnrichedMetadata(t *testing.T, got, want ticketcore.Indexed) {
	t.Helper()

	assertIndexedTicketSourceMetadata(t, got, want)
	assertIndexedTicketReviewMetadata(t, got, want)
}

func assertIndexedTicketSourceMetadata(t *testing.T, got, want ticketcore.Indexed) {
	t.Helper()

	matches := got.Synopsis == want.Synopsis &&
		got.Status == want.Status &&
		got.BoardColumn == want.BoardColumn &&
		got.SourceKind == want.SourceKind &&
		got.SourceRef == want.SourceRef &&
		got.Slice == want.Slice
	if !matches {
		t.Fatalf("source metadata = %#v, want %#v", got, want)
	}
}

func assertIndexedTicketReviewMetadata(t *testing.T, got, want ticketcore.Indexed) {
	t.Helper()

	matches := got.Branch == want.Branch &&
		got.ReviewState == want.ReviewState &&
		got.ReviewRound == want.ReviewRound &&
		got.ReviewResult == want.ReviewResult &&
		got.ReviewCommentsURL == want.ReviewCommentsURL
	if !matches {
		t.Fatalf("review metadata = %#v, want %#v", got, want)
	}

	assertTicketReviewRefs(t, got.ReviewRefs, want.ReviewRefs)
	assertBlockingFindings(t, got.BlockingFindings, want.BlockingFindings)
}

func assertTicketReviewRefs(t *testing.T, got, want []ticketmetadata.ReviewRef) {
	t.Helper()

	if len(got) != len(want) {
		t.Fatalf("ReviewRefs = %#v, want %#v", got, want)
	}

	if len(got) > 0 && got[0] != want[0] {
		t.Fatalf("ReviewRefs = %#v, want %#v", got, want)
	}
}

func assertBlockingFindings(t *testing.T, got, want []string) {
	t.Helper()

	if len(got) != len(want) {
		t.Fatalf("BlockingFindings = %#v, want %#v", got, want)
	}

	if len(got) > 0 && got[0] != want[0] {
		t.Fatalf("BlockingFindings = %#v, want %#v", got, want)
	}
}

func timePtr(value time.Time) *time.Time {
	return &value
}
