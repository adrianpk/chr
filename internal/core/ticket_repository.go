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
	"unicode"
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
	renderer             TicketRenderer
	now                  func() time.Time
	nextKeyNumberFactory func(context.Context) (int, error)
}

// NewTicketRepository constructs the corresponding package service or value.
func NewTicketRepository(projectDir string, opts ...TicketRepositoryOption) TicketRepository {
	repository := &fileTicketRepository{
		projectDir: projectDir,
		keyPrefix:  DefaultTicketKeyPrefix,
		renderer:   NewTicketRenderer(),
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
func WithTicketRepositoryRenderer(renderer TicketRenderer) TicketRepositoryOption {
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
	ticket := Ticket{
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
		return TicketStatusBacklog
	}

	return status
}

func (r *fileTicketRepository) ticketNumber(path string) (int, bool, error) {
	number, ok := keyNumberFromText(r.keyPrefix, filepath.Base(path))
	if ok {
		return number, true, nil
	}

	content, err := readWorkflowFile(path)
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
	parts := []string{UserPath(r.projectDir, "tickets")}

	slug := slugify(useCase)
	if slug != "" {
		parts = append(parts, slug)
	}

	parts = append(parts, key+"-"+truncatedName(title)+".md")

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

func truncatedName(value string) string {
	const maxLength = 48

	slug := slugify(value)
	if len(slug) <= maxLength {
		return slug
	}

	return strings.Trim(slug[:maxLength], "-")
}

func slugify(value string) string {
	value = strings.TrimSpace(value)

	var output strings.Builder

	lastDash := false

	for _, char := range value {
		if unicode.IsLetter(char) || unicode.IsDigit(char) {
			output.WriteRune(char)

			lastDash = false

			continue
		}

		if !lastDash && output.Len() > 0 {
			output.WriteByte('-')

			lastDash = true
		}
	}

	return strings.Trim(output.String(), "-")
}
