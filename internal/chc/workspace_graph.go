package chc

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

type WorkspaceNodeKind string

const (
	WorkspaceNodeProject WorkspaceNodeKind = "project"
	WorkspaceNodeCase    WorkspaceNodeKind = "case"
	WorkspaceNodeDump    WorkspaceNodeKind = "dump"
	WorkspaceNodeQuiz    WorkspaceNodeKind = "quiz"
	WorkspaceNodeSpec    WorkspaceNodeKind = "spec"
	WorkspaceNodePlan    WorkspaceNodeKind = "plan"
	WorkspaceNodeTicket  WorkspaceNodeKind = "ticket"
	WorkspaceNodeReport  WorkspaceNodeKind = "report"
	WorkspaceNodeUnknown WorkspaceNodeKind = "unknown"
)

type WorkspaceGraph struct {
	Project                  *WorkspaceNode
	Cases                    []WorkspaceNode
	ProjectWide              []WorkspaceNode
	CaseDocuments            map[string][]WorkspaceNode
	Diagnostics              []WorkspaceGraphDiagnostic
	UnresolvedCaseReferences []WorkspaceCaseReference
}

type WorkspaceNode struct {
	Kind         WorkspaceNodeKind
	Path         string
	RelativePath string
	Slug         string
	Title        string
	Name         string
	Description  string
	Status       string
	DocumentType string
	SourcePath   string
	CuratedPath  string
	CaseSlug     string
	SourceHash   string
	ModifiedAt   time.Time
	Size         int64
}

type WorkspaceGraphDiagnostic struct {
	Path    string
	Kind    WorkspaceNodeKind
	Message string
}

type WorkspaceCaseReference struct {
	CaseSlug   string
	References []WorkspaceNode
}

type WorkspaceGraphBuilder struct {
	projectDir string
}

func NewWorkspaceGraphBuilder(projectDir string) WorkspaceGraphBuilder {
	return WorkspaceGraphBuilder{projectDir: projectDir}
}

func (b WorkspaceGraphBuilder) Build(ctx context.Context) (WorkspaceGraph, error) {
	graph := WorkspaceGraph{
		CaseDocuments: map[string][]WorkspaceNode{},
	}

	err := b.loadProject(ctx, &graph)
	if err != nil {
		return WorkspaceGraph{}, err
	}

	err = b.loadCases(ctx, &graph)
	if err != nil {
		return WorkspaceGraph{}, err
	}

	err = b.loadDocuments(ctx, &graph)
	if err != nil {
		return WorkspaceGraph{}, err
	}

	graph.finalize()

	return graph, nil
}

func (b WorkspaceGraphBuilder) loadProject(ctx context.Context, graph *WorkspaceGraph) error {
	err := ctx.Err()
	if err != nil {
		return err
	}

	path := ProjectMetadataPath(b.projectDir)

	info, statErr := os.Stat(path)
	if errors.Is(statErr, os.ErrNotExist) {
		return nil
	}

	if statErr != nil {
		return fmt.Errorf("stat project metadata: %w", statErr)
	}

	content, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("read project metadata: %w", err)
	}

	metadata, err := ParseProjectMetadata(string(content))
	if err != nil {
		graph.addDiagnostic(path, WorkspaceNodeProject, err)

		return nil
	}

	graph.Project = &WorkspaceNode{
		Kind:         WorkspaceNodeProject,
		Path:         path,
		RelativePath: b.relativePath(path),
		Slug:         "project",
		Title:        metadata.Name,
		Name:         metadata.Name,
		Description:  metadata.Description,
		SourceHash:   contentHash(content),
		ModifiedAt:   info.ModTime().UTC(),
		Size:         info.Size(),
	}

	return nil
}

