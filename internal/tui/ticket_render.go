package tui

import (
	"fmt"
	"strings"

	"forge.adrianpk.com/chr/internal/core"
)

func (m *Model) renderTicketBoard() string {
	filteredTickets := m.filteredTickets()

	var output strings.Builder

	output.WriteString("Board\n")
	output.WriteString(m.renderBoardHeaderColumnLine())
	fmt.Fprintf(&output, "  Tickets: %d", len(filteredTickets))

	if m.ticketFilter != "" {
		fmt.Fprintf(&output, "  Filter: %q", m.ticketFilter)
	}

	output.WriteString("\n\n")

	if len(filteredTickets) == 0 {
		output.WriteString("No tickets in this column.\n\n")
		output.WriteString(m.renderSelectedTicketDetail())

		return strings.TrimRight(output.String(), "\n")
	}

	for _, ticket := range filteredTickets {
		prefix := "  "
		if ticket.Index == m.selectedTicket {
			prefix = "> "
		}

		output.WriteString(prefix)
		output.WriteString(renderTicketRow(ticket))
		output.WriteString("\n")
	}

	output.WriteString("\n")
	output.WriteString(m.renderSelectedTicketDetail())

	return strings.TrimRight(output.String(), "\n")
}

func (m *Model) renderBoardHeaderColumnLine() string {
	if m.boardAttentionActive() {
		return fmt.Sprintf(
			"  Attention (%d)\n",
			len(m.attentionTickets()),
		)
	}

	return fmt.Sprintf(
		"  %s (%d/%d)\n",
		core.BoardColumnTitle(m.activeBoardColumn()),
		m.activeBoardColumnPosition()+1,
		len(core.NormalBoardColumns),
	)
}

func (m *Model) renderTicketList() string {
	filteredTickets := m.filteredTickets()
	if len(filteredTickets) == 0 {
		return "Tickets\n  No tickets match the current view."
	}

	var output strings.Builder

	output.WriteString("Tickets\n")
	output.WriteString("  Key     Status      Priority  Type      Title\n")

	for _, ticket := range filteredTickets {
		prefix := "  "
		if ticket.Index == m.selectedTicket {
			prefix = "> "
		}

		output.WriteString(prefix)
		output.WriteString(renderTicketRow(ticket))
		output.WriteString("\n")
	}

	output.WriteString("\n")
	output.WriteString(m.renderSelectedTicketDetail())

	return strings.TrimRight(output.String(), "\n")
}

func renderTicketRow(ticket TicketView) string {
	if len(ticket.AttentionReasons) > 0 {
		return fmt.Sprintf(
			"! %-7s %-11s %-9s %-9s %s",
			ticket.Key,
			ticket.Status,
			ticket.Priority,
			ticket.Type,
			ticket.Title,
		)
	}

	return fmt.Sprintf(
		"%-7s %-11s %-9s %-9s %s",
		ticket.Key,
		ticket.Status,
		ticket.Priority,
		ticket.Type,
		ticket.Title,
	)
}

func (m *Model) renderSelectedTicketDetail() string {
	ticket, ok := m.selectedTicketView()
	if !ok {
		return "Selected Ticket\n  No ticket selected."
	}

	var output strings.Builder

	output.WriteString("Selected Ticket\n")
	fmt.Fprintf(&output, "  Key: %s\n", ticket.Key)
	fmt.Fprintf(&output, "  Title: %s\n", ticket.Title)

	if ticket.Synopsis != "" {
		fmt.Fprintf(&output, "  Synopsis: %s\n", ticket.Synopsis)
	}

	fmt.Fprintf(&output, "  Status: %s\n", ticket.Status)
	fmt.Fprintf(&output, "  Priority: %s\n", ticket.Priority)
	fmt.Fprintf(&output, "  Severity: %s\n", ticket.Severity)
	fmt.Fprintf(&output, "  Type: %s\n", ticket.Type)
	fmt.Fprintf(&output, "  Use Case: %s\n", ticket.UseCase)
	fmt.Fprintf(&output, "  Project Area: %s\n", ticket.ProjectArea)
	renderTicketSourceDetail(&output, ticket)
	renderTicketReviewDetail(&output, ticket)
	renderTicketEvidenceDetail(&output, ticket)
	fmt.Fprintf(&output, "  Checklist: %d done / %d total\n", ticket.ChecklistDone, ticket.ChecklistTotal)
	fmt.Fprintf(&output, "  Comments: %d\n", ticket.CommentCount)
	fmt.Fprintf(&output, "  Path: %s\n", ticketDisplayPath(ticket.RelativePath))
	renderTicketBodySection(&output, ticket.BodyLines)

	if len(ticket.AttentionReasons) > 0 {
		output.WriteString("  Attention:\n")

		for _, reason := range ticket.AttentionReasons {
			fmt.Fprintf(&output, "    - %s\n", reason)
		}
	}

	return output.String()
}

func renderTicketEvidenceDetail(output *strings.Builder, ticket TicketView) {
	if ticket.Evidence.Status == "" {
		return
	}

	fmt.Fprintf(output, "  Evidence: %s\n", ticket.Evidence.Status)

	if ticket.Evidence.Error != "" {
		fmt.Fprintf(output, "    Error: %s\n", ticket.Evidence.Error)

		return
	}

	for _, item := range ticket.Evidence.Items {
		fmt.Fprintf(output, "    %s: %s (%s)\n", item.Label, item.Status, checkedStateLabel(item.Checked))
	}
}

func checkedStateLabel(checked bool) string {
	if checked {
		return "done"
	}

	return "pending"
}

func renderTicketReviewDetail(output *strings.Builder, ticket TicketView) {
	if !hasTicketReviewDetail(ticket) {
		return
	}

	renderTicketReviewStringDetail(output, "Branch", ticket.Branch)
	renderTicketReviewStringDetail(output, "Review State", ticket.ReviewState)

	if ticket.ReviewRound > 0 {
		fmt.Fprintf(output, "  Review Round: %d\n", ticket.ReviewRound)
	}

	renderTicketReviewStringDetail(output, "Review Result", ticket.ReviewResult)
	renderTicketReviewStringDetail(output, "Review Comments", ticket.ReviewCommentsURL)
}

func hasTicketReviewDetail(ticket TicketView) bool {
	return ticket.Branch != "" ||
		ticket.ReviewState != "" ||
		ticket.ReviewResult != "" ||
		ticket.ReviewCommentsURL != "" ||
		ticket.ReviewRound > 0
}

func renderTicketReviewStringDetail(output *strings.Builder, label, value string) {
	if value == "" {
		return
	}

	fmt.Fprintf(output, "  %s: %s\n", label, value)
}

func renderTicketSourceDetail(output *strings.Builder, ticket TicketView) {
	if ticket.SourceKind == "" && ticket.SourceRef == "" && ticket.Slice == "" {
		return
	}

	if ticket.SourceKind != "" {
		fmt.Fprintf(output, "  Source: %s\n", ticket.SourceKind)
	}

	if ticket.SourceRef != "" {
		fmt.Fprintf(output, "  Source Ref: %s\n", ticket.SourceRef)
	}

	if ticket.Slice != "" {
		fmt.Fprintf(output, "  Slice: %s\n", ticket.Slice)
	}
}

func renderTicketBodySection(output *strings.Builder, lines []string) {
	if len(lines) == 0 {
		return
	}

	output.WriteString("  Body:\n")

	for _, line := range lines {
		fmt.Fprintf(output, "    %s\n", line)
	}
}
