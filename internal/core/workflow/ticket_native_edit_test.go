package workflow

import (
	"encoding/json"
	"errors"
	"maps"
	"strings"
	"testing"

	opscore "forge.adrianpk.com/chr/internal/core/ops"
	ticketcore "forge.adrianpk.com/chr/internal/core/ticket"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

// Native edits retain unknown envelope fields and reject unproven lifecycle moves.
func TestTicketNativeEditing(t *testing.T) {
	store := workflowTestStore(t)
	record := workflowTestTask("CHR-7")
	record.Domain = map[string]json.RawMessage{"unknown": json.RawMessage(`{"retained":true}`)}

	created, err := store.Create(t.Context(), "Native ticket", record)
	if err != nil {
		t.Fatal(err)
	}

	opts := TicketBoardReadOptions{Layout: store.layout, ProjectID: store.projectID, NativeWorkflow: true}

	original, err := ReadTicketDocument(t.Context(), opts, created.Task.Key)
	if err != nil {
		t.Fatal(err)
	}

	body := "Native full description.\n\nAdditional retained detail."

	saved, err := UpdateTicketDocument(t.Context(), opts, TicketEditRequest{
		Key: created.Task.Key, Revision: original.Revision, Body: &body, Binding: ticketEditorBinding("native-save"),
		Fields: map[string]string{"title": "Native edited", "status": string(ticketcore.StatusBacklog), "priority": "high", "severity": "high", "type": "feature", "owner": "Maintainer", "assignee": "Developer"},
	})
	if err != nil || saved.CommittedRevision == "" || saved.CommittedRevision != saved.Revision || saved.Description != body || saved.Fields["assignee"] != "Developer" || saved.Ticket.Priority != "high" {
		t.Fatal("native save lost body or metadata", saved.Fields, err)
	}

	checkNativeTicketIdentity(t, store, created)

	checkNativeTicketRejected(t, opts, saved, TicketEditRequest{Fields: map[string]string{"status": string(ticketcore.StatusDone)}, Binding: ticketEditorBinding("native-done")})

	if _, err := UpdateTicketDocument(t.Context(), opts, TicketEditRequest{Key: created.Task.Key, Revision: original.Revision, Body: &body, Binding: ticketEditorBinding("native-stale")}); !errors.Is(err, workflowstate.ErrConflict) {
		t.Fatal("native stale edit was accepted", err)
	}
}

func checkNativeTicketIdentity(t *testing.T, store *Store, created Record) {
	t.Helper()

	read, err := store.Read(t.Context(), created.Path)
	if err != nil || !strings.Contains(string(read.Domain["unknown"]), `"retained": true`) || read.Task.SetID != "" || read.Task.Key != created.Task.Key {
		t.Fatal("native edit changed immutable relationships or unknown metadata", err)
	}
}

// Display edits cannot restore stale status or case values from a derived body.
func TestTicketNativeMetadataIdentity(t *testing.T) {
	store := workflowTestStore(t)
	record := workflowTestTask("CHR-7")
	record.Task.Status, record.Task.UseCase = string(ticketcore.StatusBacklog), "retained-case"

	body, err := ticketcore.NewRenderer().Render(ticketcore.Ticket{Key: "CHR-7", Title: "Actual title", Status: ticketcore.StatusReadyForDev, UseCase: "outdated-case", CreatedAt: record.CreatedAt, UpdatedAt: record.UpdatedAt, Body: "Original body."})
	if err != nil {
		t.Fatal(err)
	}

	record.Body = []byte(body)

	created, err := store.Create(t.Context(), "Outdated envelope title", record)
	if err != nil {
		t.Fatal(err)
	}

	opts := TicketBoardReadOptions{Layout: store.layout, ProjectID: store.projectID, NativeWorkflow: true}

	document, err := ReadTicketDocument(t.Context(), opts, created.Task.Key)
	if err != nil || document.Fields["title"] != "Actual title" {
		t.Fatal("a derived title shadowed the document", document.Fields, err)
	}

	_, err = UpdateTicketDocument(t.Context(), opts, TicketEditRequest{Key: created.Task.Key, Revision: document.Revision, Fields: map[string]string{"owner": "Maintainer"}, Binding: ticketEditorBinding("metadata")})
	if err != nil {
		t.Fatal(err)
	}

	current, err := store.Read(t.Context(), created.Path)
	if err != nil || current.Task.Status != string(ticketcore.StatusBacklog) || current.Task.UseCase != "retained-case" {
		t.Fatal("metadata edit restored a stale workflow projection", err)
	}
}

// Approved descriptions and claimed tasks remain owned by their existing workflow.
func TestTicketNativeOwnership(t *testing.T) {
	for _, owner := range []string{"contract", "claim"} {
		t.Run(owner, func(t *testing.T) {
			store := workflowTestStore(t)

			record := workflowTestTask("CHR-7")
			if owner == "contract" {
				record.Domain = map[string]json.RawMessage{"execution_contract": json.RawMessage(`{"body":"Approved instructions"}`)}
			} else {
				record.Task.Phase, record.Task.Status = "running", string(ticketcore.StatusInProgress)
				record.Task.AttemptID, record.Task.ClaimGeneration = "owned-attempt", 1
			}

			created, err := store.Create(t.Context(), "Native owner", record)
			if err != nil {
				t.Fatal(err)
			}

			opts := TicketBoardReadOptions{Layout: store.layout, ProjectID: store.projectID, NativeWorkflow: true}

			document, err := ReadTicketDocument(t.Context(), opts, created.Task.Key)
			if err != nil {
				t.Fatal(err)
			}

			body := "Replacement instructions."
			checkNativeTicketRejected(t, opts, document, TicketEditRequest{Body: &body, Binding: ticketEditorBinding("owned-body")})

			if owner == "contract" {
				checkNativeTicketMetadata(t, store, opts, document, created)
			}
		})
	}
}

func checkNativeTicketMetadata(t *testing.T, store *Store, opts TicketBoardReadOptions, document TicketDocument, original Record) {
	t.Helper()

	_, err := UpdateTicketDocument(t.Context(), opts, TicketEditRequest{Key: document.Ticket.Key, Revision: document.Revision, Fields: map[string]string{"title": "Display edit", "priority": "high"}, Binding: ticketEditorBinding("display-only")})
	if err != nil {
		t.Fatal(err)
	}

	record, err := store.Read(t.Context(), original.Path)
	if err != nil || string(record.Body) != string(original.Body) || record.Task.ContractHash != original.Task.ContractHash {
		t.Fatal("display edit rewrote the approved description or contract", err)
	}
}

func checkNativeTicketRejected(t *testing.T, opts TicketBoardReadOptions, document TicketDocument, request TicketEditRequest) {
	t.Helper()

	before := nativeQueueImages(t, opts.Layout.Root())

	request.Key, request.Revision = document.Ticket.Key, document.Revision
	if _, err := UpdateTicketDocument(t.Context(), opts, request); err == nil {
		t.Fatal("unsafe native change was accepted")
	}

	if !maps.Equal(before, nativeQueueImages(t, opts.Layout.Root())) {
		t.Fatal("rejected mutation changed existing images")
	}
}

// Both authority modes create canonical independent tickets with the same retained fields.
func TestTicketDocumentCreation(t *testing.T) {
	for _, native := range []bool{false, true} {
		t.Run(map[bool]string{false: "unadopted", true: "native"}[native], func(t *testing.T) {
			layout, err := opscore.ResolveLayout(opscore.LayoutOptions{ProjectDir: t.TempDir(), Namespace: "default"})
			if err != nil {
				t.Fatal(err)
			}

			if native {
				layout = workflowTestStore(t).layout
			}

			opts := TicketBoardReadOptions{Layout: layout, ProjectID: "project-1", NativeWorkflow: native}
			body := "Full new description.\n\nSecond paragraph."

			created, err := CreateTicketDocument(t.Context(), opts, "CHR-88", TicketEditRequest{
				Fields: map[string]string{"title": "New ticket", "priority": "high", "type": "bug", "owner": "Maintainer", "assignee": "Developer", "status": string(ticketcore.StatusReadyForDev)},
				Body:   &body, Binding: ticketEditorBinding("create"),
			})
			if err != nil {
				t.Fatal("creation lost its retained fields or authority", created.Fields, err)
			}

			checkNewTicketDocument(t, opts, created, body)
		})
	}
}

func checkNewTicketDocument(t *testing.T, opts TicketBoardReadOptions, created TicketDocument, body string) {
	t.Helper()

	if created.Description != body || created.Fields["owner"] != "Maintainer" || created.Fields["assignee"] != "Developer" || created.Ticket.NativeWorkflow != opts.NativeWorkflow {
		t.Fatal("creation lost its retained fields or authority", created.Fields)
	}

	if !strings.HasPrefix(created.Ticket.Path, opts.Layout.Root()) || created.Ticket.Slice != "" {
		t.Fatal("creation used legacy storage or enrolled slice work", created.Ticket.RelativePath)
	}
}
