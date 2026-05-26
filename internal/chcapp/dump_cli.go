package chcapp

import (
	"context"
	"fmt"
	"io"

	"github.com/urfave/cli/v3"
)

func newDumpCommand(stdout io.Writer, deps appDependencies) *cli.Command {
	return &cli.Command{
		Name:  "dump",
		Usage: "Manage workspace dumps",
		Commands: []*cli.Command{
			newDumpAppendCommand(stdout),
			newDumpListCommand(stdout),
			newDumpNewCommand(stdout),
			newDumpOpenCommand(stdout),
			newDumpSaveCommand(stdout),
			newDumpShowCommand(stdout),
		},
	}
}

func newDumpListCommand(stdout io.Writer) *cli.Command {
	return &cli.Command{
		Name:   "list",
		Usage:  "List dumps for the resolved project",
		Flags:  ticketProjectFlags(),
		Action: dumpListAction(stdout),
	}
}

func newDumpNewCommand(stdout io.Writer) *cli.Command {
	return &cli.Command{
		Name:   "new",
		Usage:  "Create a dump",
		Flags:  dumpWriteFlags(),
		Action: dumpNewAction(stdout),
	}
}

func newDumpShowCommand(stdout io.Writer) *cli.Command {
	return &cli.Command{
		Name:      "show",
		Usage:     "Show dump content",
		ArgsUsage: "<project-relative-path>",
		Flags:     ticketProjectFlags(),
		Action:    dumpShowAction(stdout),
	}
}

func newDumpSaveCommand(stdout io.Writer) *cli.Command {
	return &cli.Command{
		Name:      "save",
		Usage:     "Replace dump content",
		ArgsUsage: "<project-relative-path>",
		Flags:     dumpContentFlags(),
		Action:    dumpSaveAction(stdout),
	}
}

func newDumpAppendCommand(stdout io.Writer) *cli.Command {
	return &cli.Command{
		Name:      "append",
		Usage:     "Append content to a dump",
		ArgsUsage: "<project-relative-path>",
		Flags:     dumpContentFlags(),
		Action:    dumpAppendAction(stdout),
	}
}

func newDumpOpenCommand(stdout io.Writer) *cli.Command {
	return &cli.Command{
		Name:      "open",
		Usage:     "Print the editor command for a dump",
		ArgsUsage: "<project-relative-path>",
		Flags:     ticketProjectFlags(),
		Action:    dumpOpenAction(stdout),
	}
}

func dumpContentFlags() []cli.Flag {
	return append(ticketProjectFlags(),
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

func dumpListAction(stdout io.Writer) cli.ActionFunc {
	return func(ctx context.Context, cmd *cli.Command) error {
		return runDumpList(ctx, cmd, stdout)
	}
}

func dumpNewAction(stdout io.Writer) cli.ActionFunc {
	return func(ctx context.Context, cmd *cli.Command) error {
		return runDumpNew(ctx, cmd, stdout)
	}
}

func dumpShowAction(stdout io.Writer) cli.ActionFunc {
	return func(ctx context.Context, cmd *cli.Command) error {
		return runDumpShow(ctx, cmd, stdout)
	}
}

func dumpSaveAction(stdout io.Writer) cli.ActionFunc {
	return func(ctx context.Context, cmd *cli.Command) error {
		return runDumpSave(ctx, cmd, stdout)
	}
}

func dumpAppendAction(stdout io.Writer) cli.ActionFunc {
	return func(ctx context.Context, cmd *cli.Command) error {
		return runDumpAppend(ctx, cmd, stdout)
	}
}

func dumpOpenAction(stdout io.Writer) cli.ActionFunc {
	return func(ctx context.Context, cmd *cli.Command) error {
		return runDumpOpen(ctx, cmd, stdout)
	}
}

func runDumpList(ctx context.Context, cmd *cli.Command, stdout io.Writer) error {
	if cmd.NArg() != 0 {
		return fmt.Errorf("dump list does not accept positional arguments")
	}

	operations, _, err := serviceFromCommand(ctx, cmd)
	if err != nil {
		return err
	}

	result, err := operations.ListDumps()
	if err != nil {
		return err
	}

	for _, dump := range result.Dumps {
		_, err = fmt.Fprintf(stdout, "%s\t%s\t%s\t%s\n", dump.RelativePath, dump.Scope.CaseSlug, dump.Status, dump.Title)
		if err != nil {
			return fmt.Errorf("write dump list result: %w", err)
		}
	}

	return nil
}

func runDumpNew(ctx context.Context, cmd *cli.Command, stdout io.Writer) error {
	if cmd.NArg() != 0 {
		return fmt.Errorf("dump new does not accept positional arguments")
	}

	operations, _, err := serviceFromCommand(ctx, cmd)
	if err != nil {
		return err
	}

	result, err := operations.CreateDump(DumpCreateRequest{
		Title:   cmd.String("title"),
		Content: cmd.String("content"),
		Scope:   DumpScope{CaseSlug: cmd.String("use-case")},
	})
	if err != nil {
		return err
	}

	_, err = fmt.Fprintln(stdout, result.Message)
	if err != nil {
		return fmt.Errorf("write dump new result: %w", err)
	}

	return nil
}

func runDumpShow(ctx context.Context, cmd *cli.Command, stdout io.Writer) error {
	if cmd.NArg() != 1 {
		return fmt.Errorf("dump show requires exactly one path argument")
	}

	operations, _, err := serviceFromCommand(ctx, cmd)
	if err != nil {
		return err
	}

	result, err := operations.ReadDump(DumpReadRequest{RelativePath: cmd.Args().First()})
	if err != nil {
		return err
	}

	_, err = fmt.Fprint(stdout, result.Content)
	if err != nil {
		return fmt.Errorf("write dump show result: %w", err)
	}

	return nil
}

func runDumpSave(ctx context.Context, cmd *cli.Command, stdout io.Writer) error {
	if cmd.NArg() != 1 {
		return fmt.Errorf("dump save requires exactly one path argument")
	}

	operations, _, err := serviceFromCommand(ctx, cmd)
	if err != nil {
		return err
	}

	result, err := operations.SaveDump(DumpSaveRequest{
		RelativePath: cmd.Args().First(),
		Content:      cmd.String("content"),
	})
	if err != nil {
		return err
	}

	_, err = fmt.Fprintln(stdout, result.Message)
	if err != nil {
		return fmt.Errorf("write dump save result: %w", err)
	}

	return nil
}

func runDumpAppend(ctx context.Context, cmd *cli.Command, stdout io.Writer) error {
	if cmd.NArg() != 1 {
		return fmt.Errorf("dump append requires exactly one path argument")
	}

	operations, _, err := serviceFromCommand(ctx, cmd)
	if err != nil {
		return err
	}

	result, err := operations.AppendDump(DumpAppendRequest{
		RelativePath: cmd.Args().First(),
		Content:      cmd.String("content"),
	})
	if err != nil {
		return err
	}

	_, err = fmt.Fprintln(stdout, result.Message)
	if err != nil {
		return fmt.Errorf("write dump append result: %w", err)
	}

	return nil
}

func runDumpOpen(ctx context.Context, cmd *cli.Command, stdout io.Writer) error {
	if cmd.NArg() != 1 {
		return fmt.Errorf("dump open requires exactly one path argument")
	}

	operations, _, err := serviceFromCommand(ctx, cmd)
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

	_, err = fmt.Fprintf(stdout, "editor: %s %v\n", result.EditorCommand.Program, result.EditorCommand.Args)
	if err != nil {
		return fmt.Errorf("write dump open result: %w", err)
	}

	return nil
}
