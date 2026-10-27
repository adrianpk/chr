package core

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	opsnamespace "forge.adrianpk.com/chr/internal/core/ops/namespace"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

// First and concurrent later captures share one destination even when the
// caller's active namespace changes. Every contribution and raw source survives.
func TestDumpMutationCapture(t *testing.T) {
	project := t.TempDir()
	t.Setenv(opsnamespace.UserEnv, "operator-space")

	first, err := MutateDump(t.Context(), project, dumpCaptureRequest("first", "First material."), nil)
	if err != nil || first.Destination != "ops/default/dump/00000-project-context/artifact.md" {
		t.Fatal("first capture did not use default project namespace", first, err)
	}

	t.Setenv(opsnamespace.UserEnv, "another-space")

	var group sync.WaitGroup
	for _, id := range []string{"second", "third"} {
		group.Go(func() {
			result, err := MutateDump(t.Context(), project, dumpCaptureRequest(id, id+" material."), nil)
			if err != nil || result.Destination != first.Destination {
				t.Error("concurrent capture changed destination", result, err)
			}
		})
	}

	group.Wait()

	checkDumpMutationContributions(t, project, first.Destination)
}

func checkDumpMutationContributions(t *testing.T, project, reference string) {
	t.Helper()

	observed, err := ObserveProjectDump(t.Context(), project)
	if err != nil || observed.Destination != reference || len(observed.Candidates) != 1 || len(observed.Components) != 3 {
		t.Fatal("capture fragmented the logical dump", observed, err)
	}

	document, err := ReadDumpDocument(t.Context(), project, reference)
	if err != nil {
		t.Fatal(err)
	}

	for _, text := range []string{"First material.", "second material.", "third material."} {
		if strings.Count(document.Content, text) != 1 {
			t.Fatal("a contribution was lost or duplicated", text, document.Content)
		}
	}
}

func dumpCaptureRequest(id, content string) DumpMutationRequest {
	return DumpMutationRequest{OperationID: id, Action: "capture", Title: "Project Context", Content: content}
}

// Capture adopts an existing identity in its original namespace. Full-body Save
// replaces its editable body, keeps unknown frontmatter/metadata, and fences stale
// edits after another contribution instead of appending the whole edited body.
func TestDumpMutationSave(t *testing.T) {
	project := t.TempDir()
	reference := projectDumpTestArtifact(t, project, "team", "Existing", "---\ntitle: Existing\ncustom:\n  retained: true\n---\nOriginal material.\n")
	patchProjectDumpTestMetadata(t, project, reference, map[string]any{"custom": map[string]any{"retained": true}})

	result, err := MutateDump(t.Context(), project, dumpCaptureRequest("adopt", "Captured material."), nil)
	if err != nil || result.Destination != reference {
		t.Fatal("capture replaced an existing identity", result, err)
	}

	before, err := ReadDumpDocument(t.Context(), project, reference)
	if err != nil {
		t.Fatal(err)
	}

	request := DumpMutationRequest{OperationID: "save", Action: "save", Reference: reference, ExpectedRevision: before.Revision, Content: "Replacement body.\n"}
	if _, err := MutateDump(t.Context(), project, request, nil); err != nil {
		t.Fatal(err)
	}

	after, err := ReadDumpDocument(t.Context(), project, reference)
	if err != nil || !strings.Contains(after.Content, "custom:\n  retained: true") || !strings.HasSuffix(after.Content, "Replacement body.\n") || strings.Contains(after.Content, "Captured material.") {
		t.Fatal("save failed body replacement or frontmatter retention", after, err)
	}

	checkDumpMutationSaveHistory(t, project, before)

	request.OperationID = "stale-save"
	if _, err := MutateDump(t.Context(), project, request, nil); !errors.Is(err, workflowstate.ErrConflict) {
		t.Fatal("stale body save accepted", err)
	}
}

func checkDumpMutationSaveHistory(t *testing.T, project string, before DumpDocumentObservation) {
	t.Helper()

	r := &fileAuthoringSessionRepository{projectDir: project}

	history, err := r.readAuthoringDocument(projectDumpHistoryRef(before.Artifact) + "/artifact.md")
	if err != nil || string(history) != before.Content {
		t.Fatal("save did not preserve the exact previous image", err)
	}

	metadata, err := r.readAuthoringDocument(filepath.Join(filepath.Dir(before.Artifact.Reference), "metadata.json"))
	if err != nil || !strings.Contains(string(metadata), "\"custom\"") || !strings.Contains(string(metadata), "\"dump_captures\"") {
		t.Fatal("save lost unknown metadata or capture provenance", err)
	}
}

// Named captures validate an existing scope, preserve its namespace and remain
// independent documents rather than acquiring a case singleton rule.
func TestDumpMutationCase(t *testing.T) {
	project := t.TempDir()
	if _, err := NewCaseMetadataService(project).Save(CaseMetadata{Name: "Named Work"}); err != nil {
		t.Fatal(err)
	}

	var references []string

	for _, id := range []string{"first", "second"} {
		request := dumpCaptureRequest(id, "Scoped material.")
		request.CaseSlug, request.Namespace = "named-work", "team"

		result, err := MutateDump(t.Context(), project, request, nil)
		if err != nil || !strings.HasPrefix(result.Destination, "ops/team/dump/") {
			t.Fatal("case capture lost scope namespace", result, err)
		}

		references = append(references, result.Destination)
	}

	if references[0] == references[1] {
		t.Fatal("case captures were made into a singleton")
	}

	request := dumpCaptureRequest("unknown", "Unroutable material.")

	request.CaseSlug, request.Namespace = "unknown", "team"
	if _, err := MutateDump(t.Context(), project, request, nil); err == nil {
		t.Fatal("unknown case created implicitly")
	}

	if _, err := os.Stat(ProjectMetadataPath(project)); !errors.Is(err, os.ErrNotExist) {
		t.Fatal("scoped captures published a general pointer", err)
	}
}

// Append accepts one observed source and preserves a retry after a later edit.
// Reusing that identity with a different payload remains a conflict.
func TestDumpMutationAppend(t *testing.T) {
	project := t.TempDir()

	first, err := MutateDump(t.Context(), project, dumpCaptureRequest("first", "Initial material."), nil)
	if err != nil {
		t.Fatal(err)
	}

	document, err := ReadDumpDocument(t.Context(), project, first.Destination)
	if err != nil {
		t.Fatal(err)
	}

	request := DumpMutationRequest{OperationID: "append", Action: "append", Reference: first.Destination, ExpectedRevision: document.Revision, Content: "Appended material."}

	appended, err := MutateDump(t.Context(), project, request, nil)
	if err != nil {
		t.Fatal(err)
	}

	if _, err := MutateDump(t.Context(), project, dumpCaptureRequest("later", "Later material."), nil); err != nil {
		t.Fatal(err)
	}

	replayed, err := MutateDump(t.Context(), project, request, nil)
	if err != nil || replayed != appended {
		t.Fatal("retry lost its original result", replayed, err)
	}

	request.Content = "Changed append payload."
	if _, err := MutateDump(t.Context(), project, request, nil); !errors.Is(err, workflowstate.ErrConflict) {
		t.Fatal("append operation identity was reused", err)
	}
}
