package authoring

import (
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
	workspacecore "forge.adrianpk.com/chr/internal/core/workspace"
)

func telegramOperationCaseRequest() CaseMutationRequest {
	return CaseMutationRequest{Binding: workflowstate.MutationBinding{RequestID: "case-request", Operator: "maintainer"}, ProjectID: "project-1", Namespace: "default", Metadata: CaseMetadata{Name: "Example case", Description: "Original description", Context: "Original context"}}
}

// TestTelegramOperationCaseReplay returns the original case after reopening and a later independent edit.
func TestTelegramOperationCaseReplay(t *testing.T) {
	owner := NewCaseMetadataService(t.TempDir())
	request := telegramOperationCaseRequest()

	first, err := owner.SaveBound(t.Context(), request)
	if err != nil {
		t.Fatal(err)
	}

	changed := first.Metadata

	changed.Description = "Later edit"
	if _, err := owner.Save(changed); err != nil {
		t.Fatal(err)
	}

	reopened := NewCaseMetadataService(owner.projectDir)

	again, err := reopened.SaveBound(t.Context(), request)
	if err != nil || !reflect.DeepEqual(again, first) {
		t.Fatal("case replay changed original result", again, first, err)
	}

	request.Metadata.Name = "A different request"
	if _, err := reopened.SaveBound(t.Context(), request); !errors.Is(err, workflowstate.ErrConflict) {
		t.Fatal("changed case arguments accepted", err)
	}

	cases, err := owner.LoadAll()
	if err != nil || len(cases) != 1 || cases[0].Description != "Later edit" {
		t.Fatal("replay duplicated or overwrote case", cases, err)
	}
}

// TestTelegramOperationCaseCrash recovers both write boundaries through the existing metadata owner.
func TestTelegramOperationCaseCrash(t *testing.T) {
	for index, name := range []string{"case image", "receipt"} {
		t.Run(name, func(t *testing.T) {
			owner := NewCaseMetadataService(t.TempDir())
			owner.afterMutationReplace = func(n int) error {
				if n == index {
					return errors.New("lost acknowledgement")
				}

				return nil
			}

			partial, err := owner.SaveBound(t.Context(), telegramOperationCaseRequest())
			if err == nil {
				t.Fatal("fault did not fire")
			}

			reopened := NewCaseMetadataService(owner.projectDir)

			result, err := reopened.SaveBound(t.Context(), telegramOperationCaseRequest())
			if err != nil || result.Metadata.Slug != partial.Metadata.Slug {
				t.Fatal("case recovery changed allocation", result, partial, err)
			}

			checkTelegramOperationCaseCount(t, reopened)
		})
	}
}

// TestTelegramOperationCaseConflict preserves a divergent pending image, including against ordinary saves.
func TestTelegramOperationCaseConflict(t *testing.T) {
	owner := NewCaseMetadataService(t.TempDir())
	owner.afterMutationReplace = func(int) error { return errors.New("lost acknowledgement") }

	partial, err := owner.SaveBound(t.Context(), telegramOperationCaseRequest())
	if err == nil {
		t.Fatal("fault did not fire")
	}

	target := filepath.Join(workspacecore.UserPath(owner.projectDir), partial.Metadata.SourcePath)
	if err := os.WriteFile(target, []byte("divergent"), 0o644); err != nil {
		t.Fatal(err)
	}

	reopened := NewCaseMetadataService(owner.projectDir)
	if _, err := reopened.SaveBound(t.Context(), telegramOperationCaseRequest()); !errors.Is(err, workflowstate.ErrConflict) {
		t.Fatal("divergent case accepted", err)
	}

	if _, err := reopened.Save(CaseMetadata{Name: "Another case"}); !errors.Is(err, workflowstate.ErrConflict) {
		t.Fatal("ordinary save bypassed pending journal", err)
	}

	data, err := os.ReadFile(target)
	if err != nil || string(data) != "divergent" {
		t.Fatal("divergent case was overwritten", err)
	}
}

