package core

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"forge.adrianpk.com/chr/internal/core/contenthash"
	opsmetadata "forge.adrianpk.com/chr/internal/core/ops/metadata"
	quizcore "forge.adrianpk.com/chr/internal/core/quiz"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

// Manual creates allocate distinct documents in the same scope and honor direct
// approval. Duplicate operations acknowledge their original committed identity.
func TestSpecMutationCreate(t *testing.T) {
	project := t.TempDir()
	request := specTestRequest("create-one")

	first := specTestMutate(t, project, request)
	if again := specTestMutate(t, project, request); again != first {
		t.Fatal(first, again)
	}

	request.OperationID, request.Status = "create-two", "approved"

	second := specTestMutate(t, project, request)
	if second.Reference == first.Reference {
		t.Fatal("independent specs merged")
	}

	for _, want := range []struct {
		result SpecMutationResult
		status string
	}{{first, "draft"}, {second, "approved"}} {
		document := specTestRead(t, project, want.result.Reference)
		if document.Status != want.status || document.metadata["status"] != want.status || document.Body != request.Body {
			t.Fatal(document)
		}
	}

	request.Body = "Changed arguments"
	if _, err := MutateSpec(t.Context(), project, request); !errors.Is(err, workflowstate.ErrConflict) {
		t.Fatal("changed retry accepted", err)
	}
}

// Saving a generated identity preserves unknown metadata, original lifecycle
// timestamps and retained provenance, and rejects stale or cross-scope writes.
func TestSpecMutationSave(t *testing.T) {
	project := t.TempDir()
	document := seedSpecGeneratedDocument(t, project)
	request := specTestRequest("save-generated")
	request.Action, request.Namespace, request.Reference, request.ExpectedRevision = "save", "", document.Reference, document.Revision
	request.Title, request.Body, request.Status = "Corrected scope", "Corrected opening text.\n\nExtra requirements.", "approved"
	result := specTestMutate(t, project, request)

	after := specTestRead(t, project, result.Reference)
	if after.Status != "approved" || after.Body != request.Body || after.provenance != document.provenance || after.metadata["custom"] == nil || !strings.Contains(after.front, "custom_front: retained") {
		t.Fatal(after)
	}

	for _, key := range []string{"created_at", "source_request_id", "source_response_id"} {
		beforeFields := specTestFields(t, document.front)

		afterFields := specTestFields(t, after.front)
		if beforeFields[key] != afterFields[key] {
			t.Fatal("provenance changed", key)
		}
	}

	request.OperationID = "stale-save"
	if _, err := MutateSpec(t.Context(), project, request); !errors.Is(err, workflowstate.ErrConflict) {
		t.Fatal(err)
	}
}

// Approval aliases normalize only in the read projection. A metadata-only edit
// also changes the revision, so an editor cannot overwrite it unnoticed.
func TestSpecLifecycleAliases(t *testing.T) {
	for _, status := range []string{"accepted", "final", "approved", "draft"} {
		t.Run(status, func(t *testing.T) {
			project := t.TempDir()
			result := specTestMutate(t, project, specTestRequest("alias"))
			document := specTestRead(t, project, result.Reference)
			content := strings.Replace(string(document.images[0].After), "status: "+opsmetadata.Quote("draft"), "status: "+opsmetadata.Quote(status), 1)

			path := filepath.Join(project, result.Reference)
			if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
				t.Fatal(err)
			}

			after := specTestRead(t, project, result.Reference)

			want := "approved"
			if status == "draft" {
				want = "draft"
			}

			if after.Status != want || string(after.images[0].After) != content {
				t.Fatal(after)
			}
		})
	}
}

// Competing saves from one observation commit exactly one replacement.
func TestSpecMutationConcurrent(t *testing.T) {
	project := t.TempDir()
	result := specTestMutate(t, project, specTestRequest("initial"))
	document := specTestRead(t, project, result.Reference)
	results := make(chan error, 2)

	var workers sync.WaitGroup
	for _, id := range []string{"left", "right"} {
		workers.Go(func() {
			request := specTestRequest(id)
			request.Action, request.Namespace, request.Reference, request.ExpectedRevision = "save", "", document.Reference, document.Revision

			_, err := MutateSpec(t.Context(), project, request)
			results <- err
		})
	}

	workers.Wait()
	close(results)

	committed, conflicted := 0, 0

	for err := range results {
		if err == nil {
			committed++
		} else if errors.Is(err, workflowstate.ErrConflict) {
			conflicted++
		} else {
			t.Fatal(err)
		}
	}

	if committed != 1 || conflicted != 1 {
		t.Fatal(committed, conflicted)
	}
}

// Every journal image can be interrupted and recovered by the same explicit
// operation. Reads fail closed during partial publication; retry creates once.
func TestSpecMutationRecovery(t *testing.T) {
	for stop := range 3 {
		t.Run(string(rune('0'+stop)), func(t *testing.T) { checkSpecMutationRecovery(t, stop) })
	}
}

