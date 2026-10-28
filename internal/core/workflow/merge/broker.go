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

// Call binds the immutable input and retained evidence to a driver call.
type Call struct {
	Protocol string                    `json:"protocol"`
	Input    workflowstate.MergerInput `json:"input"`
	Evidence map[string]string         `json:"evidence"`
}

// BrokerResult retains a raw proposal and its reported usage for host validation.
type BrokerResult struct {
	Proposal json.RawMessage
	Usage    *workflowstate.MergerUsage
}

// ServeBroker exchanges bounded tool messages and one terminal proposal.
func ServeBroker(ctx context.Context, tools *Tools, reader io.Reader, writer io.Writer) (result BrokerResult, err error) {
	scanner := bufio.NewScanner(reader)
	scanner.Buffer(make([]byte, 4096), workflowstate.FileLimit+1)

	encoder := json.NewEncoder(writer)

	total, messages := 0, 0
	for scanner.Scan() {
		total, messages = total+len(scanner.Bytes()), messages+1
		if total > 8*workflowstate.FileLimit || messages > workflowstate.PageLimit || result.Proposal != nil {
			return BrokerResult{}, workflowstate.ErrConflict
		}

		request, decodeErr := decodeTool(scanner.Bytes())
		if decodeErr != nil {
			return BrokerResult{}, decodeErr
		}

		if request.Operation != "result" {
			if err := writeToolResponse(ctx, tools, encoder, request); err != nil {
				return BrokerResult{}, err
			}

			continue
		}

		result, err = writeTerminal(encoder, request)
		if err != nil {
			return BrokerResult{}, err
		}
	}

	if err := brokerComplete(scanner, result.Proposal, ctx); err != nil {
		return BrokerResult{}, err
	}

	return result, nil
}

func decodeTool(data []byte) (toolRequest, error) {
	var request toolRequest
	if err := CheckJSON(data); err != nil {
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

func terminal(request toolRequest) (BrokerResult, error) {
	if request.Path != "" || request.Content != "" || request.Check != nil || len(request.Result) == 0 {
		return BrokerResult{}, workflowstate.ErrConflict
	}

	return BrokerResult{Proposal: request.Result, Usage: request.Usage}, nil
}

func brokerComplete(scanner *bufio.Scanner, proposal json.RawMessage, ctx context.Context) error {
	if scanner.Err() != nil || proposal == nil || ctx.Err() != nil {
		return fmt.Errorf("%w: merger driver did not finish a bounded proposal", workflowstate.ErrConflict)
	}

	return nil
}

func writeToolResponse(ctx context.Context, tools *Tools, encoder *json.Encoder, request toolRequest) error {
	response, err := tools.apply(ctx, request)
	if err != nil {
		response.Error = err.Error()
		if len(response.Error) > 4096 {
			response.Error = response.Error[:4096]
		}
	}

	return encoder.Encode(response)
}

func writeTerminal(encoder *json.Encoder, request toolRequest) (BrokerResult, error) {
	result, err := terminal(request)
	if err != nil {
		return BrokerResult{}, err
	}

	return result, encoder.Encode(toolResponse{})
}