func (b WorkspaceGraphBuilder) loadCases(ctx context.Context, graph *WorkspaceGraph) error {
	root := CaseMetadataDirectory(b.projectDir)

	err := walkMarkdownFiles(ctx, root, func(path string) error {
		info, statErr := os.Stat(path)
		if statErr != nil {
			return fmt.Errorf("stat case metadata: %w", statErr)
		}

		content, readErr := os.ReadFile(path)
		if readErr != nil {
			return fmt.Errorf("read case metadata: %w", readErr)
		}

		metadata, parseErr := ParseCaseMetadata(string(content))
		if parseErr != nil {
			graph.addDiagnostic(path, WorkspaceNodeCase, parseErr)

			return nil
		}

		slug := caseSlugFromPath(path)
		graph.Cases = append(graph.Cases, WorkspaceNode{
			Kind:         WorkspaceNodeCase,
			Path:         path,
			RelativePath: b.relativePath(path),
			Slug:         slug,
			Title:        metadata.Name,
			Name:         metadata.Name,
			Description:  metadata.Description,
			SourceHash:   contentHash(content),
			ModifiedAt:   info.ModTime().UTC(),
			Size:         info.Size(),
		})

		return nil
	})
	if err != nil {
		return fmt.Errorf("load cases: %w", err)
	}

	return nil
}

func (b WorkspaceGraphBuilder) loadDocuments(ctx context.Context, graph *WorkspaceGraph) error {
	root := UserPath(b.projectDir)

	err := walkMarkdownFiles(ctx, root, func(path string) error {
		if path == ProjectMetadataPath(b.projectDir) || isCaseMetadataPath(b.projectDir, path) {
			return nil
		}

		if isArchivedGenerationPath(b.projectDir, path) {
			return nil
		}

		node, err := b.documentNode(path)
		if err != nil {
			graph.addDiagnostic(path, WorkspaceNodeUnknown, err)

			return nil
		}

		if node.Kind == WorkspaceNodeUnknown {
			return nil
		}

		if node.CaseSlug == "" {
			graph.ProjectWide = append(graph.ProjectWide, node)

			return nil
		}

		if graph.hasCase(node.CaseSlug) {
			graph.CaseDocuments[node.CaseSlug] = append(graph.CaseDocuments[node.CaseSlug], node)

			return nil
		}

		graph.addUnresolvedCaseReference(node.CaseSlug, node)

		return nil
	})
	if err != nil {
		return fmt.Errorf("load workspace documents: %w", err)
	}

	return nil
}

func (b WorkspaceGraphBuilder) documentNode(path string) (WorkspaceNode, error) {
	info, err := os.Stat(path)
	if err != nil {
		return WorkspaceNode{}, fmt.Errorf("stat workspace document: %w", err)
	}

	content, err := os.ReadFile(path)
	if err != nil {
		return WorkspaceNode{}, fmt.Errorf("read workspace document: %w", err)
	}

	kind := workspaceKindForPath(b.projectDir, path)
	if kind == WorkspaceNodeTicket {
		return b.ticketNode(path, info, content)
	}

	fields := map[string]string{}

	body := string(content)
	if strings.HasPrefix(string(content), "---\n") {
		frontmatter, nextBody, splitErr := splitMarkdownFrontmatter(string(content))
		if splitErr != nil {
			return WorkspaceNode{}, splitErr
		}

		fields = parseSimpleYAMLFields(frontmatter)
		body = nextBody
	}

	title := firstNonEmpty(
		unquoteMetadataString(fields["title"]),
		unquoteMetadataString(fields["name"]),
		firstMarkdownHeading(body),
		strings.TrimSuffix(filepath.Base(path), filepath.Ext(path)),
	)

	return WorkspaceNode{
		Kind:         kind,
		Path:         path,
		RelativePath: b.relativePath(path),
		Slug:         strings.TrimSuffix(filepath.Base(path), filepath.Ext(path)),
		Title:        title,
		Status:       unquoteMetadataString(fields["status"]),
		DocumentType: unquoteMetadataString(fields["type"]),
		SourcePath: firstNonEmpty(
			unquoteMetadataString(fields["source"]),
			unquoteMetadataString(fields["raw_path"]),
		),
		CuratedPath: firstNonEmpty(
			unquoteMetadataString(fields["curated"]),
			unquoteMetadataString(fields["curated_path"]),
		),
		CaseSlug:   documentCaseSlug(fields),
		SourceHash: contentHash(content),
		ModifiedAt: info.ModTime().UTC(),
		Size:       info.Size(),
	}, nil
}

