package app

import (
	"fmt"

	authoringcore "forge.adrianpk.com/chr/internal/core/authoring"
	opscore "forge.adrianpk.com/chr/internal/core/ops"
	ticketcore "forge.adrianpk.com/chr/internal/core/ticket"
	workflowcore "forge.adrianpk.com/chr/internal/core/workflow"
	workspaceconfig "forge.adrianpk.com/chr/internal/core/workspace/config"
)

// ReadTicketBoard uses the selected authority without rebuilding the legacy index.
func (s service) ReadTicketBoard(offset int) (ticketcore.Board, error) {
	return s.readTicketBoardPage(offset, "")
}

func (s service) readTicketBoardPage(offset int, caseSlug string) (ticketcore.Board, error) {
	opts, err := s.ticketDocumentOptions()
	if err != nil {
		return ticketcore.Board{}, err
	}

	opts.Offset, opts.CaseSlug = offset, caseSlug

	return workflowcore.ReadTicketBoard(s.ctx, opts)
}

func (s service) ticketDocumentOptions() (workflowcore.TicketBoardReadOptions, error) {
	config, err := workspaceconfig.Load(s.projectContext.ProjectDir)
	if err != nil {
		return workflowcore.TicketBoardReadOptions{}, err
	}

	layout, err := opscore.ResolveLayout(opscore.LayoutOptions{ProjectDir: s.projectContext.ProjectDir, Namespace: s.workflowNamespace})
	if err != nil {
		return workflowcore.TicketBoardReadOptions{}, err
	}

	return workflowcore.TicketBoardReadOptions{
		Layout: layout, ProjectID: s.projectContext.ProjectID,
		NativeWorkflow: workspaceconfig.WorkAuthority(config.Work) == workspaceconfig.WorkStateAuthorityTickets,
	}, nil
}

// ReadTicketDocument resolves one concrete document behind the Board.
func (s service) ReadTicketDocument(key string) (workflowcore.TicketDocument, error) {
	opts, err := s.ticketDocumentOptions()
	if err != nil {
		return workflowcore.TicketDocument{}, err
	}

	return workflowcore.ReadTicketDocument(s.ctx, opts, key)
}

// SaveTicketDocument applies metadata, body or discussion through shared guards.
func (s service) SaveTicketDocument(request workflowcore.TicketEditRequest) (workflowcore.TicketDocument, error) {
	opts, err := s.ticketDocumentOptions()
	if err != nil {
		return workflowcore.TicketDocument{}, err
	}

	if err := s.validateTicketScopeChange(request); err != nil {
		return workflowcore.TicketDocument{}, err
	}

	if err := s.validateTicketEditorReview(request); err != nil {
		return workflowcore.TicketDocument{}, err
	}

	return workflowcore.UpdateTicketDocument(s.ctx, opts, request)
}

func (s service) validateTicketEditorReview(request workflowcore.TicketEditRequest) error {
	if request.Fields["status"] != string(ticketcore.StatusReadyForReview) {
		return nil
	}

	document, err := s.ReadTicketDocument(request.Key)
	if err != nil || document.Ticket.NativeWorkflow {
		return err
	}

	return s.validateReadyForReviewEvidence(document.Ticket, string(ticketcore.StatusReadyForReview))
}

func (s service) validateTicketScopeChange(request workflowcore.TicketEditRequest) error {
	slug, present := request.Fields["use_case"]
	if !present || slug == "" {
		return nil
	}

	current, err := s.ReadTicketDocument(request.Key)
	if err != nil || current.Fields["use_case"] == slug {
		return err
	}

	return s.validateTicketDocumentScope(slug)
}

// CreateBoardTicket creates an independent canonical ticket in the selected authority.
func (s service) CreateBoardTicket(request workflowcore.TicketEditRequest) (workflowcore.TicketDocument, error) {
	opts, err := s.ticketDocumentOptions()
	if err != nil {
		return workflowcore.TicketDocument{}, err
	}

	if err := s.validateTicketDocumentScope(request.Fields["use_case"]); err != nil {
		return workflowcore.TicketDocument{}, err
	}

	config, err := workspaceconfig.Load(s.projectContext.ProjectDir)
	if err != nil {
		return workflowcore.TicketDocument{}, err
	}

	number, err := s.reserveNextTicketKeyNumber(config.TicketKeyPrefix)
	if err != nil {
		return workflowcore.TicketDocument{}, err
	}

	return workflowcore.CreateTicketDocument(s.ctx, opts, fmt.Sprintf("%s-%d", config.TicketKeyPrefix, number), request)
}

func (s service) validateTicketDocumentScope(slug string) error {
	if slug == "" {
		return nil
	}

	cases, err := authoringcore.NewCaseMetadataService(s.projectContext.ProjectDir).LoadAll()
	if err != nil {
		return err
	}

	for _, item := range cases {
		if item.Slug == slug {
			return nil
		}
	}

	return fmt.Errorf("select an existing use case")
}
