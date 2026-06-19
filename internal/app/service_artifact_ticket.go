package app

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"forge.adrianpk.com/chc/internal/core"
)

// ArtifactOperationResult is the TUI-facing result for artifact creation
// actions.
type ArtifactOperationResult struct {
	Message           string
	Tickets           []core.IndexedTicket
	AuthoringSessions []core.AuthoringSession
}

// CreateAuthoringSessionResult identifies a persisted authoring session and
// the files created for it.
type CreateAuthoringSessionResult struct {
	Session core.AuthoringSession
	Paths   core.AuthoringSessionPaths
}

// CreateTicketRequest carries the metadata and body for a new repository
// ticket.
type CreateTicketRequest struct {
	Title       string
	Synopsis    string
	Status      string
	Priority    string
	Severity    string
	Type        string
	Tags        []string
	UseCase     string
	ProjectArea string
	SourceKind  string
	SourceRef   string
	Slice       string
	Body        string
}

// CreateTicketResult identifies the created ticket.
type CreateTicketResult struct {
	Key  string
	Path string
}

func (s service) CreateTicket(request CreateTicketRequest) (CreateTicketResult, error) {
	if strings.TrimSpace(request.Title) == "" {
		return CreateTicketResult{}, errors.New("ticket title is required")
	}

	if strings.TrimSpace(request.Synopsis) == "" {
		return CreateTicketResult{}, errors.New("ticket synopsis is required")
	}

	config, err := core.LoadWorkspaceConfig(s.projectContext.ProjectDir)
	if err != nil {
		return CreateTicketResult{}, err
	}

	nextNumber, err := s.reserveNextTicketKeyNumber(config.TicketKeyPrefix)
	if err != nil {
		return CreateTicketResult{}, err
	}

	repository := core.NewTicketRepository(
		s.projectContext.ProjectDir,
		core.WithTicketRepositoryKeyPrefix(config.TicketKeyPrefix),
		core.WithTicketRepositoryNextKeyNumber(nextNumber),
	)

	result, err := repository.CreateTicket(s.ctx, core.CreateTicketOptions{
		Title:       request.Title,
		Synopsis:    request.Synopsis,
		Status:      request.Status,
		Priority:    request.Priority,
		Severity:    request.Severity,
		Type:        request.Type,
		Tags:        request.Tags,
		UseCase:     request.UseCase,
		ProjectArea: request.ProjectArea,
		SourceKind:  request.SourceKind,
		SourceRef:   request.SourceRef,
		Slice:       request.Slice,
		Body:        request.Body,
	})
	if err != nil {
		return CreateTicketResult{}, err
	}

	err = s.refreshCreatedTicketIndex(result.Path)
	if err != nil {
		return CreateTicketResult{}, err
	}

	return CreateTicketResult{
		Key:  result.Key,
		Path: result.Path,
	}, nil
}

func (s service) reserveNextTicketKeyNumber(keyPrefix string) (int, error) {
	seed, err := core.NextTicketKeyNumber(s.ctx, s.projectContext.ProjectDir, keyPrefix)
	if err != nil {
		return 0, err
	}

	repository, err := core.NewTicketKeySequenceRepository(s.ctx, s.env, s.assetsFS)
	if err != nil {
		return 0, err
	}

	defer func() {
		//nolint:errcheck // The caller only needs the reserved key number.
		_ = repository.Close()
	}()

	return repository.ReserveNextTicketKey(s.ctx, core.ReserveNextTicketKeyOptions{
		ProjectID: s.projectContext.ProjectID,
		Prefix:    keyPrefix,
		Seed:      seed,
		UpdatedAt: time.Now().UTC(),
	})
}

