package authoring

import (
	"context"
	"errors"
	"maps"
	"os"
	"path/filepath"
	"strings"
	"testing"

	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

// Invalid capture inputs, unresolved general ownership and missing/stale edit
// bindings fail without changing canonical artifact images.
func TestDumpMutationGuards(t *testing.T) {
	for _, name := range []string{"identity", "empty", "oversize", "namespace", "multiple", "missing target", "raw edit", "missing revision", "traversal", "cross document"} {
		t.Run(name, func(t *testing.T) { checkDumpMutationGuard(t, name) })
	}
}

func checkDumpMutationGuard(t *testing.T, name string) {
	t.Helper()

	project := t.TempDir()
	request := dumpCaptureRequest("guard", "Captured material.")

	switch name {
	case "identity":
		request.OperationID = ""
	case "empty":
		request.Content = "  "
	case "oversize":
		request.Content = strings.Repeat("x", 65537)
	case "namespace":
		request.Namespace = "operator-space"
	case "multiple":
		projectDumpTestArtifact(t, project, "default", "First", "First material.")
		projectDumpTestArtifact(t, project, "other", "Second", "Second material.")
	case "missing target":
		prepareProjectDumpConflict(t, project, "missing")
	default:
		request = prepareDumpEditGuard(t, project, name)
	}

	before := projectDumpTestImage(t, project)
	if _, err := MutateDump(t.Context(), project, request, nil); err == nil {
		t.Fatal("invalid write accepted", name)
	}

	if !maps.Equal(before, projectDumpTestImage(t, project)) {
		t.Fatal("invalid write changed canonical bytes")
	}
}

func prepareDumpEditGuard(t *testing.T, project, name string) DumpMutationRequest {
	t.Helper()

	body := "Original material."
	if name == "raw edit" {
		body = "---\ntype: dump_raw\n---\nOriginal raw material."
	}

	reference := projectDumpTestArtifact(t, project, "default", "General", body)

	observed, err := ReadDumpDocument(t.Context(), project, reference)
	if err != nil {
		t.Fatal(err)
	}

	request := DumpMutationRequest{OperationID: "edit-guard", Action: "save", Reference: reference, ExpectedRevision: observed.Revision, Content: "Replacement material."}

	switch name {
	case "missing revision":
		request.ExpectedRevision = ""
	case "traversal":
		request.Reference = "../outside/artifact.md"
	case "cross document":
		other := projectDumpTestArtifact(t, project, "other", "Scoped", "---\ncase: another-case\n---\nScoped material.")
		request.Reference = other
	}

	return request
}

// Provider preparation cannot reclassify a capture's scope or unlink its raw
// component. All failed preparations leave canonical documents unpublished.
func TestDumpMutationPreparedScope(t *testing.T) {
	for _, kind := range []string{"scope", "raw reference", "document kind"} {
		t.Run(kind, func(t *testing.T) {
			project := t.TempDir()
			request := dumpCaptureRequest("scope", "Material.")

			prepare := func(ctx context.Context, destination, raw string) (DumpCapturePreparation, error) {
				prepared, err := defaultDumpCapture(request)(ctx, destination, raw)
				prepared.Content = invalidDumpPreparedContent(kind, prepared.Content, raw)

				return prepared, err
			}
			if _, err := MutateDump(t.Context(), project, request, prepare); err == nil {
				t.Fatal("provider changed capture ownership")
			}

			if _, err := os.Stat(ProjectMetadataPath(project)); !errors.Is(err, os.ErrNotExist) {
				t.Fatal("invalid preparation published pointer", err)
			}
		})
	}
}

func invalidDumpPreparedContent(kind, content, raw string) string {
	switch kind {
	case "scope":
		return strings.Replace(content, "type: dump\n", "type: dump\ncase_slug: hidden-case\n", 1)
	case "document kind":
		return strings.Replace(content, "type: dump\n", "type: plan\n", 1)
	default:
		return strings.ReplaceAll(content, raw, "ops/default/dump/foreign/artifact.md")
	}
}

// Superseded references edit the resolved destination with its observed binding,
// retaining the original superseded body as historical material.
func TestDumpMutationAlias(t *testing.T) {
	project, sources := prepareProjectDumpReconciliation(t)

	preview, err := PreviewProjectDumpReconciliation(t.Context(), project)
	if err != nil {
		t.Fatal(err)
	}

	_, err = ReconcileProjectDump(t.Context(), project, ProjectDumpReconciliationRequest{OperationID: "reconcile", ExpectedRevision: preview.Revision, Destination: preview.Destination})
	if err != nil {
		t.Fatal(err)
	}

	observed, err := ReadDumpDocument(t.Context(), project, sources[0])
	if err != nil || observed.Artifact.Reference != preview.Destination {
		t.Fatal("alias read did not return canonical identity", observed, err)
	}

	request := DumpMutationRequest{OperationID: "save-alias", Action: "save", Reference: sources[0], ExpectedRevision: observed.Revision, Content: "Updated canonical material.\n"}

	result, err := MutateDump(t.Context(), project, request, nil)
	if err != nil || result.Destination != preview.Destination {
		t.Fatal("alias save did not retain canonical identity", result, err)
	}

	retained, err := os.ReadFile(filepath.Join(project, filepath.FromSlash(sources[0])))
	if err != nil || string(retained) != "Recent material.\n" {
		t.Fatal("alias save rewrote the superseded original", err)
	}
}

// Legacy case_slug declarations are scoped identities rather than general dump
// candidates; creating general context must not adopt them accidentally.
func TestDumpMutationLegacyCase(t *testing.T) {
	project := t.TempDir()
	old := projectDumpTestArtifact(t, project, "default", "Scoped", "---\ncase_slug: named-work\n---\nCase material.")

	result, err := MutateDump(t.Context(), project, dumpCaptureRequest("general", "General material."), nil)
	if err != nil || result.Destination == old {
		t.Fatal("general capture adopted a legacy scoped identity", result, err)
	}
}

// An exact copy of another project's document cannot supply a local Save guard.
func TestDumpMutationProjectBinding(t *testing.T) {
	project, other := t.TempDir(), t.TempDir()
	reference := projectDumpTestArtifact(t, project, "default", "General", "Original material.")
	projectDumpTestArtifact(t, other, "default", "General", "Original material.")

	foreign, err := ReadDumpDocument(t.Context(), other, reference)
	if err != nil {
		t.Fatal(err)
	}

	request := DumpMutationRequest{OperationID: "foreign", Action: "save", Reference: reference, ExpectedRevision: foreign.Revision, Content: "Unbound replacement."}
	if _, err := MutateDump(t.Context(), project, request, nil); !errors.Is(err, workflowstate.ErrConflict) {
		t.Fatal("foreign document revision accepted", err)
	}
}

// A caller can bind a named capture to its observed case metadata image as well
// as validating that the case still exists under authoring exclusion.
func TestDumpMutationCaseRevision(t *testing.T) {
	project := t.TempDir()
	owner := NewCaseMetadataService(project)

	metadata, err := owner.Save(CaseMetadata{Name: "Named Work"})
	if err != nil {
		t.Fatal(err)
	}

	request := dumpCaptureRequest("case-revision", "Scoped material.")
	request.CaseSlug, request.Namespace, request.ExpectedRevision = metadata.Slug, "team", metadata.SourceHash

	metadata.Description = "Changed case context."
	if _, err := owner.Save(metadata); err != nil {
		t.Fatal(err)
	}

	if _, err := MutateDump(t.Context(), project, request, nil); !errors.Is(err, workflowstate.ErrConflict) {
		t.Fatal("stale case capture accepted", err)
	}
}