// TestTelegramOperationCaseGuards rejects stale updates, changed failed requests and escaped paths.
func TestTelegramOperationCaseGuards(t *testing.T) {
	owner := NewCaseMetadataService(t.TempDir())
	request := telegramOperationCaseRequest()

	first, err := owner.SaveBound(t.Context(), request)
	if err != nil {
		t.Fatal(err)
	}

	update := request
	update.Binding.RequestID = "update"
	update.Metadata = first.Metadata

	update.Metadata.Description = "Updated description"
	if _, err := owner.SaveBound(t.Context(), update); err != nil {
		t.Fatal(err)
	}

	update.Binding.RequestID = "stale"
	if _, err := owner.SaveBound(t.Context(), update); !errors.Is(err, workflowstate.ErrConflict) {
		t.Fatal("stale image accepted", err)
	}

	update.Metadata.SourcePath = filepath.Join(owner.projectDir, "escaped.md")
	if _, err := owner.SaveBound(t.Context(), update); err == nil {
		t.Fatal("escaped path accepted")
	}

	invalid := request
	invalid.Binding.RequestID = "failed"

	invalid.Metadata.Name = ""
	if _, err := owner.SaveBound(t.Context(), invalid); err == nil {
		t.Fatal("invalid name accepted")
	}

	invalid.Metadata.Name = "Different"
	if _, err := owner.SaveBound(t.Context(), invalid); !errors.Is(err, workflowstate.ErrConflict) {
		t.Fatal("failed request arguments replaced", err)
	}
}

// TestTelegramOperationCaseConcurrent serializes caller-neutral creation through separate owners.
func TestTelegramOperationCaseConcurrent(t *testing.T) {
	root := t.TempDir()
	results := make(chan error, 2)

	for range 2 {
		go func() {
			_, err := NewCaseMetadataService(root).SaveBound(t.Context(), telegramOperationCaseRequest())
			results <- err
		}()
	}

	first, second := <-results, <-results
	if first != nil || second != nil {
		t.Fatal("concurrent case mutation failed", first, second)
	}

	cases, err := NewCaseMetadataService(root).LoadAll()
	if err != nil || len(cases) != 1 || !strings.Contains(cases[0].SourcePath, "example-case.md") {
		t.Fatal("case concurrency duplicated allocation", cases, err)
	}
}

func checkTelegramOperationCaseCount(t *testing.T, owner CaseMetadataService) {
	t.Helper()

	cases, err := owner.LoadAll()
	if err != nil || len(cases) != 1 {
		t.Fatal("case recovery duplicated effect", cases, err)
	}
}

// TestTelegramOperationCaseReadScope refuses file and directory symlinks outside the case owner.
func TestTelegramOperationCaseReadScope(t *testing.T) {
	outside := NewCaseMetadataService(t.TempDir())

	metadata, err := outside.Save(CaseMetadata{Name: "Outside"})
	if err != nil {
		t.Fatal(err)
	}

	for _, kind := range []string{"file", "directory"} {
		t.Run(kind, func(t *testing.T) {
			owner := NewCaseMetadataService(t.TempDir())

			linkTelegramOperationCase(t, owner, outside, metadata, kind)

			if _, err := owner.LoadBounded(t.Context(), 4096); err == nil {
				t.Fatal("case read escaped its owner")
			}
		})
	}
}

func linkTelegramOperationCase(t *testing.T, owner, outside CaseMetadataService, metadata CaseMetadata, kind string) {
	t.Helper()

	root := CaseMetadataDirectory(owner.projectDir)
	if err := os.MkdirAll(filepath.Dir(root), 0o755); err != nil {
		t.Fatal(err)
	}

	target, link := CaseMetadataDirectory(outside.projectDir), root
	if kind == "file" {
		if err := os.MkdirAll(root, 0o755); err != nil {
			t.Fatal(err)
		}

		target, link = metadata.SourcePath, filepath.Join(root, "outside.md")
	}

	if err := os.Symlink(target, link); err != nil {
		t.Fatal(err)
	}
}
