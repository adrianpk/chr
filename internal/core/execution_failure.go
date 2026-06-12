package core

import (
	"errors"
	"fmt"
	"strings"
)

// ExecutionFailureKind defines a package boundary type.
type ExecutionFailureKind string

const (
	// ExecutionFailurePreflight defines a package boundary constant.
	ExecutionFailurePreflight ExecutionFailureKind = "preflight"
	// ExecutionFailureRuntime defines a package boundary constant.
	ExecutionFailureRuntime ExecutionFailureKind = "runtime"
	// ExecutionFailureHarness defines a package boundary constant.
	ExecutionFailureHarness ExecutionFailureKind = "harness"
	// ExecutionFailureWorker defines a package boundary constant.
	ExecutionFailureWorker ExecutionFailureKind = "worker"
	// ExecutionFailureValidation defines a package boundary constant.
	ExecutionFailureValidation ExecutionFailureKind = "validation"
	// ExecutionFailureResultCapture defines a package boundary constant.
	ExecutionFailureResultCapture ExecutionFailureKind = "result_capture"
)

// ExecutionFailure describes a package boundary value.
type ExecutionFailure struct {
	Kind ExecutionFailureKind
	Err  error
}

// NewExecutionFailure constructs the corresponding package service or value.
func NewExecutionFailure(kind ExecutionFailureKind, err error) ExecutionFailure {
	return ExecutionFailure{
		Kind: kind,
		Err:  err,
	}
}

// Error executes the corresponding package operation.
func (e ExecutionFailure) Error() string {
	if e.Err == nil {
		return string(e.Kind)
	}

	return fmt.Sprintf("%s: %v", e.Kind, e.Err)
}

// Unwrap executes the corresponding package operation.
func (e ExecutionFailure) Unwrap() error {
	return e.Err
}

// ClassifyExecutionFailure executes the corresponding package operation.
func ClassifyExecutionFailure(err error) ExecutionFailureKind {
	var failure ExecutionFailure
	if errors.As(err, &failure) {
		return failure.Kind
	}

	return ExecutionFailurePreflight
}

// ExecutionFailureKindFromReason executes the corresponding package operation.
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
