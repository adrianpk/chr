package workspace

import (
	"context"
	"slices"
	"testing"
)

func TestBoardQueryReturnsNormalColumnsAndCounts(t *testing.T) {
	index := newTestTicketIndex(t)
	upsertBoardQueryTicket(t, index, "CHC-3", TicketStatusDone)
	upsertBoardQueryTicket(t, index, "CHC-1", TicketStatusBacklog)
	upsertBoardQueryTicket(t, index, "CHC-2", TicketStatusReadyForDev)

	board, err := NewBoardQuery(index).ProjectBoard(context.Background(), BoardQueryOptions{
		ProjectID: "project-1",
	})
	if err != nil {
		t.Fatalf("ProjectBoard returned error: %v", err)
	}

	if board.TotalCount != 3 || board.NormalCount != 3 || board.AttentionCount != 0 {
		t.Fatalf("board counts = %#v, want three normal tickets", board)
	}

	if len(board.Columns) != len(NormalBoardColumns) {
		t.Fatalf("columns = %d, want %d", len(board.Columns), len(NormalBoardColumns))
	}

	assertBoardColumn(t, board, TicketStatusBacklog, "Backlog", []string{"CHC-1"})
	assertBoardColumn(t, board, TicketStatusReadyForDev, "Ready For Dev", []string{"CHC-2"})
	assertBoardColumn(t, board, TicketStatusDone, "Done", []string{"CHC-3"})
}

func TestBoardQueryMapsLegacyStatusesToBoardColumns(t *testing.T) {
	index := newTestTicketIndex(t)
	upsertBoardQueryTicketWithMetadata(t, index, IndexedTicket{
		Key:    "CHC-1",
		Status: TicketStatusReadyLegacy,
	})

	board, err := NewBoardQuery(index).ProjectBoard(context.Background(), BoardQueryOptions{
		ProjectID: "project-1",
	})
	if err != nil {
		t.Fatalf("ProjectBoard returned error: %v", err)
	}

	assertBoardColumn(t, board, TicketStatusReadyForDev, "Ready For Dev", []string{"CHC-1"})
}

func TestBoardQuerySeparatesAttentionTickets(t *testing.T) {
	index := newTestTicketIndex(t)
	upsertBoardQueryTicket(t, index, "CHC-1", TicketStatusBacklog)
	upsertBoardQueryTicket(t, index, "CHC-2", TicketStatusBlockedLegacy)

	board, err := NewBoardQuery(index).ProjectBoard(context.Background(), BoardQueryOptions{
		ProjectID: "project-1",
	})
	if err != nil {
		t.Fatalf("ProjectBoard returned error: %v", err)
	}

	if board.TotalCount != 2 || board.NormalCount != 1 || board.AttentionCount != 1 {
		t.Fatalf("board counts = %#v, want one normal and one attention ticket", board)
	}

	if len(board.Attention) != 1 || board.Attention[0].Ticket.Key != "CHC-2" {
		t.Fatalf("attention tickets = %#v, want CHC-2", board.Attention)
	}

	wantReasons := []string{
		"status is blocked",
		"status is outside normal Board columns: blocked",
	}

	if !slices.Equal(board.Attention[0].Reasons, wantReasons) {
		t.Fatalf("attention reasons = %#v, want blocked status reasons", board.Attention[0].Reasons)
	}

	if got := indexedTicketKeys(board.Tickets); !slices.Equal(got, []string{"CHC-1", "CHC-2"}) {
		t.Fatalf("board tickets = %#v, want normal tickets followed by attention tickets", got)
	}
}

func TestBoardQueryFlagsReviewStateAttention(t *testing.T) {
	index := newTestTicketIndex(t)
	upsertBoardQueryTicketWithMetadata(t, index, IndexedTicket{
		Key:         "CHC-1",
		Status:      TicketStatusReadyForDev,
		ReviewState: "changes_requested",
	})

	board, err := NewBoardQuery(index).ProjectBoard(context.Background(), BoardQueryOptions{
		ProjectID: "project-1",
	})
	if err != nil {
		t.Fatalf("ProjectBoard returned error: %v", err)
	}

	if board.NormalCount != 0 || board.AttentionCount != 1 {
		t.Fatalf("board counts = %#v, want one attention ticket", board)
	}

	wantReasons := []string{"review changes requested"}
	if !slices.Equal(board.Attention[0].Reasons, wantReasons) {
		t.Fatalf("attention reasons = %#v, want review attention", board.Attention[0].Reasons)
	}
}

func TestBoardQueryRejectsInvalidOptions(t *testing.T) {
	_, err := NewBoardQuery(nil).ProjectBoard(context.Background(), BoardQueryOptions{})
	if err == nil {
		t.Fatal("ProjectBoard returned nil error")
	}
}

func assertBoardColumn(t *testing.T, board Board, columnID, title string, wantKeys []string) {
	t.Helper()

	for _, column := range board.Columns {
		if column.ID != columnID {
			continue
		}

		if column.Title != title {
			t.Fatalf("column title = %q, want %q", column.Title, title)
		}

		gotKeys := indexedTicketKeys(column.Tickets)
		if !slices.Equal(gotKeys, wantKeys) {
			t.Fatalf("column %s keys = %#v, want %#v", columnID, gotKeys, wantKeys)
		}

		if column.Count != len(wantKeys) {
			t.Fatalf("column %s count = %d, want %d", columnID, column.Count, len(wantKeys))
		}

		return
	}

	t.Fatalf("missing board column %s", columnID)
}

func upsertBoardQueryTicket(t *testing.T, index TicketIndex, key, status string) {
	t.Helper()

	upsertBoardQueryTicketWithMetadata(t, index, IndexedTicket{
		Key:    key,
		Status: status,
	})
}

func upsertBoardQueryTicketWithMetadata(t *testing.T, index TicketIndex, metadata IndexedTicket) {
	t.Helper()

	ticket := testIndexedTicket(metadata.Key, metadata.Tags)
	ticket.Status = metadata.Status
	ticket.BoardColumn, _ = BoardTicketStatusFor(metadata.Status)
	ticket.ReviewState = metadata.ReviewState

	err := index.UpsertTicket(context.Background(), ticket)
	if err != nil {
		t.Fatalf("UpsertTicket returned error: %v", err)
	}
}
