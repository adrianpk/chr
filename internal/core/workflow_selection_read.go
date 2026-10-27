package core

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"path/filepath"
	"sort"
	"strings"

	"forge.adrianpk.com/chr/internal/core/contenthash"
	opsfile "forge.adrianpk.com/chr/internal/core/ops/file"
	opsmetadata "forge.adrianpk.com/chr/internal/core/ops/metadata"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

func (s *WorkflowStore) readWorkflowSelection(ctx context.Context) (workflowSelection, error) {
	return s.readWorkflowSelectionSources(ctx, true)
}

func (s *WorkflowStore) readWorkflowSelectionSources(ctx context.Context, unprepared bool) (workflowSelection, error) {
	snapshot, err := s.readSnapshot(ctx)
	if err != nil {
		return workflowSelection{}, err
	}

	selection := workflowSelection{snapshot: snapshot, records: map[string]WorkflowRecord{}, sets: map[string]WorkflowRecord{}, slices: map[string]WorkflowSlice{}, evidenceHashes: map[string]string{}}
	paths := map[string]bool{}

	for _, record := range snapshot.Records {
		if record.Set != nil {
			selection.sets[record.ID] = record
			for _, descriptor := range record.Set.Slices {
				selection.slices[descriptor.ID] = descriptor
			}
		}

		if record.Task == nil {
			continue
		}

		ticket, parseErr := workflowIndexedTicket(ctx, s, record)
		if parseErr != nil {
			return workflowSelection{}, parseErr
		}

		ticket.CreatedAt = &record.CreatedAt
		selection.tickets = append(selection.tickets, ticket)
		selection.records[ticket.Key] = record
		paths[ticket.RelativePath] = true
	}

	if len(selection.slices) > workflowstate.RecordLimit {
		return workflowSelection{}, fmt.Errorf("workflow slice observation bound exceeded")
	}

	if err := s.appendWorkflowUnprepared(ctx, &selection, paths, unprepared); err != nil {
		return workflowSelection{}, err
	}

	sort.Slice(selection.tickets, func(i, j int) bool { return compareIndexedTicketsByKey(selection.tickets[i], selection.tickets[j]) < 0 })

	if err := s.hashWorkflowEvidence(ctx, &selection); err != nil {
		return workflowSelection{}, err
	}

	return selection, nil
}

func (s *WorkflowStore) readUnpreparedTickets(ctx context.Context, selection *workflowSelection, paths map[string]bool) error {
	count := 0

	err := filepath.WalkDir(filepath.Join(s.layout.Root(), "ticket"), func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}

		if err := ctx.Err(); err != nil {
			return err
		}

		if entry.IsDir() || !strings.HasSuffix(path, ".md") || entry.Name() == "TEMPLATE.md" {
			return nil
		}

		count++
		if count > workflowstate.RecordLimit {
			return fmt.Errorf("workflow ticket observation bound exceeded")
		}

		return s.appendUnpreparedTicket(selection, paths, path)
	})
	if errors.Is(err, fs.ErrNotExist) {
		return nil
	}

	return err
}

func (selection workflowSelection) hash() string {
	images := []string{workflowSnapshotHash(selection.snapshot)}
	for _, ticket := range selection.tickets {
		images = append(images, ticket.RelativePath+"\x00"+ticket.ContentHash)
	}

	for path, hash := range selection.evidenceHashes {
		images = append(images, path+"\x00"+hash)
	}

	sort.Strings(images)

	return contenthash.Sum([]byte(strings.Join(images, "\n")))
}

func (s *WorkflowStore) appendUnpreparedTicket(selection *workflowSelection, paths map[string]bool, path string) error {
	relative, err := filepath.Rel(s.layout.Root(), path)
	if err != nil {
		return err
	}

	if paths[relative] {
		return nil
	}

	if _, err := s.safePath(relative); err != nil {
		return err
	}

	content, err := opsfile.Read(path)
	if err != nil {
		return err
	}

	if !strings.HasPrefix(string(content), "---\n") {
		return nil
	}

	ticket, err := workflowUnpreparedTicket(content)
	if err != nil {
		return err
	}

	if ticket.Key == "" {
		return nil
	}

	ticket.ProjectID, ticket.RelativePath = s.projectID, relative

	ticket.Namespace = s.layout.Namespace()

	return selection.appendObservedTicket(ticket)
}

func (selection *workflowSelection) appendObservedTicket(ticket IndexedTicket) error {
	for _, old := range selection.tickets {
		if old.Key == ticket.Key {
			return fmt.Errorf("duplicate observed ticket key: %s", ticket.Key)
		}
	}

	selection.tickets = append(selection.tickets, ticket)
	if len(selection.tickets) > workflowstate.RecordLimit {
		return fmt.Errorf("workflow preview record bound exceeded")
	}

	return nil
}

func workflowUnpreparedTicket(content []byte) (IndexedTicket, error) {
	frontmatter, err := parseTicketFrontmatter(string(content))
	if err != nil {
		return IndexedTicket{}, err
	}

	metadata := frontmatter.Metadata

	created, err := opsmetadata.OptionalTimestamp(metadata["created_at"])
	if err != nil {
		return IndexedTicket{}, err
	}

	return IndexedTicket{IndexedTicketIdentity: IndexedTicketIdentity{Key: opsmetadata.Unquote(metadata["key"])}, IndexedTicketClassification: IndexedTicketClassification{Status: opsmetadata.Unquote(metadata["status"]), UseCase: opsmetadata.Unquote(metadata["use_case"]), Priority: opsmetadata.Unquote(metadata["priority"]), Severity: opsmetadata.Unquote(metadata["severity"]), Type: opsmetadata.Unquote(metadata["type"]), ProjectArea: opsmetadata.Unquote(metadata["project_area"]), Tags: opsmetadata.ParseList(metadata["tags"])}, IndexedTicketTimes: IndexedTicketTimes{CreatedAt: created}, IndexedTicketProjection: IndexedTicketProjection{ContentHash: contenthash.Sum(content)}}, nil
}

func (s *WorkflowStore) hashWorkflowEvidence(ctx context.Context, selection *workflowSelection) error {
	for _, record := range selection.records {
		for _, relative := range record.Task.Evidence {
			if err := ctx.Err(); err != nil {
				return err
			}

			if _, found := selection.evidenceHashes[relative]; found {
				continue
			}

			if len(selection.evidenceHashes) >= workflowstate.RecordLimit {
				return fmt.Errorf("workflow evidence observation bound exceeded")
			}

			hash, err := s.workflowEvidenceHash(relative)
			if err != nil {
				return err
			}

			selection.evidenceHashes[relative] = hash
		}
	}

	return nil
}

func (s *WorkflowStore) workflowEvidenceHash(relative string) (string, error) {
	path, err := s.safePath(relative)
	if err != nil {
		return "", err
	}

	data, err := opsfile.Read(path)
	if errors.Is(err, fs.ErrNotExist) {
		return "missing", nil
	}

	if err != nil {
		return "", err
	}

	return contenthash.Sum(data), nil
}

func (s *WorkflowStore) appendWorkflowUnprepared(ctx context.Context, selection *workflowSelection, paths map[string]bool, enabled bool) error {
	if !enabled {
		return nil
	}

	return s.readUnpreparedTickets(ctx, selection, paths)
}
