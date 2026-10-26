package app

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/urfave/cli/v3"

	"forge.adrianpk.com/chr/internal/core"
)

func (factory commandFactory) ticketBoardCommand() *cli.Command {
	return &cli.Command{
		Name:   "board",
		Usage:  "Inspect Board columns for the resolved project",
		Flags:  ticketProjectFlags(),
		Action: factory.runTicketBoard,
	}
}

func (factory commandFactory) ticketListCommand() *cli.Command {
	return &cli.Command{
		Name:   "list",
		Usage:  "List tickets for the resolved project",
		Flags:  ticketProjectFlags(),
		Action: factory.runTicketList,
	}
}

func (factory commandFactory) ticketShowCommand() *cli.Command {
	return &cli.Command{
		Name:   "show",
		Usage:  "Show one ticket by key",
		Flags:  ticketProjectFlags(),
		Action: factory.runTicketShow,
	}
}

func (factory commandFactory) runTicketShow(ctx context.Context, cmd *cli.Command) error {
	if cmd.NArg() != 1 {
		return fmt.Errorf("ticket show requires exactly one ticket key argument")
	}

	env, err := newUserEnvironmentFromCommand(cmd)
	if err != nil {
		return err
	}

	projectContext, index, err := factory.openTicketIndexForCommand(ctx, cmd, env)
	if err != nil {
		return err
	}

	defer func() {
		//nolint:errcheck // Command result already comes from the main operation.
		_ = index.Close()
	}()

	err = refreshProjectTicketIndex(ctx, index, projectContext)
	if err != nil {
		return err
	}

	ticket, err := index.GetProjectTicketByKey(ctx, projectContext.ProjectID, cmd.Args().First())
	if err != nil {
		return err
	}

	return printTicketShow(factory.io.Stdout, ticket, ticketEvidenceSummary(ticket))
}

func printTicketShow(stdout io.Writer, ticket core.IndexedTicket, evidence core.EvidenceStrategySummary) error {
	lines := []string{
		"key: " + ticket.Key,
		"title: " + ticket.Title,
		"status: " + ticket.Status,
		"priority: " + ticket.Priority,
		"severity: " + ticket.Severity,
		"type: " + ticket.Type,
		"use_case: " + ticket.UseCase,
		"project_area: " + ticket.ProjectArea,
		"branch: " + ticket.Branch,
		"review_state: " + ticket.ReviewState,
		fmt.Sprintf("review_round: %d", ticket.ReviewRound),
		"review_result: " + ticket.ReviewResult,
		"review_comments_url: " + ticket.ReviewCommentsURL,
		"path: " + ticket.RelativePath,
	}

	lines = append(lines, ticketEvidenceShowLines(evidence)...)

	for _, line := range lines {
		_, err := fmt.Fprintln(stdout, line)
		if err != nil {
			return fmt.Errorf("write ticket show result: %w", err)
		}
	}

	return nil
}

func (factory commandFactory) runTicketList(ctx context.Context, cmd *cli.Command) error {
	if cmd.NArg() != 0 {
		return fmt.Errorf("ticket list does not accept positional arguments")
	}

	env, err := newUserEnvironmentFromCommand(cmd)
	if err != nil {
		return err
	}

	projectContext, err := newCommandProjectContextResolver(env, factory.deps.assetsFS).Resolve(ctx, commandProjectContextOptions{
		ExplicitDir:           cmd.String("dir"),
		UseExplicitDir:        cmd.IsSet("dir"),
		AllowCurrentDirectory: true,
	})
	if err != nil {
		return err
	}

	index, err := core.NewTicketIndex(ctx, env, factory.deps.assetsFS)
	if err != nil {
		return err
	}

	defer func() {
		//nolint:errcheck // Command result already comes from the main operation.
		_ = index.Close()
	}()

	err = refreshProjectTicketIndex(ctx, index, projectContext)
	if err != nil {
		return err
	}

	tickets, err := index.ListProjectTickets(ctx, projectContext.ProjectID)
	if err != nil {
		return err
	}

	return printTicketList(factory.io.Stdout, tickets)
}

func (factory commandFactory) runTicketBoard(ctx context.Context, cmd *cli.Command) error {
	if cmd.NArg() != 0 {
		return fmt.Errorf("ticket board does not accept positional arguments")
	}

	env, err := newUserEnvironmentFromCommand(cmd)
	if err != nil {
		return err
	}

	projectContext, err := newCommandProjectContextResolver(env, factory.deps.assetsFS).Resolve(ctx, commandProjectContextOptions{
		ExplicitDir:           cmd.String("dir"),
		UseExplicitDir:        cmd.IsSet("dir"),
		AllowCurrentDirectory: true,
	})
	if err != nil {
		return err
	}

	board, err := newService(ctx, env, projectContext, factory.deps.assetsFS).Board()
	if err != nil {
		return err
	}

	return printTicketBoard(factory.io.Stdout, board)
}

