package core

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	opscore "forge.adrianpk.com/chr/internal/core/ops"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

// Concrete reads retain every actual plan, including same-scope documents and
// other namespaces, without manufacturing scope placeholders or tracker rows.
func TestPlanCatalogDocuments(t *testing.T) {
	project := t.TempDir()

	for _, namespace := range []string{"default", "research"} {
		for _, title := range []string{"First plan", "Second plan"} {
			planCatalogSeed(t, project, namespace, title)
		}
	}

	documents, err := ListPlanDocuments(t.Context(), project)
	if err != nil || len(documents) != 4 {
		t.Fatal(documents, err)
	}

	for _, document := range documents {
		checkPlanCatalogDocument(t, project, document)
	}
}

func planCatalogSeed(t *testing.T, project, namespace, title string) string {
	t.Helper()

	layout, err := opscore.ResolveOpsLayout(opscore.ResolveOpsLayoutOptions{ProjectDir: project, Namespace: namespace})
	if err != nil {
		t.Fatal(err)
	}

	artifact, err := opscore.CreateOpsArtifact(opscore.CreateOpsArtifactOptions{Layout: layout, Type: "plan", Title: title, Producer: "test"})
	if err != nil {
		t.Fatal(err)
	}

	path := filepath.Join(artifact.Dir, "artifact.md")
	if err := os.WriteFile(path, []byte("---\ntitle: "+title+"\ncase_slug: same-scope\n---\n\n# "+title+"\n\nFull retained body.\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	relative, err := filepath.Rel(project, path)
	if err != nil {
		t.Fatal(err)
	}

	return filepath.ToSlash(relative)
}

// Corrupt or foreign documents fail explicitly; body selection cannot escape
// the plan family, read retained tracker files, or traverse symlinks.
func TestPlanCatalogGuards(t *testing.T) {
	for _, kind := range []string{"symlink", "metadata", "frontmatter", "size", "missing envelope"} {
		t.Run(kind, func(t *testing.T) { checkPlanCatalogGuard(t, kind) })
	}

	project := t.TempDir()
	for _, reference := range []string{"../outside.md", "ops/default/tracker/00001-map/artifact.md", "ops/default/plan/00001-plan/files/tracker.md", "ops/default/plan/not-an-artifact/artifact.md"} {
		if _, err := ReadPlanDocument(t.Context(), project, reference); err == nil {
			t.Fatal("invalid plan accepted", reference)
		}
	}
}

func checkPlanCatalogGuard(t *testing.T, kind string) {
	t.Helper()

	project := t.TempDir()
	reference := planCatalogSeed(t, project, "default", "Guarded plan")
	path := filepath.Join(project, reference)

	mutatePlanCatalogSource(t, path, kind)

	if _, err := ListPlanDocuments(t.Context(), project); err == nil {
		t.Fatal("invalid source hidden")
	}
}

func checkPlanCatalogDocument(t *testing.T, project string, document PlanDocument) {
	t.Helper()

	if strings.Contains(document.Body, "case_slug:") || !strings.Contains(document.Body, "Full retained body.") {
		t.Fatal(document)
	}

	selected, err := ReadPlanDocument(t.Context(), project, document.Reference)
	if err != nil || selected.Revision != document.Revision || selected.Body != document.Body {
		t.Fatal(selected, err)
	}
}

func planCatalogSymlink(t *testing.T, path string) {
	t.Helper()

	if err := os.Remove(path); err != nil {
		t.Fatal(err)
	}

	if err := os.Symlink(filepath.Join(t.TempDir(), "outside.md"), path); err != nil {
		t.Fatal(err)
	}
}

func mutatePlanCatalogSource(t *testing.T, path, kind string) {
	t.Helper()

	data := []byte("broken")

	switch kind {
	case "symlink":
		planCatalogSymlink(t, path)
	case "metadata":
		path = filepath.Join(filepath.Dir(path), "metadata.json")
		data = []byte(`{"type":"tracker"}`)
	case "frontmatter":
		data = []byte("---\ntitle: broken\n")
	case "size":
		data = []byte(strings.Repeat("x", workflowstate.FileLimit+1))
	case "missing envelope":
		if err := os.Remove(filepath.Join(filepath.Dir(path), "metadata.json")); err != nil {
			t.Fatal(err)
		}
	}

	if kind != "symlink" && kind != "missing envelope" {
		if err := os.WriteFile(path, data, 0o644); err != nil {
			t.Fatal(err)
		}
	}
}
