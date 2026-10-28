package execution_test

import (
	"fmt"
	"testing"

	executioncore "forge.adrianpk.com/chr/internal/core/execution"
)

func TestExecutionFailurePrefixesErrorWithKind(t *testing.T) {
	err := executioncore.NewError(executioncore.FailureHarness, fmt.Errorf("docker failed"))

	if err.Error() != "harness: docker failed" {
		t.Fatalf("Error = %q, want harness prefix", err.Error())
	}
}

func TestExecutionFailureKindFromReasonParsesKnownKind(t *testing.T) {
	kind := executioncore.FailureKindFromReason("worker: codex failed")

	if kind != executioncore.FailureWorker {
		t.Fatalf("kind = %q, want worker", kind)
	}
}
