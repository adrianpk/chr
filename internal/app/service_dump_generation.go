package app

import (
	"context"
	"encoding/json"
	"net/http"

	harnesshost "forge.adrianpk.com/chr/internal/core/harness/host"
	"forge.adrianpk.com/chr/internal/harness"
)

func (p openAIDumpCurationProvider) executeBoundDumpAPI(ctx context.Context, request DumpCurationRequest, httpRequest *http.Request, trace dumpCurationOpenAITrace) (dumpCurationOpenAIExecutionResult, error) {
	var result dumpCurationOpenAIExecutionResult

	schema, err := json.Marshal(dumpCurationJSONSchema())
	if err != nil {
		return result, err
	}

	call := p.dumpAPIRunRequest(request, trace)
	call.OutputSchema = schema
	_, err = harnesshost.RunAPIGeneration(ctx, call, func(context.Context) ([]byte, error) {
		var err error

		result, err = p.executeOpenAIRequest(httpRequest, trace.StartedAt)

		return result.Output, err
	})

	return result, err
}

func (p openAIDumpCurationProvider) dumpAPIRunRequest(request DumpCurationRequest, trace dumpCurationOpenAITrace) harness.AgentRunRequest {
	return harness.AgentRunRequest{ID: dumpCurationRequestID(request), Purpose: "dump_curation", ProjectDir: request.ProjectDir, Prompt: string(trace.RequestBody), OutputName: "output.json"}
}
