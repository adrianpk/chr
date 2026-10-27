package authoring

import (
	"encoding/json"
	"errors"
	"maps"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	opsnamespace "forge.adrianpk.com/chr/internal/core/ops/namespace"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

// Preview orders valid creation timestamps before uncertain evidence, then
// relative paths; apply retains exact originals and replays one receipt.
func TestProjectDumpReconcile(t *testing.T) {
	project, sources := prepareProjectDumpReconciliation(t)
	before := projectDumpTestImage(t, project)

	preview, err := PreviewProjectDumpReconciliation(t.Context(), project)
	if err != nil {
		t.Fatal("creation evidence did not determine the retained identity", preview, err)
	}

	if !maps.Equal(before, projectDumpTestImage(t, project)) {
		t.Fatal("preview modified source documents")
	}

	checkProjectDumpReconciliationPreview(t, preview, sources)

	request := ProjectDumpReconciliationRequest{OperationID: "reconcile-1", ExpectedRevision: preview.Revision, Destination: preview.Destination}

	result, err := ReconcileProjectDump(t.Context(), project, request)
	if err != nil || result.Destination != preview.Destination {
		t.Fatal(result, err)
	}

	checkProjectDumpReconciliationImages(t, project, preview, before)

	after := projectDumpTestImage(t, project)

	replay, err := ReconcileProjectDump(t.Context(), project, request)
	if err != nil || replay != result || !maps.Equal(after, projectDumpTestImage(t, project)) {
		t.Fatal("retry repeated the composition", replay, err)
	}

	request.Destination = sources[0]
	if _, err := ReconcileProjectDump(t.Context(), project, request); !errors.Is(err, workflowstate.ErrConflict) {
		t.Fatal("operation identity accepted another target", err)
	}
}

func checkProjectDumpReconciliationPreview(t *testing.T, preview ProjectDumpReconciliationPreview, sources []string) {
	t.Helper()

	if preview.Destination != sources[1] || len(preview.Sources) != 3 || preview.Sources[2].TimestampIssue == "" {
		t.Fatal("creation evidence did not determine the retained identity", preview)
	}

	if !strings.HasPrefix(preview.Content, "---\ntitle: Old context\ncustom:\n  retained: true\n---\n") || strings.Index(preview.Content, "Old material.") > strings.Index(preview.Content, "Recent material.") {
		t.Fatal("composition lost the target frontmatter or chronological ordering", preview.Content)
	}
}

func prepareProjectDumpReconciliation(t *testing.T) (string, []string) {
	t.Helper()

	project := t.TempDir()
	recent := projectDumpTestArtifact(t, project, "default", "Recent", "Recent material.\n")
	old := projectDumpTestArtifact(t, project, "team", "Old", "---\ntitle: Old context\ncustom:\n  retained: true\n---\nOld material.\n")
	unknown := projectDumpTestArtifact(t, project, "default", "Unknown", "Undated material.\n")
	patchProjectDumpTestMetadata(t, project, old, map[string]any{"created_at": "2020-01-01T00:00:00Z", "custom": map[string]any{"retained": true}})
	patchProjectDumpTestMetadata(t, project, unknown, map[string]any{"created_at": "invalid"})
	projectDumpTestArtifact(t, project, "team", "Raw", "---\ntype: dump_raw\n---\nRaw original.\n")
	projectDumpTestArtifact(t, project, "team", "Scoped", "---\ncase: named-work\n---\nScoped original.\n")
	writeTestFile(t, ProjectMetadataPath(project), "---\nschema_version: 1\nname: Example\ndescription: Context.\ncustom:\n  retained: true\n---\nProject context.\n")

	return project, []string{recent, old, unknown}
}

func patchProjectDumpTestMetadata(t *testing.T, project, reference string, updates map[string]any) {
	t.Helper()

	path := filepath.Join(project, filepath.Dir(filepath.FromSlash(reference)), "metadata.json")

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}

	var fields map[string]any
	if err := json.Unmarshal(data, &fields); err != nil {
		t.Fatal(err)
	}

	maps.Copy(fields, updates)

	encoded, err := json.MarshalIndent(fields, "", "  ")
	if err != nil {
		t.Fatal(err)
	}

	writeTestFile(t, path, string(encoded))
}

func checkProjectDumpReconciliationImages(t *testing.T, project string, preview ProjectDumpReconciliationPreview, before map[string]string) {
	t.Helper()

	r := &fileSessionRepository{projectDir: project}
	for _, source := range preview.Sources {
		checkProjectDumpOriginalImages(t, r, source, before)

		resolved, err := ResolveProjectDumpReference(t.Context(), project, source.Reference)
		if err != nil || resolved.Reference != preview.Destination {
			t.Fatal("old reference no longer resolves", source.Reference, resolved, err)
		}
	}

	for path, original := range before {
		if projectDumpReconciliationChangedPath(project, path, preview) {
			continue
		}

		data, err := os.ReadFile(path)
		if err != nil || string(data) != original {
			t.Fatal("a retained source or raw/scoped component changed", path, err)
		}
	}

	checkProjectDumpReconciliationMetadata(t, project, preview)
}

