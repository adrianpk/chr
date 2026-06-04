package chc

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// AssistedAuthoringContextPackageOptions describes a package boundary value.
type AssistedAuthoringContextPackageOptions struct {
	ProjectDir       string
	TargetPath       string
	Session          AuthoringSession
	SessionPaths     AuthoringSessionPaths
	TranscriptEvents []AssistedTranscriptEvent
}

// AssistedAuthoringContextPackage describes a package boundary value.
type AssistedAuthoringContextPackage struct {
	ProjectDir        string
	WorkspaceDir      string
	Session           AuthoringSession
	SessionPaths      AuthoringSessionPaths
	AgentProtocol     AgentProtocolContext
	SelectedDocuments []ContextDocument
	TranscriptSummary AssistedTranscriptSummary
	Sources           []ContextSourceInventoryEntry
}

// AssistedTranscriptSummary describes a package boundary value.
type AssistedTranscriptSummary struct {
	QuestionCount       int
	AnswerCount         int
	ResultCount         int
	StateCount          int
	UnansweredQuestions []AssistedTranscriptQuestionGap
	LastResultSummary   string
}

// AssistedTranscriptQuestionGap describes a package boundary value.
type AssistedTranscriptQuestionGap struct {
	TurnID     string
	TurnIndex  int
	QuestionID string
	Prompt     string
}

// LoadAssistedAuthoringContextPackage loads the corresponding package data.
func LoadAssistedAuthoringContextPackage(
	opts AssistedAuthoringContextPackageOptions,
) (AssistedAuthoringContextPackage, error) {
	projectDir, err := cleanAbsProjectDir(opts.ProjectDir)
	if err != nil {
		return AssistedAuthoringContextPackage{}, err
	}

	err = ValidateAuthoringSession(opts.Session)
	if err != nil {
		return AssistedAuthoringContextPackage{}, err
	}

	sessionPaths := normalizeAuthoringContextSessionPaths(projectDir, opts.Session.ID, opts.SessionPaths)

	agentProtocol, err := LoadAgentProtocolContext(AgentProtocolContextOptions{
		ProjectDir: projectDir,
		TargetPath: opts.TargetPath,
	})
	if err != nil {
		return AssistedAuthoringContextPackage{}, err
	}

	documents, documentSources, err := loadAuthoringSelectedDocuments(projectDir, opts.Session)
	if err != nil {
		return AssistedAuthoringContextPackage{}, err
	}

	summary, transcriptSources, err := summarizeAssistedTranscript(projectDir, sessionPaths, opts.Session.ID, opts.TranscriptEvents)
	if err != nil {
		return AssistedAuthoringContextPackage{}, err
	}

	sources := append([]ContextSourceInventoryEntry{}, agentProtocol.Sources...)
	sources = append(sources, authoringSessionContextSources(projectDir, sessionPaths)...)
	sources = append(sources, transcriptSources...)
	sources = append(sources, documentSources...)

	return AssistedAuthoringContextPackage{
		ProjectDir:        projectDir,
		WorkspaceDir:      filepath.Join(projectDir, DirectoryName),
		Session:           opts.Session,
		SessionPaths:      sessionPaths,
		AgentProtocol:     agentProtocol,
		SelectedDocuments: documents,
		TranscriptSummary: summary,
		Sources:           sources,
	}, nil
}

// RenderAssistedAuthoringContextPackage renders the corresponding package data.
func RenderAssistedAuthoringContextPackage(pkg AssistedAuthoringContextPackage) string {
	var output strings.Builder

	output.WriteString("# Assisted Authoring Context\n\n")
	writeAssistedContextAgent(&output, pkg.AgentProtocol.AgentIdentity)
	writeAssistedContextProject(&output, pkg)
	writeAssistedContextSession(&output, pkg)
	writeAssistedContextArtifacts(&output, pkg.SessionPaths)
	writeAssistedContextTranscript(&output, pkg.TranscriptSummary)
	writeAssistedContextProtocols(&output, pkg.AgentProtocol)
	writeAssistedContextDocuments(&output, pkg.SelectedDocuments)
	writeAssistedContextConversationBoundary(&output)
	writeAssistedContextSourceInventory(&output, pkg.Sources)

	return output.String()
}

