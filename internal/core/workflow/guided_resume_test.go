package workflow

import (
	"testing"

	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

// TestGuidedUseCaseRecoveryResumeMarker preserves interrupted set-closing
// recovery while allowing an interrupted task or validation to continue normally.
func TestGuidedUseCaseRecoveryResumeMarker(t *testing.T) {
	for _, test := range []struct {
		name   string
		change func(*workflowInvocationFile)
		want   string
	}{
		{"task", func(run *workflowInvocationFile) { run.Lane = &workflowstate.Lane{} }, "resumed"},
		{"validation", func(run *workflowInvocationFile) { run.Cursor = 1; run.Lane = &workflowstate.Lane{} }, "resumed"},
		{"closing", func(run *workflowInvocationFile) { run.Cursor, run.ConfirmedCursor = 1, 1 }, "interrupted"},
		{"operator pause", func(run *workflowInvocationFile) { run.Outcome = "paused" }, "paused"},
	} {
		t.Run(test.name, func(t *testing.T) {
			run := workflowInvocationFile{Invocation: Invocation{InvocationExecution: InvocationExecution{Status: workflowstate.InvocationPending, Outcome: "interrupted"}, InvocationInput: InvocationInput{Members: []Member{{}}, SetImages: []InputImage{{Path: "slice-set/original/artifact.md"}}}}}
			test.change(&run)
			restoreWorkflowInterruptedContinuation(&run)

			if run.Outcome != test.want {
				t.Fatal("resume selected the wrong native recovery owner", run.Outcome)
			}
		})
	}
}
