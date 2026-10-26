package app

import (
	"context"
	"fmt"

	"github.com/urfave/cli/v3"
)

func (factory commandFactory) dumpCommand() *cli.Command {
	return &cli.Command{
		Name:  "dump",
		Usage: "Manage workspace dumps",
		Commands: []*cli.Command{
			factory.dumpAppendCommand(),
			factory.dumpListCommand(),
			factory.dumpNewCommand(),
			factory.dumpOpenCommand(),
			factory.dumpSaveCommand(),
			factory.dumpShowCommand(),
		},
	}
}

func (factory commandFactory) dumpListCommand() *cli.Command {
	return &cli.Command{
		Name:   "list",
		Usage:  "List dumps for the resolved project",
		Flags:  ticketProjectFlags(),
		Action: factory.runDumpList,
	}
}

func (factory commandFactory) dumpNewCommand() *cli.Command {
	return &cli.Command{
		Name:   "new",
		Usage:  "Create a dump",
		Flags:  dumpWriteFlags(),
		Action: factory.runDumpNew,
	}
}

func (factory commandFactory) dumpShowCommand() *cli.Command {
	return &cli.Command{
		Name:      "show",
		Usage:     "Show dump content",
		ArgsUsage: "<project-relative-path>",
		Flags:     append(ticketProjectFlags(), &cli.BoolFlag{Name: "revision", Usage: "print the observed document revision"}),
		Action:    factory.runDumpShow,
	}
}

func (factory commandFactory) dumpSaveCommand() *cli.Command {
	return &cli.Command{
		Name:      "save",
		Usage:     "Replace dump content",
		ArgsUsage: "<project-relative-path>",
		Flags:     dumpContentFlags(),
		Action:    factory.runDumpSave,
	}
}

func (factory commandFactory) dumpAppendCommand() *cli.Command {
	return &cli.Command{
		Name:      "append",
		Usage:     "Append content to a dump",
		ArgsUsage: "<project-relative-path>",
		Flags:     dumpContentFlags(),
		Action:    factory.runDumpAppend,
	}
}

func (factory commandFactory) dumpOpenCommand() *cli.Command {
	return &cli.Command{
		Name:      "open",
		Usage:     "Print the editor command for a dump",
		ArgsUsage: "<project-relative-path>",
		Flags:     ticketProjectFlags(),
		Action:    factory.runDumpOpen,
	}
}

func dumpContentFlags() []cli.Flag {
	return append(ticketProjectFlags(),
		&cli.StringFlag{Name: "operation-id", Usage: "stable identity for retrying the same write"},
		&cli.StringFlag{Name: "revision", Usage: "observed revision; retain with operation-id for edit retries"},
		&cli.StringFlag{
			Name:  "content",
			Usage: "dump markdown content",
		},
	)
}

func dumpWriteFlags() []cli.Flag {
	return append(dumpContentFlags(),
		&cli.StringFlag{
			Name:  "title",
			Usage: "dump title",
		},
		&cli.StringFlag{
			Name:  "use-case",
			Usage: "use-case slug",
		},
	)
}

func (factory commandFactory) runDumpList(ctx context.Context, cmd *cli.Command) error {
	if cmd.NArg() != 0 {
		return fmt.Errorf("dump list does not accept positional arguments")
	}

	operations, _, err := factory.serviceFromCommand(ctx, cmd)
	if err != nil {
		return err
	}

	result, err := operations.ListDumps()
	if err != nil {
		return err
	}

	for _, dump := range result.Dumps {
		_, err = fmt.Fprintf(factory.io.Stdout,
			"%s\t%s\t%s\t%s\n", dump.RelativePath, dump.Scope.CaseSlug, dump.Status, dump.Title)
		if err != nil {
			return fmt.Errorf("write dump list result: %w", err)
		}
	}

	return nil
}

func (factory commandFactory) runDumpNew(ctx context.Context, cmd *cli.Command) error {
	if cmd.NArg() != 0 {
		return fmt.Errorf("dump new does not accept positional arguments")
	}

	operations, _, err := factory.serviceFromCommand(ctx, cmd)
	if err != nil {
		return err
	}

	result, err := operations.CreateDump(DumpCreateRequest{
		OperationID: dumpRequestID(cmd.String("operation-id")), ExpectedRevision: cmd.String("revision"),
		Title:   cmd.String("title"),
		Content: cmd.String("content"),
		Scope:   DumpScope{CaseSlug: cmd.String("use-case")},
	})
	if err != nil {
		return err
	}

	_, err = fmt.Fprintln(factory.io.Stdout,
		result.Message)
	if err != nil {
		return fmt.Errorf("write dump new result: %w", err)
	}

	return nil
}

