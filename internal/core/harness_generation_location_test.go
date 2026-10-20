package core

import "testing"

// A host retains configuration from one project while authorizing execution in
// its owned worktree. Recovery never follows replacement project routing.
func TestHarnessGenerationBindingLocation(t *testing.T) {
	project, original := generationBindingFixture(t)
	host, worktree := t.TempDir(), t.TempDir()
	location := HarnessGenerationLocation{ConfigDir: project, ExecutionDir: worktree, ProjectRef: "sample-project"}

	operation, err := FreezeHarnessGeneration(t.Context(), host, "review-operation", "plan_generation", generationBindingResolver(original.program, "original-model"), location)
	if err != nil {
		t.Fatal(err)
	}

	ctx := WithHarnessGeneration(t.Context(), operation)

	request := generationBindingRequest(worktree)
	if _, err := RunAgent(ctx, generationBindingProcessRunner{}, request); err != nil {
		t.Fatal("owned worktree lost its original project binding", err)
	}

	replacement := HarnessGenerationLocation{ConfigDir: t.TempDir(), ExecutionDir: t.TempDir(), ProjectRef: "replacement"}

	retained, err := FreezeHarnessGeneration(t.Context(), host, operation.ID, operation.Purpose, nil, replacement)
	if err != nil || retained != operation {
		t.Fatal("host recovery followed replacement routing", err)
	}

	observed, bound := HarnessGenerationLocationFromContext(WithHarnessGeneration(t.Context(), retained))
	if !bound || observed != location {
		t.Fatal("retained host routing changed", observed)
	}
}
