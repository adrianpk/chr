package app

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"strings"
	"testing"

	"forge.adrianpk.com/chr/internal/core"
)

// The shared operation enforces current local authority and exposes one fresh
// snapshot to an independently constructed service without restarting either.
func TestHarnessSelectionServiceChange(t *testing.T) {
	operations, project, configDir, command := harnessSelectionFixture(t)
	t.Setenv(core.ActiveOperatorIDEnv, "operator")
	t.Setenv("CHR_SPEC_GENERATION_BACKEND", "openai")

	initial, err := operations.InspectHarnessSelections(t.Context(), "spec_generation")
	if err != nil || initial.Snapshot.Worker != "" || len(initial.Profiles) != 3 {
		t.Fatal("choices not inspectable", initial, err)
	}

	choice := "writing"
	request := core.HarnessSelectionChangeRequest{Binding: core.WorkflowMutationBinding{RequestID: "select-both", Operator: "operator"}, ProjectRef: "fixture", ExpectedRevision: initial.Snapshot.Revision, Worker: &choice, Generation: &choice}
	authority := localHarnessSelectionAuthority(project, request.Binding.Operator)

	accepted, err := operations.ChangeHarnessSelections(t.Context(), request, authority)
	if err != nil {
		t.Fatal(err)
	}

	second := newService(t.Context(), operations.env, operations.projectContext, nil)

	checkHarnessSelectionServiceStatus(t, second, accepted, configDir, command)

	checkHarnessSelectionServiceReplay(t, second, request, authority, accepted)

	if _, err := os.Stat(configDir + "/spawned"); !os.IsNotExist(err) {
		t.Fatal("switching called the model executable")
	}
}

func checkHarnessSelectionServiceStatus(t *testing.T, second service, accepted core.HarnessSelectionSnapshot, configDir, command string) {
	t.Helper()

	status, err := second.InspectHarnessSelections(t.Context(), "spec_generation")
	if err != nil || status.Snapshot != accepted || status.Worker.Profile != "writing" || status.Generation.Profile != "writing" || len(status.Problems) != 0 {
		t.Fatal("second service retained stale defaults", status, err)
	}

	data, err := json.Marshal(status)
	if err != nil || strings.Contains(string(data), configDir) || strings.Contains(string(data), command) {
		t.Fatal("status exposed private configuration")
	}
}

func checkHarnessSelectionServiceReplay(t *testing.T, second service, request core.HarnessSelectionChangeRequest, authority func(context.Context) error, accepted core.HarnessSelectionSnapshot) {
	t.Helper()

	replay, err := second.ChangeHarnessSelections(t.Context(), request, authority)
	if err != nil || replay != accepted {
		t.Fatal("retry did not retain the response", replay, err)
	}

	t.Setenv(core.ActiveOperatorIDEnv, "different-operator")

	if _, err := second.ChangeHarnessSelections(t.Context(), request, authority); err == nil {
		t.Fatal("revoked actor replayed an accepted change")
	}
}

// Profile/readiness checks happen before publishing, and an exact foreign project
// cannot use the selected service's configuration as its mutation target.
func TestHarnessSelectionServiceRejects(t *testing.T) {
	for _, name := range []string{"missing", "pi-writing", "missing-executable", "foreign-project", "no-authority"} {
		t.Run(name, func(t *testing.T) { checkHarnessSelectionServiceRejection(t, name) })
	}
}

func checkHarnessSelectionServiceRejection(t *testing.T, name string) {
	t.Helper()

	operations, project, _, _ := harnessSelectionFixture(t)
	t.Setenv(core.ActiveOperatorIDEnv, "operator")

	initial, err := operations.InspectHarnessSelections(t.Context(), "spec_generation")
	if err != nil {
		t.Fatal(err)
	}

	request := core.HarnessSelectionChangeRequest{Binding: core.WorkflowMutationBinding{RequestID: "invalid", Operator: "operator"}, ProjectRef: "fixture", ExpectedRevision: initial.Snapshot.Revision, Generation: &name}
	authority := localHarnessSelectionAuthority(project, "operator")

	if name == "foreign-project" {
		request.ProjectRef = "foreign"
		request.Generation = nil
		request.Worker = harnessChoiceTestPointer("writing")
	}

	if name == "no-authority" {
		authority = nil
	}

	if _, err := operations.ChangeHarnessSelections(t.Context(), request, authority); err == nil {
		t.Fatal("invalid selection accepted")
	}

	config, err := core.LoadWorkspaceConfig(project)
	if err != nil || config.Revision() != initial.Snapshot.Revision {
		t.Fatal("invalid request changed configuration", config.Harnesses, err)
	}
}

// An inspection retains an invalid explicit reference visibly rather than
// silently projecting a healthy legacy default.
func TestHarnessSelectionStatusInvalid(t *testing.T) {
	for _, mode := range []string{"explicit missing", "invalid legacy backend"} {
		t.Run(mode, func(t *testing.T) { checkHarnessSelectionStatusInvalid(t, mode) })
	}
}

func checkHarnessSelectionStatusInvalid(t *testing.T, mode string) {
	t.Helper()

	operations, project, _, _ := harnessSelectionFixture(t)
	if mode == "explicit missing" {
		harnessSelectionWrite(t, core.UserPath(project, "config.yml"), "harnesses:\n  worker: missing\n")
	} else {
		t.Setenv("CHR_SPEC_GENERATION_BACKEND", "unsupported")
	}

	status, err := operations.InspectHarnessSelections(t.Context(), "spec_generation")
	if err != nil || len(status.Problems) != 1 {
		t.Fatal("inspection hid invalid selection", status, err)
	}

	if mode == "explicit missing" {
		if status.Worker.Profile != "missing" || status.Worker.Source != "project" || status.Worker.Readiness.Ready() {
			t.Fatal("invalid reference fell back", status.Worker)
		}
	} else if status.Generation.Readiness.Ready() {
		t.Fatal("invalid legacy backend appeared ready")
	}
}

// A waiting writer must recheck live authorization after acquiring exclusion.
func TestHarnessSelectionAuthorityRechecked(t *testing.T) {
	operations, _, _, _ := harnessSelectionFixture(t)
	count := 0
	request := core.HarnessSelectionChangeRequest{Binding: core.WorkflowMutationBinding{RequestID: "revoked", Operator: "operator"}, ProjectRef: "fixture", ExpectedRevision: core.WorkflowContentHash(nil), Worker: harnessChoiceTestPointer("writing")}

	_, err := operations.ChangeHarnessSelections(t.Context(), request, func(_ context.Context) error {
		count++
		if count > 1 {
			return errors.New("actor revoked")
		}

		return nil
	})
	if err == nil || count != 2 {
		t.Fatal("authorization was not rechecked", count, err)
	}

	config, err := core.LoadWorkspaceConfig(operations.projectContext.ProjectDir)
	if err != nil || config.Harnesses.Worker != "" {
		t.Fatal("revoked request changed configuration", err)
	}
}

func harnessChoiceTestPointer(value string) *string { return &value }
