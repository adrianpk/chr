package core

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	authoringcore "forge.adrianpk.com/chr/internal/core/authoring"
	workspacecore "forge.adrianpk.com/chr/internal/core/workspace"
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

func TestWorkspaceGraphBuilderBuildCachedIndexPreservesDumpPairMetadata(t *testing.T) {
	projectDir := t.TempDir()

	writeWorkspaceGraphFile(t, authoringcore.CaseMetadataPath(projectDir, "blog"), authoringcore.RenderCaseMetadata(authoringcore.CaseMetadata{
		Name: "Blog Publishing",
	}))
	writeWorkspaceGraphFile(t, filepath.Join(projectDir, workspacecore.DirectoryName, workspacecore.UserDirectoryName, "dumps", "blog-notes.md"), `---
title: "Blog Notes"
status: "curated"
type: "dump"
source: "ops/default/dump/00000-blog-notes-raw/artifact.md"
case: "blog"
---

# Blog Notes
`)
	writeWorkspaceGraphFile(t, filepath.Join(projectDir, workspacecore.DirectoryName, workspacecore.UserDirectoryName, "dumps", "blog-notes-raw.md"), `---
title: "Blog Notes Raw Capture"
status: "raw"
type: "dump_raw"
curated: "ops/default/dump/00000-blog-notes/artifact.md"
case: "blog"
---

# Blog Notes Raw Capture
`)

	_, err := NewWorkspaceGraphBuilder(projectDir).BuildCachedIndex(context.Background())
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

	rawID := second.Index.NodesByPath["ops/default/dump/00000-blog-notes-raw/artifact.md"]

	raw := second.Index.NodesByID[rawID]
	if raw.DocumentType != "dump_raw" || raw.CuratedPath != "ops/default/dump/00000-blog-notes/artifact.md" {
		t.Fatalf("cached raw dump metadata = %#v", raw)
	}
}

func TestWorkspaceGraphBuilderBuildCachedIndexPreservesQuizPairMetadata(t *testing.T) {
	projectDir := t.TempDir()

	writeWorkspaceGraphFile(t, authoringcore.CaseMetadataPath(projectDir, "blog"), authoringcore.RenderCaseMetadata(authoringcore.CaseMetadata{
		Name: "Blog Publishing",
	}))
	writeWorkspaceGraphFile(t, filepath.Join(projectDir, workspacecore.DirectoryName, workspacecore.UserDirectoryName, "quizzes", "blog-quiz.md"), `---
title: "Blog Quiz"
status: "curated"
raw_path: "ops/default/quiz/00000-blog-quiz-raw/artifact.md"
case: "blog"
---

# Blog Quiz
`)
	writeWorkspaceGraphFile(t, filepath.Join(projectDir, workspacecore.DirectoryName, workspacecore.UserDirectoryName, "quizzes", "blog-quiz-raw.md"), `---
title: "Blog Quiz Raw Capture"
status: "raw"
curated_path: "ops/default/quiz/00000-blog-quiz/artifact.md"
case: "blog"
---

# Blog Quiz Raw Capture
`)

	_, err := NewWorkspaceGraphBuilder(projectDir).BuildCachedIndex(context.Background())
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

	rawID := second.Index.NodesByPath["ops/default/quiz/00000-blog-quiz-raw/artifact.md"]

	raw := second.Index.NodesByID[rawID]
	if raw.CuratedPath != "ops/default/quiz/00000-blog-quiz/artifact.md" {
		t.Fatalf("cached raw quiz metadata = %#v", raw)
	}
}

func TestWorkspaceGraphBuilderBuildCachedIndexRewritesStaleCache(t *testing.T) {
	projectDir := t.TempDir()
	writeCaseRootedWorkspaceGraphFixture(t, projectDir)

	first, err := NewWorkspaceGraphBuilder(projectDir).BuildCachedIndex(context.Background())
	if err != nil {
		t.Fatalf("first BuildCachedIndex returned error: %v", err)
	}

	writeWorkspaceGraphFile(t, filepath.Join(projectDir, workspacecore.DirectoryName, workspacecore.UserDirectoryName, "specs", "blog-spec.md"), `---
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

	specID := second.Index.NodesByPath["ops/default/spec/00000-blog-spec/artifact.md"]
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
