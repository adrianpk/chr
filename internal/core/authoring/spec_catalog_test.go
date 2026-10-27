package authoring

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	opsmetadata "forge.adrianpk.com/chr/internal/core/ops/metadata"
	quizcore "forge.adrianpk.com/chr/internal/core/quiz"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

// Catalog reads preserve every actual artifact across namespaces and never
// manufacture placeholder cards from missing files or operational directories.
func TestSpecCatalogDocuments(t *testing.T) {
	project := t.TempDir()

	for _, namespace := range []string{"default", "research"} {
		request := specTestRequest(namespace)
		request.Namespace = namespace
		specTestMutate(t, project, request)
	}

	if err := os.MkdirAll(filepath.Join(project, "ops/default/spec/stable"), 0o755); err != nil {
		t.Fatal(err)
	}

	documents, err := ListSpecDocuments(t.Context(), project)
	if err != nil || len(documents) != 2 {
		t.Fatal(documents, err)
	}

	if !strings.Contains(documents[0].Reference, "/default/") || !strings.Contains(documents[1].Reference, "/research/") {
		t.Fatal(documents)
	}
}

// An in-flight generation cannot replace a manual edit made after observing
// its target, including when generation originally observed absence.
func TestSpecGenerationRevision(t *testing.T) {
	project := t.TempDir()
	spec := GeneratedSpec{Scope: quizcore.Scope{ProjectID: "sample-project"}, Title: "Generated", Body: "Original generated body.", CreatedAt: time.Now().UTC()}

	paths, err := GeneratedSpecLayout(project, spec.Scope)
	if err != nil {
		t.Fatal(err)
	}

	spec.ExpectedRevision, err = SpecTargetRevision(t.Context(), project, paths.RelativePath)
	if err != nil {
		t.Fatal(err)
	}

	if _, err := WriteGeneratedSpec(t.Context(), project, spec); err != nil {
		t.Fatal(err)
	}

	if _, err := WriteGeneratedSpec(t.Context(), project, spec); !errors.Is(err, workflowstate.ErrConflict) {
		t.Fatal("stale absence accepted", err)
	}

	document := specTestRead(t, project, paths.RelativePath)
	spec.ExpectedRevision = document.Revision
	request := specTestRequest("manual-edit")
	request.Action, request.Namespace, request.Reference, request.ExpectedRevision = "save", "", document.Reference, document.Revision
	specTestMutate(t, project, request)

	if _, err := WriteGeneratedSpec(t.Context(), project, spec); !errors.Is(err, workflowstate.ErrConflict) {
		t.Fatal("generation overwrote editor", err)
	}

	if after := specTestRead(t, project, document.Reference); after.Body != request.Body {
		t.Fatal(after)
	}
}

// A malformed, excessive or symlinked source fails the catalog read instead of
// silently returning an empty collection that could hide retained documents.
func TestSpecCatalogReadGuards(t *testing.T) {
	for _, kind := range []string{"symlink", "metadata", "bound"} {
		t.Run(kind, func(t *testing.T) { checkSpecCatalogGuard(t, kind) })
	}
}

func checkSpecCatalogGuard(t *testing.T, kind string) {
	t.Helper()

	project := t.TempDir()
	result := specTestMutate(t, project, specTestRequest("guard"))
	path := filepath.Join(project, result.Reference)

	switch kind {
	case "symlink":
		if err := os.Remove(path); err != nil {
			t.Fatal(err)
		}

		if err := os.Symlink(filepath.Join(t.TempDir(), "outside.md"), path); err != nil {
			t.Fatal(err)
		}
	case "metadata":
		path = filepath.Join(filepath.Dir(path), "metadata.json")
		if err := os.WriteFile(path, []byte("broken"), 0o644); err != nil {
			t.Fatal(err)
		}
	case "bound":
		if err := os.WriteFile(path, []byte(strings.Repeat("x", workflowstate.FileLimit+1)), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	if _, err := ListSpecDocuments(t.Context(), project); err == nil {
		t.Fatal("invalid source hidden")
	}
}

// A matching revision cannot authorize reassigning an existing generated
// identity whose retained scope differs from the requested generation scope.
func TestSpecGenerationScope(t *testing.T) {
	project := t.TempDir()
	spec := GeneratedSpec{Scope: quizcore.Scope{ProjectID: "sample-project"}, Title: "Scoped", Body: "Retained requirements.", CreatedAt: time.Now().UTC()}

	paths, err := WriteGeneratedSpec(t.Context(), project, spec)
	if err != nil {
		t.Fatal(err)
	}

	before := specTestRead(t, project, paths.RelativePath)

	content := strings.Replace(string(before.images[0].After), "case_slug: "+opsmetadata.Quote(""), "case_slug: "+opsmetadata.Quote("other-case"), 1)
	if err := os.WriteFile(paths.Path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	spec.ExpectedRevision, err = SpecTargetRevision(t.Context(), project, paths.RelativePath)
	if err != nil {
		t.Fatal(err)
	}

	if _, err := WriteGeneratedSpec(t.Context(), project, spec); !errors.Is(err, workflowstate.ErrConflict) {
		t.Fatal("generated scope changed", err)
	}
}
