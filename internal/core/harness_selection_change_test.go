package core

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"forge.adrianpk.com/chr/internal/core/contenthash"
	harnesscore "forge.adrianpk.com/chr/internal/core/harness"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
	workspacecore "forge.adrianpk.com/chr/internal/core/workspace"
)

// A single accepted image preserves unrelated bytes, commits both roles and
// retains its original response even after later changes or unavailable profiles.
func TestHarnessSelectionChange(t *testing.T) {
	project := t.TempDir()
	original := "# retain comment\nunknown_extension:\n  custom: untouched\nwork:\n  max_active_workers: 2\n"
	writeHarnessSelectionTestFile(t, workspacecore.UserPath(project, "config.yml"), original)

	request := harnessSelectionTestRequest("first", contenthash.Sum([]byte(original)), "alpha")

	result, err := ChangeHarnessSelections(t.Context(), project, harnessSelectionTestOptions(request))
	if err != nil || result.Worker != "alpha" || result.Generation != "alpha" || result.ChangedBy != "operator" || result.RequestID != "first" {
		t.Fatal("atomic selection failed", result, err)
	}

	data := readHarnessSelectionTestConfig(t, project)
	if !strings.HasPrefix(string(data), original) || result.Revision != contenthash.Sum(data) {
		t.Fatal("changed unrelated configuration or reported a partial revision")
	}

	checkHarnessSelectionChangeReplay(t, project, request, result)
}

func checkHarnessSelectionChangeReplay(t *testing.T, project string, request HarnessSelectionChangeRequest, result harnesscore.SelectionSnapshot) {
	t.Helper()

	next := harnessSelectionTestRequest("second", result.Revision, "beta")
	if _, err := ChangeHarnessSelections(t.Context(), project, harnessSelectionTestOptions(next)); err != nil {
		t.Fatal(err)
	}

	options := harnessSelectionTestOptions(request)
	options.Validate = func(context.Context, WorkspaceConfig) error { return errors.New("profile is now unavailable") }

	replay, err := ChangeHarnessSelections(t.Context(), project, options)
	if err != nil || replay != result {
		t.Fatal("lost-response replay rebound the request", replay, err)
	}

	changed := request

	changed.Generation = harnessSelectionTestChoice("other")
	if _, err := ChangeHarnessSelections(t.Context(), project, harnessSelectionTestOptions(changed)); !errors.Is(err, workflowstate.ErrConflict) {
		t.Fatal("request ID exchanged choices", err)
	}

	config, err := LoadWorkspaceConfig(project)
	if err != nil || config.Harnesses.Worker != "beta" {
		t.Fatal("replay rewrote current selections", config.Harnesses, err)
	}
}

// Invalid, unauthorized, stale and unprepared requests leave the configuration
// and accepted-request evidence untouched. Expected revisions cover all fields.
func TestHarnessSelectionChangeRejects(t *testing.T) {
	cases := []struct {
		name   string
		change func(*HarnessSelectionChangeOptions)
	}{
		{"stale unrelated edit", func(o *HarnessSelectionChangeOptions) { o.Request.ExpectedRevision = contenthash.Sum([]byte("older")) }},
		{"unauthorized", func(o *HarnessSelectionChangeOptions) {
			o.Authorize = func(context.Context) error { return errors.New("denied") }
		}},
		{"missing authority", func(o *HarnessSelectionChangeOptions) { o.Authorize = nil }},
		{"unprepared", func(o *HarnessSelectionChangeOptions) {
			o.Validate = func(context.Context, WorkspaceConfig) error { return errors.New("unready") }
		}},
		{"path input", func(o *HarnessSelectionChangeOptions) {
			o.Request.Worker = harnessSelectionTestChoice("/private/driver")
		}},
		{"absent roles", func(o *HarnessSelectionChangeOptions) { o.Request.Worker = nil; o.Request.Generation = nil }},
		{"unbounded actor", func(o *HarnessSelectionChangeOptions) { o.Request.Binding.Operator = strings.Repeat("a", 129) }},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			project := t.TempDir()
			original := "unknown: retained\n"
			writeHarnessSelectionTestFile(t, workspacecore.UserPath(project, "config.yml"), original)

			options := harnessSelectionTestOptions(harnessSelectionTestRequest("reject", contenthash.Sum([]byte(original)), "alpha"))
			tc.change(&options)

			if _, err := ChangeHarnessSelections(t.Context(), project, options); err == nil {
				t.Fatal("invalid request accepted")
			}

			if string(readHarnessSelectionTestConfig(t, project)) != original {
				t.Fatal("rejection changed configuration")
			}

			for _, relative := range []string{harnessSelectionPending, harnessSelectionReceiptPath("reject")} {
				if _, err := os.Stat(workspacecore.UserPath(project, relative)); !errors.Is(err, os.ErrNotExist) {
					t.Fatal("rejection retained an accepted request", err)
				}
			}
		})
	}
}

