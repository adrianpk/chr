package core

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"
)

var opsArtifactIDPattern = regexp.MustCompile(`^(\d+)-`)

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
func CreateOpsArtifact(opts CreateOpsArtifactOptions) (OpsArtifact, error) {
	if _, err := opts.Layout.ArtifactRoot(opts.Type); err != nil {
		return OpsArtifact{}, err
	}
	if strings.TrimSpace(opts.Title) == "" {
		return OpsArtifact{}, fmt.Errorf("ops artifact title is required")
	}
	if strings.TrimSpace(opts.Producer) == "" {
		return OpsArtifact{}, fmt.Errorf("ops artifact producer is required")
	}

	now := time.Now
	if opts.Now != nil {
		now = opts.Now
	}

	lockPath := filepath.Join(opts.Layout.Root(), "lock", "id.lock")
	if err := acquireOpsArtifactLock(lockPath); err != nil {
		return OpsArtifact{}, err
	}
	defer func() { _ = os.Remove(lockPath) }()

	slug := opsArtifactSlug(opts.Title)
	if slug == "" {
		slug = "artifact"
	}
	id := fmt.Sprintf("%05d-%s", nextOpsArtifactNumber(opts.Layout.Root()), slug)
	dir, err := opts.Layout.ArtifactDirectory(opts.Type, id)
	if err != nil {
		return OpsArtifact{}, err
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return OpsArtifact{}, fmt.Errorf("create ops artifact directory: %w", err)
	}

	timestamp := now().UTC().Format(time.RFC3339)
	metadata := map[string]any{
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
	encoded, err := json.MarshalIndent(metadata, "", "  ")
	if err != nil {
		return OpsArtifact{}, fmt.Errorf("encode ops artifact metadata: %w", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "artifact.md"), []byte("# "+strings.TrimSpace(opts.Title)+"\n"), 0o644); err != nil {
		return OpsArtifact{}, fmt.Errorf("write ops artifact body: %w", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "metadata.json"), append(encoded, '\n'), 0o644); err != nil {
		return OpsArtifact{}, fmt.Errorf("write ops artifact metadata: %w", err)
	}
	for _, path := range []string{"files", "evidence", filepath.Join("extension", strings.TrimSpace(opts.Producer))} {
		if err := os.MkdirAll(filepath.Join(dir, path), 0o755); err != nil {
			return OpsArtifact{}, fmt.Errorf("create ops artifact envelope: %w", err)
		}
	}

	return OpsArtifact{ID: id, Type: opts.Type, Dir: dir}, nil
}

func acquireOpsArtifactLock(path string) error {
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

func nextOpsArtifactNumber(namespaceRoot string) int {
	maximum := 0
	for artifactType := range opsArtifactTypes {
		entries, err := os.ReadDir(filepath.Join(namespaceRoot, artifactType))
		if err != nil {
			continue
		}
		for _, entry := range entries {
			if match := opsArtifactIDPattern.FindStringSubmatch(entry.Name()); len(match) == 2 {
				var number int
				_, _ = fmt.Sscanf(match[1], "%d", &number)
				maximum = max(maximum, number)
			}
		}
	}
	return maximum + 1
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
func RefreshOpsArtifactMetadata(dir string, status string, now time.Time) error {
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
