package core

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	authoringcore "forge.adrianpk.com/chr/internal/core/authoring"
	opsfile "forge.adrianpk.com/chr/internal/core/ops/file"
	opsmetadata "forge.adrianpk.com/chr/internal/core/ops/metadata"
	workspacecore "forge.adrianpk.com/chr/internal/core/workspace"
)

// AuthoringSessionsDirectory defines a package boundary constant.
const AuthoringSessionsDirectory = "authoring/sessions"

// ErrAuthoringSessionNotFound identifies a package boundary error.
var ErrAuthoringSessionNotFound = errors.New("authoring session not found")

// AuthoringSessionRepository defines a package boundary interface.
type AuthoringSessionRepository interface {
	CreateSession(ctx context.Context, session authoringcore.Session) (AuthoringSessionPaths, error)
	UpdateSession(ctx context.Context, session authoringcore.Session) (AuthoringSessionPaths, error)
	GetSession(ctx context.Context, id string) (authoringcore.Session, error)
	ListSessions(ctx context.Context) ([]authoringcore.Session, error)
	AppendTranscriptEvents(ctx context.Context, sessionID string, events []AssistedTranscriptEvent) error
	ListTranscriptEvents(ctx context.Context, sessionID string) ([]AssistedTranscriptEvent, error)
	ObserveContinuation(context.Context, string) (AuthoringSnapshot, error)
	ContinueAuthoring(context.Context, AuthoringContinuationRequest, AuthoringGenerator, ...AuthoringCommitGuard) (AuthoringContinuationResult, error)
}

// AuthoringSessionPaths describes a package boundary value.
type AuthoringSessionPaths struct {
	SessionDir     string
	SessionPath    string
	TranscriptPath string
	ArtifactsPath  string
}

type fileAuthoringSessionRepository struct {
	projectDir            string
	afterMutationReplace  func(int) error
	afterGuidedCheckpoint func() error
}

// NewAuthoringSessionRepository constructs the corresponding package service or value.
func NewAuthoringSessionRepository(projectDir string) AuthoringSessionRepository {
	return &fileAuthoringSessionRepository{projectDir: projectDir}
}

// AuthoringSessionLayout executes the corresponding package operation.
func AuthoringSessionLayout(projectDir, id string) AuthoringSessionPaths {
	sessionDir := filepath.Join(projectDir, workspacecore.DirectoryName, filepath.FromSlash(AuthoringSessionsDirectory), id)

	return AuthoringSessionPaths{
		SessionDir:     sessionDir,
		SessionPath:    filepath.Join(sessionDir, "session.md"),
		TranscriptPath: filepath.Join(sessionDir, "transcript.md"),
		ArtifactsPath:  filepath.Join(sessionDir, "artifacts.md"),
	}
}

// CreateSession executes the corresponding package operation.
func (r *fileAuthoringSessionRepository) createSession(ctx context.Context, session authoringcore.Session) (AuthoringSessionPaths, error) {
	err := ctx.Err()
	if err != nil {
		return AuthoringSessionPaths{}, err
	}

	err = ValidateAuthoringSession(session)
	if err != nil {
		return AuthoringSessionPaths{}, err
	}

	paths := AuthoringSessionLayout(r.projectDir, session.ID)

	err = os.MkdirAll(paths.SessionDir, 0o755)
	if err != nil {
		return AuthoringSessionPaths{}, fmt.Errorf("create authoring session directory: %w", err)
	}

	content, err := RenderAuthoringSession(session, paths)
	if err != nil {
		return AuthoringSessionPaths{}, err
	}

	err = writeNewAuthoringFile(paths.SessionPath, content)
	if err != nil {
		return AuthoringSessionPaths{}, err
	}

	err = writeNewAuthoringFile(paths.TranscriptPath, "# Transcript\n\n")
	if err != nil {
		return AuthoringSessionPaths{}, err
	}

	err = writeNewAuthoringFile(paths.ArtifactsPath, "# Artifacts\n\n")
	if err != nil {
		return AuthoringSessionPaths{}, err
	}

	return paths, nil
}

