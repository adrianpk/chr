package core

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"

	executioncore "forge.adrianpk.com/chr/internal/core/execution"
	harnessruntime "forge.adrianpk.com/chr/internal/core/harness/runtime"
	projectcore "forge.adrianpk.com/chr/internal/core/project"
	workspaceconfig "forge.adrianpk.com/chr/internal/core/workspace/config"
)

// RuntimeArtifactCleanupPreview describes a package boundary value.
type RuntimeArtifactCleanupPreview struct {
	ArtifactDir       string
	ReferencedRunDirs []string
	Candidates        []RuntimeArtifactCleanupCandidate
}

// RuntimeArtifactCleanupCandidate describes a package boundary value.
type RuntimeArtifactCleanupCandidate struct {
	Path string
}

// RuntimeArtifactCleanupPreviewService describes a package boundary value.
type RuntimeArtifactCleanupPreviewService struct {
	repository executioncore.Repository
}

// NewRuntimeArtifactCleanupPreviewService constructs the corresponding package service or value.
func NewRuntimeArtifactCleanupPreviewService(
	repository executioncore.Repository,
) RuntimeArtifactCleanupPreviewService {
	return RuntimeArtifactCleanupPreviewService{repository: repository}
}

// Preview executes the corresponding package operation.
func (s RuntimeArtifactCleanupPreviewService) Preview(
	ctx context.Context,
	projectDir string,
	projectID string,
) (RuntimeArtifactCleanupPreview, error) {
	config, err := workspaceconfig.Load(projectDir)
	if err != nil {
		return RuntimeArtifactCleanupPreview{}, err
	}

	config.Runtime = harnessruntime.FillConfigDefaults(config.Runtime)
	artifactDir := projectcore.ResolvePath(projectDir, config.Runtime.ArtifactDir)

	records, err := s.repository.ListProjectExecutions(ctx, projectID)
	if err != nil {
		return RuntimeArtifactCleanupPreview{}, err
	}

	referenced := referencedRunDirs(records)

	candidates, err := unreferencedRuntimeArtifactDirs(artifactDir, referenced)
	if err != nil {
		return RuntimeArtifactCleanupPreview{}, err
	}

	return RuntimeArtifactCleanupPreview{
		ArtifactDir:       artifactDir,
		ReferencedRunDirs: sortedReferencedRunDirs(referenced),
		Candidates:        candidates,
	}, nil
}

func referencedRunDirs(records []executioncore.Record) map[string]struct{} {
	referenced := map[string]struct{}{}

	for _, record := range records {
		path := filepath.Clean(record.Artifacts.RunDir)
		if path == "." || path == "" {
			continue
		}

		referenced[path] = struct{}{}
	}

	return referenced
}

func sortedReferencedRunDirs(referenced map[string]struct{}) []string {
	paths := make([]string, 0, len(referenced))

	for path := range referenced {
		paths = append(paths, path)
	}

	sort.Strings(paths)

	return paths
}

func unreferencedRuntimeArtifactDirs(
	artifactDir string,
	referenced map[string]struct{},
) ([]RuntimeArtifactCleanupCandidate, error) {
	info, err := os.Stat(artifactDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}

		return nil, fmt.Errorf("inspect runtime artifact directory %q: %w", artifactDir, err)
	}

	if !info.IsDir() {
		return nil, fmt.Errorf("runtime artifact path is not a directory: %s", artifactDir)
	}

	leafDirs, err := runtimeArtifactLeafDirs(artifactDir)
	if err != nil {
		return nil, err
	}

	candidates := make([]RuntimeArtifactCleanupCandidate, 0, len(leafDirs))

	for _, leafDir := range leafDirs {
		if leafDir == artifactDir {
			continue
		}

		_, ok := referenced[leafDir]
		if ok {
			continue
		}

		candidates = append(candidates, RuntimeArtifactCleanupCandidate{Path: leafDir})
	}

	sort.SliceStable(candidates, func(i, j int) bool {
		return candidates[i].Path < candidates[j].Path
	})

	return candidates, nil
}

func runtimeArtifactLeafDirs(root string) ([]string, error) {
	children := map[string]int{}
	dirs := []string{}

	err := filepath.WalkDir(root, func(path string, entry os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}

		if !entry.IsDir() {
			return nil
		}

		cleanPath := filepath.Clean(path)
		dirs = append(dirs, cleanPath)

		parent := filepath.Dir(cleanPath)
		if cleanPath != root {
			children[parent]++
		}

		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("walk runtime artifact directory: %w", err)
	}

	leafDirs := []string{}

	for _, dir := range dirs {
		if children[dir] == 0 {
			leafDirs = append(leafDirs, dir)
		}
	}

	sort.Strings(leafDirs)

	return leafDirs, nil
}
