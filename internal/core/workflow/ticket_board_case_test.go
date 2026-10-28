package workflow

import (
	"fmt"
	"path/filepath"
	"testing"
)

// TestTicketBoardCase filters before pagination and leaves other cases untouched.
func TestTicketBoardCase(t *testing.T) {
	store := workflowTestStore(t)

	for i, useCase := range []string{"other", "checkout", "checkout"} {
		record := workflowTestTask(fmt.Sprintf("CHR-%d", i+1))
		record.Task.UseCase = useCase

		data, err := encodeWorkflowRecord(record)
		if err != nil {
			t.Fatal(err)
		}

		writeTicketFile(t, filepath.Join(store.layout.Root(), record.Path), string(data))
		writeTicketFile(t, filepath.Join(store.layout.Root(), filepath.Dir(record.Path), "artifact.md"), string(record.Body))
	}

	opts := TicketBoardReadOptions{Layout: store.layout, ProjectID: store.projectID, NativeWorkflow: true, CaseSlug: "checkout", Limit: 1}
	for offset, key := range []string{"CHR-2", "CHR-3"} {
		opts.Offset = offset

		page, err := ReadTicketBoard(t.Context(), opts)
		if err != nil || page.TotalCount != 2 || len(page.Tickets) != 1 || page.Tickets[0].Key != key {
			t.Fatal("case filter or page boundary lost the selected ticket", err, page)
		}
	}

	opts.CaseSlug, opts.Offset, opts.Limit = "", 0, 3

	all, err := ReadTicketBoard(t.Context(), opts)
	if err != nil || all.TotalCount != 3 {
		t.Fatal("case query changed the project Board", err)
	}
}
