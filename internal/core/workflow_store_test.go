package core

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func workflowTestStore(t *testing.T) *WorkflowStore {
	t.Helper()

	dir := t.TempDir()
	if output, err := exec.Command("git", "init", "-q", dir).CombinedOutput(); err != nil {
		t.Fatalf("git init: %v: %s", err, output)
	}

	layout, err := ResolveOpsLayout(ResolveOpsLayoutOptions{ProjectDir: dir, Namespace: "default"})
	if err != nil {
		t.Fatal(err)
	}

	return NewWorkflowStore(layout, "project-1")
}

// TestWorkflowRoundTrip protects content, unknown fields and stale-writer rejection.
func TestWorkflowRoundTrip(t *testing.T) {
	s := workflowTestStore(t)
	ctx := context.Background()

	r, err := s.Create(ctx, "Original ticket", workflowTestTask("CHR-1"))
	if err != nil {
		t.Fatal(err)
	}

	if r.ID == r.Task.Key || string(r.Body) != "# Original body\n" {
		t.Fatal("artifact identity replaced ticket key or body")
	}

	old := r
	r.Metadata["custom"] = json.RawMessage(`{"native":true}`)
	r.Metadata["title"] = json.RawMessage(`"Renamed ticket"`)
	r.Domain["extra"] = json.RawMessage(`"preserved"`)
	r.Task.Status = TicketStatusBacklog

	written, err := s.Write(ctx, r, r.Revision)
	if err != nil {
		t.Fatal(err)
	}

	if written.Revision != 2 || string(written.Domain["extra"]) != `"preserved"` || !strings.Contains(string(written.Metadata["custom"]), `"native": true`) || !written.CreatedAt.Equal(old.CreatedAt) {
		t.Fatal("round trip lost revision, creation time or unknown fields")
	}

	if _, err := s.Write(ctx, old, old.Revision); !errors.Is(err, ErrWorkflowConflict) {
		t.Fatalf("stale writer: %v", err)
	}

	testWorkflowManualEdit(t, s, written)
}

func testWorkflowManualEdit(t *testing.T, s *WorkflowStore, written WorkflowRecord) {
	t.Helper()

	ctx := context.Background()
	path := filepath.Join(s.layout.Root(), written.Path)

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}

	if err := os.WriteFile(path, append(data, ' '), 0o644); err != nil {
		t.Fatal(err)
	}

	if _, err := s.Write(ctx, written, written.Revision); !errors.Is(err, ErrWorkflowConflict) {
		t.Fatalf("manual edit with unchanged revision: %v", err)
	}
}

// TestWorkflowAdoption retains exact native body and compatible frontmatter.
func TestWorkflowAdoption(t *testing.T) {
	s := workflowTestStore(t)
	ctx := context.Background()
	relative := "ticket/backlog/CHR-1.md"
	path := filepath.Join(s.layout.Root(), relative)
	content := "---\nkey: CHR-1\ncreated_at: 2026-01-02T03:04:05Z\nstatus: backlog\ntitle: Original title\ncustom: |\n  Preserve this block\nreview_refs:\n  - type: pull_request\n    url: https://example.org/pr/1\n---\n\nBody  with spaces.\n"

	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}

	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	if _, err := s.Read(ctx, relative); !errors.Is(err, ErrNotWorkflow) {
		t.Fatalf("unadopted document became executable: %v", err)
	}

	task := workflowTestTask("CHR-1").Task

	r, err := s.Adopt(ctx, relative, *task, workflowHash([]byte(content)))
	if err != nil {
		t.Fatal(err)
	}

	workflowTestLegacyRace(t, ctx, path, []byte(content))

	r.Task.Evidence = []string{"old-evidence"}
	r = workflowTestWrite(t, s, r)
	r.Task.Evidence = nil

	r = workflowTestWrite(t, s, r)
	workflowTestLegacyGuard(t, ctx, path)

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(string(data), "custom: |\n  Preserve this block\nreview_refs:") || strings.Contains(string(data), "task_evidence:") || string(r.Body) != "\nBody  with spaces.\n" || r.ID != "CHR-1" {
		t.Fatal("adoption changed native content, identity or cleared-field semantics")
	}
}

func workflowTestWrite(t *testing.T, s *WorkflowStore, r WorkflowRecord) WorkflowRecord {
	t.Helper()

	written, err := s.Write(context.Background(), r, r.Revision)
	if err != nil {
		t.Fatal(err)
	}

	return written
}

func workflowTestLegacyGuard(t *testing.T, ctx context.Context, path string) {
	t.Helper()

	if err := NewTicketStatusUpdater().UpdateStatus(ctx, UpdateTicketStatusOptions{Path: path, Status: TicketStatusBacklog}); err == nil {
		t.Fatal("legacy writer bypassed workflow guards")
	}
}

func workflowTestLegacyRace(t *testing.T, ctx context.Context, path string, observed []byte) {
	t.Helper()

	if err := writeLegacyTicketFile(ctx, path, observed, []byte("stale replacement")); err == nil {
		t.Fatal("pre-adoption reader overwrote an adopted record")
	}
}

// TestWorkflowRelocation proves checkout paths do not participate in identity.
func TestWorkflowRelocation(t *testing.T) {
	store := workflowTestStore(t)
	r := workflowTestCreate(t, store, "CHR-1")

	destination := filepath.Join(t.TempDir(), "moved")
	if err := os.Rename(store.layout.ProjectDir(), destination); err != nil {
		t.Fatal(err)
	}

	layout, err := ResolveOpsLayout(ResolveOpsLayoutOptions{ProjectDir: destination, Namespace: "default"})
	if err != nil {
		t.Fatal(err)
	}

	got, err := NewWorkflowStore(layout, "project-1").Read(context.Background(), r.Path)
	if err != nil {
		t.Fatal(err)
	}

	if got.ID != r.ID || got.Task.Key != r.Task.Key || got.SourceHash != r.SourceHash || !got.CreatedAt.Equal(r.CreatedAt) {
		t.Fatal("relocation changed durable identity or content")
	}
}
