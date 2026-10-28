package workflow

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	generationcore "forge.adrianpk.com/chr/internal/core/authoring/generation"
	ticketcore "forge.adrianpk.com/chr/internal/core/ticket"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

// TestWorkflowLooseChangedInput rejects unguarded rank edits and changed execution
// or description images instead of converting either into a scheduling update.
func TestWorkflowLooseChangedInput(t *testing.T) {
	for _, change := range []string{"description", "title", "contract", "phase", "raw rank"} {
		t.Run(change, func(t *testing.T) {
			f := workflowLooseQueueFixture(t, false)
			record := workflowLooseRecord(t, f.store, "PRJ-91")
			before := workflowTestRead(t, f.store, f.run.ID)
			changeWorkflowLooseInput(t, f.store, record, change)

			if _, err := f.store.ReconcileWorkflowMembership(t.Context(), before.ID); !errors.Is(err, workflowstate.ErrConflict) {
				t.Fatal("changed frozen execution was silently accepted", err)
			}

			if current := workflowTestRead(t, f.store, before.ID); current.Revision != before.Revision || current.Cursor != before.Cursor || len(current.Attempts) != 1 {
				t.Fatal("conflicting input advanced or replayed work")
			}
		})
	}
}

func changeWorkflowLooseInput(t *testing.T, store *Store, record workflowstate.Record, change string) {
	t.Helper()

	switch change {
	case "description":
		record.Body = append(record.Body, []byte("Changed requirement\n")...)
	case "title":
		record.Body = []byte(strings.ReplaceAll(string(record.Body), "title: Independent task", "title: Another requirement"))
	case "contract":
		var contract generationcore.GeneratedWorkflowTask
		if err := json.Unmarshal(record.Domain["execution_contract"], &contract); err != nil {
			t.Fatal(err)
		}

		contract.Body = "Another requirement"

		data, err := json.Marshal(contract)
		if err != nil {
			t.Fatal(err)
		}

		record.Domain["execution_contract"] = data
	case "phase":
		record.Task.Phase, record.Task.Status = "blocked", string(ticketcore.StatusBlockedLegacy)
	case "raw rank":
		path := filepath.Join(store.layout.Root(), filepath.Dir(record.Path), "artifact.md")
		if err := os.WriteFile(path, workflowLooseBody(record.Task.Key, "critical"), 0o600); err != nil {
			t.Fatal(err)
		}

		return
	}

	if _, err := store.Write(t.Context(), record, record.Revision); err != nil {
		t.Fatal(err)
	}
}

// TestWorkflowLooseRetainedProof ensures archived evidence remains mandatory
// before claiming another independent ticket, even after queue reselection.
func TestWorkflowLooseRetainedProof(t *testing.T) {
	f := workflowLooseQueueFixture(t, false)
	workflowFinalizeFrontier(t, f)

	run := workflowAdvanceFrontier(t, f)
	if err := os.WriteFile(filepath.Join(f.store.layout.Root(), run.Deliveries[0].DeliveryRef), []byte("{}\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	if _, err := f.store.ClaimWorkflow(t.Context(), run.ID); !errors.Is(err, workflowstate.ErrConflict) {
		t.Fatal("altered delivery admitted a successor", err)
	}
}
