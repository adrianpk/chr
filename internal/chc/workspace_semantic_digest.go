package chc

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

const workspaceSemanticDigestCacheSchemaVersion = 1

// WorkspaceSemanticDigestStatus defines a package boundary type.
type WorkspaceSemanticDigestStatus string

const (
	// WorkspaceSemanticDigestReady defines a package boundary constant.
	WorkspaceSemanticDigestReady WorkspaceSemanticDigestStatus = "ready"
	// WorkspaceSemanticDigestPending defines a package boundary constant.
	WorkspaceSemanticDigestPending WorkspaceSemanticDigestStatus = "pending"
	// WorkspaceSemanticDigestStale defines a package boundary constant.
	WorkspaceSemanticDigestStale WorkspaceSemanticDigestStatus = "stale"
)

// WorkspaceSemanticDigest describes a package boundary value.
type WorkspaceSemanticDigest struct {
	NodeID      string
	Path        string
	Kind        WorkspaceNodeKind
	ModelID     string
	PromptHash  string
	SourceHash  string
	Digest      string
	GeneratedAt time.Time
}

// WorkspaceSemanticDigestState describes a package boundary value.
type WorkspaceSemanticDigestState struct {
	NodeID string
	Status WorkspaceSemanticDigestStatus
	Digest WorkspaceSemanticDigest
}

// WorkspaceSemanticDigestCache describes a package boundary value.
type WorkspaceSemanticDigestCache struct {
	Entries map[string]WorkspaceSemanticDigest
}

// WorkspaceSemanticDigestGenerator defines a package boundary interface.
type WorkspaceSemanticDigestGenerator interface {
	GenerateWorkspaceSemanticDigest(
		ctx context.Context,
		request WorkspaceSemanticDigestRequest,
	) (string, error)
}

// DeterministicWorkspaceSemanticDigestGenerator describes a package boundary value.
type DeterministicWorkspaceSemanticDigestGenerator struct{}

// WorkspaceSemanticDigestRequest describes a package boundary value.
type WorkspaceSemanticDigestRequest struct {
	NodeID     string
	Node       WorkspaceNode
	ModelID    string
	PromptHash string
}

// WorkspaceSemanticDigestGeneration describes a package boundary value.
type WorkspaceSemanticDigestGeneration struct {
	NodeID string
	Status WorkspaceSemanticDigestStatus
	Digest WorkspaceSemanticDigest
}

// WorkspaceSemanticIndexWorker describes a package boundary value.
type WorkspaceSemanticIndexWorker struct {
	Generator  WorkspaceSemanticDigestGenerator
	ModelID    string
	PromptHash string
	Now        func() time.Time
}

// WorkspaceSemanticIndexWorkerResult describes a package boundary value.
type WorkspaceSemanticIndexWorkerResult struct {
	Cache          WorkspaceSemanticDigestCache
	Generated      []WorkspaceSemanticDigestGeneration
	ReadyCount     int
	PendingCount   int
	StaleCount     int
	GeneratedCount int
}

type workspaceSemanticDigestCacheFile struct {
	SchemaVersion int                       `json:"schema_version"`
	WrittenAt     time.Time                 `json:"written_at"`
	Entries       []WorkspaceSemanticDigest `json:"entries"`
}

// WorkspaceSemanticDigestCachePath executes the corresponding package operation.
func WorkspaceSemanticDigestCachePath(projectDir string) string {
	return UserPath(projectDir, "tmp", "cache", "workspace-index", "semantic-digests.json")
}

// NewWorkspaceSemanticDigestCache constructs the corresponding package service or value.
func NewWorkspaceSemanticDigestCache() WorkspaceSemanticDigestCache {
	return WorkspaceSemanticDigestCache{
		Entries: map[string]WorkspaceSemanticDigest{},
	}
}

