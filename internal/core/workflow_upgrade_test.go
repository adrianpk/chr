package core

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	generationcore "forge.adrianpk.com/chr/internal/core/authoring/generation"
	"forge.adrianpk.com/chr/internal/core/contenthash"
)

func workflowUpgradeFixture(t *testing.T, body string) (*WorkflowStore, WorkflowUpgradeOptions) {
	t.Helper()

	store := workflowTestStore(t)
	if body != "" {
		writeWorkspaceConfig(t, store.layout.ProjectDir(), body)
	}

	prior, err := ResolveWorkflowConfiguration(store.layout.ProjectDir(), "")
	if err != nil {
		t.Fatal(err)
	}

	return store, WorkflowUpgradeOptions{ExpectedRevision: prior.Revision, StateAuthority: WorkStateAuthorityTickets, CheckQuiescent: func(context.Context) error { return nil }}
}

func workflowUpgradeValue(config WorkflowConfiguration, key string) string {
	for _, value := range config.Values {
		if value.Key == key {
			return value.Value
		}
	}

	return ""
}

// TestWorkflowUpgradePolicy keeps absent and explicit legacy review choices,
// records their sources before the marker and changes review only on request.
func TestWorkflowUpgradePolicy(t *testing.T) {
	cases := []struct{ name, body, choice, review string }{
		{"absent", "", "", "human_review"},
		{"explicit human", "review:\n  mode: human_review\n", "", "human_review"},
		{"explicit agentic", "review:\n  mode: agentic_review\n", "", "agentic_review"},
		{"explicit trusted", "review:\n  mode: trusted_auto_merge\n", "", "trusted_auto_merge"},
		{"new choice", "", "trusted_auto_merge", "trusted_auto_merge"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			store, opts := workflowUpgradeFixture(t, tc.body)
			opts.ReviewMode = ReviewWorkflowMode(tc.choice)

			prior, err := ResolveWorkflowConfiguration(store.layout.ProjectDir(), "")
			if err != nil {
				t.Fatal(err)
			}

			result, err := store.UpgradeWorkflowConfiguration(t.Context(), opts)
			if err != nil || result.Current.Version != 1 || !reflect.DeepEqual(result.Prior, prior) {
				t.Fatal("prior policy or version changed", result, err)
			}

			if workflowUpgradeValue(result.Current, "review.mode") != tc.review || workflowUpgradeValue(result.Prior, "work.state_authority") != "tracker_compat" || workflowUpgradeValue(result.Current, "work.state_authority") != "tickets" {
				t.Fatal("upgrade changed policy implicitly", result)
			}

			checkWorkflowUpgradeSnapshot(t, store, opts, result)
		})
	}
}

func checkWorkflowUpgradeSnapshot(t *testing.T, store *WorkflowStore, opts WorkflowUpgradeOptions, result WorkflowUpgradeResult) {
	t.Helper()

	data, err := os.ReadFile(filepath.Join(store.layout.ProjectDir(), result.Snapshot))
	if err != nil {
		t.Fatal(err)
	}

	var snapshot workflowUpgradeJournal
	if err := json.Unmarshal(data, &snapshot); err != nil || !reflect.DeepEqual(snapshot.Prior, result.Prior) {
		t.Fatal("prior policy snapshot lost", err)
	}

	again, err := store.UpgradeWorkflowConfiguration(t.Context(), opts)
	if err != nil || !reflect.DeepEqual(again, result) {
		t.Fatal("upgrade retry changed its result", again, err)
	}

	current, err := ResolveWorkflowConfiguration(store.layout.ProjectDir(), "")
	if err != nil || !reflect.DeepEqual(current, result.Current) {
		t.Fatal("stored policy differs from upgrade receipt", current, err)
	}
}

// TestWorkflowUpgradePreserves leaves unrelated configuration, explicit budgets
// and unselected legacy history intact, including explicit tracker authority.
func TestWorkflowUpgradePreserves(t *testing.T) {
	body := "# Project settings\nnotifications:\n  telegram_chat_id: private-channel\nwork:\n  workspace_mode: project\n  max_worker_attempts: 7\nreview:\n  mode: agentic_review\n  forgejo_token_env: PRIVATE_TOKEN\n"
	store, opts := workflowUpgradeFixture(t, body)
	opts.StateAuthority = WorkStateAuthorityTrackerCompat
	history := workflowLegacyTicketFixture(t, store, generationcore.GeneratedWorkflowTask{ID: "old", Title: "History", Outcome: "Done"}, 101)

	result, err := store.UpgradeWorkflowConfiguration(t.Context(), opts)
	if err != nil || result.Adopted != nil || workflowUpgradeValue(result.Current, "work.workspace_mode") != "project" || workflowUpgradeValue(result.Current, "work.max_worker_attempts") != "7" {
		t.Fatal("unrequested authority or budget transition", result, err)
	}

	checkWorkflowUpgradeUnrelated(t, store, result, history)
}

