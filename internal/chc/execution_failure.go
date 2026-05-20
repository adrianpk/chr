package chc

import (
	"errors"
	"fmt"
	"strings"
)

type ExecutionFailureKind string

const (
	ExecutionFailurePreflight     ExecutionFailureKind = "preflight"
	ExecutionFailureRuntime       ExecutionFailureKind = "runtime"
	ExecutionFailureHarness       ExecutionFailureKind = "harness"
	ExecutionFailureWorker        ExecutionFailureKind = "worker"
	ExecutionFailureValidation    ExecutionFailureKind = "validation"
	ExecutionFailureResultCapture ExecutionFailureKind = "result_capture"
)

type ExecutionFailure struct {
	Kind ExecutionFailureKind
	Err  error
}

func NewExecutionFailure(kind ExecutionFailureKind, err error) ExecutionFailure {
	return ExecutionFailure{
		Kind: kind,
		Err:  err,
	}
}

func (e ExecutionFailure) Error() string {
	if e.Err == nil {
		return string(e.Kind)
	}

	return fmt.Sprintf("%s: %v", e.Kind, e.Err)
}

func (e ExecutionFailure) Unwrap() error {
	return e.Err
}

func ClassifyExecutionFailure(err error) ExecutionFailureKind {
	var failure ExecutionFailure
	if errors.As(err, &failure) {
		return failure.Kind
	}

	return ExecutionFailurePreflight
}

func ExecutionFailureKindFromReason(reason string) ExecutionFailureKind {
	kind, _, found := strings.Cut(reason, ":")
	if !found {
		return ""
	}

	switch ExecutionFailureKind(kind) {
	case ExecutionFailurePreflight,
		ExecutionFailureRuntime,
		ExecutionFailureHarness,
		ExecutionFailureWorker,
		ExecutionFailureValidation,
		ExecutionFailureResultCapture:
		return ExecutionFailureKind(kind)
	default:
		return ""
	}
}
