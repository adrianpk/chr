package app

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"
	"time"

	"forge.adrianpk.com/chr/internal/core"
)

// RunWorkflowManual supervises one harness and reconciles its exact Git report.
// The harness receives frozen public inputs; the host keeps claim credentials.
func (s service) RunWorkflowManual(ctx context.Context, id string, harness []string) (result core.WorkflowTaskResult, err error) {
	if err := validateWorkflowManualHarness(harness); err != nil {
		return result, err
	}

	claim, err := s.ClaimWorkflow(ctx, id)
	if err != nil {
		return result, err
	}
	defer func() { err = errors.Join(err, claim.Close()) }()

	if _, err := s.PrepareWorkflowManualTask(ctx, claim, harness); err != nil {
		return result, s.checkpointWorkflowManual(ctx, claim, err)
	}

	command, err := s.ExecuteWorkflowManualTask(ctx, claim)
	if err != nil {
		return result, err
	}

	candidate, err := parseWorkflowManualReport(command)
	if err != nil {
		return result, s.checkpointWorkflowManual(ctx, claim, err)
	}

	result, err = s.ReconcileWorkflowTask(ctx, claim, candidate)
	if err != nil {
		err = s.checkpointWorkflowManual(ctx, claim, err)
	}

	return result, err
}

func validateWorkflowManualHarness(harness []string) error {
	if len(harness) < 1 || len(harness) > 64 || len(strings.Join(harness, " ")) > 4096 {
		return fmt.Errorf("manual harness requires 1..64 arguments within 4096 bytes")
	}

	if harness[0] == "" {
		return fmt.Errorf("manual harness requires an executable")
	}

	for _, argument := range harness {
		if strings.ContainsRune(argument, '\x00') {
			return fmt.Errorf("manual harness arguments must not contain NUL")
		}
	}

	return nil
}

func (s service) checkpointWorkflowManual(ctx context.Context, claim *core.WorkflowClaim, failure error) error {
	cleanup, cancel := context.WithTimeout(context.WithoutCancel(ctx), 5*time.Second)
	defer cancel()

	return errors.Join(failure, s.CheckpointWorkflow(cleanup, claim))
}

func parseWorkflowManualReport(command core.WorkflowCommandResult) (core.WorkflowManualCandidate, error) {
	if command.Truncated || len(command.Stdout) > 4096 {
		return core.WorkflowManualCandidate{}, fmt.Errorf("manual candidate report exceeds 4096 bytes")
	}

	decoder := json.NewDecoder(bytes.NewBufferString(command.Stdout))
	decoder.DisallowUnknownFields()

	candidate := core.WorkflowManualCandidate{}
	if err := decoder.Decode(&candidate); err != nil {
		return candidate, fmt.Errorf("decode manual candidate: %w", err)
	}

	if err := decoder.Decode(new(any)); !errors.Is(err, io.EOF) {
		return candidate, fmt.Errorf("manual harness must return one candidate object")
	}

	return candidate, nil
}
