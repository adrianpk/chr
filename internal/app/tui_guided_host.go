package app

import (
	"context"

	"forge.adrianpk.com/chr/internal/tui"
)

// Startup owns recovery and execution; terminal refresh remains a pure query.
func runManagedTUI(ctx context.Context, options tui.Options, operations service) error {
	ctx, cancel := context.WithCancel(ctx)
	host := newGuidedWorkHost(ctx)

	defer func() {
		cancel()
		host.close()
	}()

	operations.ctx, operations.guidedWorkHost = ctx, host
	host.start(func(ctx context.Context) service {
		scoped := operations
		scoped.ctx = ctx

		return scoped
	})

	options.GuidedExecutor = tuiGuidedExecutor{operations: operations}

	return tui.Run(ctx, options)
}
