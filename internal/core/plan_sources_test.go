package core

import (
	"maps"
	"os"
	"path/filepath"
	"strings"
	"testing"

	quizcore "forge.adrianpk.com/chr/internal/core/quiz"
)

// Observe real owners without allocating runtime state; only selected sources
// can affect the snapshot, and newly added material changes its binding.
func TestPlanSources(t *testing.T) {
	project := t.TempDir()
	writeTestFile(t, ProjectMetadataPath(project), RenderProjectMetadata(ProjectMetadata{Name: "Examples", Description: "Plan a greeting.", Context: "Keep publication explicit."}))
	writeTestFile(t, CaseMetadataPath(project, "chosen"), RenderCaseMetadata(CaseMetadata{Name: "Chosen case", Description: "Print a greeting."}))

	general := projectDumpTestArtifact(t, project, "team", "General", "# General\n\nPreserve explicit user actions.\n")
	projectDumpTestArtifact(t, project, "team", "Other case", "---\ncase: other\n---\nUnrelated discovery.\n")

	scope := quizcore.QuizScope{ProjectID: "example", CaseSlug: "chosen"}
	before := projectDumpTestImage(t, project)

	observed, err := ObservePlanSources(t.Context(), project, scope)
	if err != nil || !observed.HasDiscovery {
		t.Fatal("observe selected discovery", observed, err)
	}

	for _, input := range observed.Inputs {
		if strings.Contains(input.Content, "Unrelated discovery") {
			t.Fatal("unrelated case entered plan context")
		}
	}

	if !maps.Equal(before, projectDumpTestImage(t, project)) {
		t.Fatal("source observation wrote artifacts")
	}

	if _, err := os.Stat(filepath.Join(project, authoringRuntimeRef("owner.lock"))); !os.IsNotExist(err) {
		t.Fatal("observation acquired a mutation owner", err)
	}

	checkPlanSourceChanges(t, project, scope, general, observed.Hash)
}

func checkPlanSourceChanges(t *testing.T, project string, scope quizcore.QuizScope, general, previousHash string) {
	t.Helper()
	writeTestFile(t, filepath.Join(project, general), "# General\n\nPrint hello and verify stdout.\n")

	changed, err := ObservePlanSources(t.Context(), project, scope)
	if err != nil || changed.Hash == previousHash {
		t.Fatal("edited discovery did not invalidate observation", err)
	}

	projectDumpTestArtifact(t, project, "team", "Chosen discovery", "---\ncase: chosen\n---\nNo network access.\n")

	added, err := ObservePlanSources(t.Context(), project, scope)
	if err != nil || added.Hash == changed.Hash {
		t.Fatal("new scoped discovery did not invalidate observation", err)
	}
}

// Presence does not count headings or known placeholders as useful discovery;
// scope-relative semantic sufficiency is deliberately left to the provider.
func TestPlanDiscoveryPresence(t *testing.T) {
	for _, item := range []struct {
		name, body string
		want       bool
	}{
		{"empty", "", false},
		{"headings", "# Context\n## Constraints\n", false},
		{"placeholder", "# Context\n\nTODO\n- TBD.\n<!-- pending -->\n", false},
		{"short", "Print hello. Verify stdout equals hello.", true},
	} {
		t.Run(item.name, func(t *testing.T) {
			if got := usablePlanDiscovery(item.body); got != item.want {
				t.Fatalf("presence = %v, want %v", got, item.want)
			}
		})
	}
}
