package workflow

import (
	"context"
	"errors"
	"reflect"
	"testing"

	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

type unavailableNativePreparation struct{}

func (unavailableNativePreparation) PrepareWorkflowIntegration(context.Context, PreparationQuery) (workflowstate.IntegrationProof, error) {
	return workflowstate.IntegrationProof{}, errors.New("publisher credential unavailable")
}

// An explicit, revision-bound continuation retries only failed preparation.
// The original worker and failed attempt remain retained; receipt replay cannot
// prepare twice, and an unbound continuation cannot resume the failure.
func TestNativeHostPreparationRetry(t *testing.T) {
	host := newNativeHostFixture(t)
	f, provider := host.f, host.provider
	entry := nativeRoleAdmit(t, f.store, f.run.ID)
	entry = nativeRoleTransition(t, f.store, entry, "preparing")
	nativeRoleTransition(t, f.store, entry, "validating")

	if _, err := f.store.PrepareWorkflowSliceIntegration(t.Context(), f.run.ID, provider, unavailableNativePreparation{}); err == nil {
		t.Fatal("missing credentials did not block preparation")
	}

	before := workflowTestRead(t, f.store, f.run.ID)

	ports := ContinuationPorts{Publication: PublicationPorts{Provider: provider, Preparer: f.git, Observer: NewWorkflowIntegrationGit(host.profile.Repository, f.remote)}}
	if _, err := f.store.ContinueWorkflow(t.Context(), f.run.ID, ContinuationRequest{StepLimit: 1}, ports); err != nil {
		t.Fatal(err)
	}

	blocked := workflowTestRead(t, f.store, f.run.ID)
	if blocked.PreparationCount != before.PreparationCount || len(blocked.Attempts) != len(before.Attempts) {
		t.Fatal("unbound continuation retried preparation")
	}

	request := workflowstate.ControlMutationRequest{Binding: workflowstate.MutationBinding{RequestID: "native-preparation-retry", Operator: "operator"}, Action: "continue", ID: f.run.ID, ExpectedRevision: blocked.Revision, StepLimit: 1}

	control := LifecycleControlPorts{Execution: ports}
	if _, err := f.store.ControlWorkflowLifecycle(t.Context(), request, control, nil); err != nil {
		t.Fatal("explicit continuation could not recover preparation", err)
	}

	after := workflowTestRead(t, f.store, f.run.ID)
	if after.Preparation == nil || after.Preparation.State != "validated" || after.PreparationCount != 2 || len(after.Attempts) != len(before.Attempts)+1 {
		t.Fatal("preparation retry did not retain both attempts")
	}

	checkNativeHostRetryReplay(t, host, request, control, after)
}

func checkNativeHostRetryReplay(t *testing.T, host nativeHostFixture, request workflowstate.ControlMutationRequest, control LifecycleControlPorts, after Invocation) {
	t.Helper()

	if _, err := host.f.store.ControlWorkflowLifecycle(t.Context(), request, control, nil); err != nil {
		t.Fatal(err)
	}

	replayed := workflowTestRead(t, host.f.store, host.f.run.ID)
	if replayed.PreparationCount != after.PreparationCount || len(replayed.Attempts) != len(after.Attempts) || !reflect.DeepEqual(replayed.Attempts[0], host.f.run.Attempts[0]) {
		t.Fatal("retry replay repeated preparation or changed original worker proof")
	}
}
