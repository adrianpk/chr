package workspace

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
	kind := ExecutionFailureKindFromReason("worker: codex failed")

	if kind != ExecutionFailureWorker {
		t.Fatalf("kind = %q, want worker", kind)
	}
}