// UpdateSession executes the corresponding package operation.
func (r *fileAuthoringSessionRepository) updateSession(ctx context.Context, session authoringcore.Session) (AuthoringSessionPaths, error) {
	err := ctx.Err()
	if err != nil {
		return AuthoringSessionPaths{}, err
	}

	err = ValidateAuthoringSession(session)
	if err != nil {
		return AuthoringSessionPaths{}, err
	}

	paths := AuthoringSessionLayout(r.projectDir, session.ID)

	_, err = os.Stat(paths.SessionPath)
	if errors.Is(err, os.ErrNotExist) {
		return AuthoringSessionPaths{}, fmt.Errorf("%w: %s", ErrAuthoringSessionNotFound, session.ID)
	}

	if err != nil {
		return AuthoringSessionPaths{}, fmt.Errorf("stat authoring session: %w", err)
	}

	content, err := RenderAuthoringSession(session, paths)
	if err != nil {
		return AuthoringSessionPaths{}, err
	}

	err = os.WriteFile(paths.SessionPath, []byte(content), 0o644)
	if err != nil {
		return AuthoringSessionPaths{}, fmt.Errorf("update authoring session: %w", err)
	}

	return paths, nil
}

// GetSession executes the corresponding package operation.
func (r *fileAuthoringSessionRepository) getSession(ctx context.Context, id string) (authoringcore.Session, error) {
	err := ctx.Err()
	if err != nil {
		return authoringcore.Session{}, err
	}

	paths := AuthoringSessionLayout(r.projectDir, id)

	content, err := opsfile.Read(paths.SessionPath)
	if errors.Is(err, os.ErrNotExist) {
		return authoringcore.Session{}, fmt.Errorf("%w: %s", ErrAuthoringSessionNotFound, id)
	}

	if err != nil {
		return authoringcore.Session{}, fmt.Errorf("read authoring session: %w", err)
	}

	session, err := ParseAuthoringSession(string(content))
	if err != nil {
		return authoringcore.Session{}, err
	}

	return session, nil
}

// ListSessions lists the requested data.
func (r *fileAuthoringSessionRepository) listSessions(ctx context.Context) ([]authoringcore.Session, error) {
	sessionsDir := filepath.Join(r.projectDir, workspacecore.DirectoryName, filepath.FromSlash(AuthoringSessionsDirectory))
	sessions := []authoringcore.Session{}

	visited := 0

	err := filepath.WalkDir(sessionsDir, func(path string, entry os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}

		visited++
		if visited > 16384 {
			return fmt.Errorf("authoring session inventory exceeds its bound")
		}

		return r.appendListedAuthoringSession(ctx, &sessions, path, entry)
	})
	if errors.Is(err, os.ErrNotExist) {
		return sessions, nil
	}

	if err != nil {
		return nil, fmt.Errorf("list authoring sessions: %w", err)
	}

	return sessions, nil
}

// AppendTranscriptEvents executes the corresponding package operation.
func (r *fileAuthoringSessionRepository) appendTranscriptEvents(ctx context.Context, sessionID string, events []AssistedTranscriptEvent) error {
	err := ctx.Err()
	if err != nil {
		return err
	}

	if sessionID == "" {
		return fmt.Errorf("authoring session id is required")
	}

	paths := AuthoringSessionLayout(r.projectDir, sessionID)

	err = requireAuthoringSessionFile(paths.SessionPath, sessionID)
	if err != nil {
		return err
	}

	var output bytes.Buffer

	for _, event := range events {
		if event.SessionID != sessionID {
			return fmt.Errorf("assisted transcript event session id does not match session")
		}

		rendered, err := RenderAssistedTranscriptEvent(event)
		if err != nil {
			return err
		}

		output.WriteString(rendered)
		output.WriteString("\n")
	}

	file, err := os.OpenFile(paths.TranscriptPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return fmt.Errorf("open authoring transcript: %w", err)
	}

	defer func() {
		//nolint:errcheck // Write errors are returned by the main write call.
		_ = file.Close()
	}()

	_, err = file.WriteString(output.String())
	if err != nil {
		return fmt.Errorf("append authoring transcript: %w", err)
	}

	return nil
}

// ListTranscriptEvents lists the requested data.
func (r *fileAuthoringSessionRepository) listTranscriptEvents(ctx context.Context, sessionID string) ([]AssistedTranscriptEvent, error) {
	err := ctx.Err()
	if err != nil {
		return nil, err
	}

	if sessionID == "" {
		return nil, fmt.Errorf("authoring session id is required")
	}

	paths := AuthoringSessionLayout(r.projectDir, sessionID)

	err = requireAuthoringSessionFile(paths.SessionPath, sessionID)
	if err != nil {
		return nil, err
	}

	content, err := opsfile.Read(paths.TranscriptPath)
	if errors.Is(err, os.ErrNotExist) {
		return []AssistedTranscriptEvent{}, nil
	}

	if err != nil {
		return nil, fmt.Errorf("read authoring transcript: %w", err)
	}

	return ParseAssistedTranscriptEvents(string(content))
}

