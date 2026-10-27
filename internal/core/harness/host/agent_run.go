package host

import (
	"context"
	"fmt"
	"path/filepath"

	harnesstrace "forge.adrianpk.com/chr/internal/core/harness/trace"
	workspacecore "forge.adrianpk.com/chr/internal/core/workspace"
	"forge.adrianpk.com/chr/internal/harness"
)

// RunAgent applies host-owned artifact layout and tracing to an injected runner.
func RunAgent(ctx context.Context, runner harness.AgentRunner, request harness.AgentRunRequest) (harness.AgentRunResult, error) {
	if runner == nil {
		return harness.AgentRunResult{}, fmt.Errorf("agent runner is required")
	}

	request, err := prepareBoundGenerationRequest(ctx, request)
	if err != nil {
		return harness.AgentRunResult{}, err
	}

	if err := harness.ValidateAgentRunRequest(request); err != nil {
		return harness.AgentRunResult{}, err
	}

	request.ID = GenerationRequestID(ctx, request.ID)

	if request.RunDir == "" {
		request.RunDir = workspacecore.UserPath(request.ProjectDir, "tmp", "agent-runs", request.ID)
	}

	metadata, err := startGenerationRun(ctx, request)
	if err != nil {
		return harness.AgentRunResult{}, err
	}

	request.OnStart = func(result harness.AgentRunResult) { recordRequest(ctx, request, result) }
	result, err := runner.RunAgent(ctx, request)
	err = finishGenerationRun(request, result, metadata, err)

	if !request.DeferResponseTrace {
		status := "ready"
		if err != nil {
			status = "error"
		}

		RecordResponse(ctx, request, result, status, err)
	}

	return result, err
}

func recordRequest(ctx context.Context, request harness.AgentRunRequest, result harness.AgentRunResult) {
	harnesstrace.Record(ctx, harnesstrace.LoadOptions{ProjectDir: request.ProjectDir}, map[string]any{
		"event": "llm_request", "llm_call_id": request.ID, "request_id": request.ID,
		"action": request.Purpose, "provider": result.Provider, "model": result.Model,
		"prompt_path": filepath.Join(request.RunDir, "prompt.md"), "output_path": result.OutputPath,
		"request_payload_bytes": len(request.Prompt), "request_payload": request.Prompt,
	})
}

// RecordResponse records an agent response trace event.
func RecordResponse(ctx context.Context, request harness.AgentRunRequest, result harness.AgentRunResult, status string, runErr error) {
	request.ID = GenerationRequestID(ctx, request.ID)

	event := map[string]any{
		"event":                  "llm_response",
		"llm_call_id":            result.ID,
		"request_id":             request.ID,
		"action":                 request.Purpose,
		"provider":               result.Provider,
		"model":                  result.Model,
		"latency_ms":             result.Latency.Milliseconds(),
		"status":                 status,
		"output_payload_bytes":   len(result.Output),
		"response_payload_bytes": len(result.Stdout) + len(result.Stderr),
		"output_payload":         string(result.Output),
		"response_payload":       string(result.Output),
		"stdout":                 string(result.Stdout),
		"stderr":                 string(result.Stderr),
	}
	if runErr != nil {
		event["error"] = runErr.Error()
	}

	harnesstrace.Record(ctx, harnesstrace.LoadOptions{ProjectDir: request.ProjectDir}, event)
}
