package app

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"encoding/pem"
	"io/fs"
	"maps"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strconv"
	"testing"

	"forge.adrianpk.com/chr/internal/core"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

func workflowCLIFixture(t *testing.T) (service, core.WorkflowMaterializationResult, string) {
	t.Helper()

	project := newTestWorkspace(t)
	serviceWorkflowDeliveryGit(t, project, "init", "-q")
	writeCLIWorkspaceConfig(t, project, "ticket:\n  key_prefix: PRJ\nwork:\n  state_authority: tickets\n")
	writeServiceWorkspaceFile(t, filepath.Join(project, "ops", "default", "plan", "approved.md"), "approved\n")

	operations := newTestService(t, project)
	operations.projectContext.ProjectID = projectID(project)
	plan := serviceWorkflowPlanFixture()
	plan.UseCase = "delivery"
	plan.Slices[0].Tasks[0].Validation = workflowstate.Validation{Required: true, Commands: [][]string{{"git", "show", "HEAD:result.txt"}}, TimeoutSeconds: 30}

	hierarchy, err := operations.MaterializeWorkflowPlan(t.Context(), plan)
	if err != nil {
		t.Fatal(err)
	}

	serviceWorkflowPublishCoordinator(t, project)
	serviceWorkflowDeliveryGit(t, project, "-c", "user.name=Test", "-c", "user.email=test@example.com", "commit", "--allow-empty", "-qm", "Fixture")

	target := serviceWorkflowDependencyGit(t, project, "rev-parse", "HEAD")
	remote := filepath.Join(t.TempDir(), "canonical.git")
	serviceWorkflowDeliveryGit(t, project, "init", "--bare", remote)
	serviceWorkflowDeliveryGit(t, project, "push", remote, target+":refs/heads/dev")
	workflowCLIHTTPRemote(t, project, remote)

	return operations, hierarchy, remote
}

func workflowCLIHTTPRemote(t *testing.T, project, remote string) {
	t.Helper()

	serviceWorkflowDeliveryGit(t, project, "--git-dir", remote, "update-server-info")

	server := httptest.NewTLSServer(http.StripPrefix("/test/repo", http.FileServer(http.Dir(remote))))
	t.Cleanup(server.Close)

	certificate := filepath.Join(t.TempDir(), "fixture-ca.pem")
	if err := os.WriteFile(certificate, pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: server.Certificate().Raw}), 0o600); err != nil {
		t.Fatal(err)
	}

	serviceWorkflowDeliveryGit(t, project, "remote", "add", "origin", server.URL+"/test/repo")
	serviceWorkflowDeliveryGit(t, project, "config", "http.sslCAInfo", certificate)
}

func workflowCLIImage(t *testing.T, root string) map[string][32]byte {
	t.Helper()

	image := map[string][32]byte{}

	if err := filepath.WalkDir(root, func(path string, entry fs.DirEntry, failure error) error {
		if failure != nil || entry.IsDir() {
			return failure
		}

		data, err := os.ReadFile(path)
		image[path] = sha256.Sum256(data)

		return err
	}); err != nil {
		t.Fatal(err)
	}

	return image
}

func workflowCLIPreview(t *testing.T, project string, args ...string) (WorkflowWorkPreview, error) {
	t.Helper()

	var output, stderr bytes.Buffer

	command := append([]string{"work", args[0], "--dir", project, "--preview", "--json"}, args[1:]...)
	err := run(t.Context(), command, &output, &stderr)

	var preview WorkflowWorkPreview
	if err == nil {
		err = json.Unmarshal(output.Bytes(), &preview)
	}

	return preview, err
}

// TestWorkflowCLIExtents resolves all explicit extents through the real root
// parser and proves their previews preserve every Git and operational file.
func TestWorkflowCLIExtents(t *testing.T) {
	operations, hierarchy, _ := workflowCLIFixture(t)
	project := operations.projectContext.ProjectDir
	before := workflowCLIImage(t, project)

	cases := []struct {
		name, identity, extent string
		total                  int
	}{
		{"ticket", hierarchy.Tasks[0].Task.Key, "ticket", 1},
		{"slice", hierarchy.Set.Set.Slices[0].ID, "slice", 3},
		{"slice-set", hierarchy.Set.ID, "slice_set", 6},
		{"use-case", "delivery", "use_case", 6},
		{"loose", "", "loose", 0},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			args := []string{tc.name, "--limit", "1"}
			if tc.identity != "" {
				args = append(args, tc.identity)
			}

			preview, err := workflowCLIPreview(t, project, args...)
			if err != nil || preview.Preview.Selector.Extent != tc.extent || preview.Preview.Total != tc.total || len(preview.Preview.Members) > 1 || preview.Policy.Mode != workflowstate.ReviewHuman {
				t.Fatalf("extent preview: %+v, %v", preview, err)
			}
		})
	}

	if !maps.Equal(before, workflowCLIImage(t, project)) {
		t.Fatal("preview changed repository files, refs, objects or operational records")
	}
}

// TestWorkflowCLIPages binds every page to the exact remote target while leaving
// a newly observed remote commit absent from the local object database.
func TestWorkflowCLIPages(t *testing.T) {
	operations, hierarchy, remote := workflowCLIFixture(t)
	project := operations.projectContext.ProjectDir

	first, err := workflowCLIPreview(t, project, "slice-set", "--limit", "1", hierarchy.Set.ID)
	if err != nil {
		t.Fatal(err)
	}

	page, err := workflowCLIPreview(t, project, "slice-set", "--limit", "1", "--offset", strconv.Itoa(first.Preview.NextOffset), "--snapshot", first.Preview.SnapshotHash, hierarchy.Set.ID)
	if err != nil || page.Preview.Members[0].Key == first.Preview.Members[0].Key || page.Preview.SnapshotHash != first.Preview.SnapshotHash {
		t.Fatal("pagination lost frozen selection", err)
	}

	next := workflowCLIAdvanceTarget(t, project, remote)

	before := workflowCLIImage(t, project)
	if _, err := workflowCLIPreview(t, project, "slice-set", "--snapshot", first.Preview.SnapshotHash, hierarchy.Set.ID); err == nil {
		t.Fatal("stale target-bound page accepted")
	}

	preview, err := workflowCLIPreview(t, project, "slice-set", hierarchy.Set.ID)
	if err != nil || preview.Target.Commit != next || !maps.Equal(before, workflowCLIImage(t, project)) {
		t.Fatal("target observation fetched objects or failed to report current target", err)
	}
}

func workflowCLIAdvanceTarget(t *testing.T, project, remote string) string {
	t.Helper()

	clone := filepath.Join(t.TempDir(), "target")
	serviceWorkflowDeliveryGit(t, project, "clone", "--branch", "dev", remote, clone)
	serviceWorkflowDeliveryGit(t, clone, "-c", "user.name=Test", "-c", "user.email=test@example.com", "commit", "--allow-empty", "-qm", "Moved target")
	serviceWorkflowDeliveryGit(t, clone, "push", "origin", "dev")
	serviceWorkflowDeliveryGit(t, project, "--git-dir", remote, "update-server-info")

	return serviceWorkflowDependencyGit(t, clone, "rev-parse", "HEAD")
}
