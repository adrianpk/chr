package app

import (
	"bytes"
	"encoding/json"
	"maps"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"forge.adrianpk.com/chr/internal/core"
)

func workflowTrackerCLI(t *testing.T, project, id, action string, extra ...string) (core.WorkflowTrackerStatus, error) {
	t.Helper()

	args := append([]string{"work", "tracker", action, "--dir", project, "--json"}, extra...)
	args = append(args, id)

	var output bytes.Buffer

	err := run(t.Context(), args, &output, &bytes.Buffer{})

	var result core.WorkflowTrackerStatus
	if decodeErr := json.Unmarshal(output.Bytes(), &result); decodeErr != nil {
		t.Fatal("CLI omitted projection status", output.String(), decodeErr, err)
	}

	return result, err
}

// TestWorkflowTrackerCLI exercises real root parsing and the shared owner,
// including failure output, hash-bound permission, and execution-free repair.
func TestWorkflowTrackerCLI(t *testing.T) {
	operations, hierarchy, _ := workflowCLIFixture(t)
	project := operations.projectContext.ProjectDir

	first, err := workflowTrackerCLI(t, project, hierarchy.Set.ID, "export")
	if err != nil || first.Status != "current" {
		t.Fatal("CLI export failed", first, err)
	}

	checkWorkflowTrackerCLIService(t, operations, hierarchy.Set.ID, first)

	edited := []byte("Preserve manual notes\n")

	path := filepath.Join(project, "ops", "default", first.Path)
	if err := os.WriteFile(path, edited, 0o600); err != nil {
		t.Fatal(err)
	}

	stale, err := workflowTrackerCLI(t, project, hierarchy.Set.ID, "repair")
	if err == nil || stale.Reason != "projection_stale" || stale.ObservedHash != core.WorkflowContentHash(edited) {
		t.Fatal("CLI omitted failure provenance", stale, err)
	}

	before := workflowCLIImage(t, project)
	if status, err := workflowTrackerCLI(t, project, hierarchy.Set.ID, "status"); err != nil || status.Status != "stale" {
		t.Fatal("CLI stale status failed", status, err)
	}

	if !maps.Equal(before, workflowCLIImage(t, project)) {
		t.Fatal("CLI status wrote project files")
	}

	checkWorkflowTrackerCLIRepair(t, project, hierarchy.Set.ID, before, edited)
}

func checkWorkflowTrackerCLIService(t *testing.T, operations service, id string, first core.WorkflowTrackerStatus) {
	t.Helper()

	status, err := operations.InspectWorkflowTracker(t.Context(), id)
	if err != nil || status.SourceHash != first.SourceHash || status.OutputHash != first.OutputHash {
		t.Fatal("CLI and common service differ", status, err)
	}
}

func checkWorkflowTrackerCLIRepair(t *testing.T, project, id string, before map[string][32]byte, edited []byte) {
	t.Helper()

	repaired, err := workflowTrackerCLI(t, project, id, "repair", "--replace-hash", core.WorkflowContentHash(edited))
	if err != nil || repaired.Status != "current" {
		t.Fatal("explicit CLI replacement failed", repaired, err)
	}

	after := workflowCLIImage(t, project)
	projectionRoot := filepath.Join(project, "ops", "default", "tracker") + string(filepath.Separator)

	for _, image := range []map[string][32]byte{before, after} {
		for path := range image {
			if strings.HasPrefix(path, projectionRoot) {
				delete(image, path)
			}
		}
	}

	if !maps.Equal(before, after) {
		t.Fatal("tracker repair changed domain or invocation")
	}
}
