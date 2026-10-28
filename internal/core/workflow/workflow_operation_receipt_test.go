package workflow

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

type operationReceiptFixture struct {
	path    string
	image   []byte
	foreign []byte
	read    func() (bool, error)
}

func operationReceiptTestFixture(t *testing.T, operation string) operationReceiptFixture {
	t.Helper()

	store := workflowTestStore(t)
	request := telegramOperationMutationRequest()

	result, err := store.ApplyRecordMutation(t.Context(), request)
	if err != nil {
		t.Fatal(err)
	}

	mutationImage, err := os.ReadFile(filepath.Join(store.layout.Root(), result.ReceiptRef))
	if err != nil {
		t.Fatal(err)
	}

	normalized, err := store.normalizeMutation(request)
	if err != nil {
		t.Fatal(err)
	}

	control := workflowstate.ControlMutationRequest{Binding: request.Binding, Action: "pause", ID: "run-1", ExpectedRevision: 7}
	commit := store.newWorkflowControlCommit(control, normalized.Digest, nil)

	files, err := commit.files(t.Context(), control.ExpectedRevision+1)
	if err != nil {
		t.Fatal(err)
	}

	fixture := operationReceiptFixture{path: filepath.Join(store.layout.Root(), result.ReceiptRef), image: mutationImage, foreign: files[0].After}

	fixture.read = func() (bool, error) {
		retained, err := store.readMutationReceipt(normalized.Request, normalized.Digest)

		return retained.Found, err
	}
	if operation == "control" {
		fixture.image, fixture.foreign = fixture.foreign, fixture.image
		fixture.read = func() (bool, error) {
			_, found, err := store.readWorkflowControlReceipt(control, normalized.Digest)

			return found, err
		}
	}

	return fixture
}

// A decoded file is not a confirmed operation. Each real receipt reader must
// reject malformed images and the other operation's correctly checksummed schema.
func TestOperationReceiptGuards(t *testing.T) {
	cases := []struct {
		name     string
		image    func(operationReceiptFixture) []byte
		found    bool
		failure  bool
		conflict bool
	}{
		{"confirmed", func(f operationReceiptFixture) []byte { return f.image }, true, false, false},
		{"missing", func(operationReceiptFixture) []byte { return nil }, false, false, false},
		{"malformed", func(operationReceiptFixture) []byte { return []byte("{") }, false, true, false},
		{"null", func(operationReceiptFixture) []byte { return []byte("null") }, false, true, true},
		{"foreign schema", func(f operationReceiptFixture) []byte { return f.foreign }, false, true, true},
	}
	for _, operation := range []string{"record", "control"} {
		for _, test := range cases {
			t.Run(operation+"/"+test.name, func(t *testing.T) {
				fixture := operationReceiptTestFixture(t, operation)
				fixture.replace(t, test.image(fixture))

				found, err := fixture.read()
				if found != test.found || (err != nil) != test.failure || errors.Is(err, workflowstate.ErrConflict) != test.conflict {
					t.Fatal("receipt presence or failure changed", found, err)
				}
			})
		}
	}
}

func (fixture operationReceiptFixture) replace(t *testing.T, image []byte) {
	t.Helper()

	if image == nil {
		if err := os.Remove(fixture.path); err != nil {
			t.Fatal(err)
		}

		return
	}

	if err := os.WriteFile(fixture.path, image, 0o644); err != nil {
		t.Fatal(err)
	}
}