func checkWorkflowUpgradeUnrelated(t *testing.T, store *WorkflowStore, result WorkflowUpgradeResult, history WorkflowAdoptionTask) {
	t.Helper()

	image, err := os.ReadFile(UserPath(store.layout.ProjectDir(), "config.yml"))
	if err != nil || !strings.Contains(string(image), "notifications:\n  telegram_chat_id: private-channel\n") || !strings.Contains(string(image), "  forgejo_token_env: PRIVATE_TOKEN\n") {
		t.Fatal("unrelated configuration lost", err)
	}

	snapshot, err := os.ReadFile(filepath.Join(store.layout.ProjectDir(), result.Snapshot))
	if err != nil || strings.Contains(string(snapshot), "PRIVATE_TOKEN") || strings.Contains(string(snapshot), "private-channel") {
		t.Fatal("snapshot exposed unrelated configuration", err)
	}

	historical, err := os.ReadFile(filepath.Join(store.layout.Root(), history.Path))
	if err != nil || contenthash.Sum(historical) != history.SourceHash {
		t.Fatal("unselected history changed", err)
	}
}

// TestWorkflowUpgradeAdoption preserves selected legacy identities and original
// history while leaving an unrelated ticket outside the upgraded workflow.
func TestWorkflowUpgradeAdoption(t *testing.T) {
	store, selection := workflowPlanAdoptionFixture(t)

	prior, err := ResolveWorkflowConfiguration(store.layout.ProjectDir(), "")
	if err != nil {
		t.Fatal(err)
	}

	opts := WorkflowUpgradeOptions{ExpectedRevision: prior.Revision, StateAuthority: WorkStateAuthorityTickets, Adoption: &selection, CheckQuiescent: func(context.Context) error { return nil }}
	history := workflowLegacyTicketFixture(t, store, generationcore.GeneratedWorkflowTask{ID: "old", Title: "Historical", Outcome: "Preserved"}, 101)

	result, err := store.UpgradeWorkflowConfiguration(t.Context(), opts)
	if err != nil || result.Adopted == nil {
		t.Fatal("selected adoption failed", err)
	}

	checkAdoptedWorkflowTasks(t, store, *result.Adopted, selection)
	checkWorkflowUpgradeSnapshot(t, store, opts, result)

	data, err := os.ReadFile(filepath.Join(store.layout.Root(), history.Path))
	if err != nil || contenthash.Sum(data) != history.SourceHash {
		t.Fatal("upgrade converted unselected history", err)
	}
}

// TestWorkflowUpgradeRejects refuses malformed, stale and unobserved upgrades
// before either policy or selected domain records are changed.
func TestWorkflowUpgradeRejects(t *testing.T) {
	cases := []struct {
		name   string
		change func(*WorkflowUpgradeOptions)
	}{
		{"stale config", func(o *WorkflowUpgradeOptions) { o.ExpectedRevision = strings.Repeat("a", 64) }},
		{"invalid revision", func(o *WorkflowUpgradeOptions) { o.ExpectedRevision = strings.Repeat("z", 64) }},
		{"missing authority", func(o *WorkflowUpgradeOptions) { o.StateAuthority = "" }},
		{"unknown review", func(o *WorkflowUpgradeOptions) { o.ReviewMode = "unknown" }},
		{"no ownership observation", func(o *WorkflowUpgradeOptions) { o.CheckQuiescent = nil }},
		{"active legacy execution", func(o *WorkflowUpgradeOptions) {
			o.CheckQuiescent = func(context.Context) error { return ErrWorkflowOwned }
		}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			store, opts := workflowUpgradeFixture(t, "")
			tc.change(&opts)

			if _, err := store.UpgradeWorkflowConfiguration(t.Context(), opts); err == nil {
				t.Fatal("unsafe upgrade accepted")
			}

			checkWorkflowUpgradeAbsent(t, store)
		})
	}
}

func checkWorkflowUpgradeAbsent(t *testing.T, store *WorkflowStore) {
	t.Helper()

	for _, path := range []string{"config.yml", workflowUpgradePending, "workflow-upgrades"} {
		if _, err := os.Stat(UserPath(store.layout.ProjectDir(), path)); !errors.Is(err, os.ErrNotExist) {
			t.Fatal("rejected upgrade wrote policy or snapshot", path, err)
		}
	}
}

// TestWorkflowUpgradeAmbiguous ensures selection ambiguity and active legacy
// attempts are rejected before the project is marked version one.
func TestWorkflowUpgradeAmbiguous(t *testing.T) {
	cases := []struct {
		name   string
		change func(*testing.T, *WorkflowStore, *WorkflowPlanAdoptionOptions)
	}{
		{"duplicate key", duplicateLegacyWorkflowTicket},
		{"active attempt", markLegacyWorkflowAttempt},
		{"foreign use case", markLegacyWorkflowUseCase},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			store, selection := workflowPlanAdoptionFixture(t)
			tc.change(t, store, &selection)

			opts := WorkflowUpgradeOptions{ExpectedRevision: contenthash.Sum(nil), StateAuthority: WorkStateAuthorityTickets, Adoption: &selection, CheckQuiescent: func(context.Context) error { return nil }}
			if _, err := store.UpgradeWorkflowConfiguration(t.Context(), opts); err == nil {
				t.Fatal("ambiguous upgrade accepted")
			}

			checkWorkflowUpgradeAbsent(t, store)
		})
	}
}
