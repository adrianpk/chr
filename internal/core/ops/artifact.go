package ops

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
)

var artifactIDPattern = regexp.MustCompile(`^(\d+)-`)

// HasArtifactNumberPrefix reports the existing numeric artifact path prefix.
func HasArtifactNumberPrefix(name string) bool { return artifactIDPattern.MatchString(name) }

// Artifact is one allocated shared artifact envelope.
type Artifact struct {
	ID   string
	Type string
	Dir  string
}

// ArtifactOptions configures a new shared artifact envelope.
type ArtifactOptions struct {
	Layout      Layout
	Type        string
	Title       string
	Producer    string
	ExternalIDs []string
	Now         func() time.Time
}

// CreateArtifact allocates a namespace-global ID and creates its envelope.
func CreateArtifact(opts ArtifactOptions) (artifact Artifact, err error) {
	if err := validateArtifactOptions(opts); err != nil {
		return Artifact{}, err
	}

	lockPath := filepath.Join(opts.Layout.Root(), "lock", "id.lock")
	if err := AcquireArtifactLock(lockPath); err != nil {
		return Artifact{}, err
	}

	defer func() {
		if removeErr := os.Remove(lockPath); removeErr != nil && !errors.Is(removeErr, os.ErrNotExist) && err == nil {
			err = fmt.Errorf("remove ops artifact allocation lock: %w", removeErr)
		}
	}()

	id := fmt.Sprintf("%05d-%s", NextArtifactNumber(opts.Layout.Root()), ArtifactSlugOrDefault(opts.Title))

	dir, err := opts.Layout.ArtifactDirectory(opts.Type, id)
	if err != nil {
		return Artifact{}, err
	}

	if err := CreateArtifactEnvelope(dir, opts.Producer); err != nil {
		return Artifact{}, err
	}

	metadata := NewArtifactMetadata(opts, id, artifactNow(opts))
	if err := writeArtifactFiles(dir, opts.Title, metadata); err != nil {
		return Artifact{}, err
	}

	return Artifact{ID: id, Type: opts.Type, Dir: dir}, nil
}

// EnsureArtifactEnvelope creates or refreshes a known shared artifact envelope.
func EnsureArtifactEnvelope(opts ArtifactOptions, id string) (Artifact, error) {
	if err := validateArtifactOptions(opts); err != nil {
		return Artifact{}, err
	}

	if strings.TrimSpace(id) == "" || filepath.Base(id) != id {
		return Artifact{}, fmt.Errorf("invalid ops artifact id: %s", id)
	}

	dir, err := opts.Layout.ArtifactDirectory(opts.Type, id)
	if err != nil {
		return Artifact{}, err
	}

	if err := CreateArtifactEnvelope(dir, opts.Producer); err != nil {
		return Artifact{}, err
	}

	if err := ensureArtifactBody(filepath.Join(dir, "artifact.md"), opts.Title); err != nil {
		return Artifact{}, err
	}

	metadata := NewArtifactMetadata(opts, id, artifactNow(opts))
	if err := ensureArtifactMetadata(filepath.Join(dir, "metadata.json"), metadata); err != nil {
		return Artifact{}, err
	}

	return Artifact{ID: id, Type: opts.Type, Dir: dir}, nil
}

func validateArtifactOptions(opts ArtifactOptions) error {
	if _, err := opts.Layout.ArtifactRoot(opts.Type); err != nil {
		return err
	}

	if strings.TrimSpace(opts.Title) == "" {
		return fmt.Errorf("ops artifact title is required")
	}

	if strings.TrimSpace(opts.Producer) == "" {
		return fmt.Errorf("ops artifact producer is required")
	}

	return nil
}

func artifactNow(opts ArtifactOptions) time.Time {
	if opts.Now != nil {
		return opts.Now().UTC()
	}

	return time.Now().UTC()
}

// NewArtifactMetadata constructs the original shared envelope metadata.
func NewArtifactMetadata(opts ArtifactOptions, id string, now time.Time) map[string]any {
	timestamp := now.Format(time.RFC3339)

	return map[string]any{
		"schema_version": "ops-layout/v1",
		"id":             id,
		"type":           opts.Type,
		"namespace":      opts.Layout.Namespace(),
		"title":          strings.TrimSpace(opts.Title),
		"status":         "created",
		"revision":       1,
		"created_at":     timestamp,
		"updated_at":     timestamp,
		"producer":       map[string]string{"name": strings.TrimSpace(opts.Producer), "version": "0.0.0"},
		"external_ids":   append([]string(nil), opts.ExternalIDs...),
		"sources":        []any{},
		"continuations":  []any{},
		"files":          []string{"artifact.md"},
	}
}

// CreateArtifactEnvelope creates the shared envelope directories.
func CreateArtifactEnvelope(dir, producer string) error {
	paths := []string{"", "files", "evidence", filepath.Join("extension", strings.TrimSpace(producer))}
	for _, path := range paths {
		if err := os.MkdirAll(filepath.Join(dir, path), 0o755); err != nil {
			return fmt.Errorf("create ops artifact envelope: %w", err)
		}
	}

	return nil
}

