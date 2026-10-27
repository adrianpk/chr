package authoring

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

	opsfile "forge.adrianpk.com/chr/internal/core/ops/file"
	opsmetadata "forge.adrianpk.com/chr/internal/core/ops/metadata"
	ticketmetadata "forge.adrianpk.com/chr/internal/core/ticket/metadata"
	workspacecore "forge.adrianpk.com/chr/internal/core/workspace"
)

// SessionsDirectory defines a package boundary constant.
const SessionsDirectory = "authoring/sessions"

// ErrSessionNotFound identifies a package boundary error.
var ErrSessionNotFound = errors.New("authoring session not found")

// SessionRepository defines a package boundary interface.
type SessionRepository interface {
	CreateSession(ctx context.Context, session Session) (SessionPaths, error)
	UpdateSession(ctx context.Context, session Session) (SessionPaths, error)
	GetSession(ctx context.Context, id string) (Session, error)
	ListSessions(ctx context.Context) ([]Session, error)
	AppendTranscriptEvents(ctx context.Context, sessionID string, events []AssistedTranscriptEvent) error
	ListTranscriptEvents(ctx context.Context, sessionID string) ([]AssistedTranscriptEvent, error)
	ObserveContinuation(context.Context, string) (Snapshot, error)
	ContinueAuthoring(context.Context, ContinuationRequest, Generator, ...CommitGuard) (ContinuationResult, error)
}

// SessionPaths describes a package boundary value.
type SessionPaths struct {
	SessionDir     string
	SessionPath    string
	TranscriptPath string
	ArtifactsPath  string
}

type fileSessionRepository struct {
	projectDir            string
	afterMutationReplace  func(int) error
	afterGuidedCheckpoint func() error
}

// NewSessionRepository constructs the corresponding package service or value.
func NewSessionRepository(projectDir string) SessionRepository {
	return &fileSessionRepository{projectDir: projectDir}
}

// SessionLayout executes the corresponding package operation.
func SessionLayout(projectDir, id string) SessionPaths {
	sessionDir := filepath.Join(projectDir, workspacecore.DirectoryName, filepath.FromSlash(SessionsDirectory), id)

	return SessionPaths{
		SessionDir:     sessionDir,
		SessionPath:    filepath.Join(sessionDir, "session.md"),
		TranscriptPath: filepath.Join(sessionDir, "transcript.md"),
		ArtifactsPath:  filepath.Join(sessionDir, "artifacts.md"),
	}
}

// CreateSession executes the corresponding package operation.
func (r *fileSessionRepository) createSession(ctx context.Context, session Session) (SessionPaths, error) {
	err := ctx.Err()
	if err != nil {
		return SessionPaths{}, err
	}

	err = ValidateSession(session)
	if err != nil {
		return SessionPaths{}, err
	}

	paths := SessionLayout(r.projectDir, session.ID)

	err = os.MkdirAll(paths.SessionDir, 0o755)
	if err != nil {
		return SessionPaths{}, fmt.Errorf("create authoring session directory: %w", err)
	}

	content, err := RenderSession(session, paths)
	if err != nil {
		return SessionPaths{}, err
	}

	err = writeNewAuthoringFile(paths.SessionPath, content)
	if err != nil {
		return SessionPaths{}, err
	}

	err = writeNewAuthoringFile(paths.TranscriptPath, "# Transcript\n\n")
	if err != nil {
		return SessionPaths{}, err
	}

	err = writeNewAuthoringFile(paths.ArtifactsPath, "# Artifacts\n\n")
	if err != nil {
		return SessionPaths{}, err
	}

	return paths, nil
}

// UpdateSession executes the corresponding package operation.
func (r *fileSessionRepository) updateSession(ctx context.Context, session Session) (SessionPaths, error) {
	err := ctx.Err()
	if err != nil {
		return SessionPaths{}, err
	}

	err = ValidateSession(session)
	if err != nil {
		return SessionPaths{}, err
	}

	paths := SessionLayout(r.projectDir, session.ID)

	_, err = os.Stat(paths.SessionPath)
	if errors.Is(err, os.ErrNotExist) {
		return SessionPaths{}, fmt.Errorf("%w: %s", ErrSessionNotFound, session.ID)
	}

	if err != nil {
		return SessionPaths{}, fmt.Errorf("stat authoring session: %w", err)
	}

	content, err := RenderSession(session, paths)
	if err != nil {
		return SessionPaths{}, err
	}

	err = os.WriteFile(paths.SessionPath, []byte(content), 0o644)
	if err != nil {
		return SessionPaths{}, fmt.Errorf("update authoring session: %w", err)
	}

	return paths, nil
}

// GetSession executes the corresponding package operation.
func (r *fileSessionRepository) getSession(ctx context.Context, id string) (Session, error) {
	err := ctx.Err()
	if err != nil {
		return Session{}, err
	}

	paths := SessionLayout(r.projectDir, id)

	content, err := opsfile.Read(paths.SessionPath)
	if errors.Is(err, os.ErrNotExist) {
		return Session{}, fmt.Errorf("%w: %s", ErrSessionNotFound, id)
	}

	if err != nil {
		return Session{}, fmt.Errorf("read authoring session: %w", err)
	}

	session, err := ParseSession(string(content))
	if err != nil {
		return Session{}, err
	}

	return session, nil
}

