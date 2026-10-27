package app

import (
	"context"
	"fmt"

	"github.com/urfave/cli/v3"

	ticketcore "forge.adrianpk.com/chr/internal/core/ticket"
)

func (factory commandFactory) ticketNewCommand() *cli.Command {
	return &cli.Command{
		Name:   "new",
		Usage:  "Create a ticket in the current workspace",
		Flags:  ticketNewFlags(),
		Action: factory.runTicketNew,
	}
}

func (factory commandFactory) ticketCommentCommand() *cli.Command {
	return &cli.Command{
		Name:   "comment",
		Usage:  "Append a comment to one ticket",
		Flags:  ticketCommentFlags(),
		Action: factory.runTicketComment,
	}
}

func (factory commandFactory) ticketStatusCommand() *cli.Command {
	return &cli.Command{
		Name:   "status",
		Usage:  "Update one ticket status",
		Flags:  ticketProjectFlags(),
		Action: factory.runTicketStatus,
	}
}

func ticketCommentFlags() []cli.Flag {
	return append(ticketProjectFlags(),
		&cli.StringFlag{
			Name:  "author",
			Usage: "comment author",
			Value: "user",
		},
	)
}

func ticketNewFlags() []cli.Flag {
	return append(ticketProjectFlags(),
		&cli.StringFlag{
			Name:  "priority",
			Usage: "ticket priority",
		},
		&cli.StringFlag{
			Name:  "synopsis",
			Usage: "ticket synopsis",
		},
		&cli.StringFlag{
			Name:  "severity",
			Usage: "ticket severity",
		},
		&cli.StringFlag{
			Name:  "type",
			Usage: "ticket type",
		},
		&cli.StringSliceFlag{
			Name:  "tag",
			Usage: "ticket tag",
		},
		&cli.StringFlag{
			Name:  "use-case",
			Usage: "ticket use case",
		},
		&cli.StringFlag{
			Name:  "project-area",
			Usage: "ticket project area",
		},
	)
}

func (factory commandFactory) runTicketStatus(ctx context.Context, cmd *cli.Command) error {
	if cmd.NArg() != 2 {
		return fmt.Errorf("ticket status requires exactly one ticket key and one status argument")
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

	status := cmd.Args().Get(1)

	result, err := newService(ctx, env, projectContext, factory.deps.assetsFS).MoveTicketStatus(cmd.Args().Get(0), status)
	if err != nil {
		return err
	}

	updated, err := ticketOperationResultTicket(result, cmd.Args().Get(0))
	if err != nil {
		return err
	}

	_, err = fmt.Fprintf(factory.io.Stdout,
		"ticket: %s %s\n", updated.Key, status)
	if err != nil {
		return fmt.Errorf("write ticket status result: %w", err)
	}

	return nil
}

func ticketOperationResultTicket(result TicketOperationResult, key string) (ticketcore.Indexed, error) {
	for _, ticket := range result.Tickets {
		if ticket.Key == key {
			return ticket, nil
		}
	}

	return ticketcore.Indexed{}, fmt.Errorf("ticket not found after operation: %s", key)
}

func (factory commandFactory) runTicketComment(ctx context.Context, cmd *cli.Command) error {
	if cmd.NArg() != 2 {
		return fmt.Errorf("ticket comment requires exactly one ticket key and one comment argument")
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

	result, err := newService(ctx, env, projectContext, factory.deps.assetsFS).AddTicketComment(TicketCommentRequest{
		Key:    cmd.Args().Get(0),
		Author: cmd.String("author"),
		Body:   cmd.Args().Get(1),
	})
	if err != nil {
		return err
	}

	updated, err := ticketOperationResultTicket(result, cmd.Args().Get(0))
	if err != nil {
		return err
	}

	_, err = fmt.Fprintf(factory.io.Stdout,
		"ticket: %s\n", updated.Path)
	if err != nil {
		return fmt.Errorf("write ticket comment result: %w", err)
	}

	return nil
}

func (factory commandFactory) runTicketNew(ctx context.Context, cmd *cli.Command) error {
	if cmd.NArg() != 1 {
		return fmt.Errorf("ticket new requires exactly one title argument")
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

	result, err := newService(ctx, env, projectContext, factory.deps.assetsFS).CreateTicket(CreateTicketRequest{
		Title:       cmd.Args().First(),
		Synopsis:    cmd.String("synopsis"),
		Priority:    cmd.String("priority"),
		Severity:    cmd.String("severity"),
		Type:        cmd.String("type"),
		Tags:        cmd.StringSlice("tag"),
		UseCase:     cmd.String("use-case"),
		ProjectArea: cmd.String("project-area"),
	})
	if err != nil {
		return err
	}

	_, err = fmt.Fprintf(factory.io.Stdout,
		"ticket: %s\n", result.Path)
	if err != nil {
		return fmt.Errorf("write ticket result: %w", err)
	}

	return nil
}
