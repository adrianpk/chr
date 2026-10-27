package core

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	workspacecore "forge.adrianpk.com/chr/internal/core/workspace"
)

const workspaceIndexCacheSchemaVersion = 3

var errInvalidWorkspaceIndexCache = errors.New("invalid workspace index cache")

// WorkspaceIndexCacheStatus defines a package boundary type.
type WorkspaceIndexCacheStatus string

const (
	// WorkspaceIndexCacheMiss defines a package boundary constant.
	WorkspaceIndexCacheMiss WorkspaceIndexCacheStatus = "miss"
	// WorkspaceIndexCacheHit defines a package boundary constant.
	WorkspaceIndexCacheHit WorkspaceIndexCacheStatus = "hit"
	// WorkspaceIndexCacheStale defines a package boundary constant.
	WorkspaceIndexCacheStale WorkspaceIndexCacheStatus = "stale"
)

// WorkspaceIndexCacheState describes a package boundary value.
type WorkspaceIndexCacheState struct {
	Status WorkspaceIndexCacheStatus
	Path   string
}

// CachedWorkspaceIndexResult describes a package boundary value.
type CachedWorkspaceIndexResult struct {
	Graph WorkspaceGraph
	Index WorkspaceIndex
	Cache WorkspaceIndexCacheState
}

type workspaceIndexCacheFile struct {
	SchemaVersion int            `json:"schema_version"`
	WrittenAt     time.Time      `json:"written_at"`
	Index         WorkspaceIndex `json:"index"`
}

// BuildCachedIndex builds the requested workspace artifact.
func (b WorkspaceGraphBuilder) BuildCachedIndex(ctx context.Context) (CachedWorkspaceIndexResult, error) {
	graph, err := b.Build(ctx)
	if err != nil {
		return CachedWorkspaceIndexResult{}, err
	}

	current := NewWorkspaceIndexFromGraph(graph)
	cachePath := WorkspaceIndexCachePath(b.projectDir)

	cached, err := readWorkspaceIndexCache(cachePath)
	if errors.Is(err, errInvalidWorkspaceIndexCache) {
		cached.Hit = false
		err = nil
	}

	if err != nil {
		return CachedWorkspaceIndexResult{}, err
	}

	status := WorkspaceIndexCacheMiss
	index := current

	if cached.Hit {
		if workspaceIndexCacheValid(cached.Index, current) {
			status = WorkspaceIndexCacheHit
			index = cached.Index
		} else {
			status = WorkspaceIndexCacheStale
		}
	}

	if status != WorkspaceIndexCacheHit {
		err = writeWorkspaceIndexCache(cachePath, current)
		if err != nil {
			return CachedWorkspaceIndexResult{}, err
		}
	}

	return CachedWorkspaceIndexResult{
		Graph: graph,
		Index: index,
		Cache: WorkspaceIndexCacheState{
			Status: status,
			Path:   cachePath,
		},
	}, nil
}

// WorkspaceIndexCachePath executes the corresponding package operation.
func WorkspaceIndexCachePath(projectDir string) string {
	return workspacecore.UserPath(projectDir, "tmp", "cache", "workspace-index", "structural-index.json")
}

type workspaceIndexCacheLookup struct {
	Index WorkspaceIndex
	Hit   bool
}

func readWorkspaceIndexCache(path string) (workspaceIndexCacheLookup, error) {
	content, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return workspaceIndexCacheLookup{}, nil
	}

	if err != nil {
		return workspaceIndexCacheLookup{}, fmt.Errorf("read workspace index cache: %w", err)
	}

	var cache workspaceIndexCacheFile

	if json.Unmarshal(content, &cache) != nil {
		return workspaceIndexCacheLookup{}, errInvalidWorkspaceIndexCache
	}

	if cache.SchemaVersion != workspaceIndexCacheSchemaVersion {
		return workspaceIndexCacheLookup{}, nil
	}

	return workspaceIndexCacheLookup{Index: cache.Index, Hit: true}, nil
}

func writeWorkspaceIndexCache(path string, index WorkspaceIndex) error {
	err := os.MkdirAll(filepath.Dir(path), 0o755)
	if err != nil {
		return fmt.Errorf("create workspace index cache directory: %w", err)
	}

	content, err := json.MarshalIndent(workspaceIndexCacheFile{
		SchemaVersion: workspaceIndexCacheSchemaVersion,
		WrittenAt:     time.Now().UTC(),
		Index:         index,
	}, "", "  ")
	if err != nil {
		return fmt.Errorf("encode workspace index cache: %w", err)
	}

	tmpPath := path + ".tmp"

	err = os.WriteFile(tmpPath, content, 0o644)
	if err != nil {
		return fmt.Errorf("write workspace index cache: %w", err)
	}

	err = os.Rename(tmpPath, path)
	if err != nil {
		return fmt.Errorf("replace workspace index cache: %w", err)
	}

	return nil
}

func workspaceIndexCacheValid(cached, current WorkspaceIndex) bool {
	if len(cached.NodesByID) != len(current.NodesByID) {
		return false
	}

	for id, currentNode := range current.NodesByID {
		cachedNode, ok := cached.NodesByID[id]
		if !ok {
			return false
		}

		if workspaceNodeSourceChanged(cachedNode, currentNode) {
			return false
		}
	}

	return true
}
