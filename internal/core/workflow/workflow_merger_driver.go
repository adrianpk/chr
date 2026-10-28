package workflow

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"os"

	workflowmerge "forge.adrianpk.com/chr/internal/core/workflow/merge"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

type workflowMergerDriverResult struct {
	command workflowstate.CommandResult
	data    json.RawMessage
	usage   *workflowstate.MergerUsage
	checks  []workflowstate.CommandResult
	err     error
}

func runWorkflowMergerDriver(ctx context.Context, input workflowMergerDriverInput) workflowMergerDriverResult {
	claim := input.Claim

	run, err := claim.store.ReadWorkflow(ctx, claim.InvocationID)
	if err != nil {
		return workflowMergerDriverResult{err: err}
	}

	index := len(run.Attempts[len(run.Attempts)-1].Processes)

	commandOutput, err := claim.store.workflowProcessOutput(claim, index)
	if err != nil {
		return workflowMergerDriverResult{err: err}
	}

	defer func() { _ = commandOutput.Stdout.File.Close(); _ = commandOutput.Stderr.File.Close() }() //nolint:errcheck

	gate, permit, err := os.Pipe()
	if err != nil {
		return workflowMergerDriverResult{err: err}
	}
	defer func() { _ = gate.Close(); _ = permit.Close() }() //nolint:errcheck

	return executeWorkflowMergerDriver(ctx, input, workflowMergerDriverIO{Gate: gate, Permit: permit, Index: index, Output: commandOutput})
}

func executeWorkflowMergerDriver(ctx context.Context, input workflowMergerDriverInput, execution workflowMergerDriverIO) workflowMergerDriverResult {
	claim, prepared, tools, call := input.Claim, input.Prepared, input.Tools, input.Call
	gate, permit, index := execution.Gate, execution.Permit, execution.Index
	stdout, stderr := execution.Output.Stdout, execution.Output.Stderr

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	supervisedProcess := workflowSupervisedCommand(ctx, claim, prepared, prepared.input.Worker, gate)

	process, guard := supervisedProcess.Process, supervisedProcess.Guard

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
		command, waitErr := runWorkflowSupervisedProcess(ctx, claim, prepared, workflowProcessExecution{Command: prepared.input.Worker, Process: process, Permit: permit, Index: index, Output: workflowCommandOutput{Stdout: stdout, Stderr: stderr}})
		_ = writer.Close() //nolint:errcheck

		finished <- workflowMergerDriverResult{command: command, err: waitErr}
	}()

	broker, brokerErr := exchangeWorkflowMerger(ctx, tools, call, reader, stdin)
	if brokerErr != nil {
		cancel()

		_ = reader.Close() //nolint:errcheck
	}

	result := <-finished
	result.data, result.usage, result.err = broker.Proposal, broker.Usage, errors.Join(result.err, brokerErr)

	return result
}

func exchangeWorkflowMerger(ctx context.Context, tools *workflowmerge.Tools, call workflowmerge.Call, reader io.Reader, writer io.Writer) (workflowmerge.BrokerResult, error) {
	if err := json.NewEncoder(writer).Encode(call); err != nil {
		return workflowmerge.BrokerResult{}, err
	}

	return workflowmerge.ServeBroker(ctx, tools, reader, writer)
}

type workflowMergerDriverInput struct {
	Claim    *Claim
	Prepared workflowPreparedTask
	Tools    *workflowmerge.Tools
	Call     workflowmerge.Call
}

type workflowMergerDriverIO struct {
	Gate   *os.File
	Permit *os.File
	Index  int
	Output workflowCommandOutput
}
