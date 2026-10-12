package core

import "testing"

type nativePreparationGuardOnly struct {
	WorkflowPublicationGuardProvider
}

// TestNativePreparationAdmission requires the explicit native profile, exact
// queue phase and trusted result-bound observer before claiming a host gate.
// Unsupported entrypoints cannot silently retain a legacy candidate instead.
func TestNativePreparationAdmission(t *testing.T) {
	for _, condition := range []string{"disabled", "source only", "missing observer", "queued"} {
		t.Run(condition, func(t *testing.T) { nativePreparationAdmission(t, condition) })
	}
}

func nativePreparationAdmission(t *testing.T, condition string) {
	t.Helper()

	fixture := newNativePreparationFixture(t, "rebase", workflowPreparationGate("test -f result.txt"))
	f := fixture.f

	provider := nativePreparationAdmissionProvider(t, fixture, condition)
	before := workflowTestRead(t, f.store, f.run.ID)

	if _, err := f.store.PrepareWorkflowSliceIntegration(t.Context(), f.run.ID, provider, f.git); err == nil {
		t.Fatal("unsupported native entrypoint claimed validation")
	}

	after := workflowTestRead(t, f.store, f.run.ID)
	if len(after.Attempts) != len(before.Attempts) || after.Preparation != nil {
		t.Fatal("unsupported native entrypoint consumed a preparation attempt")
	}
}

func nativePreparationAdmissionProvider(t *testing.T, fixture nativePreparationFixture, condition string) WorkflowPublicationGuardProvider {
	t.Helper()

	switch condition {
	case "disabled":
		writeWorkflowTestConfigValue(t, fixture.f.store.layout.ProjectDir(), "work", "integration.publication_profile", WorkflowPublicationDisabled)
	case "source only":
		return fixture.f.provider
	case "missing observer":
		return nativePreparationGuardOnly{fixture.provider}
	case "queued":
		nativeRoleTransition(t, fixture.f.store, fixture.entry, "queued")
	}

	return fixture.provider
}
