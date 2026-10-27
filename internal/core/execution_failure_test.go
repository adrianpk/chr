package core

import (
	"fmt"
	"testing"

	executioncore "forge.adrianpk.com/chr/internal/core/execution"
)

func TestExecutionFailurePrefixesErrorWithKind(t *testing.T) {
	err := executioncore.NewExecutionFailure(executioncore.ExecutionFailureHarness, fmt.Errorf("docker failed"))

	if err.Error() != "harness: docker failed" {
		t.Fatalf("Error = %q, want harness prefix", err.Error())
	}
}

func TestExecutionFailureKindFromReasonParsesKnownKind(t *testing.T) {
	kind := executioncore.ExecutionFailureKindFromReason("worker: codex failed")

	if kind != executioncore.ExecutionFailureWorker {
		t.Fatalf("kind = %q, want worker", kind)
	}
}