func normalizeAuthoringContextSessionPaths(
	projectDir string,
	sessionID string,
	paths AuthoringSessionPaths,
) AuthoringSessionPaths {
	if paths.SessionDir == "" {
		return AuthoringSessionLayout(projectDir, sessionID)
	}

	return paths
}

func loadAuthoringSelectedDocuments(
	projectDir string,
	session AuthoringSession,
) ([]ContextDocument, []ContextSourceInventoryEntry, error) {
	refs := append([]AuthoringArtifactRef{}, session.Inputs...)
	refs = append(refs, session.Outputs...)
	sort.SliceStable(refs, func(left, right int) bool {
		return refs[left].Path < refs[right].Path
	})

	documents := make([]ContextDocument, 0, len(refs))
	sources := make([]ContextSourceInventoryEntry, 0, len(refs))

	for _, ref := range refs {
		if ref.Path == "" {
			continue
		}

		document, source, err := loadAuthoringSelectedDocument(projectDir, ref)
		if err != nil {
			return nil, nil, err
		}

		if document.ID != "" {
			documents = append(documents, document)
		}

		sources = append(sources, source)
	}

	return documents, sources, nil
}

func loadAuthoringSelectedDocument(
	projectDir string,
	ref AuthoringArtifactRef,
) (ContextDocument, ContextSourceInventoryEntry, error) {
	path := ref.Path
	if !filepath.IsAbs(path) {
		path = filepath.Join(projectDir, filepath.FromSlash(path))
	}

	path = filepath.Clean(path)
	if !pathIsInsideProject(projectDir, path) {
		return ContextDocument{}, ContextSourceInventoryEntry{}, fmt.Errorf("authoring context document is outside project: %s", path)
	}

	relPath, err := filepath.Rel(projectDir, path)
	if err != nil {
		return ContextDocument{}, ContextSourceInventoryEntry{}, fmt.Errorf("resolve authoring context document relative path: %w", err)
	}

	relPath = filepath.ToSlash(relPath)
	id := "authoring-document:" + ref.ID
	source := ContextSourceInventoryEntry{
		ID:          id,
		Kind:        ContextSourceKindDocument,
		Class:       ContextSourceClassFresh,
		Path:        relPath,
		Description: ref.Description,
		Required:    false,
	}

	content, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			source.Included = false
			source.Reason = "file was not found"

			return ContextDocument{}, source, nil
		}

		return ContextDocument{}, ContextSourceInventoryEntry{}, fmt.Errorf("read authoring context document: %w", err)
	}

	source.Included = true

	return ContextDocument{
		ID:           id,
		Path:         path,
		RelativePath: relPath,
		Content:      string(content),
	}, source, nil
}