func projectDumpReconciliationChangedPath(project, path string, preview ProjectDumpReconciliationPreview) bool {
	if path == filepath.Join(project, filepath.FromSlash(preview.Destination)) {
		return true
	}

	for _, source := range preview.Sources {
		if path == filepath.Join(project, filepath.Dir(filepath.FromSlash(source.Reference)), "metadata.json") {
			return true
		}
	}

	return false
}

func checkProjectDumpOriginalImages(t *testing.T, r *fileSessionRepository, source ProjectDumpArtifact, before map[string]string) {
	t.Helper()

	project := r.projectDir
	for _, name := range []string{"artifact.md", "metadata.json"} {
		original := before[filepath.Join(project, filepath.Dir(filepath.FromSlash(source.Reference)), name)]

		history, err := r.readAuthoringDocument(projectDumpHistoryRef(source) + "/" + name)
		if err != nil || string(history) != original {
			t.Fatal("original image was not retained", source.Reference, name, err)
		}
	}
}

func checkProjectDumpReconciliationMetadata(t *testing.T, project string, preview ProjectDumpReconciliationPreview) {
	t.Helper()

	t.Setenv(opsnamespace.UserEnv, "another-namespace")

	observed, err := ObserveProjectDump(t.Context(), project)
	if err != nil || observed.Destination != preview.Destination || observed.ReconciliationReason != "" || len(observed.Candidates) != 1 || len(observed.Aliases) != 2 {
		t.Fatal("restart/namespace change lost the one logical destination", observed, err)
	}

	metadata, err := os.ReadFile(filepath.Join(project, filepath.Dir(filepath.FromSlash(preview.Destination)), "metadata.json"))
	if err != nil || !strings.Contains(string(metadata), "\"custom\"") || !strings.Contains(string(metadata), "\"target_source_hash\"") || !strings.Contains(string(metadata), "\"timestamp_issue\"") {
		t.Fatal("unknown metadata or reconciliation provenance was lost", string(metadata), err)
	}

	checkProjectDumpRetainedProject(t, project)
}

func checkProjectDumpRetainedProject(t *testing.T, project string) {
	t.Helper()

	projectMetadata, err := os.ReadFile(ProjectMetadataPath(project))
	if err != nil || !strings.Contains(string(projectMetadata), "custom:\n  retained: true") || !strings.Contains(string(projectMetadata), "Project context.") {
		t.Fatal("reference publication lost project metadata", err)
	}
}

// Every replacement can lose acknowledgement. A new owner recovers the same
// journal, returns the receipt and never incorporates original material twice.
func TestProjectDumpReconcileRecovery(t *testing.T) {
	for fault := range 12 {
		t.Run(string(rune('a'+fault)), func(t *testing.T) {
			checkProjectDumpReconciliationFault(t, fault)
		})
	}
}

func checkProjectDumpReconciliationFault(t *testing.T, fault int) {
	t.Helper()

	project, _ := prepareProjectDumpReconciliation(t)

	preview, err := PreviewProjectDumpReconciliation(t.Context(), project)
	if err != nil {
		t.Fatal(err)
	}

	request := ProjectDumpReconciliationRequest{OperationID: "recover", ExpectedRevision: preview.Revision, Destination: preview.Destination}
	repo := &fileSessionRepository{projectDir: project, afterMutationReplace: func(index int) error {
		if index == fault {
			return errors.New("lost acknowledgement")
		}

		return nil
	}}

	result, err := repo.reconcileProjectDump(t.Context(), request)
	if err == nil {
		t.Fatal("fault not exercised", fault)
	}

	if _, err := ObserveProjectDump(t.Context(), project); !errors.Is(err, workflowstate.ErrConflict) {
		t.Fatal("read bypassed pending recovery", err)
	}

	replayed, err := ReconcileProjectDump(t.Context(), project, request)
	if err != nil || replayed != result {
		t.Fatal("restart did not recover the committed result", replayed, result, err)
	}

	body, err := os.ReadFile(filepath.Join(project, filepath.FromSlash(preview.Destination)))
	if err != nil || string(body) != preview.Content {
		t.Fatal("recovery changed the composition", err)
	}
}

// Competing confirmations of one observation have one winner; stale source or
// target changes fail before publication rather than replacing new material.
func TestProjectDumpReconcileConcurrent(t *testing.T) {
	project, _ := prepareProjectDumpReconciliation(t)

	preview, err := PreviewProjectDumpReconciliation(t.Context(), project)
	if err != nil {
		t.Fatal(err)
	}

	var group sync.WaitGroup

	results := make(chan error, 2)

	for _, id := range []string{"first", "second"} {
		group.Go(func() {
			_, err := ReconcileProjectDump(t.Context(), project, ProjectDumpReconciliationRequest{OperationID: id, ExpectedRevision: preview.Revision, Destination: preview.Destination})
			results <- err
		})
	}

	group.Wait()
	close(results)

	winners := 0

	for err := range results {
		if err == nil {
			winners++
		} else if !errors.Is(err, workflowstate.ErrConflict) {
			t.Fatal(err)
		}
	}

	if winners != 1 {
		t.Fatal("concurrent confirmations did not have one winner", winners)
	}
}
