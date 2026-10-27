package ticket

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
	workspacecore "forge.adrianpk.com/chr/internal/core/workspace"
)

// DefaultKeyPrefix defines a package boundary constant.
const DefaultKeyPrefix = "CHR"

// Repository defines a package boundary interface.
type Repository interface {
	Create(ctx context.Context, opts CreateOptions) (CreateResult, error)
}

// CreateOptions describes a package boundary value.
type CreateOptions struct {
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

// CreateResult describes a package boundary value.
type CreateResult struct {
	Key  string
	Path string
}

// RepositoryOption defines a package boundary type.
type RepositoryOption func(*fileRepository)

type fileRepository struct {
	projectDir    string
	keyPrefix     string
	renderer      Renderer
	now           func() time.Time
	nextKeyNumber func(context.Context, string) (int, error)
}

// NewRepository stores the required key allocator and document dependencies.
func NewRepository(projectDir string, nextKeyNumber func(context.Context, string) (int, error), opts ...RepositoryOption) Repository {
	repository := &fileRepository{
		projectDir:    projectDir,
		nextKeyNumber: nextKeyNumber,
		keyPrefix:     DefaultKeyPrefix,
		renderer:      NewRenderer(),
		now:           time.Now,
	}

	for _, opt := range opts {
		opt(repository)
	}

	return repository
}

// WithKeyPrefix executes the corresponding package operation.
func WithKeyPrefix(keyPrefix string) RepositoryOption {
	return func(repository *fileRepository) {
		if keyPrefix != "" {
			repository.keyPrefix = keyPrefix
		}
	}
}

// WithRenderer executes the corresponding package operation.
func WithRenderer(renderer Renderer) RepositoryOption {
	return func(repository *fileRepository) {
		if renderer != nil {
			repository.renderer = renderer
		}
	}
}

// WithRepositoryClock executes the corresponding package operation.
func WithRepositoryClock(now func() time.Time) RepositoryOption {
	return func(repository *fileRepository) {
		if now != nil {
			repository.now = now
		}
	}
}

// Create executes the corresponding package operation.
func (r *fileRepository) Create(ctx context.Context, opts CreateOptions) (CreateResult, error) {
	err := ctx.Err()
	if err != nil {
		return CreateResult{}, err
	}

	if strings.TrimSpace(opts.Title) == "" {
		return CreateResult{}, fmt.Errorf("ticket title is required")
	}

	if strings.TrimSpace(opts.Synopsis) == "" {
		return CreateResult{}, fmt.Errorf("ticket synopsis is required")
	}

	nextNumber, err := r.nextKeyNumber(ctx, r.keyPrefix)
	if err != nil {
		return CreateResult{}, err
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

	content, err := r.renderer.Render(ticket)
	if err != nil {
		return CreateResult{}, err
	}

	path := r.ticketPath(opts.UseCase, key, opts.Title)

	err = os.MkdirAll(filepath.Dir(path), 0o755)
	if err != nil {
		return CreateResult{}, fmt.Errorf("create ticket directory: %w", err)
	}

	err = os.WriteFile(path, []byte(content), 0o644)
	if err != nil {
		return CreateResult{}, fmt.Errorf("create ticket file: %w", err)
	}

	return CreateResult{Key: key, Path: path}, nil
}

func ticketStatusOrDefault(status string) string {
	status = strings.TrimSpace(status)
	if status == "" {
		return StatusBacklog
	}

	return status
}

// NumberObservation preserves whether a ticket key was present.
type NumberObservation struct {
	Number int
	Found  bool
}

// ReadNumber observes a key in the filename before reading ticket metadata.
func ReadNumber(path, keyPrefix string) (NumberObservation, error) {
	number, ok := NumberFromText(keyPrefix, filepath.Base(path))
	if ok {
		return NumberObservation{Number: number, Found: true}, nil
	}

	content, err := opsfile.Read(path)
	if err != nil {
		return NumberObservation{}, fmt.Errorf("read ticket metadata: %w", err)
	}

	number, ok = NumberFromText(keyPrefix, string(content))
	if ok {
		return NumberObservation{Number: number, Found: true}, nil
	}

	return NumberObservation{}, nil
}

func (r *fileRepository) ticketPath(useCase, key, title string) string {
	parts := []string{workspacecore.UserPath(r.projectDir, "tickets")}

	slug := slugcore.Normalize(useCase)
	if slug != "" {
		parts = append(parts, slug)
	}

	parts = append(parts, key+"-"+slugcore.ArtifactName(title)+".md")

	return filepath.Join(parts...)
}

// NumberFromText observes the existing prefix-number spelling.
func NumberFromText(prefix, text string) (int, bool) {
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