func summarizeAssistedTranscript(
	projectDir string,
	paths AuthoringSessionPaths,
	sessionID string,
	events []AssistedTranscriptEvent,
) (AssistedTranscriptSummary, []ContextSourceInventoryEntry, error) {
	summary := AssistedTranscriptSummary{}
	questions := map[string]AssistedTranscriptQuestionGap{}
	answered := map[string]struct{}{}

	for _, event := range events {
		if event.SessionID != sessionID {
			return AssistedTranscriptSummary{}, nil, fmt.Errorf("assisted transcript event session id does not match session")
		}

		switch event.Type {
		case AssistedTranscriptEventQuestion:
			summary.QuestionCount++
			questions[event.Question.ID] = AssistedTranscriptQuestionGap{
				TurnID:     event.TurnID,
				TurnIndex:  event.TurnIndex,
				QuestionID: event.Question.ID,
				Prompt:     event.Question.Prompt,
			}
		case AssistedTranscriptEventAnswer:
			summary.AnswerCount++
			answered[event.Answer.QuestionID] = struct{}{}
		case AssistedTranscriptEventResult:
			summary.ResultCount++
			if event.Result.Summary != "" {
				summary.LastResultSummary = event.Result.Summary
			}
		case AssistedTranscriptEventState:
			summary.StateCount++
		}
	}

	for questionID, question := range questions {
		_, ok := answered[questionID]
		if !ok {
			summary.UnansweredQuestions = append(summary.UnansweredQuestions, question)
		}
	}

	sort.Slice(summary.UnansweredQuestions, func(left, right int) bool {
		return summary.UnansweredQuestions[left].TurnIndex < summary.UnansweredQuestions[right].TurnIndex
	})

	return summary, []ContextSourceInventoryEntry{assistedTranscriptSource(projectDir, paths)}, nil
}

func authoringSessionContextSources(projectDir string, paths AuthoringSessionPaths) []ContextSourceInventoryEntry {
	sessionPath := relativeContextPath(projectDir, paths.SessionPath)
	artifactsPath := relativeContextPath(projectDir, paths.ArtifactsPath)

	return []ContextSourceInventoryEntry{
		{
			ID:          "authoring-session:session",
			Kind:        ContextSourceKindDocument,
			Class:       ContextSourceClassFresh,
			Path:        sessionPath,
			Description: "Authoring session metadata.",
			Required:    true,
			Included:    true,
		},
		{
			ID:          "authoring-session:artifacts",
			Kind:        ContextSourceKindDocument,
			Class:       ContextSourceClassFresh,
			Path:        artifactsPath,
			Description: "Authoring session artifact registry.",
			Required:    false,
			Included:    true,
		},
	}
}

func assistedTranscriptSource(projectDir string, paths AuthoringSessionPaths) ContextSourceInventoryEntry {
	return ContextSourceInventoryEntry{
		ID:          "authoring-session:transcript",
		Kind:        ContextSourceKindDocument,
		Class:       ContextSourceClassFresh,
		Path:        relativeContextPath(projectDir, paths.TranscriptPath),
		Description: "Assisted authoring transcript events.",
		Required:    true,
		Included:    true,
	}
}

func relativeContextPath(projectDir, path string) string {
	relPath, err := filepath.Rel(projectDir, path)
	if err != nil {
		return filepath.ToSlash(path)
	}

	return filepath.ToSlash(relPath)
}

func writeAssistedContextAgent(output *strings.Builder, identity AgentIdentity) {
	output.WriteString("## Agent Identity\n\n")
	output.WriteString("- id: ")
	output.WriteString(identity.ID)
	output.WriteString("\n")
	output.WriteString("- title: ")
	output.WriteString(identity.Title)
	output.WriteString("\n")
	output.WriteString("- role: ")
	output.WriteString(identity.Role)
	output.WriteString("\n\n")
	writeMarkdownList(output, "Authoring Stance", []string{
		"Act as a technical analyst for project authoring.",
		"Ask focused follow-up questions when information is missing.",
		"Convert captured answers into durable Markdown artifacts.",
		"Do not rely on implicit conversation history.",
	})
}

func writeAssistedContextProject(output *strings.Builder, pkg AssistedAuthoringContextPackage) {
	output.WriteString("## Project\n\n")
	output.WriteString("- project_id: ")
	output.WriteString(pkg.Session.ProjectID)
	output.WriteString("\n")
	output.WriteString("- project_dir: ")
	output.WriteString(pkg.ProjectDir)
	output.WriteString("\n")
	output.WriteString("- workspace_dir: ")
	output.WriteString(pkg.WorkspaceDir)
	output.WriteString("\n\n")
}

