package authoring

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	opscore "forge.adrianpk.com/chr/internal/core/ops"
	slugcore "forge.adrianpk.com/chr/internal/core/ops/slug"
)

// ArtifactLayoutOptions describes a package boundary value.
type ArtifactLayoutOptions struct {
	ProjectDir   string
	SessionID    string
	Kind         ArtifactKind
	Role         ArtifactRole
	Topic        string
	Name         string
	Description  string
	AllowReplace bool
}

// ArtifactLayout describes a package boundary value.
type ArtifactLayout struct {
	Ref       ArtifactRef
	FullPath  string
	Dir       string
	Collision ArtifactCollision
}

// ArtifactCollision describes a package boundary value.
type ArtifactCollision struct {
	Exists  bool
	Path    string
	Allowed bool
}

// ArtifactInventoryEntry describes a package boundary value.
type ArtifactInventoryEntry struct {
	SessionID  string
	ArtifactID string
	Kind       ArtifactKind
	Role       ArtifactRole
	Path       string
	Exists     bool
	Collision  bool
}

// PlanArtifactLayout executes the corresponding package operation.
func PlanArtifactLayout(opts ArtifactLayoutOptions) (ArtifactLayout, error) {
	err := validateAuthoringArtifactLayoutOptions(opts)
	if err != nil {
		return ArtifactLayout{}, err
	}

	body, err := authoringArtifactPath(opts)
	if err != nil {
		return ArtifactLayout{}, err
	}

	collision, err := authoringArtifactCollision(body.Path, opts.AllowReplace)
	if err != nil {
		return ArtifactLayout{}, err
	}

	return ArtifactLayout{
		Ref: ArtifactRef{
			ID:          authoringArtifactID(opts),
			Kind:        opts.Kind,
			Role:        opts.Role,
			Path:        body.RelativePath,
			Description: strings.TrimSpace(opts.Description),
		},
		FullPath:  body.Path,
		Dir:       filepath.Dir(body.Path),
		Collision: collision,
	}, nil
}

// ArtifactInventory executes the corresponding package operation.
func ArtifactInventory(sessionID string, layouts []ArtifactLayout) []ArtifactInventoryEntry {
	entries := make([]ArtifactInventoryEntry, 0, len(layouts))

	for _, layout := range layouts {
		entries = append(entries, ArtifactInventoryEntry{
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

func validateAuthoringArtifactLayoutOptions(opts ArtifactLayoutOptions) error {
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

	if opts.Kind == ArtifactKindSession || opts.Kind == ArtifactKindTranscript {
		return fmt.Errorf("authoring artifact kind is session-owned: %s", opts.Kind)
	}

	if strings.TrimSpace(opts.Topic) == "" && strings.TrimSpace(opts.Name) == "" {
		return fmt.Errorf("authoring artifact topic or name is required")
	}

	return nil
}

// EnsureAuthoringArtifactMetadata ensures metadata for an authoring artifact body.
func EnsureAuthoringArtifactMetadata(projectDir string, ref ArtifactRef) error {
	return opscore.EnsureArtifactBodyMetadata(projectDir, ref.Path)
}

func authoringArtifactPath(opts ArtifactLayoutOptions) (opscore.NamedArtifactPath, error) {
	name := strings.TrimSpace(opts.Name)
	if name == "" {
		name = opts.Topic
	}

	fileName := authoringArtifactFileName(opts.SessionID, name)
	artifactType := authoringArtifactType(opts.Kind)

	return opscore.NamedArtifactBody(opts.ProjectDir, artifactType, strings.TrimSuffix(fileName, ".md"))
}

func authoringArtifactType(kind ArtifactKind) string {
	types := map[ArtifactKind]string{
		ArtifactKindDump:     "dump",
		ArtifactKindQuiz:     "quiz",
		ArtifactKindDraft:    "spec",
		ArtifactKindSpec:     "spec",
		ArtifactKindPlan:     "plan",
		ArtifactKindSliceSet: "plan",
		ArtifactKindSlice:    "plan",
		ArtifactKindTracker:  "tracker",
		ArtifactKindTicket:   "ticket",
	}

	artifactType, ok := types[kind]
	if !ok {
		return "report"
	}

	return artifactType
}

func authoringArtifactFileName(sessionID, name string) string {
	slug := strings.ToLower(slugcore.ArtifactName(name))
	if slug == "" {
		slug = "artifact"
	}

	sessionSlug := strings.ToLower(slugcore.ArtifactName(sessionID))
	if sessionSlug == "" {
		return slug + ".md"
	}

	return sessionSlug + "-" + slug + ".md"
}

func authoringArtifactID(opts ArtifactLayoutOptions) string {
	name := strings.TrimSpace(opts.Name)
	if name == "" {
		name = opts.Topic
	}

	return string(opts.Kind) + ":" + strings.ToLower(slugcore.ArtifactName(opts.SessionID)) + ":" + strings.ToLower(slugcore.ArtifactName(name))
}

func authoringArtifactCollision(path string, allowReplace bool) (ArtifactCollision, error) {
	info, err := os.Stat(path)
	if err == nil {
		if info.Size() == 0 {
			return ArtifactCollision{Path: path, Allowed: true}, nil
		}

		return ArtifactCollision{
			Exists:  true,
			Path:    path,
			Allowed: allowReplace,
		}, nil
	}

	if !os.IsNotExist(err) {
		return ArtifactCollision{}, fmt.Errorf("stat authoring artifact: %w", err)
	}

	return ArtifactCollision{Path: path, Allowed: true}, nil
}
