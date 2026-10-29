package tui

import (
	"strings"

	"forge.adrianpk.com/chr/internal/core"
	ticketcore "forge.adrianpk.com/chr/internal/core/ticket"
)

func (m *Model) selectedTicketView() (TicketView, bool) {
	if m.SelectedView() == ViewDash {
		return m.selectedDashTicketView()
	}

	if len(m.tickets) == 0 {
		return TicketView{}, false
	}

	for _, ticket := range m.filteredTickets() {
		if ticket.Index == m.selectedTicket {
			return ticket, true
		}
	}

	return TicketView{}, false
}

func (m *Model) selectPreviousTicket() {
	if !m.ticketSelectionActive() {
		return
	}

	filteredTickets := m.filteredTickets()
	if len(filteredTickets) == 0 {
		return
	}

	position := m.selectedTicketPosition(filteredTickets)
	if position == 0 {
		m.selectedTicket = filteredTickets[len(filteredTickets)-1].Index

		return
	}

	m.selectedTicket = filteredTickets[position-1].Index
}

func (m *Model) selectNextTicket() {
	if !m.ticketSelectionActive() {
		return
	}

	filteredTickets := m.filteredTickets()
	if len(filteredTickets) == 0 {
		return
	}

	position := m.selectedTicketPosition(filteredTickets)
	if position >= len(filteredTickets)-1 {
		m.selectedTicket = filteredTickets[0].Index

		return
	}

	m.selectedTicket = filteredTickets[position+1].Index
}

func (m *Model) ticketSelectionActive() bool {
	return m.SelectedView() == ViewBoard || m.SelectedView() == ViewTickets
}

func (m *Model) selectedTicketPosition(tickets []TicketView) int {
	for index, ticket := range tickets {
		if ticket.Index == m.selectedTicket {
			return index
		}
	}

	return 0
}

// TicketView is the terminal projection of an indexed ticket.
type TicketView struct {
	Revision          string
	Index             int
	Key               string
	Title             string
	Synopsis          string
	Status            string
	Priority          string
	Severity          string
	Type              string
	UseCase           string
	ProjectArea       string
	SourceKind        string
	SourceRef         string
	Slice             string
	Branch            string
	ReviewState       string
	ReviewRound       int
	ReviewResult      string
	ReviewCommentsURL string
	BoardColumn       string
	AttentionReasons  []string
	RelativePath      string
	ChecklistDone     int
	ChecklistTotal    int
	CommentCount      int
	BodyLines         []string
	Evidence          core.EvidenceStrategySummary
}

// NewTicketViews converts indexed tickets into TUI rows.
func NewTicketViews(tickets []ticketcore.Indexed) []TicketView {
	views := make([]TicketView, 0, len(tickets))

	for index, ticket := range tickets {
		boardColumn, ok := ticketcore.BoardStatus(ticket.Status)
		if ok && ticketcore.NormalBoardStatus(ticket.BoardColumn) {
			boardColumn = ticket.BoardColumn
		}

		view := TicketView{
			Revision:          ticket.ContentHash,
			Index:             index,
			Key:               ticket.Key,
			Title:             ticket.Title,
			Synopsis:          ticket.Synopsis,
			Status:            string(ticket.Status),
			Priority:          ticket.Priority,
			Severity:          ticket.Severity,
			Type:              ticket.Type,
			UseCase:           ticket.UseCase,
			ProjectArea:       ticket.ProjectArea,
			SourceKind:        ticket.SourceKind,
			SourceRef:         ticket.SourceRef,
			Slice:             ticket.Slice,
			Branch:            ticket.Branch,
			ReviewState:       string(ticket.ReviewState),
			ReviewRound:       ticket.ReviewRound,
			ReviewResult:      string(ticket.ReviewResult),
			ReviewCommentsURL: ticket.ReviewCommentsURL,
			BoardColumn:       string(boardColumn),
			AttentionReasons:  ticketcore.BoardAttentionReasons(ticket),
			RelativePath:      ticket.RelativePath,
		}

		detail := ticketMarkdownDetail(ticket.Path)
		view.ChecklistDone = detail.ChecklistDone
		view.ChecklistTotal = detail.ChecklistTotal
		view.CommentCount = detail.CommentCount
		view.BodyLines = detail.BodyLines
		view.Evidence = detail.Evidence

		views = append(views, view)
	}

	return views
}

// Matches reports whether a ticket row should be shown for the current filter.
func (t TicketView) Matches(filter string) bool {
	values := []string{
		t.Key,
		t.Title,
		t.Status,
		t.Priority,
		t.Severity,
		t.Type,
		t.UseCase,
		t.ProjectArea,
		t.Synopsis,
		t.SourceKind,
		t.SourceRef,
		t.Slice,
		t.Branch,
		t.ReviewState,
		t.ReviewResult,
		t.ReviewCommentsURL,
		t.RelativePath,
	}

	for _, value := range values {
		if strings.Contains(strings.ToLower(value), filter) {
			return true
		}
	}

	return false
}