func requireAuthoringSessionFile(path, sessionID string) error {
	_, err := os.Stat(path)
	if errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("%w: %s", ErrAuthoringSessionNotFound, sessionID)
	}

	if err != nil {
		return fmt.Errorf("stat authoring session: %w", err)
	}

	return nil
}

// RenderAuthoringSession renders the corresponding package data.
func RenderAuthoringSession(session authoringcore.Session, paths AuthoringSessionPaths) (string, error) {
	err := ValidateAuthoringSession(session)
	if err != nil {
		return "", err
	}

	var output bytes.Buffer

	err = renderAuthoringSessionFrontmatter(&output, session)
	if err != nil {
		return "", err
	}

	renderAuthoringSessionBody(&output, session, paths)

	return output.String(), nil
}

// ParseAuthoringSession parses the corresponding package data.
func ParseAuthoringSession(content string) (authoringcore.Session, error) {
	metadata, err := parseTicketFrontmatter(content)
	if err != nil {
		return authoringcore.Session{}, err
	}

	session, err := authoringSessionFromMetadata(metadata.Metadata)
	if err != nil {
		return authoringcore.Session{}, err
	}

	session.Guided, err = parseGuidedUseCase(metadata.Metadata["guided_state"])
	if err != nil {
		return authoringcore.Session{}, err
	}

	session.Inputs = parseAuthoringArtifactSection(content, "Inputs")
	session.Outputs = parseAuthoringArtifactSection(content, "Outputs")
	session.Provenance = parseAuthoringProvenanceSection(content)

	err = ValidateAuthoringSession(session)
	if err != nil {
		return authoringcore.Session{}, err
	}

	return session, nil
}

func renderAuthoringSessionFrontmatter(output *bytes.Buffer, session authoringcore.Session) error {
	_, err := output.WriteString("---\n")
	if err != nil {
		return fmt.Errorf("write authoring session frontmatter: %w", err)
	}

	fields := []struct {
		key    string
		value  string
		quoted bool
	}{
		{key: "id", value: session.ID, quoted: true},
		{key: "project_id", value: session.ProjectID, quoted: true},
		{key: "status", value: string(session.Status), quoted: true},
		{key: "mode", value: string(session.Mode), quoted: true},
		{key: "target", value: string(session.Target), quoted: true},
		{key: "topic", value: session.Topic, quoted: true},
		{key: "created_at", value: opsmetadata.Timestamp(session.CreatedAt)},
		{key: "started_at", value: yamlOptionalTimestamp(session.StartedAt)},
		{key: "finished_at", value: yamlOptionalTimestamp(session.FinishedAt)},
		{key: "updated_at", value: opsmetadata.Timestamp(session.UpdatedAt)},
		{key: "description", value: session.Description, quoted: true},
	}

	if session.Revision > 0 {
		fields = append(fields, struct {
			key, value string
			quoted     bool
		}{key: "namespace", value: session.Namespace, quoted: true}, struct {
			key, value string
			quoted     bool
		}{key: "revision", value: strconv.FormatInt(session.Revision, 10)})
	}

	if err := renderGuidedState(output, session.Guided); err != nil {
		return err
	}

	for _, field := range fields {
		value := field.value
		if field.quoted {
			value = opsmetadata.Quote(value)
		}

		_, err = fmt.Fprintf(output, "%s: %s\n", field.key, value)
		if err != nil {
			return fmt.Errorf("write authoring session frontmatter: %w", err)
		}
	}

	_, err = output.WriteString("---\n\n")
	if err != nil {
		return fmt.Errorf("write authoring session frontmatter: %w", err)
	}

	return nil
}

func renderAuthoringSessionBody(output *bytes.Buffer, session authoringcore.Session, paths AuthoringSessionPaths) {
	output.WriteString("# Authoring Session\n\n")
	output.WriteString("## Files\n\n")
	output.WriteString("| Role | Path |\n")
	output.WriteString("| --- | --- |\n")
	output.WriteString("| session | " + filepath.ToSlash(paths.SessionPath) + " |\n")
	output.WriteString("| transcript | " + filepath.ToSlash(paths.TranscriptPath) + " |\n")
	output.WriteString("| artifacts | " + filepath.ToSlash(paths.ArtifactsPath) + " |\n\n")
	renderAuthoringArtifactTable(output, "Inputs", session.Inputs)
	renderAuthoringArtifactTable(output, "Outputs", session.Outputs)
	renderAuthoringProvenanceTable(output, session.Provenance)
}