func writeArtifactFiles(dir, title string, metadata map[string]any) error {
	if err := ensureArtifactBody(filepath.Join(dir, "artifact.md"), title); err != nil {
		return err
	}

	encoded, err := json.MarshalIndent(metadata, "", "  ")
	if err != nil {
		return fmt.Errorf("encode ops artifact metadata: %w", err)
	}

	if err := os.WriteFile(filepath.Join(dir, "metadata.json"), append(encoded, '\n'), 0o644); err != nil {
		return fmt.Errorf("write ops artifact metadata: %w", err)
	}

	return nil
}

func ensureArtifactBody(path, title string) error {
	if _, err := os.Stat(path); err == nil {
		return nil
	} else if !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("stat ops artifact body: %w", err)
	}

	if err := os.WriteFile(path, []byte("# "+strings.TrimSpace(title)+"\n"), 0o644); err != nil {
		return fmt.Errorf("write ops artifact body: %w", err)
	}

	return nil
}

func ensureArtifactMetadata(path string, metadata map[string]any) error {
	if _, err := os.Stat(path); err == nil {
		return nil
	} else if !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("stat ops artifact metadata: %w", err)
	}

	encoded, err := json.MarshalIndent(metadata, "", "  ")
	if err != nil {
		return fmt.Errorf("encode ops artifact metadata: %w", err)
	}

	if err := os.WriteFile(path, append(encoded, '\n'), 0o644); err != nil {
		return fmt.Errorf("write ops artifact metadata: %w", err)
	}

	return nil
}

// AcquireArtifactLock creates the shared exclusive allocation lock.
// The caller owns guarded reclamation and removal after this returns.
func AcquireArtifactLock(path string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("create ops lock directory: %w", err)
	}

	file, err := os.OpenFile(path, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o644)
	if os.IsExist(err) {
		return fmt.Errorf("ops artifact allocation lock exists: %s", path)
	}

	if err != nil {
		return fmt.Errorf("create ops artifact allocation lock: %w", err)
	}

	return file.Close()
}

// NextArtifactNumber observes the next namespace-wide artifact number.
// Callers that allocate must hold the namespace allocation lock.
func NextArtifactNumber(namespaceRoot string) int {
	maximum := 0

	for artifactType := range artifactTypes {
		entries, err := os.ReadDir(filepath.Join(namespaceRoot, artifactType))
		if err != nil {
			continue
		}

		for _, entry := range entries {
			maximum = max(maximum, artifactNumber(entry.Name()))
		}
	}

	return maximum + 1
}

func artifactNumber(name string) int {
	match := artifactIDPattern.FindStringSubmatch(name)
	if len(match) != 2 {
		return 0
	}

	number, err := strconv.Atoi(match[1])
	if err != nil {
		return 0
	}

	return number
}

// ArtifactSlugOrDefault keeps the canonical ASCII ID slug and fallback.
func ArtifactSlugOrDefault(value string) string {
	slug := artifactSlug(value)
	if slug == "" {
		return "artifact"
	}

	return slug
}

func artifactSlug(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	value = regexp.MustCompile(`[^a-z0-9]+`).ReplaceAllString(value, "-")

	return strings.Trim(value, "-")
}

// ArtifactFiles returns the shared, non-private files in deterministic order.
func ArtifactFiles(dir string) ([]string, error) {
	files := []string{}

	err := filepath.WalkDir(dir, func(path string, entry os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}

		if entry.IsDir() || path == filepath.Join(dir, "metadata.json") || strings.HasPrefix(path, filepath.Join(dir, "extension")+string(filepath.Separator)) {
			return nil
		}

		relative, err := filepath.Rel(dir, path)
		if err != nil {
			return err
		}

		files = append(files, filepath.ToSlash(relative))

		return nil
	})
	if err != nil {
		return nil, err
	}

	sort.Strings(files)

	return files, nil
}

// RefreshArtifactMetadata updates the shared file index and terminal state.
func RefreshArtifactMetadata(dir, status string, now time.Time) error {
	if strings.TrimSpace(status) == "" {
		return fmt.Errorf("ops artifact status is required")
	}

	path := filepath.Join(dir, "metadata.json")

	content, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("read ops artifact metadata: %w", err)
	}

	metadata := map[string]any{}
	if err := json.Unmarshal(content, &metadata); err != nil {
		return fmt.Errorf("decode ops artifact metadata: %w", err)
	}

	files, err := ArtifactFiles(dir)
	if err != nil {
		return err
	}

	metadata["files"] = files
	metadata["status"] = strings.TrimSpace(status)
	metadata["updated_at"] = now.UTC().Format(time.RFC3339)

	encoded, err := json.MarshalIndent(metadata, "", "  ")
	if err != nil {
		return fmt.Errorf("encode refreshed ops artifact metadata: %w", err)
	}

	if err := os.WriteFile(path, append(encoded, '\n'), 0o644); err != nil {
		return fmt.Errorf("write refreshed ops artifact metadata: %w", err)
	}

	return nil
}
