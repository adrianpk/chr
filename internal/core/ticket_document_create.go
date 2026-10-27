package core

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"forge.adrianpk.com/chr/internal/core/contenthash"
	opscore "forge.adrianpk.com/chr/internal/core/ops"
	opsfile "forge.adrianpk.com/chr/internal/core/ops/file"
	opsmetadata "forge.adrianpk.com/chr/internal/core/ops/metadata"
	ticketmetadata "forge.adrianpk.com/chr/internal/core/ticket/metadata"
)

// CreateTicketDocument allocates through the existing canonical envelope owner.
// Selecting a case never creates slice membership or execution authorization.
func CreateTicketDocument(ctx context.Context, opts TicketBoardReadOptions, key string, request TicketEditRequest) (TicketDocument, error) {
	content, err := prepareTicketCreation(opts, key, request)
	if err != nil {
		return TicketDocument{}, err
	}

	if opts.NativeWorkflow {
		return createNativeTicketDocument(ctx, opts, key, request, content)
	}

	artifact, err := opscore.CreateArtifact(opscore.ArtifactOptions{Layout: opts.Layout, Type: "ticket", Title: request.Fields["title"], Producer: "chr"})
	if err != nil {
		return TicketDocument{}, err
	}

	if err := opsfile.WriteAtomic(filepath.Join(artifact.Dir, opscore.ArtifactBodyFile()), []byte(content)); err != nil {
		return TicketDocument{}, err
	}

	return ReadTicketDocument(ctx, opts, key)
}

func prepareTicketCreation(opts TicketBoardReadOptions, key string, request TicketEditRequest) (string, error) {
	if err := validateTicketCreationInput(opts, key, request); err != nil {
		return "", err
	}

	if err := request.Binding.Validate(); err != nil {
		return "", err
	}

	now := time.Now().UTC()
	ticket := Ticket{Key: key, Title: request.Fields["title"], Body: *request.Body, Synopsis: boardDescription([]byte(*request.Body)), Status: TicketStatusBacklog, Priority: "normal", Severity: "normal", Type: "task", CreatedAt: now, UpdatedAt: now}

	content, err := NewTicketRenderer().RenderTicket(ticket)
	if err != nil {
		return "", err
	}

	document := TicketDocument{content: []byte(content), Fields: map[string]string{"status": TicketStatusBacklog}, ReadOnly: map[string]string{}, Phase: "pending"}
	if opts.NativeWorkflow {
		document.native = &WorkflowRecord{}
	}

	content, err = prepareTicketEdit(document, request, now)
	if err != nil {
		return "", err
	}

	if err := validateNewTicketStatus(request); err != nil {
		return "", err
	}

	return content, nil
}

func validateTicketCreationInput(opts TicketBoardReadOptions, key string, request TicketEditRequest) error {
	if key == "" || opts.ProjectID == "" || opts.Layout.ProjectDir() == "" || request.Body == nil || strings.TrimSpace(request.Fields["title"]) == "" {
		return fmt.Errorf("ticket key, title and description are required")
	}

	return nil
}

func createNativeTicketDocument(ctx context.Context, opts TicketBoardReadOptions, key string, request TicketEditRequest, content string) (TicketDocument, error) {
	front, err := ticketmetadata.Parse(content)
	if err != nil {
		return TicketDocument{}, err
	}

	record := WorkflowRecord{Kind: "ticket", Body: []byte(content), Task: &WorkflowTask{
		Schema: WorkflowSchema, Key: key, Status: opsmetadata.Unquote(front.Metadata["status"]), Phase: "pending",
		UseCase: request.Fields["use_case"], ContractHash: contenthash.Sum([]byte(key + "\x00" + content)),
	}}

	_, err = NewWorkflowStore(opts.Layout, opts.ProjectID).ApplyRecordMutation(ctx, WorkflowMutationRequest{Binding: request.Binding, Title: request.Fields["title"], Record: record})
	if err != nil {
		return TicketDocument{}, err
	}

	return ReadTicketDocument(ctx, opts, key)
}

func validateNewTicketStatus(request TicketEditRequest) error {
	if status := request.Fields["status"]; status != "" && status != TicketStatusBacklog && status != TicketStatusReadyForDev {
		return fmt.Errorf("new tickets start in Backlog or Ready For Dev")
	}

	return nil
}
