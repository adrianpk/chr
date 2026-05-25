package chc

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"
	"unicode"
)

const DefaultTicketKeyPrefix = "CHC"

type TicketRepository interface {
	CreateTicket(ctx context.Context, opts CreateTicketOptions) (CreateTicketResult, error)
}

type CreateTicketOptions struct {
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

type TicketRepositoryOption func(*fileTicketRepository)

type fileTicketRepository struct {
	projectDir string
	keyPrefix  string
	renderer   TicketRenderer
	now        func() time.Time
}

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

func WithTicketRepositoryKeyPrefix(keyPrefix string) TicketRepositoryOption {
	return func(repository *fileTicketRepository) {
		if keyPrefix != "" {
			repository.keyPrefix = keyPrefix
		}
	}
}

func WithTicketRepositoryRenderer(renderer TicketRenderer) TicketRepositoryOption {
	return func(repository *fileTicketRepository) {
		if renderer != nil {
			repository.renderer = renderer
		}
	}
}

func WithTicketRepositoryClock(now func() time.Time) TicketRepositoryOption {
	return func(repository *fileTicketRepository) {
		if now != nil {
			repository.now = now
		}
	}
}

func (r *fileTicketRepository) CreateTicket(ctx context.Context, opts CreateTicketOptions) (CreateTicketResult, error) {
	err := ctx.Err()
	if err != nil {
		return CreateTicketResult{}, err
	}

	if strings.TrimSpace(opts.Title) == "" {
		return CreateTicketResult{}, fmt.Errorf("ticket title is required")
	}

	nextNumber, err := r.nextKeyNumber(ctx)
	if err != nil {
		return CreateTicketResult{}, err
	}

	now := r.now().UTC()
	key := fmt.Sprintf("%s-%d", r.keyPrefix, nextNumber)
	ticket := Ticket{
		Key:         key,
		Title:       opts.Title,
		Status:      ticketStatusOrDefault(opts.Status),
		Priority:    opts.Priority,
		Severity:    opts.Severity,
		Type:        opts.Type,
		Tags:        opts.Tags,
		UseCase:     opts.UseCase,
		ProjectArea: opts.ProjectArea,
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

func ticketStatusOrDefault(status string) string {
	status = strings.TrimSpace(status)
	if status == "" {
		return DefaultTicketStatus
	}

	return status
}

func (r *fileTicketRepository) nextKeyNumber(ctx context.Context) (int, error) {
	ticketsDir := filepath.Join(r.projectDir, ".chc", "tickets")
	highest := 0

	err := filepath.WalkDir(ticketsDir, func(path string, entry os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}

		err := ctx.Err()
		if err != nil {
			return err
		}

		if entry.IsDir() || filepath.Ext(path) != ".md" {
			return nil
		}

		number, found, err := r.ticketNumber(path)
		if err != nil {
			return err
		}

		if found && number > highest {
			highest = number
		}

		return nil
	})
	if errors.Is(err, os.ErrNotExist) {
		return 1, nil
	}

	if err != nil {
		return 0, fmt.Errorf("scan existing tickets: %w", err)
	}

	return highest + 1, nil
}

func (r *fileTicketRepository) ticketNumber(path string) (int, bool, error) {
	number, ok := keyNumberFromText(r.keyPrefix, filepath.Base(path))
	if ok {
		return number, true, nil
	}

	content, err := os.ReadFile(path)
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
	parts := []string{r.projectDir, ".chc", "tickets"}

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
