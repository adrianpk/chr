package core

import (
	"context"
	"errors"
	"testing"

	modelcore "forge.adrianpk.com/chr/internal/core/model"
	"forge.adrianpk.com/chr/internal/harness"
)

// Selected copies share the original two slots; switching cannot allocate a
// second pool while calls in the original selection remain active.
func TestHarnessGenerationBindingCapacityPool(t *testing.T) {
	entered, release := make(chan struct{}, 2), make(chan struct{})
	runner := generationCapacityRunner{entered: entered, release: release}
	provider := NewConfiguredStructuredModel(false)
	provider.Runner = runner
	selected := provider.WithGenerationSelection(false, "selected", runner)
	results := make(chan error, 2)
	request := structuredHarnessRequest(t)

	for _, active := range []*ConfiguredStructuredModel{provider, selected} {
		go func() {
			_, err := active.InvokeStructuredModel(t.Context(), request)
			results <- err
		}()
	}

	<-entered
	<-entered

	_, err := selected.WithGenerationSelection(false, "replacement", runner).InvokeStructuredModel(t.Context(), request)

	close(release)

	if !errors.Is(err, modelcore.ErrCapacity) {
		t.Fatal("selection created another capacity pool", err)
	}

	for range 2 {
		if err := <-results; err != nil {
			t.Fatal("original owned call failed", err)
		}
	}
}

type generationCapacityRunner struct {
	entered chan<- struct{}
	release <-chan struct{}
}

func (r generationCapacityRunner) RunAgent(ctx context.Context, _ harness.AgentRunRequest) (harness.AgentRunResult, error) {
	r.entered <- struct{}{}

	select {
	case <-r.release:
		return harness.AgentRunResult{Output: []byte(`{}`)}, nil
	case <-ctx.Done():
		return harness.AgentRunResult{}, ctx.Err()
	}
}
