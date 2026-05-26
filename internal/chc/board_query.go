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
	Attention       []BoardAttentionTicket
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

type BoardAttentionTicket struct {
	Ticket  IndexedTicket
	Reasons []string
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
	attention := make([]BoardAttentionTicket, 0)

	for _, ticket := range tickets {
		reasons := BoardAttentionReasons(ticket)
		if len(reasons) > 0 {
			attention = append(attention, BoardAttentionTicket{
				Ticket:  ticket,
				Reasons: reasons,
			})

			continue
		}

		columnID, _ := boardColumnForTicket(ticket)
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
	attention []BoardAttentionTicket,
) Board {
	board := Board{
		ProjectID:       projectID,
		Columns:         make([]BoardColumn, 0, len(NormalBoardColumns)),
		Attention:       sortedBoardAttentionTickets(attention),
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

	for _, attentionTicket := range board.Attention {
		board.Tickets = append(board.Tickets, attentionTicket.Ticket)
	}

	return board
}

func BoardAttentionReasons(ticket IndexedTicket) []string {
	reasons := []string{}

	if ticket.Status == TicketStatusBlockedLegacy {
		reasons = append(reasons, "status is blocked")
	}

	if _, ok := boardColumnForTicket(ticket); !ok {
		reasons = append(reasons, "status is outside normal Board columns: "+ticket.Status)
	}

	return reasons
}

func sortedBoardTickets(tickets []IndexedTicket) []IndexedTicket {
	sorted := append([]IndexedTicket(nil), tickets...)
	slices.SortFunc(sorted, compareIndexedTicketsByKey)

	return sorted
}

func sortedBoardAttentionTickets(tickets []BoardAttentionTicket) []BoardAttentionTicket {
	sorted := append([]BoardAttentionTicket(nil), tickets...)
	slices.SortFunc(sorted, func(left, right BoardAttentionTicket) int {
		return compareIndexedTicketsByKey(left.Ticket, right.Ticket)
	})

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