func renderAuthoringArtifactTable(output *bytes.Buffer, title string, refs []authoringcore.ArtifactRef) {
	output.WriteString("## " + title + "\n\n")
	output.WriteString("| ID | Kind | Role | Path | Description |\n")
	output.WriteString("| --- | --- | --- | --- | --- |\n")

	for _, ref := range refs {
		output.WriteString("| " + markdownCell(ref.ID) + " | " + markdownCell(string(ref.Kind)) + " | ")
		output.WriteString(markdownCell(string(ref.Role)) + " | " + markdownCell(ref.Path) + " | ")
		output.WriteString(markdownCell(ref.Description) + " |\n")
	}

	output.WriteString("\n")
}

func renderAuthoringProvenanceTable(output *bytes.Buffer, links []authoringcore.ProvenanceLink) {
	output.WriteString("## Provenance\n\n")
	output.WriteString("| Relation | Source ID | Source Kind | Source Role | Source Path | Target ID | Target Kind | Target Role | Target Path | Decision | Recorded At |\n")
	output.WriteString("| --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- |\n")

	for _, link := range links {
		values := []string{
			string(link.Relation),
			link.Source.ID,
			string(link.Source.Kind),
			string(link.Source.Role),
			link.Source.Path,
			link.Target.ID,
			string(link.Target.Kind),
			string(link.Target.Role),
			link.Target.Path,
			link.Decision,
			link.RecordedAt.UTC().Format(time.RFC3339),
		}
		output.WriteString("| " + strings.Join(markdownCells(values), " | ") + " |\n")
	}

	output.WriteString("\n")
}

func authoringSessionFromMetadata(metadata map[string]string) (authoringcore.Session, error) {
	revision, err := optionalAuthoringRevision(metadata)
	if err != nil {
		return authoringcore.Session{}, err
	}

	createdAt, err := requiredAuthoringTime(metadata["created_at"], "created_at")
	if err != nil {
		return authoringcore.Session{}, err
	}

	startedAt, err := opsmetadata.OptionalTimestamp(metadata["started_at"])
	if err != nil {
		return authoringcore.Session{}, fmt.Errorf("parse authoring session started_at: %w", err)
	}

	finishedAt, err := opsmetadata.OptionalTimestamp(metadata["finished_at"])
	if err != nil {
		return authoringcore.Session{}, fmt.Errorf("parse authoring session finished_at: %w", err)
	}

	updatedAt, err := requiredAuthoringTime(metadata["updated_at"], "updated_at")
	if err != nil {
		return authoringcore.Session{}, err
	}

	return authoringcore.Session{
		ID:          opsmetadata.Unquote(metadata["id"]),
		Namespace:   opsmetadata.Unquote(metadata["namespace"]),
		Revision:    revision,
		ProjectID:   opsmetadata.Unquote(metadata["project_id"]),
		Status:      authoringcore.SessionStatus(opsmetadata.Unquote(metadata["status"])),
		Mode:        authoringcore.SessionMode(opsmetadata.Unquote(metadata["mode"])),
		Target:      authoringcore.Target(opsmetadata.Unquote(metadata["target"])),
		Topic:       opsmetadata.Unquote(metadata["topic"]),
		CreatedAt:   createdAt,
		StartedAt:   startedAt,
		FinishedAt:  finishedAt,
		UpdatedAt:   updatedAt,
		Description: opsmetadata.Unquote(metadata["description"]),
	}, nil
}

func parseAuthoringArtifactSection(content, title string) []authoringcore.ArtifactRef {
	rows := parseMarkdownTableRows(sectionContent(content, "## "+title))
	refs := make([]authoringcore.ArtifactRef, 0, len(rows))

	for _, row := range rows {
		if len(row) != 5 {
			continue
		}

		refs = append(refs, authoringcore.ArtifactRef{
			ID:          row[0],
			Kind:        authoringcore.ArtifactKind(row[1]),
			Role:        authoringcore.ArtifactRole(row[2]),
			Path:        row[3],
			Description: row[4],
		})
	}

	return refs
}

