package app

import (
	"encoding/json"
	"maps"
	"strings"
	"testing"
)

func serviceIntegrationReady(t *testing.T) (service, string) {
	t.Helper()

	operations, id, original := serviceWorkflowPreparationReady(t)
	operations.workflowPullRequests = &serviceNativeReviewFixture{original, "policy-v1"}
	writeCLIWorkspaceConfig(t, operations.projectContext.ProjectDir, "review:\n  mode: human_review\nwork:\n  merge_strategy: rebase\n  integration:\n    publication_profile: git_atomic_dual_ref\n")
	serviceNativeReviewAdmit(t, operations, id)

	if _, err := operations.PrepareWorkflowSliceIntegration(t.Context(), id); err != nil {
		t.Fatal(err)
	}

	if _, err := operations.ReviewWorkflowSlice(t.Context(), id, nil); err != nil {
		t.Fatal(err)
	}

	return operations, id
}

// TestServiceIntegrationStatus observes a real prepared native candidate without
// publishing, repairing authority or disclosing claim and process credentials.
func TestServiceIntegrationStatus(t *testing.T) {
	operations, id := serviceIntegrationReady(t)
	project := operations.projectContext.ProjectDir
	before := workflowCLIImage(t, project)

	status, err := operations.WorkflowStatus(t.Context(), id, WorkflowStatusRequest{Limit: 8})
	if err != nil {
		t.Fatal(err)
	}

	checkServiceIntegrationProgress(t, status)
	checkServiceIntegrationIdentity(t, status.Integration)
	checkServiceIntegrationPublic(t, status)

	if !maps.Equal(before, workflowCLIImage(t, project)) || status.IntegratedUnits != 0 || status.ConfirmedTickets != 0 {
		t.Fatal("read-only native status changed authority or invented delivery")
	}
}

func checkServiceIntegrationProgress(t *testing.T, status WorkflowWorkStatus) {
	t.Helper()

	observed := status.Integration
	if observed == nil || observed.Phase != "reviewing" || status.ReviewState != "waiting_approval" || observed.QueuePosition != 1 || observed.Method != "rebase" || observed.Publication != "not_started" {
		t.Fatal("native review wait or canonical target queue was lost", observed)
	}
}

func checkServiceIntegrationIdentity(t *testing.T, observed *WorkflowIntegrationStatus) {
	t.Helper()

	if observed.OriginalSource == "" || observed.PreparedResult == "" || observed.OriginalSource == observed.PreparedResult || observed.Budget.Consumed.Worker != 3 || observed.MaximumRoles != 1 {
		t.Fatal("native status lost original/prepared identity or consumed worker accounting", observed)
	}
}

func checkServiceIntegrationPublic(t *testing.T, status WorkflowWorkStatus) {
	t.Helper()

	data, err := json.Marshal(status)
	if err != nil {
		t.Fatal(err)
	}

	var public struct {
		Integration json.RawMessage `json:"integration"`
	}
	if err := json.Unmarshal(data, &public); err != nil || len(public.Integration) == 0 {
		t.Fatal("native service status omitted retained integration facts", err)
	}

	for _, private := range []string{"claim_generation", "lease_until", "owner_token", "boot_id", "credential_identity", "owned_workspace_id"} {
		if strings.Contains(string(data), private) {
			t.Fatal("native status disclosed private ownership", private)
		}
	}
}
