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
	// Pending adapter calls share the UI lifetime. Shutdown quiesces their IO
	// without issuing an operator cancellation to retained work.
	options.QuizExecutor = newTUIQuizActionExecutor(operations)
	options.SpecExecutor = newTUISpecActionExecutor(operations)
	options.PlanExecutor = newTUIPlanActionExecutor(operations)
	options.WorkExecutor = newTUIWorkActionExecutor(operations)
	options.AgentWorkExecutor = newTUIAgentWorkRefreshExecutor(operations)

	return tui.Run(ctx, options)
}
