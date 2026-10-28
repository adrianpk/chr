package workflow

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"forge.adrianpk.com/chr/internal/core/contenthash"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

// TestWorkflowLooseMarkdown advances a guarded native ticket image without
// losing its description or custom fields; a later description edit conflicts.
func TestWorkflowLooseMarkdown(t *testing.T) {
	store := workflowTestStore(t)
	record := workflowLooseMarkdown(t, store)
	commit, _ := workflowSelectionCommit(t, store)
	run := workflowTestStart(t, store, StartRequest{StartInput: workflowstate.StartInput{
		RequestKey: "native-queue",
		Selection:  workflowstate.PreviewRequest{Selector: workflowstate.Selector{Extent: "loose"}, Target: workflowstate.PreviewTarget{Repository: "https://forge.example/test/repo", Commit: commit}},
	}}).Invocation

	var front string
	if err := json.Unmarshal(record.Metadata["frontmatter"], &front); err != nil {
		t.Fatal(err)
	}

	encoded, err := json.Marshal(strings.ReplaceAll(front, "priority: high", "priority: critical"))
	if err != nil {
		t.Fatal(err)
	}

	record.Metadata["frontmatter"] = encoded

	changed, err := store.Write(t.Context(), record, record.Revision)
	if err != nil {
		t.Fatal(err)
	}

	current, err := store.ReconcileWorkflowMembership(t.Context(), run.ID)
	if err != nil || current.Members[0].Admission.SourceHash != run.Members[0].Admission.SourceHash || current.Members[0].ExpectedSourceHash != changed.SourceHash {
		t.Fatal("native display write lost the frozen admission or accepted image", err)
	}

	changed.Body = append(changed.Body, []byte("New requirement\n")...)
	if _, err := store.Write(t.Context(), changed, changed.Revision); err != nil {
		t.Fatal(err)
	}

	if _, err := store.ReconcileWorkflowMembership(t.Context(), run.ID); !errors.Is(err, workflowstate.ErrConflict) {
		t.Fatal("native description edit became a scheduling change", err)
	}
}

func workflowLooseMarkdown(t *testing.T, store *Store) workflowstate.Record {
	t.Helper()

	record := workflowTestTask("PRJ-1")

	parsedMarkdown, err := workflowstate.RequireMarkdown(workflowLooseBody(record.Task.Key, "high"))
	front, body := parsedMarkdown.Front, parsedMarkdown.Body

	if err != nil {
		t.Fatal(err)
	}

	front += "\ncustom: |\n  Preserve this block"

	raw, err := json.Marshal(front)
	if err != nil {
		t.Fatal(err)
	}

	record.Metadata, record.Body, record.Path = map[string]json.RawMessage{"frontmatter": raw}, body, "ticket/backlog/PRJ-1.md"

	data := []byte("---\n" + front + "\n---\n" + string(body))

	path := filepath.Join(store.layout.Root(), record.Path)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}

	if err := os.WriteFile(path, data, 0o600); err != nil {
		t.Fatal(err)
	}

	observed, err := store.Adopt(t.Context(), record.Path, *record.Task, contenthash.Sum(data))
	if err != nil {
		t.Fatal(err)
	}

	return observed
}