// Clearing one role restores omission semantics and preserves the other role.
// Foreign project metadata and symlinked configuration fail closed.
func TestHarnessSelectionChangeClear(t *testing.T) {
	project := t.TempDir()

	first, err := ChangeHarnessSelections(t.Context(), project, harnessSelectionTestOptions(harnessSelectionTestRequest("select", contenthash.Sum(nil), "alpha")))
	if err != nil {
		t.Fatal(err)
	}

	request := HarnessSelectionChangeRequest{Binding: workflowstate.MutationBinding{RequestID: "clear", Operator: "operator"}, ProjectRef: "fixture", ExpectedRevision: first.Revision, Generation: harnessSelectionTestChoice("")}

	result, err := ChangeHarnessSelections(t.Context(), project, harnessSelectionTestOptions(request))
	if err != nil || result.Worker != "alpha" || result.Generation != "" {
		t.Fatal("clear did not preserve the other role", result, err)
	}

	request.ProjectRef = "foreign"

	request.ExpectedRevision = result.Revision
	if _, err := ChangeHarnessSelections(t.Context(), project, harnessSelectionTestOptions(request)); !errors.Is(err, workflowstate.ErrConflict) {
		t.Fatal("foreign project changed selections", err)
	}

	if err := os.Rename(workspacecore.UserPath(project, "config.yml"), workspacecore.UserPath(project, "original.yml")); err != nil {
		t.Fatal(err)
	}

	if err := os.Symlink("original.yml", workspacecore.UserPath(project, "config.yml")); err != nil {
		t.Fatal(err)
	}

	request.ProjectRef = "fixture"
	if _, err := ChangeHarnessSelections(t.Context(), project, harnessSelectionTestOptions(request)); err == nil {
		t.Fatal("symlinked configuration accepted")
	}
}

// The profile named "null" remains distinct from the null value that clears a
// role, including when the accepted change is reloaded from configuration.
func TestHarnessSelectionNullProfile(t *testing.T) {
	project := t.TempDir()
	request := harnessSelectionTestRequest("named-null", contenthash.Sum(nil), "null")

	result, err := ChangeHarnessSelections(t.Context(), project, harnessSelectionTestOptions(request))
	if err != nil || result.Worker != "null" || result.Generation != "null" {
		t.Fatal("named profile was cleared", result, err)
	}

	config, err := LoadWorkspaceConfig(project)
	if err != nil || config.Harnesses.Worker != "null" {
		t.Fatal("quoted profile did not survive reload", config.Harnesses, err)
	}

	request = harnessSelectionTestRequest("clear-null", result.Revision, "")

	result, err = ChangeHarnessSelections(t.Context(), project, harnessSelectionTestOptions(request))
	if err != nil || result.Worker != "" || result.Generation != "" {
		t.Fatal("null value did not clear roles", result, err)
	}
}

func harnessSelectionTestRequest(id, revision, profile string) HarnessSelectionChangeRequest {
	return HarnessSelectionChangeRequest{Binding: workflowstate.MutationBinding{RequestID: id, Operator: "operator"}, ProjectRef: "fixture", ExpectedRevision: revision, Worker: harnessSelectionTestChoice(profile), Generation: harnessSelectionTestChoice(profile)}
}

func harnessSelectionTestOptions(request HarnessSelectionChangeRequest) HarnessSelectionChangeOptions {
	return HarnessSelectionChangeOptions{Request: request, Authorize: func(context.Context) error { return nil }, Validate: func(context.Context, WorkspaceConfig) error { return nil }}
}

func writeHarnessSelectionTestFile(t *testing.T, path, body string) {
	t.Helper()

	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		t.Fatal(err)
	}

	if err := os.WriteFile(path, []byte(body), 0o600); err != nil {
		t.Fatal(err)
	}
}

func readHarnessSelectionTestConfig(t *testing.T, project string) []byte {
	t.Helper()

	data, err := readHarnessSelectionConfig(project)
	if err != nil {
		t.Fatal(err)
	}

	return data
}

func harnessSelectionTestChoice(value string) *string { return &value }
