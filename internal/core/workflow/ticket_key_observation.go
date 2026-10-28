package workflow

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	opscore "forge.adrianpk.com/chr/internal/core/ops"
	opsnamespace "forge.adrianpk.com/chr/internal/core/ops/namespace"
	ticketcore "forge.adrianpk.com/chr/internal/core/ticket"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

// NextTicketKeyNumber observes legacy tickets and native envelopes across namespaces.
func NextTicketKeyNumber(ctx context.Context, projectDir, keyPrefix string) (int, error) {
	observer := ticketKeyObserver{projectDir: projectDir, keyPrefix: ticketcore.DefaultKeyPrefix}
	if keyPrefix != "" {
		observer.keyPrefix = keyPrefix
	}

	return observer.nextKeyNumber(ctx)
}

type ticketKeyObserver struct{ projectDir, keyPrefix string }

func (r *ticketKeyObserver) nextKeyNumber(ctx context.Context) (int, error) {
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
	roots := []string{ticketcore.Directory(projectDir)}

	entries, err := os.ReadDir(filepath.Join(projectDir, opscore.DirectoryName))
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
			roots = append(roots, filepath.Join(projectDir, opscore.DirectoryName, entry.Name(), "ticket"))
		}
	}

	return roots, nil
}

func (r *ticketKeyObserver) observedTicketNumber(path string) (int, error) {
	observed, err := ticketcore.ReadNumber(path, r.keyPrefix)
	if err != nil || filepath.Base(path) != opscore.ArtifactBodyFile() {
		return observed.Number, err
	}

	metadata, err := readWorkflowEnvelope(filepath.Join(filepath.Dir(path), "metadata.json"))
	if errors.Is(err, os.ErrNotExist) {
		return observed.Number, nil
	}

	if err != nil {
		return 0, err
	}

	var task Task
	if raw, present := metadata.Metadata["workflow"]; present {
		if err := json.Unmarshal(raw, &task); err != nil {
			return 0, err
		}
	}

	metadataNumber, _ := ticketcore.NumberFromText(r.keyPrefix, task.Key)

	return max(observed.Number, metadataNumber), err
}
