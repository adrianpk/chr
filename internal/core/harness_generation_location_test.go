package core

import (
	"path/filepath"
	"testing"

	harnesshost "forge.adrianpk.com/chr/internal/core/harness/host"
)

// A host retains configuration from one project while authorizing execution in
// its owned worktree. Recovery never follows replacement project routing.
func TestHarnessGenerationBindingLocation(t *testing.T) {
	project, _ := generationBindingFixture(t)
	host, worktree := t.TempDir(), t.TempDir()
	location := harnesshost.HarnessGenerationLocation{ConfigDir: project, ExecutionDir: worktree, ProjectRef: "sample-project"}

	operation, err := FreezeHarnessGeneration(t.Context(), harnesshost.HarnessGenerationFreezeInput{Project: host, ID: "review-operation", Purpose: "plan_generation", Locations: []harnesshost.HarnessGenerationLocation{location}}, generationBindingResolver(filepath.Join(project, "fixture-driver"), "original-model"))
	if err != nil {
		t.Fatal(err)
	}

	ctx := harnesshost.WithHarnessGeneration(t.Context(), operation)

	request := generationBindingRequest(worktree)
	if _, err := harnesshost.RunAgent(ctx, generationBindingProcessRunner{}, request); err != nil {
		t.Fatal("owned worktree lost its original project binding", err)
	}

	replacement := harnesshost.HarnessGenerationLocation{ConfigDir: t.TempDir(), ExecutionDir: t.TempDir(), ProjectRef: "replacement"}

	retained, err := FreezeHarnessGeneration(t.Context(), harnesshost.HarnessGenerationFreezeInput{Project: host, ID: operation.ID, Purpose: operation.Purpose, Locations: []harnesshost.HarnessGenerationLocation{replacement}}, nil)
	if err != nil || retained != operation {
		t.Fatal("host recovery followed replacement routing", err)
	}

	observed, bound := harnesshost.HarnessGenerationLocationFromContext(harnesshost.WithHarnessGeneration(t.Context(), retained))
	if !bound || observed != location {
		t.Fatal("retained host routing changed", observed)
	}
}
