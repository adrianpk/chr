package app

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"testing"

	"forge.adrianpk.com/chr/internal/core"
	"forge.adrianpk.com/chr/internal/tui"
	"forge.adrianpk.com/chr/internal/webui"
)

type countingDumpCurator struct{ calls atomic.Int32 }

func (p *countingDumpCurator) CurateDump(_ context.Context, request DumpCurationRequest) (DumpCurationResult, error) {
	p.calls.Add(1)

	return DumpCurationResult{PolishedBody: request.Content}, nil
}

// Shared service captures preserve one identity across concurrent callers,
// namespace changes and an identical retry without invoking curation twice.
func TestServiceDumpCapture(t *testing.T) {
	operations := newTestService(t, newTestWorkspace(t))
	curator := &countingDumpCurator{}
	operations.dumpCurator = curator
	request := DumpCreateRequest{OperationID: "first", Title: "General", Content: "First material."}

	first, err := operations.CreateDump(request)
	if err != nil || !first.Committed {
		t.Fatal(first, err)
	}

	t.Setenv(core.OpsUserEnv, "another-namespace")

	replay, err := operations.CreateDump(request)
	if err != nil || replay.Mutation != first.Mutation || curator.calls.Load() != 1 {
		t.Fatal("retry repeated capture preparation", replay, err)
	}

	checkServiceDumpContributions(t, operations, first.RelativePath)
}

func checkServiceDumpContributions(t *testing.T, operations service, reference string) {
	t.Helper()

	var group sync.WaitGroup
	for _, id := range []string{"second", "third"} {
		group.Go(func() {
			result, err := operations.CreateDump(DumpCreateRequest{OperationID: id, Title: "Another capture", Content: id + " material."})
			if err != nil || !result.Committed || result.RelativePath != reference {
				t.Error(result, err)
			}
		})
	}

	group.Wait()

	catalog, err := operations.ListDumps()
	if err != nil || len(catalog.Dumps) != 1 || catalog.Dumps[0].RelativePath != reference {
		t.Fatal("catalog lost singleton identity", catalog, err)
	}

	for _, text := range []string{"First material.", "second material.", "third material."} {
		if strings.Count(catalog.Dumps[0].Content, text) != 1 {
			t.Fatal("contribution lost or duplicated", text)
		}
	}
}

// Index failure after canonical publication returns a committed result and
// content. An identical retry refreshes the projection without another capture.
func TestServiceDumpProjectionFailure(t *testing.T) {
	operations := newTestService(t, newTestWorkspace(t))
	curator := &countingDumpCurator{}
	operations.dumpCurator = curator
	cache := blockDumpIndex(t, operations)
	request := DumpCreateRequest{OperationID: "projection", Title: "General", Content: "Committed material."}

	result, err := operations.CreateDump(request)
	if err != nil || !result.Committed || !result.ProjectionStale || !strings.Contains(result.Content, request.Content) {
		t.Fatal("post-commit failure hid canonical success", result, err)
	}

	if err := os.Remove(cache); err != nil {
		t.Fatal(err)
	}

	replayed, err := operations.CreateDump(request)
	if err != nil || replayed.ProjectionStale || replayed.Mutation != result.Mutation || curator.calls.Load() != 1 {
		t.Fatal("projection retry repeated canonical effect", replayed, err)
	}
}

func blockDumpIndex(t *testing.T, operations service) string {
	t.Helper()

	cache := core.WorkspaceIndexCachePath(operations.projectContext.ProjectDir)
	if err := os.MkdirAll(cache, 0o755); err != nil {
		t.Fatal(err)
	}

	return cache
}

// Web writes compare metadata-aware revisions in the owner. Receipt replay
// precedes stale-source checking, including after a later independent edit.
func TestWebDumpReceipt(t *testing.T) {
	operations := newTestService(t, newTestWorkspace(t))

	created, err := operations.CreateDump(DumpCreateRequest{OperationID: "initial", Title: "General", Content: "Initial material."})
	if err != nil {
		t.Fatal(err)
	}

	request := webui.WriteActionRequest{ID: created.RelativePath, Revision: created.Revision, Fields: map[string]string{"request_id": "web-save", "content": "Saved material."}}
	provider := webViewProvider{}

	first, err := provider.writeDumpSave(operations, request)
	if err != nil || first.Status != webui.WriteStatusOK || first.Operation.Status != "committed" {
		t.Fatal(first, err)
	}

	checkWebDumpReplay(t, operations, request, first.Operation.ID)
}

func checkWebDumpReplay(t *testing.T, operations service, request webui.WriteActionRequest, operationID string) {
	t.Helper()

	provider := webViewProvider{}

	if _, err := operations.CreateDump(DumpCreateRequest{OperationID: "later", Title: "Later", Content: "Later material."}); err != nil {
		t.Fatal(err)
	}

	replay, err := provider.writeDumpSave(operations, request)
	if err != nil || replay.Status != webui.WriteStatusOK || replay.Operation.ID != operationID {
		t.Fatal("web receipt was blocked by a stale preflight", replay, err)
	}

	request.Fields["request_id"] = "new-stale-save"
	if stale, err := provider.writeDumpSave(operations, request); err != nil || stale.Status != webui.WriteStatusStale {
		t.Fatal("new stale save accepted", stale, err)
	}
}

