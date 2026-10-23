package app

import (
	"context"
	"errors"
	"log/slog"
	"sync"
	"time"

	"forge.adrianpk.com/chr/internal/core"
)

const guidedHostActivationLimit = 16

type guidedWorkActivator interface {
	ready() error
	activate(service, string) error
}

// guidedWorkHost schedules existing native continuations on the managed host's
// lifetime. Its in-memory map excludes duplicate dispatch; native claims remain
// the cross-process authority and retain all budgets and stop intentions.
type guidedWorkHost struct {
	ctx      context.Context
	cancel   context.CancelFunc
	mu       sync.Mutex
	active   map[string]bool
	progress map[string]bool
	attempts map[string]string
	wait     sync.WaitGroup
	scanWait sync.WaitGroup
}

func newGuidedWorkHost(ctx context.Context) *guidedWorkHost {
	ctx, cancel := context.WithCancel(ctx)

	return &guidedWorkHost{ctx: ctx, cancel: cancel, active: make(map[string]bool), progress: make(map[string]bool), attempts: make(map[string]string)}
}

func (host *guidedWorkHost) ready() error {
	if host.ctx.Err() != nil {
		return &core.GuidedPrerequisiteError{Action: "Start a managed Web host, then retry this retained operation.", Cause: host.ctx.Err()}
	}

	return nil
}

func (host *guidedWorkHost) activate(operations service, id string) error {
	host.mu.Lock()
	defer host.mu.Unlock()

	if err := host.ready(); err != nil {
		return err
	}

	key := operations.projectContext.ProjectDir + "\x00" + operations.workflowNamespace + "\x00" + id
	if host.active[key] {
		return nil
	}

	if len(host.active) >= guidedHostActivationLimit {
		return &core.GuidedPrerequisiteError{Action: "Managed host capacity is full. Retry this retained admission when capacity is available."}
	}

	host.active[key] = true

	host.wait.Add(1)
	go func() {
		defer host.wait.Done()
		defer host.release(key)

		failure := host.run(operations, id)
		host.acknowledgeStop(host.ctx, operations, id)

		if guidedHostFailure(failure) && host.ctx.Err() == nil {
			host.reportFailure(host.ctx, operations, id)
		}
	}()

	return nil
}

func guidedHostFailure(err error) bool {
	return err != nil && !errors.Is(err, core.ErrWorkflowOwned) && !errors.Is(err, core.ErrWorkflowCapacity) && !errors.Is(err, core.ErrWorkflowConflict) && !errors.Is(err, core.ErrGuidedUseCaseStopped) && !errors.Is(err, core.ErrWorkflowStop)
}

func (host *guidedWorkHost) release(key string) {
	host.mu.Lock()
	defer host.mu.Unlock()

	delete(host.active, key)
}

func (host *guidedWorkHost) run(operations service, id string) error {
	// The native owner applies the original finite per-task and integration
	// budgets. This extra bound limits host scheduling across the whole extent.
	for range core.WorkflowRecordLimit {
		if host.ctx.Err() != nil {
			return host.ctx.Err()
		}

		run, err := operations.ReadWorkflow(host.ctx, id)
		if err != nil || run.Admission == nil || run.Admission.Guided == nil {
			return errors.Join(err, core.ErrWorkflowConflict)
		}

		check := func(ctx context.Context) error { return operations.checkGuidedActivation(ctx, run) }

		progress, err := operations.ResumeRetainedWorkGuarded(host.ctx, id, run.Revision, check)
		if err != nil || progress.Steps == 0 || !retainedWorkRunnable(progress.Status) {
			return err
		}
	}

	return errors.New("guided host checkpoint scheduling bound reached")
}

func (host *guidedWorkHost) reportFailure(ctx context.Context, operations service, id string) {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	run, err := operations.ReadWorkflow(ctx, id)
	if err == nil && run.Admission != nil && run.Admission.Guided != nil {
		scope := core.GuidedUseCaseScope{RunID: run.Admission.Guided.RunID, ProjectID: operations.projectContext.ProjectID, Namespace: operations.workflowNamespace}
		err = core.RetainGuidedHostBlocker(ctx, operations.projectContext.ProjectDir, scope, id)
	}

	if err != nil {
		slog.ErrorContext(ctx, "Guided work activation stopped", "invocation", id, "code", "host_blocker_retention_failed")
	}
}

func (host *guidedWorkHost) close() {
	host.mu.Lock()
	host.cancel()
	host.mu.Unlock()
	host.scanWait.Wait()
	host.wait.Wait()
}