func (b WorkspaceGraphBuilder) ticketNode(path string, info os.FileInfo, content []byte) (WorkspaceNode, error) {
	fields, err := parseTicketFrontmatter(string(content))
	if err != nil {
		return WorkspaceNode{}, err
	}

	metadata := fields.Metadata

	return WorkspaceNode{
		Kind:         WorkspaceNodeTicket,
		Path:         path,
		RelativePath: b.relativePath(path),
		Slug:         unquoteMetadataString(metadata["key"]),
		Title:        firstNonEmpty(unquoteMetadataString(metadata["title"]), filepath.Base(path)),
		Status:       unquoteMetadataString(metadata["status"]),
		CaseSlug:     unquoteMetadataString(metadata["use_case"]),
		SourceHash:   contentHash(content),
		ModifiedAt:   info.ModTime().UTC(),
		Size:         info.Size(),
	}, nil
}

func (b WorkspaceGraphBuilder) relativePath(path string) string {
	relativePath, err := filepath.Rel(b.projectDir, path)
	if err != nil {
		return filepath.ToSlash(path)
	}

	return filepath.ToSlash(relativePath)
}

func (g *WorkspaceGraph) finalize() {
	sortWorkspaceNodes(g.Cases)
	sortWorkspaceNodes(g.ProjectWide)

	for slug, nodes := range g.CaseDocuments {
		sortWorkspaceNodes(nodes)

		g.CaseDocuments[slug] = nodes
	}

	sort.SliceStable(g.Diagnostics, func(i, j int) bool {
		return g.Diagnostics[i].Path < g.Diagnostics[j].Path
	})

	sort.SliceStable(g.UnresolvedCaseReferences, func(i, j int) bool {
		return g.UnresolvedCaseReferences[i].CaseSlug < g.UnresolvedCaseReferences[j].CaseSlug
	})
}

func (g WorkspaceGraph) CaseView(slug string) ([]WorkspaceNode, bool) {
	if !g.hasCase(slug) {
		return nil, false
	}

	return append([]WorkspaceNode(nil), g.CaseDocuments[slug]...), true
}

func (g WorkspaceGraph) hasCase(slug string) bool {
	for _, current := range g.Cases {
		if current.Slug == slug {
			return true
		}
	}

	return false
}

func (g *WorkspaceGraph) addDiagnostic(path string, kind WorkspaceNodeKind, err error) {
	g.Diagnostics = append(g.Diagnostics, WorkspaceGraphDiagnostic{
		Path:    filepath.ToSlash(path),
		Kind:    kind,
		Message: err.Error(),
	})
}

func (g *WorkspaceGraph) addUnresolvedCaseReference(slug string, node WorkspaceNode) {
	for index := range g.UnresolvedCaseReferences {
		if g.UnresolvedCaseReferences[index].CaseSlug == slug {
			g.UnresolvedCaseReferences[index].References = append(
				g.UnresolvedCaseReferences[index].References,
				node,
			)

			return
		}
	}

	g.UnresolvedCaseReferences = append(g.UnresolvedCaseReferences, WorkspaceCaseReference{
		CaseSlug:   slug,
		References: []WorkspaceNode{node},
	})
}

func walkMarkdownFiles(ctx context.Context, root string, visit func(path string) error) error {
	err := filepath.WalkDir(root, func(path string, entry os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}

		err := ctx.Err()
		if err != nil {
			return err
		}

		if entry.IsDir() {
			return nil
		}

		if !isWorkspaceMarkdownPath(path) {
			return nil
		}

		return visit(path)
	})
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}

	return err
}

