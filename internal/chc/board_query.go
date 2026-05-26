package chc

import (
	"context"
	"fmt"
	"slices"
	"strings"
)

var NormalBoardColumns = []string{
	TicketStatusBacklog,
	TicketStatusReadyForDev,
	TicketStatusInProgress,
	TicketStatusReadyForReview,
	TicketStatusReadyForMerge,
	TicketStatusDone,
}

type BoardQuery struct {
	Index TicketIndex
}

type BoardQueryOptions struct {
	ProjectID string
}

type Board struct {
	ProjectID       string
	Columns         []BoardColumn
	Tickets         []IndexedTicket
	TotalCount      int
	NormalCount     int
	AttentionCount  int
	NormalColumnIDs []string
}

type BoardColumn struct {
	ID      string
	Title   string
	Count   int
	Tickets []IndexedTicket
}

func NewBoardQuery(index TicketIndex) BoardQuery {
	return BoardQuery{Index: index}
}

func (q BoardQuery) ProjectBoard(ctx context.Context, opts BoardQueryOptions) (Board, error) {
	if strings.TrimSpace(opts.ProjectID) == "" {
		return Board{}, fmt.Errorf("project id is required")
	}

	if q.Index == nil {
		return Board{}, fmt.Errorf("ticket index is required")
	}

	tickets, err := q.Index.ListProjectTickets(ctx, opts.ProjectID)
	if err != nil {
		return Board{}, err
	}

	return buildBoard(opts.ProjectID, tickets), nil
}

func buildBoard(projectID string, tickets []IndexedTicket) Board {
	columns := newBoardColumns()
	attention := make([]IndexedTicket, 0)

	for _, ticket := range tickets {
		columnID, ok := boardColumnForTicket(ticket)
		if !ok {
			attention = append(attention, ticket)

			continue
		}

		columns[columnID] = append(columns[columnID], ticket)
	}

	return boardFromColumns(projectID, columns, attention)
}

func newBoardColumns() map[string][]IndexedTicket {
	columns := make(map[string][]IndexedTicket, len(NormalBoardColumns))
	for _, columnID := range NormalBoardColumns {
		columns[columnID] = []IndexedTicket{}
	}

	return columns
}

func boardColumnForTicket(ticket IndexedTicket) (string, bool) {
	if IsNormalBoardTicketStatus(ticket.BoardColumn) {
		return ticket.BoardColumn, true
	}

	return BoardTicketStatusFor(ticket.Status)
}

func boardFromColumns(
	projectID string,
	columns map[string][]IndexedTicket,
	attention []IndexedTicket,
) Board {
	board := Board{
		ProjectID:       projectID,
		Columns:         make([]BoardColumn, 0, len(NormalBoardColumns)),
		Tickets:         make([]IndexedTicket, 0),
		TotalCount:      len(attention),
		AttentionCount:  len(attention),
		NormalColumnIDs: append([]string(nil), NormalBoardColumns...),
	}

	for _, columnID := range NormalBoardColumns {
		tickets := sortedBoardTickets(columns[columnID])
		board.Columns = append(board.Columns, BoardColumn{
			ID:      columnID,
			Title:   BoardColumnTitle(columnID),
			Count:   len(tickets),
			Tickets: tickets,
		})
		board.Tickets = append(board.Tickets, tickets...)
		board.TotalCount += len(tickets)
		board.NormalCount += len(tickets)
	}

	board.Tickets = append(board.Tickets, sortedBoardTickets(attention)...)

	return board
}

func sortedBoardTickets(tickets []IndexedTicket) []IndexedTicket {
	sorted := append([]IndexedTicket(nil), tickets...)
	slices.SortFunc(sorted, compareIndexedTicketsByKey)

	return sorted
}

func BoardColumnTitle(columnID string) string {
	words := strings.Split(strings.ReplaceAll(columnID, "_", " "), " ")
	for index, word := range words {
		if word == "" {
			continue
		}

		words[index] = strings.ToUpper(word[:1]) + word[1:]
	}

	return strings.Join(words, " ")
}
