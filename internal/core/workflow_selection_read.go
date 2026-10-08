package core

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"path/filepath"
	"sort"
	"strings"
)

func (s *WorkflowStore) readWorkflowSelection(ctx context.Context) (workflowSelection, error) {
	snapshot, err := s.readSnapshot(ctx)
	if err != nil {
		return workflowSelection{}, err
	}

	selection := workflowSelection{snapshot: snapshot, records: map[string]WorkflowRecord{}, sets: map[string]WorkflowRecord{}, slices: map[string]WorkflowSlice{}}
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

	if err := s.readUnpreparedTickets(ctx, &selection, paths); err != nil {
		return workflowSelection{}, err
	}

	sort.Slice(selection.tickets, func(i, j int) bool { return compareIndexedTicketsByKey(selection.tickets[i], selection.tickets[j]) < 0 })

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
		if count > WorkflowRecordLimit {
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

	sort.Strings(images)

	return workflowHash([]byte(strings.Join(images, "\n")))
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

	content, err := readWorkflowFile(path)
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
	if len(selection.tickets) > WorkflowRecordLimit {
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

	created, err := optionalTicketTime(metadata["created_at"])
	if err != nil {
		return IndexedTicket{}, err
	}

	return IndexedTicket{Key: unquoteMetadataString(metadata["key"]), Status: unquoteMetadataString(metadata["status"]), UseCase: unquoteMetadataString(metadata["use_case"]), Priority: unquoteMetadataString(metadata["priority"]), Severity: unquoteMetadataString(metadata["severity"]), Type: unquoteMetadataString(metadata["type"]), ProjectArea: unquoteMetadataString(metadata["project_area"]), Tags: parseMetadataStringList(metadata["tags"]), CreatedAt: created, ContentHash: workflowHash(content)}, nil
}
