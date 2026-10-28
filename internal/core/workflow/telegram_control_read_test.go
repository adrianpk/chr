package workflow

import (
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"forge.adrianpk.com/chr/internal/core/contenthash"
)

// TestTelegramControlReceiptMissing proves a receipt probe cannot reserve or
// execute an unaccepted request.
func TestTelegramControlReceiptMissing(t *testing.T) {
	store, request := telegramLifecycleJournalFixture(t, "pause")
	before := workflowTestRead(t, store, request.ID)

	observed, err := store.ReadWorkflowControlReceipt(t.Context(), request)
	if err != nil || observed.Found || observed.Result.RequestID != "" || !reflect.DeepEqual(before, workflowTestRead(t, store, request.ID)) {
		t.Fatal("receipt probe changed an unaccepted request", observed.Result, observed.Found, err)
	}

	path, err := store.safePath(filepath.Join("run", "operation-bindings", contenthash.Sum([]byte(request.Binding.RequestID))+".json"))
	if err != nil {
		t.Fatal(err)
	}

	if _, err := os.Stat(path); !errors.Is(err, os.ErrNotExist) {
		t.Fatal("receipt probe reserved a mutation binding", err)
	}
}
