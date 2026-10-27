package ticket

import (
	"context"
	"fmt"
	"slices"
	"strings"
)

// NormalBoardColumns defines a package boundary variable.
var NormalBoardColumns = []string{
	string(StatusBacklog),
	string(StatusReadyForDev),
	string(StatusInProgress),
	string(StatusReadyForReview),
	string(StatusReadyForMerge),
	string(StatusDone),
}

// BoardQuery describes a package boundary value.
type BoardQuery struct {
	Index Index
}

// BoardQueryOptions describes a package boundary value.
type BoardQueryOptions struct {
	ProjectID string
}

// Board describes a package boundary value.
type Board struct {
	Revision        string
	Offset          int
	LoadedCount     int
	NextOffset      int
	ProjectID       string
	Columns         []BoardColumn
	Attention       []BoardAttentionTicket
	Tickets         []Indexed
	TotalCount      int
	NormalCount     int
	AttentionCount  int
	NormalColumnIDs []string
}

// BoardColumn describes a package boundary value.
type BoardColumn struct {
	ID      string
	Title   string
	Count   int
	Tickets []Indexed
}

// BoardAttentionTicket describes a package boundary value.
type BoardAttentionTicket struct {
	Ticket  Indexed
	Reasons []string
}

// NewBoardQuery constructs the corresponding package service or value.
func NewBoardQuery(index Index) BoardQuery {
	return BoardQuery{Index: index}
}

// ProjectBoard executes the corresponding package operation.
func (q BoardQuery) ProjectBoard(ctx context.Context, opts BoardQueryOptions) (Board, error) {
	if strings.TrimSpace(opts.ProjectID) == "" {
		return Board{}, fmt.Errorf("project id is required")
	}

	if q.Index == nil {
		return Board{}, fmt.Errorf("ticket index is required")
	}

	tickets, err := q.Index.ListProject(ctx, opts.ProjectID)
	if err != nil {
		return Board{}, err
	}

	return BuildBoard(opts.ProjectID, tickets), nil
}

// BuildBoard projects ticket status and review attention without changing the input records.
func BuildBoard(projectID string, tickets []Indexed) Board {
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

func newBoardColumns() map[string][]Indexed {
	columns := make(map[string][]Indexed, len(NormalBoardColumns))
	for _, columnID := range NormalBoardColumns {
		columns[columnID] = []Indexed{}
	}

	return columns
}

func boardColumnForTicket(ticket Indexed) (string, bool) {
	if NormalBoardStatus(ticket.BoardColumn) {
		return string(ticket.BoardColumn), true
	}

	status, ok := BoardStatus(ticket.Status)

	return string(status), ok
}

func boardFromColumns(projectID string, columns map[string][]Indexed, attention []BoardAttentionTicket) Board {
	board := Board{
		ProjectID:       projectID,
		Columns:         make([]BoardColumn, 0, len(NormalBoardColumns)),
		Attention:       sortedBoardAttentionTickets(attention),
		Tickets:         make([]Indexed, 0),
		TotalCount:      len(attention),
		AttentionCount:  len(attention),
		NormalColumnIDs: append([]string(nil), NormalBoardColumns...),
	}

	for _, columnID := range NormalBoardColumns {
		tickets := SortedBoardTickets(columns[columnID])
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

// BoardAttentionReasons executes the corresponding package operation.
func BoardAttentionReasons(ticket Indexed) []string {
	reasons := append([]string(nil), ticket.AttentionReasons...)

	if ticket.Status == StatusBlockedLegacy {
		reasons = append(reasons, "status is blocked")
	}

	if _, ok := boardColumnForTicket(ticket); !ok {
		reasons = append(reasons, "status is outside normal Board columns: "+string(ticket.Status))
	}

	switch ticket.ReviewState {
	case "changes_requested":
		reasons = append(reasons, "review changes requested")
	case "insufficient_evidence":
		reasons = append(reasons, "review evidence is insufficient")
	case "needs_human_decision":
		reasons = append(reasons, "review needs human decision")
	}

	return reasons
}

// SortedBoardTickets returns a sorted copy using the shared ticket key order.
func SortedBoardTickets(tickets []Indexed) []Indexed {
	sorted := append([]Indexed(nil), tickets...)
	slices.SortFunc(sorted, CompareKeys)

	return sorted
}

func sortedBoardAttentionTickets(tickets []BoardAttentionTicket) []BoardAttentionTicket {
	sorted := append([]BoardAttentionTicket(nil), tickets...)
	slices.SortFunc(sorted, func(left, right BoardAttentionTicket) int {
		return CompareKeys(left.Ticket, right.Ticket)
	})

	return sorted
}

// BoardColumnTitle executes the corresponding package operation.
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
