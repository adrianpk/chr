package chc

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

const workspaceIndexCacheSchemaVersion = 3

var errInvalidWorkspaceIndexCache = errors.New("invalid workspace index cache")

type WorkspaceIndexCacheStatus string

const (
	WorkspaceIndexCacheMiss  WorkspaceIndexCacheStatus = "miss"
	WorkspaceIndexCacheHit   WorkspaceIndexCacheStatus = "hit"
	WorkspaceIndexCacheStale WorkspaceIndexCacheStatus = "stale"
)

type WorkspaceIndexCacheState struct {
	Status WorkspaceIndexCacheStatus
	Path   string
}

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

func (b WorkspaceGraphBuilder) BuildCachedIndex(ctx context.Context) (CachedWorkspaceIndexResult, error) {
	graph, err := b.Build(ctx)
	if err != nil {
		return CachedWorkspaceIndexResult{}, err
	}

	current := NewWorkspaceIndexFromGraph(graph)
	cachePath := WorkspaceIndexCachePath(b.projectDir)

	cached, cacheHit, err := readWorkspaceIndexCache(cachePath)
	if errors.Is(err, errInvalidWorkspaceIndexCache) {
		cacheHit = false
		err = nil
	}

	if err != nil {
		return CachedWorkspaceIndexResult{}, err
	}

	status := WorkspaceIndexCacheMiss
	index := current

	if cacheHit {
		if workspaceIndexCacheValid(cached, current) {
			status = WorkspaceIndexCacheHit
			index = cached
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

func WorkspaceIndexCachePath(projectDir string) string {
	return UserPath(projectDir, "tmp", "cache", "workspace-index", "structural-index.json")
}

func readWorkspaceIndexCache(path string) (WorkspaceIndex, bool, error) {
	content, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return WorkspaceIndex{}, false, nil
	}

	if err != nil {
		return WorkspaceIndex{}, false, fmt.Errorf("read workspace index cache: %w", err)
	}

	var cache workspaceIndexCacheFile

	if json.Unmarshal(content, &cache) != nil {
		return WorkspaceIndex{}, false, errInvalidWorkspaceIndexCache
	}

	if cache.SchemaVersion != workspaceIndexCacheSchemaVersion {
		return WorkspaceIndex{}, false, nil
	}

	return cache.Index, true, nil
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
