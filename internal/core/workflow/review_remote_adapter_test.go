package workflow

import (
	"context"
	"errors"
	"strings"
)

type fakeRemoteReviewRunner struct {
	outputs     map[string][]byte
	outputErrs  map[string]error
	inputs      map[string][]byte
	inputErrors map[string]error
}

func newFakeRemoteReviewRunner() *fakeRemoteReviewRunner {
	return &fakeRemoteReviewRunner{
		outputs:     map[string][]byte{},
		outputErrs:  map[string]error{},
		inputs:      map[string][]byte{},
		inputErrors: map[string]error{},
	}
}

func (runner *fakeRemoteReviewRunner) Output(_ context.Context, command string, args ...string) ([]byte, error) {
	key := remoteReviewCommandKey(command, args...)
	if err := runner.outputErrs[key]; err != nil {
		return nil, err
	}

	output, ok := runner.outputs[key]
	if !ok {
		return nil, errors.New("missing output for " + key)
	}

	return output, nil
}

func (runner *fakeRemoteReviewRunner) Input(_ context.Context, command string, input []byte, args ...string) ([]byte, error) {
	key := remoteReviewCommandKey(command, args...)
	if err := runner.inputErrors[key]; err != nil {
		return nil, err
	}

	runner.inputs[key] = input

	return []byte(`{}`), nil
}

func remoteReviewCommandKey(command string, args ...string) string {
	return strings.Join(append([]string{command}, args...), " ")
}
