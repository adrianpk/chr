package core

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	opsfile "forge.adrianpk.com/chr/internal/core/ops/file"
	slugcore "forge.adrianpk.com/chr/internal/core/ops/slug"
	ticketcore "forge.adrianpk.com/chr/internal/core/ticket"
	workspacecore "forge.adrianpk.com/chr/internal/core/workspace"
)

// DefaultTicketKeyPrefix defines a package boundary constant.
const DefaultTicketKeyPrefix = "CHR"

// TicketRepository defines a package boundary interface.
type TicketRepository interface {
	CreateTicket(ctx context.Context, opts CreateTicketOptions) (CreateTicketResult, error)
}

// CreateTicketOptions describes a package boundary value.
type CreateTicketOptions struct {
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

// CreateTicketResult describes a package boundary value.
type CreateTicketResult struct {
	Key  string
	Path string
}

// TicketRepositoryOption defines a package boundary type.
type TicketRepositoryOption func(*fileTicketRepository)

type fileTicketRepository struct {
	projectDir           string
	keyPrefix            string
	renderer             ticketcore.TicketRenderer
	now                  func() time.Time
	nextKeyNumberFactory func(context.Context) (int, error)
}

// NewTicketRepository constructs the corresponding package service or value.
func NewTicketRepository(projectDir string, opts ...TicketRepositoryOption) TicketRepository {
	repository := &fileTicketRepository{
		projectDir: projectDir,
		keyPrefix:  DefaultTicketKeyPrefix,
		renderer:   ticketcore.NewTicketRenderer(),
		now:        time.Now,
	}

	for _, opt := range opts {
		opt(repository)
	}

	return repository
}

// WithTicketRepositoryKeyPrefix executes the corresponding package operation.
func WithTicketRepositoryKeyPrefix(keyPrefix string) TicketRepositoryOption {
	return func(repository *fileTicketRepository) {
		if keyPrefix != "" {
			repository.keyPrefix = keyPrefix
		}
	}
}

// WithTicketRepositoryRenderer executes the corresponding package operation.
func WithTicketRepositoryRenderer(renderer ticketcore.TicketRenderer) TicketRepositoryOption {
	return func(repository *fileTicketRepository) {
		if renderer != nil {
			repository.renderer = renderer
		}
	}
}

// WithTicketRepositoryClock executes the corresponding package operation.
func WithTicketRepositoryClock(now func() time.Time) TicketRepositoryOption {
	return func(repository *fileTicketRepository) {
		if now != nil {
			repository.now = now
		}
	}
}

// WithTicketRepositoryNextKeyNumber executes the corresponding package operation.
func WithTicketRepositoryNextKeyNumber(nextNumber int) TicketRepositoryOption {
	return func(repository *fileTicketRepository) {
		if nextNumber > 0 {
			repository.nextKeyNumberFactory = func(context.Context) (int, error) {
				return nextNumber, nil
			}
		}
	}
}

// CreateTicket executes the corresponding package operation.
func (r *fileTicketRepository) CreateTicket(ctx context.Context, opts CreateTicketOptions) (CreateTicketResult, error) {
	err := ctx.Err()
	if err != nil {
		return CreateTicketResult{}, err
	}

	if strings.TrimSpace(opts.Title) == "" {
		return CreateTicketResult{}, fmt.Errorf("ticket title is required")
	}

	if strings.TrimSpace(opts.Synopsis) == "" {
		return CreateTicketResult{}, fmt.Errorf("ticket synopsis is required")
	}

	nextNumber, err := r.allocateKeyNumber(ctx)
	if err != nil {
		return CreateTicketResult{}, err
	}

	now := r.now().UTC()
	key := fmt.Sprintf("%s-%d", r.keyPrefix, nextNumber)
	ticket := ticketcore.Ticket{
		Key:         key,
		Title:       opts.Title,
		Synopsis:    opts.Synopsis,
		Status:      ticketStatusOrDefault(opts.Status),
		Priority:    opts.Priority,
		Severity:    opts.Severity,
		Type:        opts.Type,
		Tags:        opts.Tags,
		UseCase:     opts.UseCase,
		ProjectArea: opts.ProjectArea,
		SourceKind:  opts.SourceKind,
		SourceRef:   opts.SourceRef,
		Slice:       opts.Slice,
		Body:        opts.Body,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	content, err := r.renderer.RenderTicket(ticket)
	if err != nil {
		return CreateTicketResult{}, err
	}

	path := r.ticketPath(opts.UseCase, key, opts.Title)

	err = os.MkdirAll(filepath.Dir(path), 0o755)
	if err != nil {
		return CreateTicketResult{}, fmt.Errorf("create ticket directory: %w", err)
	}

	err = os.WriteFile(path, []byte(content), 0o644)
	if err != nil {
		return CreateTicketResult{}, fmt.Errorf("create ticket file: %w", err)
	}

	return CreateTicketResult{Key: key, Path: path}, nil
}

// NextTicketKeyNumber executes the corresponding package operation.
func NextTicketKeyNumber(ctx context.Context, projectDir, keyPrefix string) (int, error) {
	repository := &fileTicketRepository{
		projectDir: projectDir,
		keyPrefix:  DefaultTicketKeyPrefix,
	}

	WithTicketRepositoryKeyPrefix(keyPrefix)(repository)

	return repository.nextKeyNumber(ctx)
}

func (r *fileTicketRepository) allocateKeyNumber(ctx context.Context) (int, error) {
	if r.nextKeyNumberFactory != nil {
		return r.nextKeyNumberFactory(ctx)
	}

	return r.nextKeyNumber(ctx)
}

func ticketStatusOrDefault(status string) string {
	status = strings.TrimSpace(status)
	if status == "" {
		return ticketcore.TicketStatusBacklog
	}

	return status
}

func (r *fileTicketRepository) ticketNumber(path string) (int, bool, error) {
	number, ok := keyNumberFromText(r.keyPrefix, filepath.Base(path))
	if ok {
		return number, true, nil
	}

	content, err := opsfile.Read(path)
	if err != nil {
		return 0, false, fmt.Errorf("read ticket metadata: %w", err)
	}

	number, ok = keyNumberFromText(r.keyPrefix, string(content))
	if ok {
		return number, true, nil
	}

	return 0, false, nil
}

func (r *fileTicketRepository) ticketPath(useCase, key, title string) string {
	parts := []string{workspacecore.UserPath(r.projectDir, "tickets")}

	slug := slugcore.Normalize(useCase)
	if slug != "" {
		parts = append(parts, slug)
	}

	parts = append(parts, key+"-"+slugcore.ArtifactName(title)+".md")

	return filepath.Join(parts...)
}

func keyNumberFromText(prefix, text string) (int, bool) {
	pattern := regexp.MustCompile(regexp.QuoteMeta(prefix) + `-(\d+)`)

	match := pattern.FindStringSubmatch(text)
	if len(match) != 2 {
		return 0, false
	}

	number, err := strconv.Atoi(match[1])
	if err != nil {
		return 0, false
	}

	return number, true
}
