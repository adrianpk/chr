package app

import (
	"testing"

	workflowcore "forge.adrianpk.com/chr/internal/core/workflow"
)

// Native publication must use the production provider's preparation capability,
// rather than succeeding only when a test injects its own native implementation.
func TestServiceNativeProvider(t *testing.T) {
	operations, id, _ := serviceWorkflowPreparationReady(t)
	operations.workflowPullRequests = nil
	writeCLIWorkspaceConfig(t, operations.projectContext.ProjectDir, "review:\n  mode: trusted_auto_merge\nwork:\n  merge_strategy: rebase\n  integration:\n    publication_profile: git_atomic_dual_ref\n")

	store, provider, err := operations.workflowReviewDependencies(id)
	if err != nil {
		t.Fatal(err)
	}

	if _, ok := provider.(workflowcore.NativePreparationProvider); !ok {
		t.Fatal("production provider cannot prepare native publication")
	}

	run, err := operations.ReadWorkflow(t.Context(), id)
	if err != nil {
		t.Fatal(err)
	}

	ports := workflowcore.PublicationPorts{Provider: provider}
	bindWorkflowNativePublicationPorts(&ports, run)

	if any(ports.NativePublisher) != any(provider) || any(ports.NativeObserver) != any(provider) {
		t.Fatal("service bypassed the configured native publisher or observer")
	}

	if _, ok := ports.NativePublisher.(workflowcore.GitPublicationInspector); !ok {
		t.Fatal("configured publisher cannot inspect its native submission")
	}

	lifecycle, err := operations.workflowLifecycleControlPorts(t.Context(), store, WorkflowControlRequest{Action: "continue", ID: id, StepLimit: 1})
	if err != nil {
		t.Fatal(err)
	}

	if _, ok := lifecycle.Execution.Publication.Provider.(workflowcore.NativePreparationProvider); !ok {
		t.Fatal("bound continuation bypassed the production native provider")
	}
}