func isCaseMetadataPath(projectDir, path string) bool {
	relativePath, err := filepath.Rel(CaseMetadataDirectory(projectDir), path)
	if err != nil {
		return false
	}

	return !strings.HasPrefix(relativePath, ".."+string(filepath.Separator)) &&
		relativePath != ".." &&
		!filepath.IsAbs(relativePath)
}

func workspaceKindForPath(projectDir, path string) WorkspaceNodeKind {
	relativePath, err := filepath.Rel(UserPath(projectDir), path)
	if err != nil {
		return WorkspaceNodeUnknown
	}

	parts := strings.Split(filepath.ToSlash(relativePath), "/")
	if len(parts) == 0 {
		return WorkspaceNodeUnknown
	}

	switch parts[0] {
	case "dumps":
		return WorkspaceNodeDump
	case "quizzes":
		return WorkspaceNodeQuiz
	case "specs":
		return WorkspaceNodeSpec
	case "plans":
		return WorkspaceNodePlan
	case "tickets":
		return WorkspaceNodeTicket
	case "reports":
		return WorkspaceNodeReport
	default:
		return WorkspaceNodeUnknown
	}
}

// IsRawDumpNode reports whether a dump node is the raw capture sibling.
func IsRawDumpNode(node WorkspaceNode) bool {
	if node.Kind != WorkspaceNodeDump {
		return false
	}

	if node.DocumentType == "dump_raw" {
		return true
	}

	return node.Status == "raw" && strings.TrimSpace(node.CuratedPath) != ""
}

// IsCuratedDumpNode reports whether a dump node is a primary curated dump.
func IsCuratedDumpNode(node WorkspaceNode) bool {
	return node.Kind == WorkspaceNodeDump && !IsRawDumpNode(node)
}

// IsRawQuizNode reports whether a quiz node is the raw capture sibling.
func IsRawQuizNode(node WorkspaceNode) bool {
	if node.Kind != WorkspaceNodeQuiz {
		return false
	}

	if strings.HasSuffix(node.Slug, "-raw") {
		return true
	}

	return node.Status == "raw" && strings.TrimSpace(node.CuratedPath) != ""
}

// IsCuratedQuizNode reports whether a quiz node is a primary curated quiz.
func IsCuratedQuizNode(node WorkspaceNode) bool {
	return node.Kind == WorkspaceNodeQuiz && !IsRawQuizNode(node)
}

func isArchivedGenerationPath(projectDir, path string) bool {
	relativePath, err := filepath.Rel(UserPath(projectDir), path)
	if err != nil {
		return false
	}

	parts := strings.Split(filepath.ToSlash(relativePath), "/")

	return len(parts) >= 3 &&
		(parts[0] == "plans" || parts[0] == "specs") &&
		parts[1] == "archive"
}

func isWorkspaceMarkdownPath(path string) bool {
	extension := strings.ToLower(filepath.Ext(path))

	return extension == ".md" || extension == ".markdown"
}

func documentCaseSlug(fields map[string]string) string {
	return firstNonEmpty(
		unquoteMetadataString(fields["case"]),
		unquoteMetadataString(fields["case_slug"]),
		unquoteMetadataString(fields["use_case"]),
	)
}

func firstMarkdownHeading(content string) string {
	for _, line := range strings.Split(content, "\n") {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "# ") {
			return strings.TrimSpace(strings.TrimPrefix(trimmed, "# "))
		}
	}

	return ""
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}

	return ""
}

func sortWorkspaceNodes(nodes []WorkspaceNode) {
	sort.SliceStable(nodes, func(i, j int) bool {
		if nodes[i].Kind == nodes[j].Kind {
			return nodes[i].RelativePath < nodes[j].RelativePath
		}

		return nodes[i].Kind < nodes[j].Kind
	})
}
