package app

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"forge.adrianpk.com/chr/internal/core"
)

// Two real Pi peers share the original model capacity across live profile changes.
func TestPiGenerationCapacity(t *testing.T) {
	operations, driver := piGenerationFixture(t, `{"ok":true}`)

	ctx, cancel := context.WithTimeout(t.Context(), 5*time.Second)
	defer cancel()

	first, firstCtx, err := operations.bindGeneration(ctx, "capacity-first", "conversation")
	if err != nil {
		t.Fatal(err)
	}

	switchGenerationFixture(t, operations, "other")

	second, secondCtx, err := operations.bindGeneration(ctx, "capacity-second", "conversation")
	if err != nil {
		t.Fatal(err)
	}

	harnessSelectionWrite(t, filepath.Join(driver, "block"), "owned wait")

	results := make(chan error, 2)

	var peers sync.WaitGroup

	defer func() { cancel(); peers.Wait() }()

	for i, call := range []struct {
		s   service
		ctx context.Context
	}{{first, firstCtx}, {second, secondCtx}} {
		peers.Go(func() {
			_, err := call.s.conversationModel.InvokeStructuredModel(call.ctx, piCapacityRequest(call.s.projectContext.ProjectDir, "owned-"+string(rune('a'+i))))
			results <- err
		})
	}

	waitPiGenerationCalls(t, driver, 2)

	_, err = second.conversationModel.InvokeStructuredModel(secondCtx, piCapacityRequest(second.projectContext.ProjectDir, "excess"))

	harnessSelectionWrite(t, filepath.Join(driver, "release"), "finish owned calls")

	if !errors.Is(err, core.ErrStructuredModelCapacity) {
		t.Error("Pi selection created another capacity pool", err)
	}

	for range 2 {
		if err := <-results; err != nil {
			t.Error("owned Pi call failed", err)
		}
	}

	checkPiPrivateCallState(t, driver)
}

func checkPiPrivateCallState(t *testing.T, driver string) {
	t.Helper()

	calls := piGenerationCalls(t, driver)
	if len(calls) != 2 || calls[0]["agent_dir"] == calls[1]["agent_dir"] || calls[0]["home"] == calls[1]["home"] {
		t.Fatal("Pi calls shared private state", calls)
	}

	for _, call := range calls {
		dir, ok := call["agent_dir"].(string)
		if !ok {
			t.Fatal("missing private state observation")
		}

		if _, err := os.Stat(dir); !os.IsNotExist(err) {
			t.Fatal("owned Pi state survived its completed call", err)
		}
	}
}

func piCapacityRequest(project, id string) core.StructuredModelRequest {
	return core.StructuredModelRequest{ID: id, Purpose: "capacity", ProjectDir: project, Instructions: "Return one bounded JSON object.", Input: json.RawMessage(`{"scope":"owned"}`), Schema: map[string]any{"type": "object"}}
}

func waitPiGenerationCalls(t *testing.T, driver string, count int) {
	t.Helper()

	deadline := time.Now().Add(4 * time.Second)
	for time.Now().Before(deadline) {
		if len(piGenerationCalls(t, driver)) == count {
			return
		}

		time.Sleep(5 * time.Millisecond)
	}

	t.Fatal("owned Pi peers did not reach their prepared calls")
}
