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

var opsArtifactIDPattern = regexp.MustCompile(`^(\d+)-`)

// HasOpsArtifactNumberPrefix reports the existing numeric artifact path prefix.
func HasOpsArtifactNumberPrefix(name string) bool { return opsArtifactIDPattern.MatchString(name) }

// OpsArtifact is one allocated shared artifact envelope.
type OpsArtifact struct {
	ID   string
	Type string
	Dir  string
}

// CreateOpsArtifactOptions configures a new shared artifact envelope.
type CreateOpsArtifactOptions struct {
	Layout      OpsLayout
	Type        string
	Title       string
	Producer    string
	ExternalIDs []string
	Now         func() time.Time
}

// CreateOpsArtifact allocates a namespace-global ID and creates its envelope.
func CreateOpsArtifact(opts CreateOpsArtifactOptions) (artifact OpsArtifact, err error) {
	if err := validateOpsArtifactOptions(opts); err != nil {
		return OpsArtifact{}, err
	}

	lockPath := filepath.Join(opts.Layout.Root(), "lock", "id.lock")
	if err := AcquireOpsArtifactLock(lockPath); err != nil {
		return OpsArtifact{}, err
	}

	defer func() {
		if removeErr := os.Remove(lockPath); removeErr != nil && !errors.Is(removeErr, os.ErrNotExist) && err == nil {
			err = fmt.Errorf("remove ops artifact allocation lock: %w", removeErr)
		}
	}()

	id := fmt.Sprintf("%05d-%s", NextOpsArtifactNumber(opts.Layout.Root()), OpsArtifactSlugOrDefault(opts.Title))

	dir, err := opts.Layout.ArtifactDirectory(opts.Type, id)
	if err != nil {
		return OpsArtifact{}, err
	}

	if err := CreateOpsArtifactEnvelope(dir, opts.Producer); err != nil {
		return OpsArtifact{}, err
	}

	metadata := NewOpsArtifactMetadata(opts, id, opsArtifactNow(opts))
	if err := writeOpsArtifactFiles(dir, opts.Title, metadata); err != nil {
		return OpsArtifact{}, err
	}

	return OpsArtifact{ID: id, Type: opts.Type, Dir: dir}, nil
}

// EnsureOpsArtifactEnvelope creates or refreshes a known shared artifact envelope.
func EnsureOpsArtifactEnvelope(opts CreateOpsArtifactOptions, id string) (OpsArtifact, error) {
	if err := validateOpsArtifactOptions(opts); err != nil {
		return OpsArtifact{}, err
	}

	if strings.TrimSpace(id) == "" || filepath.Base(id) != id {
		return OpsArtifact{}, fmt.Errorf("invalid ops artifact id: %s", id)
	}

	dir, err := opts.Layout.ArtifactDirectory(opts.Type, id)
	if err != nil {
		return OpsArtifact{}, err
	}

	if err := CreateOpsArtifactEnvelope(dir, opts.Producer); err != nil {
		return OpsArtifact{}, err
	}

	if err := ensureOpsArtifactBody(filepath.Join(dir, "artifact.md"), opts.Title); err != nil {
		return OpsArtifact{}, err
	}

	metadata := NewOpsArtifactMetadata(opts, id, opsArtifactNow(opts))
	if err := ensureOpsArtifactMetadata(filepath.Join(dir, "metadata.json"), metadata); err != nil {
		return OpsArtifact{}, err
	}

	return OpsArtifact{ID: id, Type: opts.Type, Dir: dir}, nil
}

func validateOpsArtifactOptions(opts CreateOpsArtifactOptions) error {
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

func opsArtifactNow(opts CreateOpsArtifactOptions) time.Time {
	if opts.Now != nil {
		return opts.Now().UTC()
	}

	return time.Now().UTC()
}

// NewOpsArtifactMetadata constructs the original shared envelope metadata.
func NewOpsArtifactMetadata(opts CreateOpsArtifactOptions, id string, now time.Time) map[string]any {
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

// CreateOpsArtifactEnvelope creates the shared envelope directories.
func CreateOpsArtifactEnvelope(dir, producer string) error {
	paths := []string{"", "files", "evidence", filepath.Join("extension", strings.TrimSpace(producer))}
	for _, path := range paths {
		if err := os.MkdirAll(filepath.Join(dir, path), 0o755); err != nil {
			return fmt.Errorf("create ops artifact envelope: %w", err)
		}
	}

	return nil
}

func writeOpsArtifactFiles(dir, title string, metadata map[string]any) error {
	if err := ensureOpsArtifactBody(filepath.Join(dir, "artifact.md"), title); err != nil {
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

func ensureOpsArtifactBody(path, title string) error {
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

func ensureOpsArtifactMetadata(path string, metadata map[string]any) error {
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

// AcquireOpsArtifactLock creates the shared exclusive allocation lock.
// The caller owns guarded reclamation and removal after this returns.
func AcquireOpsArtifactLock(path string) error {
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

// NextOpsArtifactNumber observes the next namespace-wide artifact number.
// Callers that allocate must hold the namespace allocation lock.
func NextOpsArtifactNumber(namespaceRoot string) int {
	maximum := 0

	for artifactType := range opsArtifactTypes {
		entries, err := os.ReadDir(filepath.Join(namespaceRoot, artifactType))
		if err != nil {
			continue
		}

		for _, entry := range entries {
			maximum = max(maximum, opsArtifactNumber(entry.Name()))
		}
	}

	return maximum + 1
}

func opsArtifactNumber(name string) int {
	match := opsArtifactIDPattern.FindStringSubmatch(name)
	if len(match) != 2 {
		return 0
	}

	number, err := strconv.Atoi(match[1])
	if err != nil {
		return 0
	}

	return number
}

// OpsArtifactSlugOrDefault keeps the canonical ASCII ID slug and fallback.
func OpsArtifactSlugOrDefault(value string) string {
	slug := opsArtifactSlug(value)
	if slug == "" {
		return "artifact"
	}

	return slug
}

func opsArtifactSlug(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	value = regexp.MustCompile(`[^a-z0-9]+`).ReplaceAllString(value, "-")

	return strings.Trim(value, "-")
}

// OpsArtifactFiles returns the shared, non-private files in deterministic order.
func OpsArtifactFiles(dir string) ([]string, error) {
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

// RefreshOpsArtifactMetadata updates the shared file index and terminal state.
func RefreshOpsArtifactMetadata(dir, status string, now time.Time) error {
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

	files, err := OpsArtifactFiles(dir)
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