// ListSessions lists the requested data.
func (r *fileSessionRepository) listSessions(ctx context.Context) ([]Session, error) {
	sessionsDir := filepath.Join(r.projectDir, workspacecore.DirectoryName, filepath.FromSlash(SessionsDirectory))
	sessions := []Session{}

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
func (r *fileSessionRepository) appendTranscriptEvents(ctx context.Context, sessionID string, events []AssistedTranscriptEvent) error {
	err := ctx.Err()
	if err != nil {
		return err
	}

	if sessionID == "" {
		return fmt.Errorf("authoring session id is required")
	}

	paths := SessionLayout(r.projectDir, sessionID)

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
func (r *fileSessionRepository) listTranscriptEvents(ctx context.Context, sessionID string) ([]AssistedTranscriptEvent, error) {
	err := ctx.Err()
	if err != nil {
		return nil, err
	}

	if sessionID == "" {
		return nil, fmt.Errorf("authoring session id is required")
	}

	paths := SessionLayout(r.projectDir, sessionID)

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
		return fmt.Errorf("%w: %s", ErrSessionNotFound, sessionID)
	}

	if err != nil {
		return fmt.Errorf("stat authoring session: %w", err)
	}

	return nil
}

// RenderSession renders the corresponding package data.
func RenderSession(session Session, paths SessionPaths) (string, error) {
	err := ValidateSession(session)
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

// ParseSession parses the corresponding package data.
func ParseSession(content string) (Session, error) {
	metadata, err := ticketmetadata.Parse(content)
	if err != nil {
		return Session{}, err
	}

	session, err := authoringSessionFromMetadata(metadata.Metadata)
	if err != nil {
		return Session{}, err
	}

	session.Guided, err = parseGuidedUseCase(metadata.Metadata["guided_state"])
	if err != nil {
		return Session{}, err
	}

	session.Inputs = parseAuthoringArtifactSection(content, "Inputs")
	session.Outputs = parseAuthoringArtifactSection(content, "Outputs")
	session.Provenance = parseAuthoringProvenanceSection(content)

	err = ValidateSession(session)
	if err != nil {
		return Session{}, err
	}

	return session, nil
}

func renderAuthoringSessionFrontmatter(output *bytes.Buffer, session Session) error {
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

func renderAuthoringSessionBody(output *bytes.Buffer, session Session, paths SessionPaths) {
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

func renderAuthoringArtifactTable(output *bytes.Buffer, title string, refs []ArtifactRef) {
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

func renderAuthoringProvenanceTable(output *bytes.Buffer, links []ProvenanceLink) {
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

func authoringSessionFromMetadata(metadata map[string]string) (Session, error) {
	revision, err := optionalAuthoringRevision(metadata)
	if err != nil {
		return Session{}, err
	}

	createdAt, err := requiredAuthoringTime(metadata["created_at"], "created_at")
	if err != nil {
		return Session{}, err
	}

	startedAt, err := opsmetadata.OptionalTimestamp(metadata["started_at"])
	if err != nil {
		return Session{}, fmt.Errorf("parse authoring session started_at: %w", err)
	}

	finishedAt, err := opsmetadata.OptionalTimestamp(metadata["finished_at"])
	if err != nil {
		return Session{}, fmt.Errorf("parse authoring session finished_at: %w", err)
	}

	updatedAt, err := requiredAuthoringTime(metadata["updated_at"], "updated_at")
	if err != nil {
		return Session{}, err
	}

	return Session{
		ID:          opsmetadata.Unquote(metadata["id"]),
		Namespace:   opsmetadata.Unquote(metadata["namespace"]),
		Revision:    revision,
		ProjectID:   opsmetadata.Unquote(metadata["project_id"]),
		Status:      SessionStatus(opsmetadata.Unquote(metadata["status"])),
		Mode:        SessionMode(opsmetadata.Unquote(metadata["mode"])),
		Target:      Target(opsmetadata.Unquote(metadata["target"])),
		Topic:       opsmetadata.Unquote(metadata["topic"]),
		CreatedAt:   createdAt,
		StartedAt:   startedAt,
		FinishedAt:  finishedAt,
		UpdatedAt:   updatedAt,
		Description: opsmetadata.Unquote(metadata["description"]),
	}, nil
}

func parseAuthoringArtifactSection(content, title string) []ArtifactRef {
	rows := parseMarkdownTableRows(sectionContent(content, "## "+title))
	refs := make([]ArtifactRef, 0, len(rows))

	for _, row := range rows {
		if len(row) != 5 {
			continue
		}

		refs = append(refs, ArtifactRef{
			ID:          row[0],
			Kind:        ArtifactKind(row[1]),
			Role:        ArtifactRole(row[2]),
			Path:        row[3],
			Description: row[4],
		})
	}

	return refs
}

func parseAuthoringProvenanceSection(content string) []ProvenanceLink {
	rows := parseMarkdownTableRows(sectionContent(content, "## Provenance"))
	links := make([]ProvenanceLink, 0, len(rows))

	for _, row := range rows {
		if len(row) != 11 {
			continue
		}

		recordedAt, err := parseAuthoringTableTime(row[10])
		if err != nil {
			continue
		}

		links = append(links, ProvenanceLink{
			Relation: ProvenanceRelation(row[0]),
			Source: ArtifactRef{
				ID:   row[1],
				Kind: ArtifactKind(row[2]),
				Role: ArtifactRole(row[3]),
				Path: row[4],
			},
			Target: ArtifactRef{
				ID:   row[5],
				Kind: ArtifactKind(row[6]),
				Role: ArtifactRole(row[7]),
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

func (r *fileSessionRepository) appendListedAuthoringSession(ctx context.Context, sessions *[]Session, path string, entry os.DirEntry) error {
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

	session, err := ParseSession(string(content))
	if err != nil {
		return err
	}

	*sessions = append(*sessions, session)

	return nil
}
