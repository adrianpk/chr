package app

import (
	"context"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"time"

	"github.com/urfave/cli/v3"

	"forge.adrianpk.com/chr/internal/core"
)

const daemonWorkflowActivationLimit = 16

func (factory commandFactory) daemonWorkflowCommand() *cli.Command {
	command := &cli.Command{Name: "workflow", Usage: "Control one retained workflow through shared work operations"}
	for _, action := range []string{"status", "continue", "pause", "cancel", "resume", "approve", "grant", "answer-decision", "reconcile"} {
		command.Commands = append(command.Commands, factory.workControlCommand(action))
	}

	return command
}

type daemonWorkflowRunner struct {
	stdout     io.Writer
	env        core.UserEnvironment
	assets     fs.FS
	repository core.DaemonStateRepository
}

func (runner daemonWorkflowRunner) projects(ctx context.Context, projects []core.DaemonProject, result *daemonForegroundLoopResult) error {
	for _, project := range projects {
		if project.Status != core.DaemonProjectStatusReady {
			continue
		}

		config, err := core.LoadWorkspaceConfig(project.Project.Path)
		if err != nil {
			return err
		}

		if core.WorkAuthority(config.Work) != core.WorkStateAuthorityTickets {
			continue
		}

		operations := newService(ctx, runner.env, commandProjectContext{ProjectDir: project.Project.Path, ProjectID: project.Project.ID}, runner.assets)
		if err := runner.workflows(ctx, operations, result); err != nil {
			return err
		}

		if result.StopRequested {
			break
		}
	}

	return nil
}

func (runner daemonWorkflowRunner) workflows(ctx context.Context, operations service, result *daemonForegroundLoopResult) error {
	activations := 0

	for offset := 0; offset >= 0; {
		page, err := operations.WorkflowStatusPage(ctx, offset, daemonWorkflowActivationLimit)
		if err != nil {
			return err
		}

		for _, status := range page.Items {
			attempted, err := runner.workflow(ctx, operations, status, result)
			if err != nil {
				return err
			}

			if attempted {
				activations++
			}

			if result.StopRequested || activations == daemonWorkflowActivationLimit {
				return nil
			}
		}

		offset = page.NextOffset
	}

	return nil
}

func (runner daemonWorkflowRunner) workflow(ctx context.Context, operations service, status WorkflowWorkStatus, result *daemonForegroundLoopResult) (bool, error) {
	stop, err := daemonStopRequested(ctx, runner.repository)
	if err != nil || stop {
		result.StopRequested = stop

		return false, err
	}

	progress, err := operations.ResumeRetainedWork(ctx, status.ID, status.Revision)
	result.WorkflowsObserved++

	result.WorkflowSteps += progress.Steps
	if progress.Steps == 0 {
		result.WorkflowsWaiting++
	}

	if progress.Attempted || err != nil {
		if writeErr := runner.progress(operations.projectContext.ProjectID, progress, err); writeErr != nil {
			return progress.Attempted, writeErr
		}
	}

	if err != nil && !daemonWorkflowWait(progress, err) {
		return progress.Attempted, err
	}

	stop, err = daemonStopRequested(ctx, runner.repository)
	result.StopRequested = stop

	return progress.Attempted, err
}

func (runner daemonWorkflowRunner) iteration(ctx context.Context, registry core.ProjectRegistry, index core.TicketIndex, result *daemonForegroundLoopResult) error {
	if err := runner.repository.UpdateDaemonHeartbeat(ctx, time.Now().UTC()); err != nil {
		return err
	}

	stop, err := daemonStopRequested(ctx, runner.repository)
	if err != nil || stop {
		result.StopRequested = stop

		return err
	}

	projects, err := core.LoadDaemonProjects(ctx, registry)
	if err != nil {
		return err
	}

	result.Iterations++

	result.ProjectsLoaded += len(projects)
	if err := refreshDaemonProjects(ctx, index, projects, result); err != nil {
		return err
	}

	return runner.projects(ctx, projects, result)
}

func daemonWorkflowWait(progress RetainedWorkProgress, failure error) bool {
	stopped := progress.Status.StopIntent != "" && (progress.Status.Status == "paused" || progress.Status.Status == "cancelled")

	return errors.Is(failure, core.ErrWorkflowOwned) || errors.Is(failure, core.ErrWorkflowCapacity) || errors.Is(failure, core.ErrWorkflowConflict) || stopped
}

func (runner daemonWorkflowRunner) progress(project string, progress RetainedWorkProgress, failure error) error {
	lines := []string{"project: " + project, "action: retained_continue", fmt.Sprintf("checkpoint_steps: %d", progress.Steps)}
	if failure != nil {
		lines = append(lines, "workflow_failure: "+failure.Error())
	}

	return printDaemonLines(runner.stdout, append(lines, workflowStatusLines(progress.Status)...))
}

// Heartbeats stay fresh during owned workers. A daemon stop yields between
// checkpoints; per-invocation pause/cancel uses the shared worker stop watcher.
func watchDaemonHeartbeat(ctx context.Context, repository core.DaemonStateRepository) (context.Context, func() error) {
	watched, cancel := context.WithCancel(ctx)
	finished := make(chan error, 1)

	go func() {
		ticker := time.NewTicker(time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-watched.Done():
				finished <- nil

				return
			case now := <-ticker.C:
				if err := repository.UpdateDaemonHeartbeat(watched, now.UTC()); err != nil {
					if watched.Err() != nil {
						err = nil
					}

					finished <- err

					cancel()

					return
				}
			}
		}
	}()

	return watched, func() error { cancel(); return <-finished }
}
