package chc

import (
	"fmt"
	"testing"
)

func TestExecutionFailurePrefixesErrorWithKind(t *testing.T) {
	err := NewExecutionFailure(ExecutionFailureHarness, fmt.Errorf("docker failed"))

	if err.Error() != "harness: docker failed" {
		t.Fatalf("Error = %q, want harness prefix", err.Error())
	}
}

func TestExecutionFailureKindFromReasonParsesKnownKind(t *testing.T) {
	kind := ExecutionFailureKindFromReason("result_capture: missing result")

	if kind != ExecutionFailureResultCapture {
		t.Fatalf("kind = %q, want result_capture", kind)
	}
}
