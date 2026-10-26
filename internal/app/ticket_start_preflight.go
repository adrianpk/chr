package app

import (
	"context"

	"github.com/urfave/cli/v3"
)

func (factory commandFactory) buildTicketStartPreflight(ctx context.Context, cmd *cli.Command) (ticketStartPreflight, error) {
	operations, err := factory.projectService(ctx, cmd)
	if err != nil {
		return ticketStartPreflight{}, err
	}

	ticket, err := operations.selectTicketStartTicket(ctx, cmd.Args().First())
	if err != nil {
		return ticketStartPreflight{}, err
	}

	return operations.prepareTicketStart(ctx, ticket, ticketStartPreflightOptions{PrepareWorktree: !cmd.Bool("dry-run")})
}