func (factory commandFactory) runDumpShow(ctx context.Context, cmd *cli.Command) error {
	if cmd.NArg() != 1 {
		return fmt.Errorf("dump show requires exactly one path argument")
	}

	operations, _, err := factory.serviceFromCommand(ctx, cmd)
	if err != nil {
		return err
	}

	result, err := operations.ReadDump(DumpReadRequest{RelativePath: cmd.Args().First()})
	if err != nil {
		return err
	}

	if cmd.Bool("revision") {
		_, err = fmt.Fprintln(factory.io.Stdout,
			result.Revision)
	} else {
		_, err = fmt.Fprint(factory.io.Stdout,
			result.Content)
	}

	if err != nil {
		return fmt.Errorf("write dump show result: %w", err)
	}

	return nil
}

func (factory commandFactory) runDumpSave(ctx context.Context, cmd *cli.Command) error {
	if cmd.NArg() != 1 {
		return fmt.Errorf("dump save requires exactly one path argument")
	}

	operations, _, err := factory.serviceFromCommand(ctx, cmd)
	if err != nil {
		return err
	}

	revision, err := dumpCLIRevision(operations, cmd)
	if err != nil {
		return err
	}

	result, err := operations.SaveDump(DumpSaveRequest{
		OperationID: dumpRequestID(cmd.String("operation-id")), ExpectedRevision: revision,
		RelativePath: cmd.Args().First(),
		Content:      cmd.String("content"),
	})
	if err != nil {
		return err
	}

	_, err = fmt.Fprintln(factory.io.Stdout,
		result.Message)
	if err != nil {
		return fmt.Errorf("write dump save result: %w", err)
	}

	return nil
}

func (factory commandFactory) runDumpAppend(ctx context.Context, cmd *cli.Command) error {
	if cmd.NArg() != 1 {
		return fmt.Errorf("dump append requires exactly one path argument")
	}

	operations, _, err := factory.serviceFromCommand(ctx, cmd)
	if err != nil {
		return err
	}

	revision, err := dumpCLIRevision(operations, cmd)
	if err != nil {
		return err
	}

	result, err := operations.AppendDump(DumpAppendRequest{
		OperationID: dumpRequestID(cmd.String("operation-id")), ExpectedRevision: revision,
		RelativePath: cmd.Args().First(),
		Content:      cmd.String("content"),
	})
	if err != nil {
		return err
	}

	_, err = fmt.Fprintln(factory.io.Stdout,
		result.Message)
	if err != nil {
		return fmt.Errorf("write dump append result: %w", err)
	}

	return nil
}

func (factory commandFactory) runDumpOpen(ctx context.Context, cmd *cli.Command) error {
	if cmd.NArg() != 1 {
		return fmt.Errorf("dump open requires exactly one path argument")
	}

	operations, _, err := factory.serviceFromCommand(ctx, cmd)
	if err != nil {
		return err
	}

	result, err := operations.OpenDump(DumpOpenRequest{RelativePath: cmd.Args().First()})
	if err != nil {
		return err
	}

	if result.EditorCommand == nil {
		return fmt.Errorf("dump editor command is missing")
	}

	_, err = fmt.Fprintf(factory.io.Stdout,
		"editor: %s %v\n", result.EditorCommand.Program, result.EditorCommand.Args)
	if err != nil {
		return fmt.Errorf("write dump open result: %w", err)
	}

	return nil
}

func dumpCLIRevision(operations service, cmd *cli.Command) (string, error) {
	if revision := cmd.String("revision"); revision != "" {
		return revision, nil
	}

	if cmd.String("operation-id") != "" {
		return "", fmt.Errorf("dump edit retries require the original --revision with --operation-id")
	}

	observed, err := operations.ReadDump(DumpReadRequest{RelativePath: cmd.Args().First()})

	return observed.Revision, err
}
