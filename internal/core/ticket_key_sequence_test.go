package core

import (
	"context"
	"testing"
	"time"

	ticketcore "forge.adrianpk.com/chr/internal/core/ticket"
)

func TestTicketKeySequenceReservesSeedThenNextNumbers(t *testing.T) {
	db := openTestSQLite(t)

	repository, err := ticketcore.NewTicketKeySequenceRepositoryFromDB(context.Background(), db, testMigrationFS())
	if err != nil {
		t.Fatalf("NewTicketKeySequenceRepositoryFromDB returned error: %v", err)
	}

	now := time.Date(2026, 5, 26, 12, 0, 0, 0, time.UTC)

	first, err := repository.ReserveNextTicketKey(context.Background(), ticketcore.ReserveNextTicketKeyOptions{
		ProjectID: "project-1",
		Prefix:    "CHR",
		Seed:      7,
		UpdatedAt: now,
	})
	if err != nil {
		t.Fatalf("ReserveNextTicketKey returned error: %v", err)
	}

	second, err := repository.ReserveNextTicketKey(context.Background(), ticketcore.ReserveNextTicketKeyOptions{
		ProjectID: "project-1",
		Prefix:    "CHR",
		Seed:      1,
		UpdatedAt: now,
	})
	if err != nil {
		t.Fatalf("ReserveNextTicketKey returned error: %v", err)
	}

	if first != 7 || second != 8 {
		t.Fatalf("reserved keys = %d, %d; want 7, 8", first, second)
	}
}

func TestTicketKeySequenceKeepsProjectPrefixesIndependent(t *testing.T) {
	db := openTestSQLite(t)

	repository, err := ticketcore.NewTicketKeySequenceRepositoryFromDB(context.Background(), db, testMigrationFS())
	if err != nil {
		t.Fatalf("NewTicketKeySequenceRepositoryFromDB returned error: %v", err)
	}

	now := time.Date(2026, 5, 26, 12, 0, 0, 0, time.UTC)

	first, err := repository.ReserveNextTicketKey(context.Background(), ticketcore.ReserveNextTicketKeyOptions{
		ProjectID: "project-1",
		Prefix:    "CHR",
		Seed:      1,
		UpdatedAt: now,
	})
	if err != nil {
		t.Fatalf("ReserveNextTicketKey returned error: %v", err)
	}

	second, err := repository.ReserveNextTicketKey(context.Background(), ticketcore.ReserveNextTicketKeyOptions{
		ProjectID: "project-1",
		Prefix:    "PRJ",
		Seed:      1,
		UpdatedAt: now,
	})
	if err != nil {
		t.Fatalf("ReserveNextTicketKey returned error: %v", err)
	}

	if first != 1 || second != 1 {
		t.Fatalf("reserved keys = %d, %d; want 1, 1", first, second)
	}
}

func TestTicketKeySequenceAdvancesWhenFilesystemSeedIsHigher(t *testing.T) {
	db := openTestSQLite(t)

	repository, err := ticketcore.NewTicketKeySequenceRepositoryFromDB(context.Background(), db, testMigrationFS())
	if err != nil {
		t.Fatalf("NewTicketKeySequenceRepositoryFromDB returned error: %v", err)
	}

	now := time.Date(2026, 5, 26, 12, 0, 0, 0, time.UTC)

	_, err = repository.ReserveNextTicketKey(context.Background(), ticketcore.ReserveNextTicketKeyOptions{
		ProjectID: "project-1",
		Prefix:    "CHR",
		Seed:      1,
		UpdatedAt: now,
	})
	if err != nil {
		t.Fatalf("ReserveNextTicketKey returned error: %v", err)
	}

	next, err := repository.ReserveNextTicketKey(context.Background(), ticketcore.ReserveNextTicketKeyOptions{
		ProjectID: "project-1",
		Prefix:    "CHR",
		Seed:      10,
		UpdatedAt: now,
	})
	if err != nil {
		t.Fatalf("ReserveNextTicketKey returned error: %v", err)
	}

	if next != 10 {
		t.Fatalf("reserved key = %d, want 10", next)
	}
}