func writeAssistedContextSession(output *strings.Builder, pkg AssistedAuthoringContextPackage) {
	session := pkg.Session

	output.WriteString("## Session\n\n")
	output.WriteString("- id: ")
	output.WriteString(session.ID)
	output.WriteString("\n")
	output.WriteString("- status: ")
	output.WriteString(string(session.Status))
	output.WriteString("\n")
	output.WriteString("- mode: ")
	output.WriteString(string(session.Mode))
	output.WriteString("\n")
	output.WriteString("- target: ")
	output.WriteString(string(session.Target))
	output.WriteString("\n")
	output.WriteString("- topic: ")
	output.WriteString(session.Topic)
	output.WriteString("\n")
	output.WriteString("- description: ")
	output.WriteString(session.Description)
	output.WriteString("\n\n")
}

func writeAssistedContextArtifacts(output *strings.Builder, paths AuthoringSessionPaths) {
	output.WriteString("## Session Artifacts\n\n")
	output.WriteString("- session: ")
	output.WriteString(filepath.ToSlash(paths.SessionPath))
	output.WriteString("\n")
	output.WriteString("- transcript: ")
	output.WriteString(filepath.ToSlash(paths.TranscriptPath))
	output.WriteString("\n")
	output.WriteString("- artifacts: ")
	output.WriteString(filepath.ToSlash(paths.ArtifactsPath))
	output.WriteString("\n\n")
}

func writeAssistedContextTranscript(output *strings.Builder, summary AssistedTranscriptSummary) {
	output.WriteString("## Transcript Summary\n\n")
	output.WriteString("- questions: ")
	fmt.Fprintf(output, "%d", summary.QuestionCount)
	output.WriteString("\n")
	output.WriteString("- answers: ")
	fmt.Fprintf(output, "%d", summary.AnswerCount)
	output.WriteString("\n")
	output.WriteString("- results: ")
	fmt.Fprintf(output, "%d", summary.ResultCount)
	output.WriteString("\n")
	output.WriteString("- state_events: ")
	fmt.Fprintf(output, "%d", summary.StateCount)
	output.WriteString("\n")
	output.WriteString("- last_result_summary: ")
	output.WriteString(summary.LastResultSummary)
	output.WriteString("\n\n")

	output.WriteString("### Unanswered Gaps\n\n")

	if len(summary.UnansweredQuestions) == 0 {
		output.WriteString("No unanswered questions were detected.\n\n")

		return
	}

	for _, question := range summary.UnansweredQuestions {
		output.WriteString("- turn ")
		fmt.Fprintf(output, "%d", question.TurnIndex)
		output.WriteString(" `")
		output.WriteString(question.QuestionID)
		output.WriteString("`: ")
		output.WriteString(question.Prompt)
		output.WriteString("\n")
	}

	output.WriteString("\n")
}

func writeAssistedContextProtocols(output *strings.Builder, ctx AgentProtocolContext) {
	output.WriteString("## Protocols\n\n")
	output.WriteString("### AGENTS.md\n\n")
	writeContextDocuments(output, ctx.AgentFiles, "No applicable AGENTS.md files were found.")
	output.WriteString("### .chc Protocol Files\n\n")
	writeContextDocuments(output, ctx.ChcFiles, "No selected .chc protocol files were found.")
}

func writeAssistedContextDocuments(output *strings.Builder, documents []ContextDocument) {
	output.WriteString("## Selected Documents\n\n")
	writeContextDocuments(output, documents, "No selected authoring documents were found.")
}

func writeAssistedContextConversationBoundary(output *strings.Builder) {
	output.WriteString("## Conversation History\n\n")
	output.WriteString("Implicit conversation history is not included. Relevant prior discussion must be captured in session transcript events or selected project documents.\n\n")
}

func writeAssistedContextSourceInventory(output *strings.Builder, sources []ContextSourceInventoryEntry) {
	output.WriteString("## Source Inventory\n\n")
	output.WriteString(RenderSourceInventory(sources))
}
