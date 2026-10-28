package merge

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"

	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

// WorkflowMergerCall binds the immutable input and retained evidence to a driver call.
type WorkflowMergerCall struct {
	Protocol string                    `json:"protocol"`
	Input    workflowstate.MergerInput `json:"input"`
	Evidence map[string]string         `json:"evidence"`
}

// ServeWorkflowMergerBroker exchanges bounded tool messages and one terminal proposal.
func ServeWorkflowMergerBroker(ctx context.Context, tools *WorkflowMergerTools, reader io.Reader, writer io.Writer) (proposal json.RawMessage, usage *workflowstate.MergerUsage, err error) {
	scanner := bufio.NewScanner(reader)
	scanner.Buffer(make([]byte, 4096), workflowstate.FileLimit+1)

	encoder := json.NewEncoder(writer)

	total, messages := 0, 0
	for scanner.Scan() {
		total, messages = total+len(scanner.Bytes()), messages+1
		if total > 8*workflowstate.FileLimit || messages > workflowstate.PageLimit || proposal != nil {
			return nil, nil, workflowstate.ErrConflict
		}

		request, decodeErr := decodeWorkflowMergerTool(scanner.Bytes())
		if decodeErr != nil {
			return nil, nil, decodeErr
		}

		if request.Operation != "result" {
			if err := writeWorkflowMergerToolResponse(ctx, tools, encoder, request); err != nil {
				return nil, nil, err
			}

			continue
		}

		proposal, usage, err = writeWorkflowMergerTerminal(encoder, request)
		if err != nil {
			return nil, nil, err
		}
	}

	if err := workflowMergerBrokerComplete(scanner, proposal, ctx); err != nil {
		return nil, nil, err
	}

	return proposal, usage, nil
}

func decodeWorkflowMergerTool(data []byte) (workflowMergerToolRequest, error) {
	var request workflowMergerToolRequest
	if err := CheckWorkflowMergerJSON(data); err != nil {
		return request, err
	}

	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()

	err := decoder.Decode(&request)
	if err != nil {
		err = fmt.Errorf("%w: invalid merger tool message", workflowstate.ErrConflict)
	}

	return request, err
}

func workflowMergerTerminal(request workflowMergerToolRequest) (json.RawMessage, *workflowstate.MergerUsage, error) {
	if request.Path != "" || request.Content != "" || request.Check != nil || len(request.Result) == 0 {
		return nil, nil, workflowstate.ErrConflict
	}

	return request.Result, request.Usage, nil
}

func workflowMergerBrokerComplete(scanner *bufio.Scanner, proposal json.RawMessage, ctx context.Context) error {
	if scanner.Err() != nil || proposal == nil || ctx.Err() != nil {
		return fmt.Errorf("%w: merger driver did not finish a bounded proposal", workflowstate.ErrConflict)
	}

	return nil
}

func writeWorkflowMergerToolResponse(ctx context.Context, tools *WorkflowMergerTools, encoder *json.Encoder, request workflowMergerToolRequest) error {
	response, err := tools.apply(ctx, request)
	if err != nil {
		response.Error = err.Error()
		if len(response.Error) > 4096 {
			response.Error = response.Error[:4096]
		}
	}

	return encoder.Encode(response)
}

func writeWorkflowMergerTerminal(encoder *json.Encoder, request workflowMergerToolRequest) (json.RawMessage, *workflowstate.MergerUsage, error) {
	proposal, usage, err := workflowMergerTerminal(request)
	if err != nil {
		return nil, nil, err
	}

	return proposal, usage, encoder.Encode(workflowMergerToolResponse{})
}