// ReadWorkspaceSemanticDigestCache executes the corresponding package operation.
func ReadWorkspaceSemanticDigestCache(path string) (WorkspaceSemanticDigestCache, error) {
	content, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return NewWorkspaceSemanticDigestCache(), nil
	}

	if err != nil {
		return WorkspaceSemanticDigestCache{}, fmt.Errorf("read workspace semantic digest cache: %w", err)
	}

	var cacheFile workspaceSemanticDigestCacheFile

	if invalidWorkspaceSemanticDigestCache(content, &cacheFile) {
		return NewWorkspaceSemanticDigestCache(), nil
	}

	if cacheFile.SchemaVersion != workspaceSemanticDigestCacheSchemaVersion {
		return NewWorkspaceSemanticDigestCache(), nil
	}

	cache := NewWorkspaceSemanticDigestCache()

	for _, entry := range cacheFile.Entries {
		if strings.TrimSpace(entry.NodeID) == "" {
			continue
		}

		cache.Entries[entry.NodeID] = entry
	}

	return cache, nil
}

// WriteWorkspaceSemanticDigestCache executes the corresponding package operation.
func WriteWorkspaceSemanticDigestCache(path string, cache WorkspaceSemanticDigestCache) error {
	err := os.MkdirAll(filepath.Dir(path), 0o755)
	if err != nil {
		return fmt.Errorf("create workspace semantic digest cache directory: %w", err)
	}

	content, err := json.MarshalIndent(workspaceSemanticDigestCacheFile{
		SchemaVersion: workspaceSemanticDigestCacheSchemaVersion,
		WrittenAt:     time.Now().UTC(),
		Entries:       cache.sortedEntries(),
	}, "", "  ")
	if err != nil {
		return fmt.Errorf("encode workspace semantic digest cache: %w", err)
	}

	tmpPath := path + ".tmp"

	err = os.WriteFile(tmpPath, content, 0o644)
	if err != nil {
		return fmt.Errorf("write workspace semantic digest cache: %w", err)
	}

	err = os.Rename(tmpPath, path)
	if err != nil {
		return fmt.Errorf("replace workspace semantic digest cache: %w", err)
	}

	return nil
}

func invalidWorkspaceSemanticDigestCache(content []byte, cacheFile *workspaceSemanticDigestCacheFile) bool {
	return json.Unmarshal(content, cacheFile) != nil
}

// WorkspaceSemanticDigestStates executes the corresponding package operation.
func WorkspaceSemanticDigestStates(
	index WorkspaceIndex,
	cache WorkspaceSemanticDigestCache,
	modelID string,
	promptHash string,
) []WorkspaceSemanticDigestState {
	states := make([]WorkspaceSemanticDigestState, 0, len(index.NodesByID))
	nodeIDs := sortedWorkspaceIndexNodeIDs(index)

	for _, nodeID := range nodeIDs {
		node := index.NodesByID[nodeID]

		digest, ok := cache.Entries[nodeID]
		if !ok {
			states = append(states, WorkspaceSemanticDigestState{
				NodeID: nodeID,
				Status: WorkspaceSemanticDigestPending,
			})

			continue
		}

		status := WorkspaceSemanticDigestReady
		if workspaceSemanticDigestStale(digest, node, modelID, promptHash) {
			status = WorkspaceSemanticDigestStale
		}

		states = append(states, WorkspaceSemanticDigestState{
			NodeID: nodeID,
			Status: status,
			Digest: digest,
		})
	}

	return states
}

// NewWorkspaceSemanticDigest constructs the corresponding package service or value.
func NewWorkspaceSemanticDigest(
	node WorkspaceNode,
	modelID string,
	promptHash string,
	digest string,
	generatedAt time.Time,
) WorkspaceSemanticDigest {
	return WorkspaceSemanticDigest{
		NodeID:      WorkspaceNodeID(node),
		Path:        filepath.ToSlash(node.RelativePath),
		Kind:        node.Kind,
		ModelID:     modelID,
		PromptHash:  promptHash,
		SourceHash:  node.SourceHash,
		Digest:      digest,
		GeneratedAt: generatedAt.UTC(),
	}
}

