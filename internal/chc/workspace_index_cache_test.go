package chc

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestWorkspaceGraphBuilderBuildCachedIndexMissWritesCache(t *testing.T) {
	projectDir := t.TempDir()
	writeCaseRootedWorkspaceGraphFixture(t, projectDir)

	result, err := NewWorkspaceGraphBuilder(projectDir).BuildCachedIndex(context.Background())
	if err != nil {
		t.Fatalf("BuildCachedIndex returned error: %v", err)
	}

	if result.Cache.Status != WorkspaceIndexCacheMiss {
		t.Fatalf("Cache.Status = %q, want %q", result.Cache.Status, WorkspaceIndexCacheMiss)
	}

	if result.Cache.Path != WorkspaceIndexCachePath(projectDir) {
		t.Fatalf("Cache.Path = %q, want %q", result.Cache.Path, WorkspaceIndexCachePath(projectDir))
	}

	assertWorkspaceIndexCacheFileExists(t, projectDir)
	assertWorkspaceIndexLookups(t, result.Index, result.Graph)
}

func TestWorkspaceGraphBuilderBuildCachedIndexHitsValidCache(t *testing.T) {
	projectDir := t.TempDir()
	writeCaseRootedWorkspaceGraphFixture(t, projectDir)

	first, err := NewWorkspaceGraphBuilder(projectDir).BuildCachedIndex(context.Background())
	if err != nil {
		t.Fatalf("first BuildCachedIndex returned error: %v", err)
	}

	second, err := NewWorkspaceGraphBuilder(projectDir).BuildCachedIndex(context.Background())
	if err != nil {
		t.Fatalf("second BuildCachedIndex returned error: %v", err)
	}

	if second.Cache.Status != WorkspaceIndexCacheHit {
		t.Fatalf("Cache.Status = %q, want %q", second.Cache.Status, WorkspaceIndexCacheHit)
	}

	if !second.Index.BuiltAt.Equal(first.Index.BuiltAt) {
		t.Fatalf("Index.BuiltAt = %s, want cached %s", second.Index.BuiltAt, first.Index.BuiltAt)
	}
}

func TestWorkspaceGraphBuilderBuildCachedIndexRewritesStaleCache(t *testing.T) {
	projectDir := t.TempDir()
	writeCaseRootedWorkspaceGraphFixture(t, projectDir)

	first, err := NewWorkspaceGraphBuilder(projectDir).BuildCachedIndex(context.Background())
	if err != nil {
		t.Fatalf("first BuildCachedIndex returned error: %v", err)
	}

	writeWorkspaceGraphFile(t, filepath.Join(projectDir, DirectoryName, "specs", "blog-spec.md"), `---
title: "Updated Blog Spec"
status: "accepted"
case: "blog"
---

# Updated Blog Spec
`)

	second, err := NewWorkspaceGraphBuilder(projectDir).BuildCachedIndex(context.Background())
	if err != nil {
		t.Fatalf("second BuildCachedIndex returned error: %v", err)
	}

	if second.Cache.Status != WorkspaceIndexCacheStale {
		t.Fatalf("Cache.Status = %q, want %q", second.Cache.Status, WorkspaceIndexCacheStale)
	}

	if second.Index.BuiltAt.Equal(first.Index.BuiltAt) {
		t.Fatalf("Index.BuiltAt = %s, want rewritten index after stale cache", second.Index.BuiltAt)
	}

	specID := second.Index.NodesByPath[filepath.ToSlash(filepath.Join(DirectoryName, "specs", "blog-spec.md"))]
	if second.Index.NodesByID[specID].Title != "Updated Blog Spec" {
		t.Fatalf("spec title = %q, want Updated Blog Spec", second.Index.NodesByID[specID].Title)
	}
}

func TestWorkspaceGraphBuilderBuildCachedIndexMissesAfterCacheDeletion(t *testing.T) {
	projectDir := t.TempDir()
	writeCaseRootedWorkspaceGraphFixture(t, projectDir)

	_, err := NewWorkspaceGraphBuilder(projectDir).BuildCachedIndex(context.Background())
	if err != nil {
		t.Fatalf("first BuildCachedIndex returned error: %v", err)
	}

	err = os.Remove(WorkspaceIndexCachePath(projectDir))
	if err != nil {
		t.Fatalf("Remove cache returned error: %v", err)
	}

	second, err := NewWorkspaceGraphBuilder(projectDir).BuildCachedIndex(context.Background())
	if err != nil {
		t.Fatalf("second BuildCachedIndex returned error: %v", err)
	}

	if second.Cache.Status != WorkspaceIndexCacheMiss {
		t.Fatalf("Cache.Status = %q, want %q", second.Cache.Status, WorkspaceIndexCacheMiss)
	}
}

func TestWorkspaceGraphBuilderBuildCachedIndexRewritesInvalidCache(t *testing.T) {
	projectDir := t.TempDir()
	writeCaseRootedWorkspaceGraphFixture(t, projectDir)
	writeWorkspaceGraphFile(t, WorkspaceIndexCachePath(projectDir), "{not-json")

	result, err := NewWorkspaceGraphBuilder(projectDir).BuildCachedIndex(context.Background())
	if err != nil {
		t.Fatalf("BuildCachedIndex returned error: %v", err)
	}

	if result.Cache.Status != WorkspaceIndexCacheMiss {
		t.Fatalf("Cache.Status = %q, want %q", result.Cache.Status, WorkspaceIndexCacheMiss)
	}

	assertWorkspaceIndexCacheFileExists(t, projectDir)
}

func assertWorkspaceIndexCacheFileExists(t *testing.T, projectDir string) {
	t.Helper()

	info, err := os.Stat(WorkspaceIndexCachePath(projectDir))
	if err != nil {
		t.Fatalf("Stat cache returned error: %v", err)
	}

	if info.IsDir() {
		t.Fatalf("cache path is a directory, want file")
	}
}
