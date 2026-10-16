package core

import (
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

// TestTelegramControlReceiptMissing proves a receipt probe cannot reserve or
// execute an unaccepted request.
func TestTelegramControlReceiptMissing(t *testing.T) {
	store, request := telegramLifecycleJournalFixture(t, "pause")
	before := workflowTestRead(t, store, request.ID)

	result, found, err := store.ReadWorkflowControlReceipt(t.Context(), request)
	if err != nil || found || result.RequestID != "" || !reflect.DeepEqual(before, workflowTestRead(t, store, request.ID)) {
		t.Fatal("receipt probe changed an unaccepted request", result, found, err)
	}

	path, err := store.safePath(filepath.Join("run", "operation-bindings", workflowHash([]byte(request.Binding.RequestID))+".json"))
	if err != nil {
		t.Fatal(err)
	}

	if _, err := os.Stat(path); !errors.Is(err, os.ErrNotExist) {
		t.Fatal("receipt probe reserved a mutation binding", err)
	}
}