func checkSpecMutationRecovery(t *testing.T, stop int) {
	t.Helper()

	project := t.TempDir()
	request := specTestRequest("recover")

	encoded, err := json.Marshal(request)
	if err != nil {
		t.Fatal(err)
	}

	digest := contenthash.Sum(append([]byte(project+"\x00spec-write\x00"), encoded...))
	r := &fileAuthoringSessionRepository{projectDir: project, afterMutationReplace: func(index int) error {
		if index == stop {
			return errors.New("interrupted")
		}

		return nil
	}}

	var partial SpecMutationResult

	err = WithAuthoringMutation(t.Context(), project, func(ctx context.Context) error { partial, err = r.commitSpec(ctx, request, digest); return err })
	if err == nil {
		t.Fatal("fault not reached")
	}

	if _, err := ReadSpecDocument(t.Context(), project, partial.Reference); !errors.Is(err, workflowstate.ErrConflict) {
		t.Fatal(err)
	}

	result := specTestMutate(t, project, request)
	if result.Reference != partial.Reference {
		t.Fatal(result, partial)
	}

	specTestRead(t, project, result.Reference)
}

func specTestRequest(id string) SpecMutationRequest {
	return SpecMutationRequest{OperationID: id, Action: "create", Namespace: "default", ProjectID: "sample-project", Title: "Independent spec", Body: "Opening content.\n\nRetained requirements."}
}

func specTestMutate(t *testing.T, project string, request SpecMutationRequest) SpecMutationResult {
	t.Helper()

	result, err := MutateSpec(t.Context(), project, request)
	if err != nil {
		t.Fatal(err)
	}

	return result
}

func specTestRead(t *testing.T, project, reference string) SpecDocument {
	t.Helper()

	document, err := ReadSpecDocument(t.Context(), project, reference)
	if err != nil {
		t.Fatal(err)
	}

	return document
}

func seedSpecGeneratedDocument(t *testing.T, project string) SpecDocument {
	t.Helper()

	spec := GeneratedSpec{Scope: quizcore.Scope{ProjectID: "sample-project"}, Title: "Generated scope", Body: "Opening text.", SourceRequestID: "source-request", SourceResponseID: "source-response", SourcePaths: []string{"docs/context.md"}, CreatedAt: time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC)}

	paths, err := WriteGeneratedSpec(t.Context(), project, spec)
	if err != nil {
		t.Fatal(err)
	}

	document := specTestRead(t, project, paths.RelativePath)
	document.metadata["custom"] = map[string]any{"keep": true}

	encoded, err := json.Marshal(document.metadata)
	if err != nil {
		t.Fatal(err)
	}

	if err := os.WriteFile(filepath.Join(filepath.Dir(paths.Path), "metadata.json"), encoded, 0o644); err != nil {
		t.Fatal(err)
	}

	content := strings.Replace(string(document.images[0].After), "---\n", "---\ncustom_front: retained\n", 1)
	if err := os.WriteFile(paths.Path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	document = specTestRead(t, project, paths.RelativePath)

	return document
}

func specTestFields(t *testing.T, front string) map[string]string {
	t.Helper()

	fields, err := ProjectDocumentFields(front)
	if err != nil {
		t.Fatal(err)
	}

	return fields
}

// Invalid scope, lifecycle, identity and capacity inputs cannot allocate a spec.
func TestSpecMutationGuards(t *testing.T) {
	cases := []struct {
		name   string
		change func(*SpecMutationRequest)
	}{
		{"title", func(r *SpecMutationRequest) { r.Title = "" }},
		{"body", func(r *SpecMutationRequest) { r.Body = "" }},
		{"body bound", func(r *SpecMutationRequest) { r.Body = strings.Repeat("x", workflowstate.FileLimit+1) }},
		{"alias write", func(r *SpecMutationRequest) { r.Status = "accepted" }},
		{"namespace", func(r *SpecMutationRequest) { r.Namespace = "../other" }},
		{"missing case", func(r *SpecMutationRequest) { r.CaseSlug = "missing" }},
		{"identity", func(r *SpecMutationRequest) { r.OperationID = "../bad" }},
		{"foreign reference", func(r *SpecMutationRequest) {
			r.Action = "save"
			r.Namespace = ""
			r.ExpectedRevision = strings.Repeat("a", 64)
			r.Reference = "ops/default/dump/00000-example/artifact.md"
		}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			request := specTestRequest("reject")
			tc.change(&request)

			if _, err := MutateSpec(t.Context(), t.TempDir(), request); err == nil {
				t.Fatal("invalid mutation accepted")
			}
		})
	}
}

// Metadata-only changes invalidate the same revision used by the body editor.
func TestSpecMetadataRevision(t *testing.T) {
	project := t.TempDir()
	result := specTestMutate(t, project, specTestRequest("before"))
	before := specTestRead(t, project, result.Reference)

	metadata := append(append([]byte(nil), before.images[1].After...), byte('\n'))
	if err := os.WriteFile(filepath.Join(project, filepath.Dir(result.Reference), "metadata.json"), metadata, 0o644); err != nil {
		t.Fatal(err)
	}

	request := specTestRequest("stale-metadata")

	request.Action, request.Namespace, request.Reference, request.ExpectedRevision = "save", "", result.Reference, before.Revision
	if _, err := MutateSpec(t.Context(), project, request); !errors.Is(err, workflowstate.ErrConflict) {
		t.Fatal(err)
	}

	after := specTestRead(t, project, result.Reference)
	if string(after.images[0].After) != string(before.images[0].After) {
		t.Fatal("stale save replaced the body")
	}
}
