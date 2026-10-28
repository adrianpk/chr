package execution

import (
	"errors"
	"fmt"
	"strings"
)

// FailureKind defines a package boundary type.
type FailureKind string

const (
	// FailurePreflight defines a package boundary constant.
	FailurePreflight FailureKind = "preflight"
	// FailureRuntime defines a package boundary constant.
	FailureRuntime FailureKind = "runtime"
	// FailureHarness defines a package boundary constant.
	FailureHarness FailureKind = "harness"
	// FailureWorker defines a package boundary constant.
	FailureWorker FailureKind = "worker"
	// FailureValidation defines a package boundary constant.
	FailureValidation FailureKind = "validation"
	// FailureResultCapture defines a package boundary constant.
	FailureResultCapture FailureKind = "result_capture"
)

// Error describes a package boundary value.
type Error struct {
	Kind FailureKind
	Err  error
}

// NewError constructs the corresponding package service or value.
func NewError(kind FailureKind, err error) Error {
	return Error{
		Kind: kind,
		Err:  err,
	}
}

// Error executes the corresponding package operation.
func (e Error) Error() string {
	if e.Err == nil {
		return string(e.Kind)
	}

	return fmt.Sprintf("%s: %v", e.Kind, e.Err)
}

// Unwrap executes the corresponding package operation.
func (e Error) Unwrap() error {
	return e.Err
}

// ClassifyFailure executes the corresponding package operation.
func ClassifyFailure(err error) FailureKind {
	var failure Error
	if errors.As(err, &failure) {
		return failure.Kind
	}

	return FailurePreflight
}

// FailureKindFromReason executes the corresponding package operation.
func FailureKindFromReason(reason string) FailureKind {
	kind, _, found := strings.Cut(reason, ":")
	if !found {
		return ""
	}

	switch FailureKind(kind) {
	case FailurePreflight,
		FailureRuntime,
		FailureHarness,
		FailureWorker,
		FailureValidation,
		FailureResultCapture:
		return FailureKind(kind)
	default:
		return ""
	}
}