func (s service) CreateDefaultTicket() (ArtifactOperationResult, error) {
	result, err := s.CreateTicket(CreateTicketRequest{
		Title:       "TUI created ticket",
		Synopsis:    "Ticket created from the TUI Board action.",
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

func (s service) refreshCreatedTicketIndex(ticketPath string) error {
	index, err := core.NewTicketIndex(s.ctx, s.env, s.assetsFS)
	if err != nil {
		return err
	}

	defer func() {
		//nolint:errcheck // The caller only needs the reconciliation result.
		_ = index.Close()
	}()

	result, err := core.NewTicketReconciler(index).ReconcileProjectTicket(s.ctx, core.ReconcileProjectTicketOptions{
		ProjectID:    s.projectContext.ProjectID,
		ProjectDir:   s.projectContext.ProjectDir,
		TicketPath:   ticketPath,
		ReconciledAt: time.Now().UTC(),
	})
	if err != nil {
		return err
	}

	if result.HasBlockers() {
		firstBlocker, _ := result.FirstBlocker()

		return errors.New("refresh created ticket index: " + firstBlocker.Err.Error())
	}

	return nil
}

func (s service) CreateTUIAuthoringSession() (ArtifactOperationResult, error) {
	now := time.Now().UTC()
	session := core.AuthoringSession{
		ID:        authoringSessionID(now),
		ProjectID: s.projectContext.ProjectID,
		Status:    core.AuthoringSessionStatusDraft,
		Mode:      core.AuthoringSessionModeTUI,
		Target:    core.AuthoringTargetSpec,
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

func (s service) CreateAuthoringSession(session core.AuthoringSession) (CreateAuthoringSessionResult, error) {
	paths, err := core.NewAuthoringSessionRepository(s.projectContext.ProjectDir).CreateSession(s.ctx, session)
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

// TicketOperationResult is the TUI-facing result for ticket mutations.
type TicketOperationResult struct {
	Message string
	Tickets []core.IndexedTicket
}

// TicketCommentRequest describes a comment to append to a ticket.
type TicketCommentRequest struct {
	Key    string
	Author string
	Body   string
}

// TicketReviewReferenceRequest describes review metadata to attach to a ticket.
type TicketReviewReferenceRequest struct {
	Key    string
	Branch string
	Ref    *core.TicketReviewRef
}

// TicketEvidenceStrategyChecklistRequest describes a ticket strategy checklist write.
type TicketEvidenceStrategyChecklistRequest struct {
	Key       string
	Checklist core.EvidenceStrategyChecklist
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

	err = core.NewTicketCommenter().AddComment(s.ctx, core.AddTicketCommentOptions{
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

func (s service) AttachTicketReviewReference(request TicketReviewReferenceRequest) (TicketOperationResult, error) {
	ticket, err := s.ticketByKey(request.Key)
	if err != nil {
		return TicketOperationResult{}, err
	}

	err = core.NewTicketReviewMetadataUpdater().UpdateReviewMetadata(s.ctx, core.UpdateTicketReviewMetadataOptions{
		Path:      ticket.Path,
		Branch:    request.Branch,
		ReviewRef: request.Ref,
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
		Message: "Attached review metadata to ticket " + ticket.Key + ".",
		Tickets: tickets,
	}, nil
}

func (s service) UpsertTicketEvidenceStrategyChecklist(request TicketEvidenceStrategyChecklistRequest) (TicketOperationResult, error) {
	ticket, err := s.ticketByKey(request.Key)
	if err != nil {
		return TicketOperationResult{}, err
	}

	err = core.NewTicketEvidenceStrategyChecklistWriter().UpsertEvidenceStrategyChecklist(
		s.ctx,
		core.UpsertTicketEvidenceStrategyChecklistOptions{
			Path:      ticket.Path,
			Checklist: request.Checklist,
			UpdatedAt: time.Now().UTC(),
		},
	)
	if err != nil {
		return TicketOperationResult{}, err
	}

	tickets, err := s.listProjectTickets()
	if err != nil {
		return TicketOperationResult{}, err
	}

	return TicketOperationResult{
		Message: "Updated evidence strategy checklist for ticket " + ticket.Key + ".",
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

	err = s.validateReadyForReviewEvidence(ticket, status)
	if err != nil {
		return TicketOperationResult{}, err
	}

	err = core.NewTicketStatusUpdater().UpdateStatus(s.ctx, core.UpdateTicketStatusOptions{
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

func (s service) validateReadyForReviewEvidence(ticket core.IndexedTicket, status string) error {
	if strings.TrimSpace(status) != core.TicketStatusReadyForReview {
		return nil
	}

	readiness, err := ticketEvidenceReadiness(ticketEvidenceRequest{
		ProjectDir: s.projectContext.ProjectDir,
		TicketPath: ticket.Path,
	})
	if err != nil {
		return err
	}

	if readiness.Allowed {
		return nil
	}

	return fmt.Errorf("ticket ready_for_review evidence gate failed: %s", readiness.BlockerReason())
}

func (s service) listProjectTickets() ([]core.IndexedTicket, error) {
	board, err := s.Board()
	if err != nil {
		return nil, err
	}

	return board.Tickets, nil
}

func (s service) Board() (core.Board, error) {
	index, err := core.NewTicketIndex(s.ctx, s.env, s.assetsFS)
	if err != nil {
		return core.Board{}, err
	}

	defer func() {
		//nolint:errcheck // List result already comes from the main operation.
		_ = index.Close()
	}()

	err = refreshProjectTicketIndex(s.ctx, index, s.projectContext)
	if err != nil {
		return core.Board{}, err
	}

	return core.NewBoardQuery(index).ProjectBoard(s.ctx, core.BoardQueryOptions{
		ProjectID: s.projectContext.ProjectID,
	})
}

func (s service) ticketByKey(key string) (core.IndexedTicket, error) {
	index, err := core.NewTicketIndex(s.ctx, s.env, s.assetsFS)
	if err != nil {
		return core.IndexedTicket{}, err
	}

	defer func() {
		//nolint:errcheck // Lookup result already comes from the main operation.
		_ = index.Close()
	}()

	err = refreshProjectTicketIndex(s.ctx, index, s.projectContext)
	if err != nil {
		return core.IndexedTicket{}, err
	}

	return index.GetProjectTicketByKey(s.ctx, s.projectContext.ProjectID, key)
}