// The application boundary requires caller bindings rather than silently
// manufacturing a revision that could authorize an overwrite of newer content.
func TestServiceDumpWriteGuards(t *testing.T) {
	operations := newTestService(t, newTestWorkspace(t))

	created, err := operations.CreateDump(DumpCreateRequest{OperationID: "initial", Title: "General", Content: "Initial material."})
	if err != nil {
		t.Fatal(err)
	}

	_, err = operations.SaveDump(DumpSaveRequest{OperationID: "unguarded", RelativePath: created.RelativePath, Content: "Unbound replacement."})
	if err == nil {
		t.Fatal("save manufactured a revision")
	}

	_, err = operations.SaveDump(DumpSaveRequest{OperationID: "guarded", ExpectedRevision: created.Revision, RelativePath: created.RelativePath, Content: "Bound replacement."})
	if err != nil {
		t.Fatal(err)
	}

	_, err = operations.AppendDump(DumpAppendRequest{OperationID: "stale", ExpectedRevision: created.Revision, RelativePath: created.RelativePath, Content: "Stale append."})
	if !errors.Is(err, core.ErrWorkflowConflict) {
		t.Fatal("stale append accepted", err)
	}
}

// Catalog and reader state comes from canonical artifacts, so namespace changes
// and a discarded index preserve scoped entries and hide superseded aliases.
func TestServiceDumpReconciliation(t *testing.T) {
	operations := newTestService(t, newTestWorkspace(t))
	project := operations.projectContext.ProjectDir
	first := serviceDumpArtifact(t, project, "team", "First", "First material.")
	second := serviceDumpArtifact(t, project, "default", "Second", "Second material.")
	serviceDumpArtifact(t, project, "team", "Named", "---\ncase: named-work\n---\nNamed material.")

	preview, err := operations.PreviewProjectDumpReconciliation()
	if err != nil || len(preview.Sources) != 2 {
		t.Fatal(preview, err)
	}

	request := core.ProjectDumpReconciliationRequest{OperationID: "reconcile", ExpectedRevision: preview.Revision, Destination: preview.Destination}
	if _, err := operations.ReconcileProjectDump(request); err != nil {
		t.Fatal(err)
	}

	t.Setenv(core.OpsUserEnv, "another-namespace")

	catalog, err := operations.ListDumps()
	if err != nil || len(catalog.Dumps) != 2 {
		t.Fatal("aliases or scoped entries confused in catalog", catalog, err)
	}

	checkServiceDumpAliases(t, operations, preview.Destination, []string{first, second})

	alias := first
	if alias == preview.Destination {
		alias = second
	}

	checkTUIDumpAliasSave(t, operations, alias, preview.Destination)
}

func checkTUIDumpAliasSave(t *testing.T, operations service, alias, destination string) {
	t.Helper()

	read, err := operations.ReadDump(DumpReadRequest{RelativePath: alias})
	if err != nil {
		t.Fatal(err)
	}

	executor := newTUIDumpActionExecutor(operations)

	result, err := executor.ExecuteDumpAction(tui.DumpActionRequest{Action: tui.DumpActionSave, OperationID: "alias-save", Revision: read.Revision, Path: alias, Content: "Updated canonical material."})
	if err != nil || result.SelectedPath != destination || !result.Committed {
		t.Fatal("TUI alias save lost its canonical selection", result, err)
	}
}

func checkServiceDumpAliases(t *testing.T, operations service, destination string, references []string) {
	t.Helper()

	for _, ref := range references {
		read, err := operations.ReadDump(DumpReadRequest{RelativePath: ref})
		if err != nil || read.RelativePath != destination || !strings.Contains(read.Content, "First material.") || !strings.Contains(read.Content, "Second material.") {
			t.Fatal("old reference did not resolve", read, err)
		}
	}
}

func serviceDumpArtifact(t *testing.T, project, namespace, title, body string) string {
	t.Helper()

	layout, err := core.ResolveOpsLayout(core.ResolveOpsLayoutOptions{ProjectDir: project, Namespace: namespace})
	if err != nil {
		t.Fatal(err)
	}

	artifact, err := core.CreateOpsArtifact(core.CreateOpsArtifactOptions{Layout: layout, Type: "dump", Title: title, Producer: "chr"})
	if err != nil {
		t.Fatal(err)
	}

	path := filepath.Join(artifact.Dir, core.OpsArtifactBodyFile())
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}

	reference, err := filepath.Rel(project, path)
	if err != nil {
		t.Fatal(err)
	}

	return filepath.ToSlash(reference)
}
