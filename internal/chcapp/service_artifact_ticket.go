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

type CreateAuthoringSessionResult struct {
	Session chc.AuthoringSession
	Paths   chc.AuthoringSessionPaths
}

type CreateTicketRequest struct {
	Title       string
	Status      string
	Priority    string
	Severity    string
	Type        string
	Tags        []string
	UseCase     string
	ProjectArea string
	Body        string
}

type CreateTicketResult struct {
	Key  string
	Path string
}

func (s service) CreateTicket(request CreateTicketRequest) (CreateTicketResult, error) {
	config, err := chc.LoadWorkspaceConfig(s.projectContext.ProjectDir)
	if err != nil {
		return CreateTicketResult{}, err
	}

	repository := chc.NewTicketRepository(
		s.projectContext.ProjectDir,
		chc.WithTicketRepositoryKeyPrefix(config.TicketKeyPrefix),
	)

	result, err := repository.CreateTicket(s.ctx, chc.CreateTicketOptions{
		Title:       request.Title,
		Status:      request.Status,
		Priority:    request.Priority,
		Severity:    request.Severity,
		Type:        request.Type,
		Tags:        request.Tags,
		UseCase:     request.UseCase,
		ProjectArea: request.ProjectArea,
		Body:        request.Body,
	})
	if err != nil {
		return CreateTicketResult{}, err
	}

	return CreateTicketResult{
		Key:  result.Key,
		Path: result.Path,
	}, nil
}

func (s service) CreateDefaultTicket() (ArtifactOperationResult, error) {
	result, err := s.CreateTicket(CreateTicketRequest{
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
		Target:    chc.AuthoringTargetSpec,
		Topic:     "TUI artifact creation",
		CreatedAt: now,
		UpdatedAt: now,
	}

	result, err := s.CreateAuthoringSession(session)
	if err != nil {
		return ArtifactOperationResult{}, err
	}

	reloaded, err := s.reloadArtifactOperationState()
	if err != nil {
		return ArtifactOperationResult{}, err
	}

	reloaded.Message = "Created authoring session " + session.ID + " at " +
		displayPath(s.projectContext.ProjectDir, result.Paths.SessionPath)

	return reloaded, nil
}

func (s service) CreateAuthoringSession(session chc.AuthoringSession) (CreateAuthoringSessionResult, error) {
	paths, err := chc.NewAuthoringSessionRepository(s.projectContext.ProjectDir).CreateSession(s.ctx, session)
	if err != nil {
		return CreateAuthoringSessionResult{}, err
	}

	return CreateAuthoringSessionResult{
		Session: session,
		Paths:   paths,
	}, nil
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

type TicketCommentRequest struct {
	Key    string
	Author string
	Body   string
}

func (s service) AddSelectedTicketComment(ticketKey string) (TicketOperationResult, error) {
	return s.AddTicketComment(TicketCommentRequest{
		Key:    ticketKey,
		Author: "tui",
		Body:   "TUI action recorded.",
	})
}

func (s service) AddTicketComment(request TicketCommentRequest) (TicketOperationResult, error) {
	ticket, err := s.ticketByKey(request.Key)
	if err != nil {
		return TicketOperationResult{}, err
	}

	err = chc.NewTicketCommenter().AddComment(s.ctx, chc.AddTicketCommentOptions{
		Path:      ticket.Path,
		Author:    request.Author,
		Body:      request.Body,
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
	return s.MoveTicketStatus(ticketKey, status)
}

func (s service) MoveTicketStatus(ticketKey, status string) (TicketOperationResult, error) {
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
