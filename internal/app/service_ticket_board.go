package app

import (
	"fmt"

	"forge.adrianpk.com/chr/internal/core"
	authoringcore "forge.adrianpk.com/chr/internal/core/authoring"
	opscore "forge.adrianpk.com/chr/internal/core/ops"
	ticketcore "forge.adrianpk.com/chr/internal/core/ticket"
)

// ReadTicketBoard uses the selected authority without rebuilding the legacy index.
func (s service) ReadTicketBoard(offset int) (core.Board, error) {
	return s.readTicketBoardPage(offset, "")
}

func (s service) readTicketBoardPage(offset int, caseSlug string) (core.Board, error) {
	opts, err := s.ticketDocumentOptions()
	if err != nil {
		return core.Board{}, err
	}

	opts.Offset, opts.CaseSlug = offset, caseSlug

	return core.ReadTicketBoard(s.ctx, opts)
}

func (s service) ticketDocumentOptions() (core.TicketBoardReadOptions, error) {
	config, err := core.LoadWorkspaceConfig(s.projectContext.ProjectDir)
	if err != nil {
		return core.TicketBoardReadOptions{}, err
	}

	layout, err := opscore.ResolveLayout(opscore.LayoutOptions{ProjectDir: s.projectContext.ProjectDir, Namespace: s.workflowNamespace})
	if err != nil {
		return core.TicketBoardReadOptions{}, err
	}

	return core.TicketBoardReadOptions{
		Layout: layout, ProjectID: s.projectContext.ProjectID,
		NativeWorkflow: core.WorkAuthority(config.Work) == core.WorkStateAuthorityTickets,
	}, nil
}

// ReadTicketDocument resolves one concrete document behind the Board.
func (s service) ReadTicketDocument(key string) (core.TicketDocument, error) {
	opts, err := s.ticketDocumentOptions()
	if err != nil {
		return core.TicketDocument{}, err
	}

	return core.ReadTicketDocument(s.ctx, opts, key)
}

// SaveTicketDocument applies metadata, body or discussion through shared guards.
func (s service) SaveTicketDocument(request core.TicketEditRequest) (core.TicketDocument, error) {
	opts, err := s.ticketDocumentOptions()
	if err != nil {
		return core.TicketDocument{}, err
	}

	if err := s.validateTicketScopeChange(request); err != nil {
		return core.TicketDocument{}, err
	}

	if err := s.validateTicketEditorReview(request); err != nil {
		return core.TicketDocument{}, err
	}

	return core.UpdateTicketDocument(s.ctx, opts, request)
}

func (s service) validateTicketEditorReview(request core.TicketEditRequest) error {
	if request.Fields["status"] != ticketcore.StatusReadyForReview {
		return nil
	}

	document, err := s.ReadTicketDocument(request.Key)
	if err != nil || document.Ticket.NativeWorkflow {
		return err
	}

	return s.validateReadyForReviewEvidence(document.Ticket, ticketcore.StatusReadyForReview)
}

func (s service) validateTicketScopeChange(request core.TicketEditRequest) error {
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
func (s service) CreateBoardTicket(request core.TicketEditRequest) (core.TicketDocument, error) {
	opts, err := s.ticketDocumentOptions()
	if err != nil {
		return core.TicketDocument{}, err
	}

	if err := s.validateTicketDocumentScope(request.Fields["use_case"]); err != nil {
		return core.TicketDocument{}, err
	}

	config, err := core.LoadWorkspaceConfig(s.projectContext.ProjectDir)
	if err != nil {
		return core.TicketDocument{}, err
	}

	number, err := s.reserveNextTicketKeyNumber(config.TicketKeyPrefix)
	if err != nil {
		return core.TicketDocument{}, err
	}

	return core.CreateTicketDocument(s.ctx, opts, fmt.Sprintf("%s-%d", config.TicketKeyPrefix, number), request)
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
