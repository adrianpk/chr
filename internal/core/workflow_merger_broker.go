package core

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"unicode/utf8"
)

type workflowMergerCall struct {
	Protocol string              `json:"protocol"`
	Input    WorkflowMergerInput `json:"input"`
	Evidence map[string]string   `json:"evidence"`
}

func (s *WorkflowStore) workflowMergerCall(input WorkflowMergerInput) (workflowMergerCall, error) {
	call := workflowMergerCall{Protocol: workflowMergerBrokerProtocol, Input: input, Evidence: map[string]string{}}
	total := 0

	for _, ref := range input.evidenceRefs() {
		if _, present := call.Evidence[ref.Path]; present {
			continue
		}

		data, err := s.readWorkflowReviewImage(ref.Path, ref.Hash)
		if err != nil {
			return call, err
		}

		total += len(data)
		if total >= WorkflowFileLimit || !utf8.Valid(data) {
			return call, ErrWorkflowConflict
		}

		call.Evidence[ref.Path] = string(data)
	}

	return call, nil
}

func serveWorkflowMergerBroker(ctx context.Context, tools *workflowMergerTools, reader io.Reader, writer io.Writer) (proposal json.RawMessage, usage *WorkflowMergerUsage, err error) {
	scanner := bufio.NewScanner(reader)
	scanner.Buffer(make([]byte, 4096), WorkflowFileLimit+1)

	encoder := json.NewEncoder(writer)

	total, messages := 0, 0
	for scanner.Scan() {
		total, messages = total+len(scanner.Bytes()), messages+1
		if total > 8*WorkflowFileLimit || messages > WorkflowPageLimit || proposal != nil {
			return nil, nil, ErrWorkflowConflict
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
	if err := checkWorkflowMergerJSON(data); err != nil {
		return request, err
	}

	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()

	err := decoder.Decode(&request)
	if err != nil {
		err = fmt.Errorf("%w: invalid merger tool message", ErrWorkflowConflict)
	}

	return request, err
}

func workflowMergerTerminal(request workflowMergerToolRequest) (json.RawMessage, *WorkflowMergerUsage, error) {
	if request.Path != "" || request.Content != "" || request.Check != nil || len(request.Result) == 0 {
		return nil, nil, ErrWorkflowConflict
	}

	return request.Result, request.Usage, nil
}

func workflowMergerBrokerComplete(scanner *bufio.Scanner, proposal json.RawMessage, ctx context.Context) error {
	if scanner.Err() != nil || proposal == nil || ctx.Err() != nil {
		return fmt.Errorf("%w: merger driver did not finish a bounded proposal", ErrWorkflowConflict)
	}

	return nil
}

func writeWorkflowMergerToolResponse(ctx context.Context, tools *workflowMergerTools, encoder *json.Encoder, request workflowMergerToolRequest) error {
	response, err := tools.apply(ctx, request)
	if err != nil {
		response.Error = workflowFailureText(err)
	}

	return encoder.Encode(response)
}

func writeWorkflowMergerTerminal(encoder *json.Encoder, request workflowMergerToolRequest) (json.RawMessage, *WorkflowMergerUsage, error) {
	proposal, usage, err := workflowMergerTerminal(request)
	if err != nil {
		return nil, nil, err
	}

	return proposal, usage, encoder.Encode(workflowMergerToolResponse{})
}
