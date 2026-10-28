package closeout

import (
	"testing"

	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

// TestProjectionTarget keeps live target binding separate from retained history.
func TestProjectionTarget(t *testing.T) {
	target := workflowstate.PreviewTarget{Repository: "https://forge.example/team/project", Commit: "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"}
	run := workflowstate.Invocation{InvocationInput: workflowstate.InvocationInput{SetImages: []workflowstate.InputImage{{Path: "plan/set-1/metadata.json"}}}}
	ProjectInvocation(&run, Projection{State: Attempt{Outcome: "completed"}, SetID: "set-1", Target: &target})

	target.Commit = "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"
	if run.ExecutionTarget == nil || run.ExecutionTarget.Commit != "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb" {
		t.Fatal("execution target lost its host-owned binding")
	}

	if len(run.Closures) != 1 || run.Closures[0].Target.Commit != "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa" {
		t.Fatal("retained closure target changed with the live binding")
	}
}
