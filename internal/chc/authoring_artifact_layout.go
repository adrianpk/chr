package chc

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

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

type AuthoringArtifactLayout struct {
	Ref       AuthoringArtifactRef
	FullPath  string
	Dir       string
	Collision AuthoringArtifactCollision
}

type AuthoringArtifactCollision struct {
	Exists  bool
	Path    string
	Allowed bool
}

type AuthoringArtifactInventoryEntry struct {
	SessionID  string
	ArtifactID string
	Kind       AuthoringArtifactKind
	Role       AuthoringArtifactRole
	Path       string
	Exists     bool
	Collision  bool
}

func PlanAuthoringArtifactLayout(opts AuthoringArtifactLayoutOptions) (AuthoringArtifactLayout, error) {
	err := validateAuthoringArtifactLayoutOptions(opts)
	if err != nil {
		return AuthoringArtifactLayout{}, err
	}

	relativePath := authoringArtifactRelativePath(opts)
	fullPath := filepath.Join(opts.ProjectDir, filepath.FromSlash(relativePath))

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

func authoringArtifactRelativePath(opts AuthoringArtifactLayoutOptions) string {
	name := strings.TrimSpace(opts.Name)
	if name == "" {
		name = opts.Topic
	}

	fileName := authoringArtifactFileName(opts.SessionID, name)
	parts := []string{DirectoryName}
	parts = append(parts, authoringArtifactDirectory(opts.Kind))
	parts = append(parts, fileName)

	return filepath.ToSlash(filepath.Join(parts...))
}

func authoringArtifactDirectory(kind AuthoringArtifactKind) string {
	directories := map[AuthoringArtifactKind]string{
		AuthoringArtifactKindDump:     "dumps",
		AuthoringArtifactKindQuiz:     "quizzes",
		AuthoringArtifactKindDraft:    "drafts",
		AuthoringArtifactKindSpec:     "specs",
		AuthoringArtifactKindPlan:     "plans",
		AuthoringArtifactKindSliceSet: "slice-sets",
		AuthoringArtifactKindSlice:    "slices",
		AuthoringArtifactKindTracker:  "trackers",
		AuthoringArtifactKindTicket:   "tickets",
	}

	directory, ok := directories[kind]
	if !ok {
		return "artifacts"
	}

	return directory
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
	_, err := os.Stat(path)
	if err == nil {
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