// GenerateWorkspaceSemanticDigest generates the requested artifact.
func (DeterministicWorkspaceSemanticDigestGenerator) GenerateWorkspaceSemanticDigest(
	_ context.Context,
	request WorkspaceSemanticDigestRequest,
) (string, error) {
	node := request.Node
	parts := []string{
		request.NodeID,
		string(node.Kind),
		node.RelativePath,
		node.Title,
		node.Status,
		node.CaseSlug,
		node.SourceHash,
		request.ModelID,
		request.PromptHash,
	}

	return contentHash([]byte(strings.Join(parts, "\x00"))), nil
}

// Generate generates the requested artifact.
func (w WorkspaceSemanticIndexWorker) Generate(
	ctx context.Context,
	index WorkspaceIndex,
	cache WorkspaceSemanticDigestCache,
) (WorkspaceSemanticIndexWorkerResult, error) {
	if w.Generator == nil {
		return WorkspaceSemanticIndexWorkerResult{}, errors.New("workspace semantic digest generator is required")
	}

	result := WorkspaceSemanticIndexWorkerResult{
		Cache: NewWorkspaceSemanticDigestCache(),
	}

	for nodeID, entry := range cache.Entries {
		result.Cache.Entries[nodeID] = entry
	}

	states := WorkspaceSemanticDigestStates(index, result.Cache, w.ModelID, w.PromptHash)
	for _, state := range states {
		err := ctx.Err()
		if err != nil {
			return WorkspaceSemanticIndexWorkerResult{}, err
		}

		result.countState(state.Status)

		if state.Status == WorkspaceSemanticDigestReady {
			continue
		}

		node := index.NodesByID[state.NodeID]

		digest, err := w.Generator.GenerateWorkspaceSemanticDigest(ctx, WorkspaceSemanticDigestRequest{
			NodeID:     state.NodeID,
			Node:       node,
			ModelID:    w.ModelID,
			PromptHash: w.PromptHash,
		})
		if err != nil {
			return WorkspaceSemanticIndexWorkerResult{}, fmt.Errorf("generate workspace semantic digest for %s: %w", state.NodeID, err)
		}

		entry := NewWorkspaceSemanticDigest(node, w.ModelID, w.PromptHash, digest, w.now())
		result.Cache.Entries[state.NodeID] = entry
		result.Generated = append(result.Generated, WorkspaceSemanticDigestGeneration{
			NodeID: state.NodeID,
			Status: state.Status,
			Digest: entry,
		})
		result.GeneratedCount++
	}

	return result, nil
}

func (w WorkspaceSemanticIndexWorker) now() time.Time {
	if w.Now != nil {
		return w.Now().UTC()
	}

	return time.Now().UTC()
}

func (r *WorkspaceSemanticIndexWorkerResult) countState(status WorkspaceSemanticDigestStatus) {
	switch status {
	case WorkspaceSemanticDigestReady:
		r.ReadyCount++
	case WorkspaceSemanticDigestPending:
		r.PendingCount++
	case WorkspaceSemanticDigestStale:
		r.StaleCount++
	}
}

func workspaceSemanticDigestStale(
	digest WorkspaceSemanticDigest,
	node WorkspaceNode,
	modelID string,
	promptHash string,
) bool {
	if digest.SourceHash != node.SourceHash {
		return true
	}

	if digest.ModelID != modelID {
		return true
	}

	return digest.PromptHash != promptHash
}

func (c WorkspaceSemanticDigestCache) sortedEntries() []WorkspaceSemanticDigest {
	entries := make([]WorkspaceSemanticDigest, 0, len(c.Entries))
	for _, entry := range c.Entries {
		entries = append(entries, entry)
	}

	sort.SliceStable(entries, func(left, right int) bool {
		return entries[left].NodeID < entries[right].NodeID
	})

	return entries
}

func sortedWorkspaceIndexNodeIDs(index WorkspaceIndex) []string {
	nodeIDs := make([]string, 0, len(index.NodesByID))
	for nodeID := range index.NodesByID {
		nodeIDs = append(nodeIDs, nodeID)
	}

	sort.Strings(nodeIDs)

	return nodeIDs
}
