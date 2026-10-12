package core

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"os"
)

type workflowMergerDriverResult struct {
	command WorkflowCommandResult
	data    json.RawMessage
	usage   *WorkflowMergerUsage
	checks  []WorkflowCommandResult
	err     error
}

func runWorkflowMergerDriver(ctx context.Context, claim *WorkflowClaim, prepared workflowPreparedTask, tools *workflowMergerTools, call workflowMergerCall) workflowMergerDriverResult {
	run, err := claim.store.ReadWorkflow(ctx, claim.InvocationID)
	if err != nil {
		return workflowMergerDriverResult{err: err}
	}

	index := len(run.Attempts[len(run.Attempts)-1].Processes)

	stdout, stderr, err := claim.store.workflowProcessOutput(claim, index)
	if err != nil {
		return workflowMergerDriverResult{err: err}
	}

	defer func() { _ = stdout.file.Close(); _ = stderr.file.Close() }() //nolint:errcheck

	gate, permit, err := os.Pipe()
	if err != nil {
		return workflowMergerDriverResult{err: err}
	}
	defer func() { _ = gate.Close(); _ = permit.Close() }() //nolint:errcheck

	return executeWorkflowMergerDriver(ctx, claim, prepared, tools, call, gate, permit, index, stdout, stderr)
}

func executeWorkflowMergerDriver(ctx context.Context, claim *WorkflowClaim, prepared workflowPreparedTask, tools *workflowMergerTools, call workflowMergerCall, gate, permit *os.File, index int, stdout, stderr *workflowCommandOutput) workflowMergerDriverResult {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	process, guard := workflowSupervisedCommand(ctx, claim, prepared, prepared.input.Worker, gate)

	stdin, err := process.StdinPipe()
	if err != nil {
		return workflowMergerDriverResult{err: err}
	}

	defer func() { _ = stdin.Close() }() //nolint:errcheck

	reader, writer := io.Pipe()
	defer func() { _ = reader.Close(); _ = writer.Close() }() //nolint:errcheck

	process.Stdout, process.Stderr = io.MultiWriter(stdout, writer), stderr
	if err := process.Start(); err != nil {
		return workflowMergerDriverResult{err: err}
	}

	if err := guard.remember(process.Process.Pid); err != nil {
		return workflowMergerDriverResult{err: errors.Join(err, permit.Close(), process.Wait())}
	}

	finished := make(chan workflowMergerDriverResult, 1)

	go func() {
		command, waitErr := runWorkflowSupervisedProcess(ctx, claim, prepared, prepared.input.Worker, process, permit, index, stdout, stderr)
		_ = writer.Close() //nolint:errcheck

		finished <- workflowMergerDriverResult{command: command, err: waitErr}
	}()

	proposal, usage, brokerErr := exchangeWorkflowMerger(ctx, tools, call, reader, stdin)
	if brokerErr != nil {
		cancel()

		_ = reader.Close() //nolint:errcheck
	}

	result := <-finished
	result.data, result.usage, result.err = proposal, usage, errors.Join(result.err, brokerErr)

	return result
}

func exchangeWorkflowMerger(ctx context.Context, tools *workflowMergerTools, call workflowMergerCall, reader io.Reader, writer io.Writer) (json.RawMessage, *WorkflowMergerUsage, error) {
	if err := json.NewEncoder(writer).Encode(call); err != nil {
		return nil, nil, err
	}

	return serveWorkflowMergerBroker(ctx, tools, reader, writer)
}
