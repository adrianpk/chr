package authoring

import (
	"errors"
	"fmt"
	"maps"
	"os"
	"path/filepath"
	"strings"
	"testing"

	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
	workspacecore "forge.adrianpk.com/chr/internal/core/workspace"
)

// Apply re-observes source membership, body, metadata and project identity inside
// the owner. A different proposed target is never accepted as a confirmation.
func TestProjectDumpReconcileStale(t *testing.T) {
	for _, change := range []string{"new capture", "body", "artifact metadata", "project metadata", "target", "history"} {
		t.Run(change, func(t *testing.T) {
			project, sources := prepareProjectDumpReconciliation(t)

			preview, err := PreviewProjectDumpReconciliation(t.Context(), project)
			if err != nil {
				t.Fatal(err)
			}

			request := ProjectDumpReconciliationRequest{OperationID: "stale", ExpectedRevision: preview.Revision, Destination: preview.Destination}

			switch change {
			case "target":
				request.Destination = sources[0]
			case "history":
				writeTestFile(t, filepath.Join(project, filepath.FromSlash(projectDumpHistoryRef(preview.Sources[0])), "artifact.md"), "Independent history bytes.")
			case "artifact metadata":
				patchProjectDumpTestMetadata(t, project, sources[0], map[string]any{"title": "Changed metadata"})
			default:
				changeProjectDumpSource(t, project, sources[0], change)
			}

			before := projectDumpTestImage(t, project)
			if _, err := ReconcileProjectDump(t.Context(), project, request); !errors.Is(err, workflowstate.ErrConflict) {
				t.Fatal("stale confirmation accepted", err)
			}

			if !maps.Equal(before, projectDumpTestImage(t, project)) {
				t.Fatal("failed confirmation changed a source image")
			}
		})
	}
}

// Explicit size/count bounds fail before publication; a larger journal may carry
// several bounded images, while each individual canonical file keeps its limit.
func TestProjectDumpReconcileBounds(t *testing.T) {
	t.Run("composed body", func(t *testing.T) {
		project := t.TempDir()
		for _, title := range []string{"First", "Second"} {
			projectDumpTestArtifact(t, project, "default", title, strings.Repeat("x", workflowstate.FileLimit/2))
		}

		before := projectDumpTestImage(t, project)
		if _, err := PreviewProjectDumpReconciliation(t.Context(), project); err == nil {
			t.Fatal("oversize composition accepted")
		}

		if !maps.Equal(before, projectDumpTestImage(t, project)) {
			t.Fatal("oversize preview modified sources")
		}
	})
	t.Run("source count", func(t *testing.T) {
		project := t.TempDir()
		for i := range projectDumpSourceLimit + 1 {
			projectDumpTestArtifact(t, project, "default", fmt.Sprintf("Source %d", i), "Context.\n")
		}

		if _, err := PreviewProjectDumpReconciliation(t.Context(), project); err == nil {
			t.Fatal("unbounded reconciliation accepted")
		}
	})
	t.Run("larger journal", checkProjectDumpLargeJournal)
}

func checkProjectDumpLargeJournal(t *testing.T) {
	t.Helper()

	project := t.TempDir()
	for _, title := range []string{"First", "Second"} {
		projectDumpTestArtifact(t, project, "default", title, strings.Repeat(title, workflowstate.FileLimit/20))
	}

	preview, err := PreviewProjectDumpReconciliation(t.Context(), project)
	if err != nil {
		t.Fatal(err)
	}

	request := ProjectDumpReconciliationRequest{OperationID: "large", ExpectedRevision: preview.Revision, Destination: preview.Destination}

	repo := &fileAuthoringSessionRepository{projectDir: project, afterMutationReplace: func(int) error { return errors.New("lost acknowledgement") }}
	if _, err := repo.reconcileProjectDump(t.Context(), request); err == nil {
		t.Fatal("journal fault not exercised")
	}

	info, err := os.Stat(filepath.Join(project, authoringRuntimeRef("pending.json")))
	if err != nil || info.Size() <= workflowstate.FileLimit {
		t.Fatal("fixture did not exercise a journal larger than a source file", err)
	}

	if _, err := ReconcileProjectDump(t.Context(), project, request); err != nil {
		t.Fatal("bounded multi-image recovery failed", err)
	}
}

// A divergent external edit is retained with its journal. Recovery must never
// overwrite it or expose the partial mutation as a completed reconciliation.
func TestProjectDumpReconcileDivergence(t *testing.T) {
	project, _ := prepareProjectDumpReconciliation(t)

	preview, err := PreviewProjectDumpReconciliation(t.Context(), project)
	if err != nil {
		t.Fatal(err)
	}

	request := ProjectDumpReconciliationRequest{OperationID: "divergence", ExpectedRevision: preview.Revision, Destination: preview.Destination}

	repo := &fileAuthoringSessionRepository{projectDir: project, afterMutationReplace: func(int) error { return errors.New("lost acknowledgement") }}
	if _, err := repo.reconcileProjectDump(t.Context(), request); err == nil {
		t.Fatal("fault not exercised")
	}

	path := filepath.Join(project, filepath.FromSlash(preview.Destination))
	writeTestFile(t, path, "Independent edit.\n")

	if _, err := ReconcileProjectDump(t.Context(), project, request); !errors.Is(err, workflowstate.ErrConflict) {
		t.Fatal("divergent recovery accepted", err)
	}

	data, err := os.ReadFile(path)
	if err != nil || string(data) != "Independent edit.\n" {
		t.Fatal("recovery overwrote independent bytes", err)
	}

	if _, err := os.Stat(filepath.Join(project, authoringRuntimeRef("pending.json"))); err != nil {
		t.Fatal("divergent journal discarded", err)
	}
}

// Journal v1 retains its target and size contract. Only v2 may carry project
// metadata and immutable dump history; neither accepts arbitrary workspace files.
func TestProjectDumpJournalPaths(t *testing.T) {
	history := "ops/default/dump/00000-context/files/history/" + strings.Repeat("a", 64) + "/artifact.md"
	for _, row := range []struct {
		path   string
		v1, v2 bool
	}{
		{authoringReceiptRef("operation"), true, true},
		{"ops/default/dump/00000-context/artifact.md", true, true},
		{filepath.Join(workspacecore.DirectoryName, workspacecore.UserDirectoryName, "project.md"), false, true},
		{history, false, true},
		{"ops/default/spec/00000-context/files/history/" + strings.Repeat("a", 64) + "/artifact.md", false, false},
		{"README.md", false, false},
		{"ops/default/dump/00000-context/files/history/arbitrary/artifact.md", false, false},
	} {
		if validAuthoringJournalVersionPath("chr-authoring-journal/v1", row.path) != row.v1 || validAuthoringJournalVersionPath(projectDumpJournalSchema, row.path) != row.v2 {
			t.Fatal("journal path boundary changed", row.path)
		}
	}
}
