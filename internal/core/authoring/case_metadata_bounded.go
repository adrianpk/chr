package authoring

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"forge.adrianpk.com/chr/internal/core/contenthash"
	opsfile "forge.adrianpk.com/chr/internal/core/ops/file"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
	workspacecore "forge.adrianpk.com/chr/internal/core/workspace"
)

// LoadBounded observes at most the owner's fixed case-record ceiling.
func (s CaseMetadataService) LoadBounded(ctx context.Context, limit int) ([]CaseMetadata, error) {
	if limit < 1 || limit > workflowstate.RecordLimit {
		return nil, fmt.Errorf("invalid case record bound")
	}

	entries, err := s.boundedCaseEntries(limit)
	if err != nil {
		return nil, err
	}

	var result []CaseMetadata

	for _, entry := range entries {
		if err := ctx.Err(); err != nil {
			return nil, err
		}

		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".md") {
			continue
		}

		metadata, err := s.loadBoundedCase(filepath.Join(CaseMetadataDirectory(s.projectDir), entry.Name()))
		if err != nil {
			return nil, err
		}

		result = append(result, metadata)
	}

	return result, nil
}

func (s CaseMetadataService) loadBoundedCase(path string) (CaseMetadata, error) {
	relative, err := filepath.Rel(workspacecore.UserPath(s.projectDir), path)
	if err != nil {
		return CaseMetadata{}, err
	}

	path, err = s.caseMutationPath(relative)
	if err != nil {
		return CaseMetadata{}, err
	}

	data, err := opsfile.Read(path)
	if err != nil {
		return CaseMetadata{}, err
	}

	metadata, err := ParseCaseMetadata(string(data))
	if err != nil {
		return metadata, err
	}

	metadata.SourcePath, metadata.SourceHash, metadata.Exists = path, contenthash.Sum(data), true
	if metadata.Slug == "" {
		metadata.Slug = CaseSlugFromPath(path)
	}

	return metadata, nil
}

func (s CaseMetadataService) boundedCaseEntries(limit int) ([]os.DirEntry, error) {
	probe, err := s.caseMutationPath("cases/.read-probe")
	if err != nil {
		return nil, err
	}

	dir, err := os.Open(filepath.Dir(probe))
	if errors.Is(err, os.ErrNotExist) {
		return nil, nil
	}

	if err != nil {
		return nil, err
	}

	entries, readErr := dir.ReadDir(limit + 1)

	closeErr := dir.Close()
	if readErr != nil && !errors.Is(readErr, io.EOF) || closeErr != nil {
		return nil, errors.Join(readErr, closeErr)
	}

	if len(entries) > limit {
		return nil, fmt.Errorf("case record capacity exceeded")
	}

	return entries, nil
}

// ReadBounded observes one canonical case without following an external source path.
func (s CaseMetadataService) ReadBounded(ctx context.Context, slug string) (CaseMetadata, error) {
	if err := ctx.Err(); err != nil {
		return CaseMetadata{}, err
	}

	if !filepath.IsLocal(slug) || filepath.Base(slug) != slug || len(slug) > 1024 {
		return CaseMetadata{}, fmt.Errorf("invalid case identity")
	}

	return s.loadBoundedCase(CaseMetadataPath(s.projectDir, slug))
}
