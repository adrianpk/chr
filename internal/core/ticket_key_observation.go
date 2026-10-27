package core

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	opscore "forge.adrianpk.com/chr/internal/core/ops"
	opsnamespace "forge.adrianpk.com/chr/internal/core/ops/namespace"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

func (r *fileTicketRepository) nextKeyNumber(ctx context.Context) (int, error) {
	roots, err := ticketKeyRoots(r.projectDir)
	if err != nil {
		return 0, err
	}

	highest := 0

	for _, root := range roots {
		files, err := boardMarkdownFiles(ctx, r.projectDir, root)
		if err != nil {
			return 0, err
		}

		for _, path := range files {
			number, err := r.observedTicketNumber(path)
			if err != nil {
				return 0, err
			}

			highest = max(highest, number)
		}
	}

	return highest + 1, nil
}

func ticketKeyRoots(projectDir string) ([]string, error) {
	roots := []string{ticketsDirectory(projectDir)}

	entries, err := os.ReadDir(filepath.Join(projectDir, opscore.OpsDirectoryName))
	if errors.Is(err, os.ErrNotExist) {
		return roots, nil
	}

	if err != nil {
		return nil, err
	}

	if len(entries) > workflowstate.RecordLimit {
		return nil, fmt.Errorf("ticket namespace observation bound exceeded")
	}

	for _, entry := range entries {
		if opsnamespace.Valid(entry.Name()) {
			roots = append(roots, filepath.Join(projectDir, opscore.OpsDirectoryName, entry.Name(), "ticket"))
		}
	}

	return roots, nil
}

func (r *fileTicketRepository) observedTicketNumber(path string) (int, error) {
	number, _, err := r.ticketNumber(path)
	if err != nil || filepath.Base(path) != opscore.OpsArtifactBodyFile() {
		return number, err
	}

	metadata, err := readWorkflowEnvelope(filepath.Join(filepath.Dir(path), "metadata.json"))
	if errors.Is(err, os.ErrNotExist) {
		return number, nil
	}

	if err != nil {
		return 0, err
	}

	var task WorkflowTask
	if raw, present := metadata.Metadata["workflow"]; present {
		if err := json.Unmarshal(raw, &task); err != nil {
			return 0, err
		}
	}

	metadataNumber, _ := keyNumberFromText(r.keyPrefix, task.Key)

	return max(number, metadataNumber), err
}
