package core

import (
	"errors"
	"maps"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// Observation is read-only; adoption retains the sole general identity and every
// byte of its raw/history components, independent of namespace or index state.
func TestProjectDumpAdoption(t *testing.T) {
	project := t.TempDir()
	raw := projectDumpTestArtifact(t, project, "team-a", "Raw capture", "---\ntype: dump_raw\nstatus: raw\n---\nOriginal capture.\n")
	reference := projectDumpTestArtifact(t, project, "team-a", "General context", "---\ntitle: General context\nsource: "+raw+"\n---\nCurated context.\n")
	projectDumpTestArtifact(t, project, "other", "Named scope", "---\ncase: named-work\n---\nCase material.\n")
	writeTestFile(t, filepath.Join(project, filepath.Dir(filepath.FromSlash(reference)), "files", "history.md"), "Retained history.\n")
	writeTestFile(t, ProjectMetadataPath(project), "---\nschema_version : 1\nname : Example\ndescription: Preserved description.\ncustom:\n  name: Retained nested name\n  project_dump_ref: untrusted-nested-ref\n---\n\nProject context.\n")

	before := projectDumpTestImage(t, project)

	metadataBefore, err := os.ReadFile(ProjectMetadataPath(project))
	if err != nil {
		t.Fatal(err)
	}

	observed, err := ObserveProjectDump(t.Context(), project)
	if err != nil || observed.Destination != reference || len(observed.Candidates) != 1 || len(observed.Components) != 1 || observed.Candidates[0].RawReference != raw {
		t.Fatal("logical components were confused with general captures", observed, err)
	}

	if !maps.Equal(before, projectDumpTestImage(t, project)) {
		t.Fatal("observation changed artifact bytes")
	}

	checkProjectDumpInspection(t, project, metadataBefore, observed)

	checkProjectDumpAdopted(t, project, observed, reference)

	if !maps.Equal(before, projectDumpTestImage(t, project)) {
		t.Fatal("adoption changed artifact components")
	}
}

func checkProjectDumpInspection(t *testing.T, project string, metadataBefore []byte, observed ProjectDumpObservation) {
	t.Helper()

	metadataAfter, err := os.ReadFile(ProjectMetadataPath(project))
	if err != nil || string(metadataAfter) != string(metadataBefore) {
		t.Fatal("observation changed project metadata", err)
	}

	if _, err := os.Stat(filepath.Join(project, authoringRuntimeRef("owner.lock"))); !errors.Is(err, os.ErrNotExist) {
		t.Fatal("read created mutation state", err)
	}

	again, err := ObserveProjectDump(t.Context(), project)
	if err != nil || again.Revision != observed.Revision {
		t.Fatal("unchanged observation revision drifted", err)
	}
}

func checkProjectDumpAdopted(t *testing.T, project string, observed ProjectDumpObservation, reference string) {
	t.Helper()

	service := NewProjectMetadataService(project)

	saved, err := service.AdoptProjectDump(t.Context(), observed.Revision)
	if err != nil || saved.ProjectDumpRef != reference || saved.Name != "Example" || saved.Context != "Project context." {
		t.Fatal("adoption lost project identity or context", saved, err)
	}

	content, err := os.ReadFile(ProjectMetadataPath(project))
	if err != nil || !strings.Contains(string(content), "custom:\n  name: Retained nested name\n  project_dump_ref: untrusted-nested-ref") {
		t.Fatal("adoption lost unknown frontmatter", string(content), err)
	}

	checkProjectDumpNamespace(t, project, reference, saved)
}

func checkProjectDumpNamespace(t *testing.T, project, reference string, saved ProjectMetadata) {
	t.Helper()

	service := NewProjectMetadataService(project)

	t.Setenv(OpsUserEnv, "new-namespace")

	observed, err := ObserveProjectDump(t.Context(), project)
	if err != nil || observed.Destination != reference || observed.ReconciliationReason != "" {
		t.Fatal("namespace change lost canonical destination", observed, err)
	}

	replayed, err := service.AdoptProjectDump(t.Context(), observed.Revision)
	if err != nil || replayed.SourceHash != saved.SourceHash {
		t.Fatal("repeated adoption rewrote the project reference", replayed, err)
	}

	saved.Name = "Updated display name"

	saved, err = service.SaveContext(t.Context(), saved)
	if err != nil || saved.ProjectDumpRef != reference {
		t.Fatal("project metadata save lost canonical reference", saved, err)
	}
}

// Empty, multiple and invalidly bound stores never acquire an accidental
// destination. Reads leave pending recovery and original artifact bytes intact.
func TestProjectDumpReconciliation(t *testing.T) {
	for _, name := range []string{"empty", "multiple", "missing", "case target", "raw target", "raw only", "pending"} {
		t.Run(name, func(t *testing.T) { checkProjectDumpConflict(t, name) })
	}
}

func checkProjectDumpConflict(t *testing.T, name string) {
	t.Helper()

	project := t.TempDir()
	prepareProjectDumpConflict(t, project, name)

	before := projectDumpTestImage(t, project)

	observed, err := ObserveProjectDump(t.Context(), project)
	if name == "pending" {
		checkProjectDumpPending(t, project, err)

		return
	}

	if err != nil {
		t.Fatal(err)
	}

	if name != "empty" && observed.ReconciliationReason == "" {
		t.Fatal("conflicting destination accepted", observed)
	}

	_, err = NewProjectMetadataService(project).AdoptProjectDump(t.Context(), observed.Revision)
	if !errors.Is(err, ErrProjectDumpReconciliation) || !maps.Equal(before, projectDumpTestImage(t, project)) {
		t.Fatal("adoption changed conflicting material", err)
	}
}

func checkProjectDumpPending(t *testing.T, project string, err error) {
	t.Helper()

	if !errors.Is(err, ErrWorkflowConflict) {
		t.Fatal("read recovered or ignored pending mutation", err)
	}

	data, readErr := os.ReadFile(filepath.Join(project, authoringRuntimeRef("pending.json")))
	if readErr != nil || string(data) != "retained pending bytes" {
		t.Fatal("read changed pending recovery", readErr)
	}
}

func prepareProjectDumpConflict(t *testing.T, project, name string) {
	t.Helper()

	switch name {
	case "multiple":
		projectDumpTestArtifact(t, project, "default", "First", "First context.\n")
		projectDumpTestArtifact(t, project, "other", "Second", "Second context.\n")
	case "missing":
		writeTestFile(t, ProjectMetadataPath(project), RenderProjectMetadata(ProjectMetadata{Name: "Example", Description: "Context.", ProjectDumpRef: "ops/default/dump/missing/artifact.md"}))
	case "case target", "raw target":
		body := "---\ncase: named-work\n---\nNamed context.\n"
		if name == "raw target" {
			body = "---\ntype: dump_raw\n---\nRaw context.\n"
		}

		reference := projectDumpTestArtifact(t, project, "default", "Scoped", body)
		writeTestFile(t, ProjectMetadataPath(project), RenderProjectMetadata(ProjectMetadata{Name: "Example", Description: "Context.", ProjectDumpRef: reference}))
	case "raw only":
		projectDumpTestArtifact(t, project, "default", "Raw", "---\ntype: dump_raw\n---\nRetained raw.\n")
	case "pending":
		writeTestFile(t, filepath.Join(project, authoringRuntimeRef("pending.json")), "retained pending bytes")
	}
}

// The complete source image fences a new candidate or metadata edit that arrives
// after observation; source binding is checked while holding the authoring owner.
func TestProjectDumpStale(t *testing.T) {
	for _, change := range []string{"new capture", "project metadata", "body", "artifact metadata", "readme"} {
		t.Run(change, func(t *testing.T) {
			project := t.TempDir()
			reference := projectDumpTestArtifact(t, project, "default", "General", "Original context.\n")

			observed, err := ObserveProjectDump(t.Context(), project)
			if err != nil {
				t.Fatal(err)
			}

			changeProjectDumpSource(t, project, reference, change)

			_, err = NewProjectMetadataService(project).AdoptProjectDump(t.Context(), observed.Revision)
			if !errors.Is(err, ErrWorkflowConflict) {
				t.Fatal("stale adoption accepted", change, err)
			}
		})
	}
}

func changeProjectDumpSource(t *testing.T, project, reference, change string) {
	t.Helper()

	switch change {
	case "new capture":
		projectDumpTestArtifact(t, project, "other", "Concurrent capture", "Other content.\n")
	case "project metadata":
		_, err := NewProjectMetadataService(project).Save(ProjectMetadata{Name: "Updated project", Description: "Changed after observation."})
		if err != nil {
			t.Fatal(err)
		}
	case "readme":
		writeTestFile(t, filepath.Join(project, "README.md"), "# Example\n\nChanged default project context.\n")
	case "body":
		writeTestFile(t, filepath.Join(project, filepath.FromSlash(reference)), "Changed body.\n")
	case "artifact metadata":
		path := filepath.Join(project, filepath.Dir(filepath.FromSlash(reference)), "metadata.json")

		content, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}

		writeTestFile(t, path, strings.Replace(string(content), "\"title\": \"General\"", "\"title\": \"Changed title\"", 1))
	}
}

