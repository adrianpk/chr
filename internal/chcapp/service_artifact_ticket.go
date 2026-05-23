package chcapp

import (
	"errors"
	"time"

	"forge.adrianpk.com/chc/internal/chc"
)

type ArtifactOperationResult struct {
	Message           string
	Tickets           []chc.IndexedTicket
	AuthoringSessions []chc.AuthoringSession
}

func (s service) CreateDefaultTicket() (ArtifactOperationResult, error) {
	config, err := chc.LoadWorkspaceConfig(s.projectContext.ProjectDir)
	if err != nil {
		return ArtifactOperationResult{}, err
	}

	repository := chc.NewTicketRepository(
		s.projectContext.ProjectDir,
		chc.WithTicketRepositoryKeyPrefix(config.TicketKeyPrefix),
	)

	result, err := repository.CreateTicket(s.ctx, chc.CreateTicketOptions{
		Title:       "TUI created ticket",
		Priority:    "normal",
		Severity:    "normal",
		Type:        "task",
		ProjectArea: "tui",
	})
	if err != nil {
		return ArtifactOperationResult{}, err
	}

	reloaded, err := s.reloadArtifactOperationState()
	if err != nil {
		return ArtifactOperationResult{}, err
	}

	reloaded.Message = "Created ticket " + result.Key + " at " +
		displayPath(s.projectContext.ProjectDir, result.Path)

	return reloaded, nil
}

func (s service) CreateTUIAuthoringSession() (ArtifactOperationResult, error) {
	now := time.Now().UTC()
	session := chc.AuthoringSession{
		ID:        authoringSessionID(now),
		ProjectID: s.projectContext.ProjectID,
		Status:    chc.AuthoringSessionStatusDraft,
		Mode:      chc.AuthoringSessionModeTUI,
		Target:    chc.AuthoringTargetDraft,
		Topic:     "TUI artifact creation",
		CreatedAt: now,
		UpdatedAt: now,
	}

	paths, err := chc.NewAuthoringSessionRepository(s.projectContext.ProjectDir).CreateSession(s.ctx, session)
	if err != nil {
		return ArtifactOperationResult{}, err
	}

	reloaded, err := s.reloadArtifactOperationState()
	if err != nil {
		return ArtifactOperationResult{}, err
	}

	reloaded.Message = "Created authoring session " + session.ID + " at " +
		displayPath(s.projectContext.ProjectDir, paths.SessionPath)

	return reloaded, nil
}

func (s service) reloadArtifactOperationState() (ArtifactOperationResult, error) {
	tickets, err := s.listProjectTickets()
	if err != nil {
		return ArtifactOperationResult{}, err
	}

	authoringSessions, err := loadTUIAuthoringSessions(s.ctx, s.projectContext)
	if err != nil {
		return ArtifactOperationResult{}, err
	}

	return ArtifactOperationResult{
		Tickets:           tickets,
		AuthoringSessions: authoringSessions,
	}, nil
}

type TicketOperationResult struct {
	Message string
	Tickets []chc.IndexedTicket
}

func (s service) AddSelectedTicketComment(ticketKey string) (TicketOperationResult, error) {
	ticket, err := s.ticketByKey(ticketKey)
	if err != nil {
		return TicketOperationResult{}, err
	}

	err = chc.NewTicketCommenter().AddComment(s.ctx, chc.AddTicketCommentOptions{
		Path:      ticket.Path,
		Author:    "tui",
		Body:      "TUI action recorded.",
		CreatedAt: time.Now().UTC(),
	})
	if err != nil {
		return TicketOperationResult{}, err
	}

	tickets, err := s.listProjectTickets()
	if err != nil {
		return TicketOperationResult{}, err
	}

	return TicketOperationResult{
		Message: "Added comment to ticket " + ticket.Key + ".",
		Tickets: tickets,
	}, nil
}

func (s service) MoveSelectedTicketStatus(ticketKey, status string) (TicketOperationResult, error) {
	if status == "" {
		return TicketOperationResult{}, errors.New("ticket status is required")
	}

	ticket, err := s.ticketByKey(ticketKey)
	if err != nil {
		return TicketOperationResult{}, err
	}

	err = chc.NewTicketStatusUpdater().UpdateStatus(s.ctx, chc.UpdateTicketStatusOptions{
		Path:      ticket.Path,
		Status:    status,
		UpdatedAt: time.Now().UTC(),
	})
	if err != nil {
		return TicketOperationResult{}, err
	}

	tickets, err := s.listProjectTickets()
	if err != nil {
		return TicketOperationResult{}, err
	}

	return TicketOperationResult{
		Message: "Moved ticket " + ticket.Key + " to " + status + ".",
		Tickets: tickets,
	}, nil
}

func (s service) listProjectTickets() ([]chc.IndexedTicket, error) {
	index, err := chc.NewTicketIndex(s.ctx, s.env, s.assetsFS)
	if err != nil {
		return nil, err
	}

	defer func() {
		//nolint:errcheck // List result already comes from the main operation.
		_ = index.Close()
	}()

	err = refreshProjectTicketIndex(s.ctx, index, s.projectContext)
	if err != nil {
		return nil, err
	}

	return index.ListProjectTickets(s.ctx, s.projectContext.ProjectID)
}

func (s service) ticketByKey(key string) (chc.IndexedTicket, error) {
	index, err := chc.NewTicketIndex(s.ctx, s.env, s.assetsFS)
	if err != nil {
		return chc.IndexedTicket{}, err
	}

	defer func() {
		//nolint:errcheck // Lookup result already comes from the main operation.
		_ = index.Close()
	}()

	err = refreshProjectTicketIndex(s.ctx, index, s.projectContext)
	if err != nil {
		return chc.IndexedTicket{}, err
	}

	return index.GetProjectTicketByKey(s.ctx, s.projectContext.ProjectID, key)
}
