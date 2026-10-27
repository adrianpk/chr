package core

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	workspacecore "forge.adrianpk.com/chr/internal/core/workspace"
)

func TestWorkspaceSemanticDigestCacheRoundTrip(t *testing.T) {
	projectDir := t.TempDir()
	cachePath := WorkspaceSemanticDigestCachePath(projectDir)
	cache := NewWorkspaceSemanticDigestCache()
	node := WorkspaceNode{
		Kind:         WorkspaceNodeSpec,
		RelativePath: filepath.ToSlash(filepath.Join(workspacecore.DirectoryName, workspacecore.UserDirectoryName, "specs", "blog.md")),
		Title:        "Blog Spec",
		SourceHash:   "source-1",
	}

	cache.Entries[WorkspaceNodeID(node)] = NewWorkspaceSemanticDigest(
		node,
		"model-a",
		"prompt-a",
		"digest-a",
		time.Date(2026, 5, 23, 12, 0, 0, 0, time.UTC),
	)

	err := WriteWorkspaceSemanticDigestCache(cachePath, cache)
	if err != nil {
		t.Fatalf("WriteWorkspaceSemanticDigestCache returned error: %v", err)
	}

	loaded, err := ReadWorkspaceSemanticDigestCache(cachePath)
	if err != nil {
		t.Fatalf("ReadWorkspaceSemanticDigestCache returned error: %v", err)
	}

	entry := loaded.Entries[WorkspaceNodeID(node)]
	if entry.Digest != "digest-a" {
		t.Fatalf("Digest = %q, want digest-a", entry.Digest)
	}

	if entry.ModelID != "model-a" || entry.PromptHash != "prompt-a" {
		t.Fatalf("entry = %#v, want model and prompt metadata", entry)
	}
}

func TestWorkspaceSemanticDigestCacheMissingAndInvalidFilesAreEmpty(t *testing.T) {
	projectDir := t.TempDir()
	cachePath := WorkspaceSemanticDigestCachePath(projectDir)

	missing, err := ReadWorkspaceSemanticDigestCache(cachePath)
	if err != nil {
		t.Fatalf("Read missing cache returned error: %v", err)
	}

	if len(missing.Entries) != 0 {
		t.Fatalf("missing cache entries = %#v, want empty", missing.Entries)
	}

	writeWorkspaceGraphFile(t, cachePath, "{not-json")

	invalid, err := ReadWorkspaceSemanticDigestCache(cachePath)
	if err != nil {
		t.Fatalf("Read invalid cache returned error: %v", err)
	}

	if len(invalid.Entries) != 0 {
		t.Fatalf("invalid cache entries = %#v, want empty", invalid.Entries)
	}
}

func TestWorkspaceSemanticDigestStatesMarkPendingReadyAndStale(t *testing.T) {
	index := WorkspaceIndex{
		NodesByID: map[string]WorkspaceNode{
			"spec:ready.md": {
				Kind:       WorkspaceNodeSpec,
				SourceHash: "source-ready",
			},
			"spec:stale.md": {
				Kind:       WorkspaceNodeSpec,
				SourceHash: "source-new",
			},
			"spec:pending.md": {
				Kind:       WorkspaceNodeSpec,
				SourceHash: "source-pending",
			},
		},
	}
	cache := NewWorkspaceSemanticDigestCache()
	cache.Entries["spec:ready.md"] = WorkspaceSemanticDigest{
		NodeID:     "spec:ready.md",
		ModelID:    "model-a",
		PromptHash: "prompt-a",
		SourceHash: "source-ready",
		Digest:     "ready",
	}
	cache.Entries["spec:stale.md"] = WorkspaceSemanticDigest{
		NodeID:     "spec:stale.md",
		ModelID:    "model-a",
		PromptHash: "prompt-a",
		SourceHash: "source-old",
		Digest:     "stale",
	}

	states := WorkspaceSemanticDigestStates(index, cache, "model-a", "prompt-a")

	assertSemanticDigestStatus(t, states, "spec:ready.md", WorkspaceSemanticDigestReady)
	assertSemanticDigestStatus(t, states, "spec:stale.md", WorkspaceSemanticDigestStale)
	assertSemanticDigestStatus(t, states, "spec:pending.md", WorkspaceSemanticDigestPending)
}

func TestDeterministicWorkspaceSemanticDigestGenerator(t *testing.T) {
	generator := DeterministicWorkspaceSemanticDigestGenerator{}
	node := WorkspaceNode{
		Kind:         WorkspaceNodeSpec,
		RelativePath: filepath.ToSlash(filepath.Join(workspacecore.DirectoryName, workspacecore.UserDirectoryName, "specs", "blog.md")),
		Title:        "Blog Spec",
		Status:       "accepted",
		CaseSlug:     "blog",
		SourceHash:   "source-1",
	}

	request := WorkspaceSemanticDigestRequest{
		NodeID:     WorkspaceNodeID(node),
		Node:       node,
		ModelID:    "model-a",
		PromptHash: "prompt-a",
	}

	first, err := generator.GenerateWorkspaceSemanticDigest(context.Background(), request)
	if err != nil {
		t.Fatalf("GenerateWorkspaceSemanticDigest returned error: %v", err)
	}

	second, err := generator.GenerateWorkspaceSemanticDigest(context.Background(), request)
	if err != nil {
		t.Fatalf("GenerateWorkspaceSemanticDigest returned error: %v", err)
	}

	if first == "" {
		t.Fatal("digest is empty")
	}

	if first != second {
		t.Fatalf("digest = %q then %q, want deterministic result", first, second)
	}
}

