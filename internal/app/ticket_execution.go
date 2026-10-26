package app

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io/fs"
	"strings"
	"time"

	"forge.adrianpk.com/chr/internal/core"
	"forge.adrianpk.com/chr/internal/harness/codex"
)

type ticketExecutionEvents struct {
	LoopID string
	Sink   core.AgentEventSink
}

var newTicketStartHarnessExecutor = func() core.HarnessExecutor {
	return core.OSHarnessExecutor{}
}

func newPendingExecutionRecord(projectID string, ticket core.IndexedTicket) core.ExecutionRecord {
	now := time.Now().UTC()

	return core.ExecutionRecord{
		ID:         executionID(projectID, ticket.Key, now),
		ProjectID:  projectID,
		TicketKey:  ticket.Key,
		TicketPath: ticket.Path,
		Status:     core.ExecutionStatusPending,
		CreatedAt:  now,
		UpdatedAt:  now,
	}
}

func ticketStartArtifactTitle(ticket core.IndexedTicket) string {
	if strings.TrimSpace(ticket.Title) == "" {
		return ticket.Key
	}

	return ticket.Key + " " + ticket.Title
}

func runTicketStartExecution(
	ctx context.Context,
	env core.UserEnvironment,
	record core.ExecutionRecord,
	result core.RuntimeValidationResult,
	events ticketExecutionEvents,
	assetsFS fs.FS,
) (core.ExecutionRecord, error) {
	repository, err := core.NewExecutionRepository(ctx, env, assetsFS)
	if err != nil {
		return core.ExecutionRecord{}, err
	}

	defer func() {
		//nolint:errcheck // Command result already comes from the main operation.
		_ = repository.Close()
	}()

	err = repository.CreateExecution(ctx, record)
	if err != nil {
		return core.ExecutionRecord{}, err
	}

	emitBranchPreparedAgentEvent(ctx, events, record)
	emitExecutionCreatedAgentEvent(ctx, events, record)

	err = core.MaterializeWorkerContextForExecution(ctx, record, result.Plan)
	if err != nil {
		return core.ExecutionRecord{}, core.NewExecutionFailure(core.ExecutionFailurePreflight, err)
	}

	invocation, err := core.NewHarnessInvocation(core.HarnessInvocationOptions{
		Worker:      codex.Runner{},
		ExecutionID: record.ID,
		TicketKey:   record.TicketKey,
		TicketPath:  record.TicketPath,
		Plan:        result.Plan,
	})
	if err != nil {
		return core.ExecutionRecord{}, err
	}

	runner := core.NewExecutionRunner(repository, newTicketStartHarnessExecutor())
	runner.EventSink = events.Sink

	executed, runErr := runner.Run(ctx, core.RunExecutionOptions{
		Record:     record,
		Invocation: invocation,
		LoopID:     events.LoopID,
	})
	if err := refreshExecutedTicketArtifactMetadata(executed, runErr); err != nil {
		return executed, err
	}

	return executed, runErr
}

func refreshExecutedTicketArtifactMetadata(executed core.ExecutionRecord, runErr error) error {
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