func projectDumpTestArtifact(t *testing.T, project, namespace, title, body string) string {
	t.Helper()

	layout, err := ResolveOpsLayout(ResolveOpsLayoutOptions{ProjectDir: project, Namespace: namespace})
	if err != nil {
		t.Fatal(err)
	}

	artifact, err := CreateOpsArtifact(CreateOpsArtifactOptions{Layout: layout, Type: "dump", Title: title, Producer: "chr", Now: func() time.Time { return time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC) }})
	if err != nil {
		t.Fatal(err)
	}

	path := filepath.Join(artifact.Dir, OpsArtifactBodyFile())
	writeTestFile(t, path, body)

	reference, err := filepath.Rel(project, path)
	if err != nil {
		t.Fatal(err)
	}

	return filepath.ToSlash(reference)
}

func projectDumpTestImage(t *testing.T, project string) map[string]string {
	t.Helper()

	image := map[string]string{}

	root := filepath.Join(project, OpsDirectoryName)
	if _, err := os.Stat(root); errors.Is(err, os.ErrNotExist) {
		return image
	}

	if err := filepath.WalkDir(root, func(path string, entry os.DirEntry, err error) error {
		if err != nil || entry.IsDir() {
			return err
		}

		data, err := os.ReadFile(path)
		image[path] = string(data)

		return err
	}); err != nil {
		t.Fatal(err)
	}

	return image
}