func TestWorkspaceSemanticIndexWorkerGeneratesPendingAndStaleDigests(t *testing.T) {
	now := time.Date(2026, 5, 23, 13, 0, 0, 0, time.UTC)
	index := WorkspaceIndex{
		NodesByID: map[string]WorkspaceNode{
			"spec:ready.md": {
				Kind:         WorkspaceNodeSpec,
				RelativePath: "specs/ready.md",
				SourceHash:   "source-ready",
			},
			"spec:stale.md": {
				Kind:         WorkspaceNodeSpec,
				RelativePath: "specs/stale.md",
				SourceHash:   "source-new",
			},
			"spec:pending.md": {
				Kind:         WorkspaceNodeSpec,
				RelativePath: "specs/pending.md",
				SourceHash:   "source-pending",
			},
		},
	}
	cache := NewWorkspaceSemanticDigestCache()
	cache.Entries["spec:ready.md"] = WorkspaceSemanticDigest{
		NodeID:     "spec:ready.md",
		ModelID:    "model-a",
		PromptHash: "prompt-a",
		SourceHash: "source-ready",
		Digest:     "ready",
	}
	cache.Entries["spec:stale.md"] = WorkspaceSemanticDigest{
		NodeID:     "spec:stale.md",
		ModelID:    "model-a",
		PromptHash: "prompt-a",
		SourceHash: "source-old",
		Digest:     "stale",
	}
	worker := WorkspaceSemanticIndexWorker{
		Generator:  DeterministicWorkspaceSemanticDigestGenerator{},
		ModelID:    "model-a",
		PromptHash: "prompt-a",
		Now: func() time.Time {
			return now
		},
	}

	result, err := worker.Generate(context.Background(), index, cache)
	if err != nil {
		t.Fatalf("Generate returned error: %v", err)
	}

	if result.ReadyCount != 1 || result.PendingCount != 1 || result.StaleCount != 1 {
		t.Fatalf("result counts = %#v, want one ready, pending, and stale", result)
	}

	if result.GeneratedCount != 2 || len(result.Generated) != 2 {
		t.Fatalf("generated = %#v, want two generated digests", result.Generated)
	}

	assertGeneratedSemanticDigest(t, result.Cache, "spec:pending.md", "source-pending", now)
	assertGeneratedSemanticDigest(t, result.Cache, "spec:stale.md", "source-new", now)
}

func TestWorkspaceIndexHealthWithSemanticDigestStates(t *testing.T) {
	index := WorkspaceIndex{
		NodesByID: map[string]WorkspaceNode{
			"spec:one.md":   {},
			"spec:two.md":   {},
			"spec:three.md": {},
		},
	}
	states := []WorkspaceSemanticDigestState{
		{NodeID: "spec:one.md", Status: WorkspaceSemanticDigestReady},
		{NodeID: "spec:two.md", Status: WorkspaceSemanticDigestPending},
		{NodeID: "spec:three.md", Status: WorkspaceSemanticDigestStale},
	}

	health := index.HealthWithSemanticDigestStates(states)

	if health.SemanticDigestPendingCount != 1 {
		t.Fatalf("SemanticDigestPendingCount = %d, want 1", health.SemanticDigestPendingCount)
	}

	if health.SemanticDigestStaleCount != 1 {
		t.Fatalf("SemanticDigestStaleCount = %d, want 1", health.SemanticDigestStaleCount)
	}
}

func TestWorkspaceSemanticDigestCachePath(t *testing.T) {
	projectDir := t.TempDir()
	want := filepath.Join(projectDir, workspacecore.DirectoryName, workspacecore.UserDirectoryName, "tmp", "cache", "workspace-index", "semantic-digests.json")

	if WorkspaceSemanticDigestCachePath(projectDir) != want {
		t.Fatalf("WorkspaceSemanticDigestCachePath = %q, want %q", WorkspaceSemanticDigestCachePath(projectDir), want)
	}

	err := os.MkdirAll(filepath.Dir(want), 0o755)
	if err != nil {
		t.Fatalf("MkdirAll returned error: %v", err)
	}
}

func assertGeneratedSemanticDigest(
	t *testing.T,
	cache WorkspaceSemanticDigestCache,
	nodeID string,
	sourceHash string,
	generatedAt time.Time,
) {
	t.Helper()

	entry := cache.Entries[nodeID]
	if entry.Digest == "" {
		t.Fatalf("digest for %s is empty", nodeID)
	}

	if entry.SourceHash != sourceHash {
		t.Fatalf("SourceHash for %s = %q, want %q", nodeID, entry.SourceHash, sourceHash)
	}

	if !entry.GeneratedAt.Equal(generatedAt) {
		t.Fatalf("GeneratedAt for %s = %s, want %s", nodeID, entry.GeneratedAt, generatedAt)
	}
}

func assertSemanticDigestStatus(
	t *testing.T,
	states []WorkspaceSemanticDigestState,
	nodeID string,
	want WorkspaceSemanticDigestStatus,
) {
	t.Helper()

	for _, state := range states {
		if state.NodeID == nodeID {
			if state.Status != want {
				t.Fatalf("status for %s = %q, want %q", nodeID, state.Status, want)
			}

			return
		}
	}

	t.Fatalf("missing digest state for %s in %#v", nodeID, states)
}
