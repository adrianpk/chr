package app

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"forge.adrianpk.com/chr/internal/core"
)

// start installs recovery in the managed server lifetime, not in any GET or poll.
func (host *guidedWorkHost) start(factory func(context.Context) service) {
	if factory == nil {
		return
	}

	host.scanWait.Add(1)
	go func() {
		defer host.scanWait.Done()

		ticker := time.NewTicker(time.Second)
		defer ticker.Stop()

		offset := 0
		for {
			offset = host.recoverPage(host.ctx, factory(host.ctx), offset)
			select {
			case <-host.ctx.Done():
				return
			case <-ticker.C:
			}
		}
	}()
}

func (host *guidedWorkHost) recoverPage(ctx context.Context, operations service, offset int) int {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	page, err := core.RecoverGuidedHostPage(ctx, operations.projectContext.ProjectDir, operations.projectContext.ProjectID, offset, core.WorkflowPageLimit)
	if err != nil {
		if host.ctx.Err() == nil {
			slog.WarnContext(ctx, "Guided recovery inventory is unavailable", "code", "guided_inventory_blocked")
		}

		return 0
	}

	for _, session := range page.Runs {
		host.recoverRun(ctx, operations, session)
	}

	if page.NextOffset < 0 || page.NextOffset > core.WorkflowRecordLimit {
		return 0
	}

	return page.NextOffset
}

func (host *guidedWorkHost) recoverRun(ctx context.Context, operations service, session core.AuthoringSession) {
	scope := core.GuidedUseCaseScope{RunID: session.ID, ProjectID: session.ProjectID, Namespace: session.Namespace}
	operations.workflowNamespace = session.Namespace

	var err error
	if session.Guided.Control != nil && session.Guided.Control.StopIntent != "" && session.Guided.Control.PendingOperation == nil {
		_, err = operations.RecoverGuidedUseCase(ctx, scope)
	} else {
		err = host.scheduleProgress(operations, scope, false)
	}

	if guidedHostFailure(err) && ctx.Err() == nil && !errors.Is(err, core.ErrGuidedUseCaseStopped) {
		slog.WarnContext(ctx, "Guided run needs recovery", "run", session.ID, "code", "guided_recovery_blocked")
	}
}
