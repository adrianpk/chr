package core

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	opscore "forge.adrianpk.com/chr/internal/core/ops"
)

// AuthoringArtifactLayoutOptions describes a package boundary value.
type AuthoringArtifactLayoutOptions struct {
	ProjectDir   string
	SessionID    string
	Kind         AuthoringArtifactKind
	Role         AuthoringArtifactRole
	Topic        string
	Name         string
	Description  string
	AllowReplace bool
}

// AuthoringArtifactLayout describes a package boundary value.
type AuthoringArtifactLayout struct {
	Ref       AuthoringArtifactRef
	FullPath  string
	Dir       string
	Collision AuthoringArtifactCollision
}

// AuthoringArtifactCollision describes a package boundary value.
type AuthoringArtifactCollision struct {
	Exists  bool
	Path    string
	Allowed bool
}

// AuthoringArtifactInventoryEntry describes a package boundary value.
type AuthoringArtifactInventoryEntry struct {
	SessionID  string
	ArtifactID string
	Kind       AuthoringArtifactKind
	Role       AuthoringArtifactRole
	Path       string
	Exists     bool
	Collision  bool
}

// PlanAuthoringArtifactLayout executes the corresponding package operation.
func PlanAuthoringArtifactLayout(opts AuthoringArtifactLayoutOptions) (AuthoringArtifactLayout, error) {
	err := validateAuthoringArtifactLayoutOptions(opts)
	if err != nil {
		return AuthoringArtifactLayout{}, err
	}

	relativePath, fullPath, err := authoringArtifactPath(opts)
	if err != nil {
		return AuthoringArtifactLayout{}, err
	}

	collision, err := authoringArtifactCollision(fullPath, opts.AllowReplace)
	if err != nil {
		return AuthoringArtifactLayout{}, err
	}

	return AuthoringArtifactLayout{
		Ref: AuthoringArtifactRef{
			ID:          authoringArtifactID(opts),
			Kind:        opts.Kind,
			Role:        opts.Role,
			Path:        relativePath,
			Description: strings.TrimSpace(opts.Description),
		},
		FullPath:  fullPath,
		Dir:       filepath.Dir(fullPath),
		Collision: collision,
	}, nil
}

// AuthoringArtifactInventory executes the corresponding package operation.
func AuthoringArtifactInventory(sessionID string, layouts []AuthoringArtifactLayout) []AuthoringArtifactInventoryEntry {
	entries := make([]AuthoringArtifactInventoryEntry, 0, len(layouts))

	for _, layout := range layouts {
		entries = append(entries, AuthoringArtifactInventoryEntry{
			SessionID:  sessionID,
			ArtifactID: layout.Ref.ID,
			Kind:       layout.Ref.Kind,
			Role:       layout.Ref.Role,
			Path:       layout.Ref.Path,
			Exists:     layout.Collision.Exists,
			Collision:  layout.Collision.Exists && !layout.Collision.Allowed,
		})
	}

	return entries
}

func validateAuthoringArtifactLayoutOptions(opts AuthoringArtifactLayoutOptions) error {
	if opts.ProjectDir == "" {
		return fmt.Errorf("authoring artifact project dir is required")
	}

	if opts.SessionID == "" {
		return fmt.Errorf("authoring artifact session id is required")
	}

	if !opts.Kind.Valid() {
		return fmt.Errorf("authoring artifact kind is invalid: %s", opts.Kind)
	}

	if !opts.Role.Valid() {
		return fmt.Errorf("authoring artifact role is invalid: %s", opts.Role)
	}

	if opts.Kind == AuthoringArtifactKindSession || opts.Kind == AuthoringArtifactKindTranscript {
		return fmt.Errorf("authoring artifact kind is session-owned: %s", opts.Kind)
	}

	if strings.TrimSpace(opts.Topic) == "" && strings.TrimSpace(opts.Name) == "" {
		return fmt.Errorf("authoring artifact topic or name is required")
	}

	return nil
}

// EnsureAuthoringArtifactMetadata ensures metadata for an authoring artifact body.
func EnsureAuthoringArtifactMetadata(projectDir string, ref AuthoringArtifactRef) error {
	return opscore.EnsureOpsArtifactBodyMetadata(projectDir, ref.Path)
}

func authoringArtifactPath(opts AuthoringArtifactLayoutOptions) (string, string, error) {
	name := strings.TrimSpace(opts.Name)
	if name == "" {
		name = opts.Topic
	}

	fileName := authoringArtifactFileName(opts.SessionID, name)
	artifactType := authoringArtifactType(opts.Kind)

	body, err := opscore.NamedOpsArtifactBody(opts.ProjectDir, artifactType, strings.TrimSuffix(fileName, ".md"))

	return body.RelativePath, body.Path, err
}

func authoringArtifactType(kind AuthoringArtifactKind) string {
	types := map[AuthoringArtifactKind]string{
		AuthoringArtifactKindDump:     "dump",
		AuthoringArtifactKindQuiz:     "quiz",
		AuthoringArtifactKindDraft:    "spec",
		AuthoringArtifactKindSpec:     "spec",
		AuthoringArtifactKindPlan:     "plan",
		AuthoringArtifactKindSliceSet: "plan",
		AuthoringArtifactKindSlice:    "plan",
		AuthoringArtifactKindTracker:  "tracker",
		AuthoringArtifactKindTicket:   "ticket",
	}

	artifactType, ok := types[kind]
	if !ok {
		return "report"
	}

	return artifactType
}

func authoringArtifactFileName(sessionID, name string) string {
	slug := strings.ToLower(truncatedName(name))
	if slug == "" {
		slug = "artifact"
	}

	sessionSlug := strings.ToLower(truncatedName(sessionID))
	if sessionSlug == "" {
		return slug + ".md"
	}

	return sessionSlug + "-" + slug + ".md"
}

func authoringArtifactID(opts AuthoringArtifactLayoutOptions) string {
	name := strings.TrimSpace(opts.Name)
	if name == "" {
		name = opts.Topic
	}

	return string(opts.Kind) + ":" + strings.ToLower(truncatedName(opts.SessionID)) + ":" + strings.ToLower(truncatedName(name))
}

func authoringArtifactCollision(path string, allowReplace bool) (AuthoringArtifactCollision, error) {
	info, err := os.Stat(path)
	if err == nil {
		if info.Size() == 0 {
			return AuthoringArtifactCollision{Path: path, Allowed: true}, nil
		}

		return AuthoringArtifactCollision{
			Exists:  true,
			Path:    path,
			Allowed: allowReplace,
		}, nil
	}

	if !os.IsNotExist(err) {
		return AuthoringArtifactCollision{}, fmt.Errorf("stat authoring artifact: %w", err)
	}

	return AuthoringArtifactCollision{Path: path, Allowed: true}, nil
}