func printTicketBoard(stdout io.Writer, board core.Board) error {
	lines := []string{
		fmt.Sprintf("total: %d", board.TotalCount),
		fmt.Sprintf("normal: %d", board.NormalCount),
		fmt.Sprintf("attention: %d", board.AttentionCount),
	}

	for _, line := range lines {
		if _, err := fmt.Fprintln(stdout, line); err != nil {
			return fmt.Errorf("write ticket board result: %w", err)
		}
	}

	for _, column := range board.Columns {
		if err := printTicketBoardColumn(stdout, column); err != nil {
			return err
		}
	}

	return nil
}

func printTicketBoardColumn(stdout io.Writer, column core.BoardColumn) error {
	_, err := fmt.Fprintf(stdout, "column: %s\t%s\t%d\n", column.ID, column.Title, column.Count)
	if err != nil {
		return fmt.Errorf("write ticket board column: %w", err)
	}

	for _, ticket := range column.Tickets {
		_, err = fmt.Fprintf(
			stdout,
			"%s\t%s\t%s\t%s\t%s\t%s\n",
			ticket.Key,
			ticket.Status,
			ticket.Priority,
			ticket.Type,
			ticket.Title,
			ticketEvidenceStatusCell(ticketEvidenceSummary(ticket)),
		)
		if err != nil {
			return fmt.Errorf("write ticket board row: %w", err)
		}
	}

	return nil
}

func printTicketList(stdout io.Writer, tickets []core.IndexedTicket) error {
	for _, ticket := range tickets {
		_, err := fmt.Fprintf(
			stdout,
			"%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\n",
			ticket.Key,
			ticket.Title,
			ticket.Status,
			ticket.Priority,
			ticket.Severity,
			ticket.Type,
			ticket.UseCase,
			ticket.RelativePath,
			ticketEvidenceStatusCell(ticketEvidenceSummary(ticket)),
		)
		if err != nil {
			return fmt.Errorf("write ticket list result: %w", err)
		}
	}

	return nil
}

func ticketEvidenceSummary(ticket core.IndexedTicket) core.EvidenceStrategySummary {
	content, err := os.ReadFile(ticket.Path)
	if err != nil {
		return core.EvidenceStrategySummary{}
	}

	return core.SummarizeEvidenceStrategyChecklist(string(content))
}

func ticketEvidenceShowLines(evidence core.EvidenceStrategySummary) []string {
	if !ticketEvidenceVisible(evidence) {
		return nil
	}

	lines := []string{"evidence: " + string(evidence.Status)}
	if evidence.Error != "" {
		return append(lines, "evidence_error: "+evidence.Error)
	}

	for _, item := range evidence.Items {
		lines = append(lines, ticketEvidenceItemLine(item))
	}

	return lines
}

func ticketEvidenceItemLine(item core.EvidenceStrategySummaryItem) string {
	return ticketEvidenceItemKey(item.Label) + ": " + item.Status + " " + ticketEvidenceItemState(item.Checked)
}

func ticketEvidenceItemKey(label string) string {
	switch label {
	case core.EvidenceStrategyChecklistLabelStrategy:
		return "evidence_strategy"
	case core.EvidenceStrategyChecklistLabelGherkin:
		return "evidence_gherkin"
	case core.EvidenceStrategyChecklistLabelTests:
		return "evidence_tests"
	case core.EvidenceStrategyChecklistLabelValidation:
		return "evidence_validation"
	case core.EvidenceStrategyChecklistLabelSummary:
		return "evidence_summary"
	default:
		return "evidence_" + strings.ToLower(strings.ReplaceAll(label, " ", "_"))
	}
}

func ticketEvidenceItemState(checked bool) string {
	if checked {
		return "done"
	}

	return "pending"
}

func ticketEvidenceStatusCell(evidence core.EvidenceStrategySummary) string {
	if !ticketEvidenceVisible(evidence) {
		return "-"
	}

	return string(evidence.Status)
}

func ticketEvidenceVisible(evidence core.EvidenceStrategySummary) bool {
	return evidence.Status != "" && evidence.Status != core.EvidenceStrategySummaryStatusMissing
}
