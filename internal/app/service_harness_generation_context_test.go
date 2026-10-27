package app

import (
	"context"
	"errors"
	"testing"
	"time"

	harnesshost "forge.adrianpk.com/chr/internal/core/harness/host"
)

// Canceling one caller must stop its call without pinning the retained operation
// to that canceled parent. A fresh caller can dispatch the same admitted binding.
func TestHarnessGenerationCallerContext(t *testing.T) {
	operations, _ := piGenerationFixture(t, `{"ok":true}`)

	parent, cancel := context.WithTimeout(t.Context(), 5*time.Second)
	defer cancel()

	bound, err := operations.bindGeneration(parent, "caller-context", "conversation")
	if err != nil {
		t.Fatal(err)
	}

	call := bound.context(parent)
	deadline, ok := call.Deadline()

	expected, _ := parent.Deadline()
	if !ok || !deadline.Equal(expected) {
		t.Fatal("generation lost its caller deadline")
	}

	cancel()

	if !errors.Is(call.Err(), context.Canceled) {
		t.Fatal("generation lost its caller cancellation", call.Err())
	}

	request := piCapacityRequest(operations.projectContext.ProjectDir, "canceled-call")
	if _, err := bound.conversationModel.InvokeStructuredModel(call, request); !errors.Is(err, context.Canceled) {
		t.Fatal("canceled generation dispatched or lost cancellation", err)
	}

	fresh := bound.context(t.Context())

	generation, ok := harnesshost.GenerationSelection(fresh)
	if !ok || generation.ID != "caller-context" || fresh.Err() != nil {
		t.Fatal("fresh caller did not retain the admitted operation", generation, fresh.Err())
	}

	request.ID = "fresh-call"
	if _, err := bound.conversationModel.InvokeStructuredModel(fresh, request); err != nil {
		t.Fatal("fresh caller inherited the canceled parent", err)
	}
}
