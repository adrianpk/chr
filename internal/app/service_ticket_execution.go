package app

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	"forge.adrianpk.com/chr/internal/core"
	executioncore "forge.adrianpk.com/chr/internal/core/execution"
	harnessruntime "forge.adrianpk.com/chr/internal/core/harness/runtime"
	ticketcore "forge.adrianpk.com/chr/internal/core/ticket"
	"forge.adrianpk.com/chr/internal/harness/codex"
)

type ticketExecutionEvents struct {
	LoopID string
	Sink   executioncore.AgentEventSink
}

var newTicketStartHarnessExecutor = func() executioncore.HarnessExecutor {
	return executioncore.OSHarnessExecutor{}
}

func newPendingExecutionRecord(projectID string, ticket ticketcore.Indexed) executioncore.Record {
	now := time.Now().UTC()

	return executioncore.Record{
		ID:         executionID(projectID, ticket.Key, now),
		ProjectID:  projectID,
		TicketKey:  ticket.Key,
		TicketPath: ticket.Path,
		Status:     executioncore.StatusPending,
		CreatedAt:  now,
		UpdatedAt:  now,
	}
}

func ticketStartArtifactTitle(ticket ticketcore.Indexed) string {
	if strings.TrimSpace(ticket.Title) == "" {
		return ticket.Key
	}

	return ticket.Key + " " + ticket.Title
}

func (s service) runTicketStartExecution(ctx context.Context, input ticketExecutionInput) (executioncore.Record, error) {
	repository, err := executioncore.NewRepository(ctx, s.env, s.assetsFS)
	if err != nil {
		return executioncore.Record{}, err
	}

	defer func() {
		//nolint:errcheck // Command result already comes from the main operation.
		_ = repository.Close()
	}()

	err = repository.CreateExecution(ctx, input.Record)
	if err != nil {
		return executioncore.Record{}, err
	}

	emitBranchPreparedAgentEvent(ctx, input.Events, input.Record)
	emitExecutionCreatedAgentEvent(ctx, input.Events, input.Record)

	err = core.MaterializeWorkerContextForExecution(ctx, input.Record, input.Runtime.Plan)
	if err != nil {
		return executioncore.Record{}, executioncore.NewFailure(executioncore.FailurePreflight, err)
	}

	invocation, err := harnessruntime.NewInvocation(harnessruntime.InvocationOptions{
		Worker:      codex.Runner{},
		ExecutionID: input.Record.ID,
		TicketKey:   input.Record.TicketKey,
		TicketPath:  input.Record.TicketPath,
		Plan:        input.Runtime.Plan,
	})
	if err != nil {
		return executioncore.Record{}, err
	}

	runner := executioncore.NewRunner(repository, newTicketStartHarnessExecutor())
	runner.EventSink = input.Events.Sink

	executed, runErr := runner.Run(ctx, executioncore.RunOptions{
		Record:     input.Record,
		Invocation: invocation,
		LoopID:     input.Events.LoopID,
	})
	if err := refreshExecutedTicketArtifactMetadata(executed, runErr); err != nil {
		return executed, err
	}

	return executed, runErr
}

func refreshExecutedTicketArtifactMetadata(executed executioncore.Record, runErr error) error {
	err := refreshTicketArtifactMetadata(executed)
	if err == nil {
		return nil
	}

	if runErr != nil {
		return fmt.Errorf("%w; refresh shared artifact metadata: %w", runErr, err)
	}

	return err
}

func executionID(projectID, ticketKey string, createdAt time.Time) string {
	sum := sha256.Sum256([]byte(projectID + "\x00" + ticketKey + "\x00" + createdAt.Format(time.RFC3339Nano)))

	return "exec-" + hex.EncodeToString(sum[:8])
}
