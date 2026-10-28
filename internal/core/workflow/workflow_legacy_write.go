package workflow

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"syscall"

	opscore "forge.adrianpk.com/chr/internal/core/ops"
	opsfile "forge.adrianpk.com/chr/internal/core/ops/file"
	opsnamespace "forge.adrianpk.com/chr/internal/core/ops/namespace"
	ticketcore "forge.adrianpk.com/chr/internal/core/ticket"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

// ErrWorkflowAuthority prevents legacy execution of an adopted task or projection.
var ErrWorkflowAuthority = errors.New("adopted workflow ticket requires ticket authority")

// CheckLegacyWorkTicket refuses adopted Markdown and JSON-backed ticket views
// before a legacy worker starts. The operation lock also fences concurrent adoption.
func CheckLegacyWorkTicket(path string) error {
	if _, err := ReadLegacyTicketFile(path); err != nil {
		return err
	}

	if filepath.Base(path) != "artifact.md" {
		return nil
	}

	data, err := opsfile.Read(filepath.Join(filepath.Dir(path), "metadata.json"))
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}

	if err != nil {
		return err
	}

	fields := map[string]json.RawMessage{}
	if err := json.Unmarshal(data, &fields); err != nil {
		return err
	}

	if _, adopted := fields["workflow"]; adopted {
		return ErrWorkflowAuthority
	}

	return nil
}

// FilterLegacyWorkTickets partitions a derived queue by its actual source images.
// It never treats an adopted projection as legacy execution authority.
func FilterLegacyWorkTickets(ctx context.Context, tickets []ticketcore.Indexed) ([]ticketcore.Indexed, int, error) {
	if len(tickets) > workflowstate.RecordLimit {
		return nil, 0, fmt.Errorf("legacy authority selection bound exceeded")
	}

	selected, excluded := []ticketcore.Indexed{}, 0

	for _, ticket := range tickets {
		if err := ctx.Err(); err != nil {
			return nil, excluded, err
		}

		err := CheckLegacyWorkTicket(ticket.Path)
		if errors.Is(err, ErrWorkflowAuthority) {
			excluded++

			continue
		}

		if err != nil {
			return nil, excluded, err
		}

		selected = append(selected, ticket)
	}

	return selected, excluded, nil
}

// ReadLegacyTicketFile reads an unadopted source image and refuses workflow authority.
// ReplaceLegacyTicketFile rechecks that image under the existing authority lock.
func ReadLegacyTicketFile(path string) ([]byte, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	parsedMarkdown, err := workflowstate.RequireMarkdown(content)
	front := parsedMarkdown.Front

	if err == nil && strings.Contains("\n"+front, "\nworkflow_schema:") {
		return nil, ErrWorkflowAuthority
	}

	return content, nil
}

// ReplaceLegacyTicketFile takes the existing repository or legacy file lock,
// rechecks adoption and source equality, then atomically replaces the image.
func ReplaceLegacyTicketFile(ctx context.Context, path string, observed, replacement []byte) error {
	root, err := legacyTicketGitRoot(path)
	if err != nil {
		return err
	}

	action := func() error {
		if err := CheckLegacyWorkTicket(path); err != nil {
			return err
		}

		current, err := ReadLegacyTicketFile(path)
		if err != nil {
			return err
		}

		if !bytes.Equal(current, observed) {
			return workflowstate.ErrConflict
		}

		return opsfile.WriteAtomic(path, replacement)
	}
	if root == "" {
		// Store requires Git coordination, so non-Git legacy tickets
		// cannot race an executable workflow adoption.
		return withLegacyTicketLock(ctx, path, action)
	}

	layout, err := legacyTicketOpsLayout(root, path)
	if err != nil {
		return err
	}

	return NewWorkflowStore(layout, "").withLock(ctx, action)
}

func withLegacyTicketLock(ctx context.Context, path string, action func() error) (err error) {
	file, err := os.OpenFile(path+".lock", os.O_CREATE|os.O_RDWR, 0o600)
	if err != nil {
		return err
	}

	defer func() { err = errors.Join(err, file.Close()) }()

	if err := opsfile.Lock(ctx, file); err != nil {
		return err
	}

	defer func() { err = errors.Join(err, syscall.Flock(int(file.Fd()), syscall.LOCK_UN)) }()

	return action()
}

func legacyTicketGitRoot(path string) (string, error) {
	absolute, err := filepath.Abs(path)
	if err != nil {
		return "", err
	}

	for dir := filepath.Dir(absolute); ; dir = filepath.Dir(dir) {
		if _, err := os.Stat(filepath.Join(dir, ".git")); err == nil {
			return dir, nil
		} else if !errors.Is(err, os.ErrNotExist) {
			return "", err
		}

		if dir == filepath.Dir(dir) {
			return "", nil
		}
	}
}

func legacyTicketOpsLayout(root, path string) (opscore.Layout, error) {
	namespace := opsnamespace.Default

	relative, err := filepath.Rel(root, path)
	if err != nil {
		return opscore.Layout{}, err
	}

	parts := strings.Split(filepath.ToSlash(relative), "/")
	if len(parts) > 2 && parts[0] == opscore.DirectoryName && opsnamespace.Valid(parts[1]) {
		namespace = parts[1]
	}

	return opscore.ResolveLayout(opscore.LayoutOptions{ProjectDir: root, Namespace: namespace})
}
