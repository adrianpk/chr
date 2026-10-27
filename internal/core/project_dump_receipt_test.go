package core

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

// An existing empty receipt directory has capacity; its normal end-of-directory
// result is not an authoring failure.
func TestProjectDumpReceiptEmpty(t *testing.T) {
	project, _ := prepareProjectDumpReconciliation(t)
	if err := os.MkdirAll(filepath.Join(project, authoringRuntimeRef("receipts")), 0o700); err != nil {
		t.Fatal(err)
	}

	preview, err := PreviewProjectDumpReconciliation(t.Context(), project)
	if err != nil {
		t.Fatal(err)
	}

	request := ProjectDumpReconciliationRequest{OperationID: "empty", ExpectedRevision: preview.Revision, Destination: preview.Destination}
	if _, err := ReconcileProjectDump(t.Context(), project, request); err != nil {
		t.Fatal("empty receipt directory rejected", err)
	}
}

// Receipts replay original committed facts after independent edits and at
// capacity. A retained operation cannot be exchanged for changed arguments.
func TestProjectDumpReconcileReceipt(t *testing.T) {
	project, _ := prepareProjectDumpReconciliation(t)

	preview, err := PreviewProjectDumpReconciliation(t.Context(), project)
	if err != nil {
		t.Fatal(err)
	}

	request := ProjectDumpReconciliationRequest{OperationID: "receipt", ExpectedRevision: preview.Revision, Destination: preview.Destination}

	result, err := ReconcileProjectDump(t.Context(), project, request)
	if err != nil {
		t.Fatal(err)
	}

	path := filepath.Join(project, filepath.FromSlash(preview.Destination))
	writeTestFile(t, path, "Independent later edit.\n")
	fillProjectDumpReceiptCapacity(t, project)

	replayed, err := ReconcileProjectDump(t.Context(), project, request)
	if err != nil || replayed != result {
		t.Fatal("receipt did not replay original committed facts at capacity", replayed, err)
	}

	data, err := os.ReadFile(path)
	if err != nil || string(data) != "Independent later edit.\n" {
		t.Fatal("receipt replay overwrote later edits", err)
	}

	request.OperationID = "new-operation"
	if _, err := ReconcileProjectDump(t.Context(), project, request); err == nil || !strings.Contains(err.Error(), "capacity exhausted") {
		t.Fatal("new operation bypassed receipt capacity", err)
	}
}

func fillProjectDumpReceiptCapacity(t *testing.T, project string) {
	t.Helper()

	for i := range workflowstate.WorkflowRecordLimit - 1 {
		path := filepath.Join(project, authoringRuntimeRef("receipts"), fmt.Sprintf("retained-%04d.json", i))
		if err := os.WriteFile(path, []byte("retained fixture"), 0o600); err != nil {
			t.Fatal(err)
		}
	}
}

// A pending authoring binding reserves its identity even before a receipt exists.
// Corrupt receipts and changed observations also fail without publishing images.
func TestProjectDumpReconcileBinding(t *testing.T) {
	for _, kind := range []string{"pending request", "corrupt receipt", "changed observation"} {
		t.Run(kind, func(t *testing.T) {
			project, _ := prepareProjectDumpReconciliation(t)

			preview, err := PreviewProjectDumpReconciliation(t.Context(), project)
			if err != nil {
				t.Fatal(err)
			}

			request := ProjectDumpReconciliationRequest{OperationID: "binding", ExpectedRevision: preview.Revision, Destination: preview.Destination}

			request = prepareProjectDumpReceiptBinding(t, project, request, kind)

			_, err = ReconcileProjectDump(t.Context(), project, request)
			if err == nil || kind != "corrupt receipt" && !errors.Is(err, workflowstate.ErrConflict) {
				t.Fatal("conflicting operation identity was reused", err)
			}
		})
	}
}

func prepareProjectDumpReceiptBinding(t *testing.T, project string, request ProjectDumpReconciliationRequest, kind string) ProjectDumpReconciliationRequest {
	t.Helper()

	switch kind {
	case "pending request":
		writeTestFile(t, filepath.Join(project, authoringRequestRef(request.OperationID)), "Existing pending operation.")
	case "corrupt receipt":
		writeTestFile(t, filepath.Join(project, authoringReceiptRef(request.OperationID)), "Invalid receipt.")
	case "changed observation":
		if _, err := ReconcileProjectDump(t.Context(), project, request); err != nil {
			t.Fatal(err)
		}

		request.ExpectedRevision = strings.Repeat("a", 64)
	}

	return request
}