func parseAuthoringProvenanceSection(content string) []authoringcore.ProvenanceLink {
	rows := parseMarkdownTableRows(sectionContent(content, "## Provenance"))
	links := make([]authoringcore.ProvenanceLink, 0, len(rows))

	for _, row := range rows {
		if len(row) != 11 {
			continue
		}

		recordedAt, err := parseAuthoringTableTime(row[10])
		if err != nil {
			continue
		}

		links = append(links, authoringcore.ProvenanceLink{
			Relation: authoringcore.ProvenanceRelation(row[0]),
			Source: authoringcore.ArtifactRef{
				ID:   row[1],
				Kind: authoringcore.ArtifactKind(row[2]),
				Role: authoringcore.ArtifactRole(row[3]),
				Path: row[4],
			},
			Target: authoringcore.ArtifactRef{
				ID:   row[5],
				Kind: authoringcore.ArtifactKind(row[6]),
				Role: authoringcore.ArtifactRole(row[7]),
				Path: row[8],
			},
			Decision:   row[9],
			RecordedAt: recordedAt,
		})
	}

	return links
}

func sectionContent(content, heading string) string {
	start := strings.Index(content, heading)
	if start < 0 {
		return ""
	}

	section := content[start+len(heading):]

	next := strings.Index(section, "\n## ")
	if next >= 0 {
		section = section[:next]
	}

	return section
}

func parseMarkdownTableRows(section string) [][]string {
	lines := strings.Split(section, "\n")
	rows := [][]string{}

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if !strings.HasPrefix(line, "|") || strings.Contains(line, "---") {
			continue
		}

		cells := strings.Split(strings.Trim(line, "|"), "|")
		parsed := make([]string, 0, len(cells))

		for _, cell := range cells {
			parsed = append(parsed, unmarkdownCell(cell))
		}

		rows = append(rows, parsed)
	}

	if len(rows) <= 1 {
		return nil
	}

	return rows[1:]
}

func writeNewAuthoringFile(path, content string) error {
	_, err := os.Stat(path)
	if err == nil {
		return fmt.Errorf("authoring file already exists: %s", path)
	}

	if !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("stat authoring file: %w", err)
	}

	err = os.WriteFile(path, []byte(content), 0o644)
	if err != nil {
		return fmt.Errorf("write authoring file: %w", err)
	}

	return nil
}

func requiredAuthoringTime(value, name string) (time.Time, error) {
	parsed, err := opsmetadata.OptionalTimestamp(value)
	if err != nil {
		return time.Time{}, fmt.Errorf("parse authoring session %s: %w", name, err)
	}

	if parsed == nil {
		return time.Time{}, nil
	}

	return *parsed, nil
}

func yamlOptionalTimestamp(value *time.Time) string {
	if value == nil {
		return opsmetadata.Quote("")
	}

	return opsmetadata.Timestamp(*value)
}

func parseAuthoringTableTime(value string) (time.Time, error) {
	parsed, err := time.Parse(time.RFC3339, value)
	if err != nil {
		return time.Time{}, err
	}

	return parsed.UTC(), nil
}

func markdownCells(values []string) []string {
	cells := make([]string, 0, len(values))
	for _, value := range values {
		cells = append(cells, markdownCell(value))
	}

	return cells
}

func markdownCell(value string) string {
	value = strings.ReplaceAll(value, "\n", " ")
	value = strings.ReplaceAll(value, "|", "\\|")

	return value
}

func unmarkdownCell(value string) string {
	value = strings.TrimSpace(value)
	value = strings.ReplaceAll(value, "\\|", "|")

	return value
}

func optionalAuthoringRevision(metadata map[string]string) (int64, error) {
	value := metadata["revision"]
	if value == "" {
		return 0, nil
	}

	revision, err := strconv.ParseInt(value, 10, 64)
	if err != nil || revision < 1 {
		return 0, fmt.Errorf("invalid authoring revision")
	}

	return revision, nil
}

func (r *fileAuthoringSessionRepository) appendListedAuthoringSession(ctx context.Context, sessions *[]authoringcore.Session, path string, entry os.DirEntry) error {
	err := ctx.Err()
	if err != nil {
		return err
	}

	if entry.IsDir() || filepath.Base(path) != "session.md" {
		return nil
	}

	relative, err := filepath.Rel(r.projectDir, path)
	if err != nil {
		return err
	}

	if _, err := r.authoringPath(relative); err != nil {
		return err
	}

	content, err := opsfile.Read(path)
	if err != nil {
		return fmt.Errorf("read authoring session: %w", err)
	}

	session, err := ParseAuthoringSession(string(content))
	if err != nil {
		return err
	}

	*sessions = append(*sessions, session)

	return nil
}
