package app

import (
	"context"
	"fmt"

	"github.com/urfave/cli/v3"
)

type dumpCLIUpdate func(service, string) (DumpOperationResult, error)

func (factory commandFactory) runDumpUpdate(ctx context.Context, cmd *cli.Command, operation string, apply dumpCLIUpdate) error {
	operations, err := factory.projectService(ctx, cmd)
	if err != nil {
		return err
	}

	revision, err := dumpCLIRevision(operations, cmd)
	if err != nil {
		return err
	}

	result, err := apply(operations, revision)
	if err != nil {
		return err
	}

	if _, err := fmt.Fprintln(factory.io.Stdout, result.Message); err != nil {
		return fmt.Errorf("write dump %s result: %w", operation, err)
	}

	return nil
}
