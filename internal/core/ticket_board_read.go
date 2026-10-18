package core

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// TicketBoardReadOptions selects authority and one bounded page of the Board.
type TicketBoardReadOptions struct {
	Layout         OpsLayout
	ProjectID      string
	NativeWorkflow bool
	Offset         int
	Limit          int
}

// ReadTicketBoard observes canonical documents and supported legacy tickets.
// It never rebuilds SQL, adopts documents or recovers a workflow transition.
func ReadTicketBoard(ctx context.Context, opts TicketBoardReadOptions) (board Board, err error) {
	if opts.Limit == 0 {
		opts.Limit = WorkflowPageLimit
	}

	if opts.ProjectID == "" || opts.Layout.ProjectDir() == "" || opts.Offset < 0 || opts.Offset > WorkflowRecordLimit || opts.Limit < 1 || opts.Limit > WorkflowPageLimit {
		return Board{}, fmt.Errorf("invalid ticket Board scope or page bounds")
	}

	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	store := NewWorkflowStore(opts.Layout, opts.ProjectID)

	read := func() error {
		var readErr error

		board, readErr = store.readTicketBoard(ctx, opts)

		return readErr
	}
	if opts.NativeWorkflow {
		err = store.withWorkflowInspection(ctx, read)
	} else {
		err = read()
	}

	return board, err
}

func (s *WorkflowStore) readTicketBoard(ctx context.Context, opts TicketBoardReadOptions) (Board, error) {
	tickets, err := s.readBoardTickets(ctx, opts.NativeWorkflow)
	if err != nil {
		return Board{}, err
	}

	return ticketBoardPage(s.projectID, tickets, opts.Offset, opts.Limit), nil
}

func (s *WorkflowStore) readBoardTickets(ctx context.Context, native bool) ([]IndexedTicket, error) {
	tickets := map[string]IndexedTicket{}
	paths := map[string]bool{}

	if native {
		if err := s.readNativeBoardTickets(ctx, tickets, paths); err != nil {
			return nil, err
		}
	}

	for _, root := range []string{filepath.Join(s.layout.Root(), "ticket"), ticketsDirectory(s.layout.ProjectDir())} {
		if err := s.readLooseBoardTickets(ctx, root, tickets, paths); err != nil {
			return nil, err
		}
	}

	all := make([]IndexedTicket, 0, len(tickets))
	for _, ticket := range tickets {
		all = append(all, ticket)
	}

	return sortedBoardTickets(all), nil
}

func (s *WorkflowStore) readNativeBoardTickets(ctx context.Context, tickets map[string]IndexedTicket, paths map[string]bool) error {
	snapshot, err := s.readSnapshot(ctx)
	if err != nil {
		return err
	}

	for _, record := range snapshot.Records {
		if record.Task == nil {
			continue
		}

		ticket, err := workflowIndexedTicket(ctx, s, record)
		if err != nil {
			return err
		}

		ticket.NativeWorkflow = true
		ticket.Key, ticket.Status, ticket.Slice = record.Task.Key, record.Task.Status, record.Task.SliceID
		ticket.ContentHash, ticket.IndexedAt = record.SourceHash, time.Time{}

		ticket.BoardColumn, _ = BoardTicketStatusFor(ticket.Status)
		if ticket.Title == "" {
			ticket.Title = record.Task.Key
		}

		if ticket.Synopsis == "" {
			ticket.Synopsis = boardDescription(record.Body)
		}

		paths[ticket.Path] = true
		tickets[ticket.Key] = ticket
	}

	return nil
}

func (s *WorkflowStore) readLooseBoardTickets(ctx context.Context, root string, tickets map[string]IndexedTicket, paths map[string]bool) error {
	files, err := boardMarkdownFiles(ctx, s.layout.ProjectDir(), root)
	if err != nil {
		return err
	}

	for _, path := range files {
		if paths[path] {
			continue
		}

		if err := s.appendLooseBoardTicket(root, path, tickets); err != nil {
			return err
		}
	}

	return nil
}

func boardMarkdownFiles(ctx context.Context, projectDir, root string) ([]string, error) {
	if err := checkBoardRoot(projectDir, root); err != nil {
		return nil, err
	}

	files := []string{}

	err := filepath.WalkDir(root, func(path string, entry fs.DirEntry, walkErr error) error {
		if path == root && errors.Is(walkErr, fs.ErrNotExist) {
			return nil
		}

		if walkErr != nil {
			return walkErr
		}

		if err := ctx.Err(); err != nil {
			return err
		}

		candidate, err := boardMarkdownCandidate(root, path, entry)
		if err != nil || !candidate {
			return err
		}

		files = append(files, path)
		if len(files) > WorkflowRecordLimit {
			return fmt.Errorf("ticket Board observation bound exceeded")
		}

		return nil
	})

	return files, err
}

func boardMarkdownCandidate(root, path string, entry fs.DirEntry) (bool, error) {
	if entry.Type()&os.ModeSymlink != 0 {
		return false, fmt.Errorf("ticket Board path cannot traverse a symlink: %s", path)
	}

	if entry.IsDir() {
		if path != root && (entry.Name() == "files" || entry.Name() == "evidence" || entry.Name() == "extension") {
			return false, filepath.SkipDir
		}

		return false, nil
	}

	return strings.HasSuffix(path, ".md") && entry.Name() != "TEMPLATE.md", nil
}

func checkBoardRoot(projectDir, root string) error {
	for parent := root; parent != projectDir; parent = filepath.Dir(parent) {
		info, err := os.Lstat(parent)
		if errors.Is(err, os.ErrNotExist) {
			continue
		}

		if err != nil {
			return err
		}

		if !info.IsDir() || info.Mode()&os.ModeSymlink != 0 {
			return fmt.Errorf("ticket Board root must use project directories: %s", parent)
		}
	}

	return nil
}

func (s *WorkflowStore) appendLooseBoardTicket(root, path string, tickets map[string]IndexedTicket) error {
	ticket, err := readBoardDocument(s.layout.ProjectDir(), path)
	if err != nil {
		return err
	}

	if old, exists := tickets[ticket.Key]; exists {
		if old.NativeWorkflow || root == ticketsDirectory(s.layout.ProjectDir()) && !strings.HasPrefix(old.Path, root+string(filepath.Separator)) {
			return nil
		}

		return fmt.Errorf("duplicate canonical Board ticket key: %s", ticket.Key)
	}

	if len(tickets) >= WorkflowRecordLimit {
		return fmt.Errorf("ticket Board observation bound exceeded")
	}

	ticket.ProjectID, ticket.Namespace = s.projectID, s.layout.Namespace()
	tickets[ticket.Key] = ticket

	return nil
}

func ticketBoardPage(projectID string, tickets []IndexedTicket, offset, limit int) Board {
	board := buildBoard(projectID, tickets)
	start, end := min(offset, len(tickets)), min(offset+limit, len(tickets))
	page := buildBoard(projectID, tickets[start:end])

	images := make([]string, 0, len(tickets))
	for _, ticket := range tickets {
		images = append(images, ticket.Key+"\x00"+ticket.ContentHash)
	}

	board.Revision = workflowHash([]byte(strings.Join(images, "\n")))

	board.Offset, board.LoadedCount, board.NextOffset = offset, end-start, -1
	if end < len(tickets) {
		board.NextOffset = end
	}

	for i := range board.Columns {
		board.Columns[i].Tickets = page.Columns[i].Tickets
	}

	board.Attention, board.Tickets = page.Attention, page.Tickets

	return board
}
