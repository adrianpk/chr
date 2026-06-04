package chc

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// WorkspaceContextScope defines a package boundary type.
type WorkspaceContextScope string

const (
	// WorkspaceContextScopeCase defines a package boundary constant.
	WorkspaceContextScopeCase WorkspaceContextScope = "case"
	// WorkspaceContextScopeProject defines a package boundary constant.
	WorkspaceContextScopeProject WorkspaceContextScope = "project"
)

// WorkspaceContextRequest describes a package boundary value.
type WorkspaceContextRequest struct {
	ProjectDir string
	Scope      WorkspaceContextScope
	CaseSlug   string
	Graph      WorkspaceGraph
}

// WorkspaceContextAssembly describes a package boundary value.
type WorkspaceContextAssembly struct {
	ProjectDir               string
	Scope                    WorkspaceContextScope
	CaseSlug                 string
	Project                  ContextDocument
	Case                     ContextDocument
	Documents                []ContextDocument
	Sources                  []ContextSourceInventoryEntry
	Diagnostics              []WorkspaceGraphDiagnostic
	UnresolvedCaseReferences []WorkspaceCaseReference
}

// NewCaseRootedWorkspaceContextRequest constructs the corresponding package service or value.
func NewCaseRootedWorkspaceContextRequest(
	projectDir string,
	graph WorkspaceGraph,
	caseSlug string,
) WorkspaceContextRequest {
	return WorkspaceContextRequest{
		ProjectDir: projectDir,
		Scope:      WorkspaceContextScopeCase,
		CaseSlug:   caseSlug,
		Graph:      graph,
	}
}

// NewProjectWideWorkspaceContextRequest constructs the corresponding package service or value.
func NewProjectWideWorkspaceContextRequest(projectDir string, graph WorkspaceGraph) WorkspaceContextRequest {
	return WorkspaceContextRequest{
		ProjectDir: projectDir,
		Scope:      WorkspaceContextScopeProject,
		Graph:      graph,
	}
}

// AssembleWorkspaceContext executes the corresponding package operation.
func AssembleWorkspaceContext(request WorkspaceContextRequest) (WorkspaceContextAssembly, error) {
	projectDir, err := cleanAbsProjectDir(request.ProjectDir)
	if err != nil {
		return WorkspaceContextAssembly{}, err
	}

	request.ProjectDir = projectDir

	err = validateWorkspaceContextRequest(request)
	if err != nil {
		return WorkspaceContextAssembly{}, err
	}

	assembly := WorkspaceContextAssembly{
		ProjectDir:               projectDir,
		Scope:                    request.Scope,
		CaseSlug:                 request.CaseSlug,
		Diagnostics:              append([]WorkspaceGraphDiagnostic(nil), request.Graph.Diagnostics...),
		UnresolvedCaseReferences: workspaceContextUnresolvedReferences(request),
	}

	err = loadWorkspaceContextProject(request, &assembly)
	if err != nil {
		return WorkspaceContextAssembly{}, err
	}

	caseFound, err := loadWorkspaceContextCase(request, &assembly)
	if err != nil || !caseFound {
		return assembly, err
	}

	for _, node := range workspaceContextDocumentNodes(request) {
		document, source, loadErr := loadWorkspaceContextNode(
			projectDir,
			workspaceContextDocumentID(node),
			node,
			false,
		)
		if loadErr != nil {
			return WorkspaceContextAssembly{}, loadErr
		}

		if document.ID != "" {
			assembly.Documents = append(assembly.Documents, document)
		}

		assembly.Sources = append(assembly.Sources, source)
	}

	return assembly, nil
}

// RenderWorkspaceContextAssembly renders the corresponding package data.
func RenderWorkspaceContextAssembly(assembly WorkspaceContextAssembly) string {
	var output strings.Builder

	output.WriteString("# Workspace Context\n\n")
	writeMarkdownField(&output, "Project Directory", assembly.ProjectDir)
	writeMarkdownField(&output, "Scope", string(assembly.Scope))

	if assembly.CaseSlug != "" {
		writeMarkdownField(&output, "Case", assembly.CaseSlug)
	}

	output.WriteString("## Project\n\n")
	writeContextDocuments(&output, []ContextDocument{assembly.Project}, "Project document was not loaded.")

	if assembly.Scope == WorkspaceContextScopeCase {
		output.WriteString("## Case\n\n")
		writeContextDocuments(&output, []ContextDocument{assembly.Case}, "Case document was not loaded.")
	}

	output.WriteString("## Workspace Documents\n\n")
	writeContextDocuments(&output, assembly.Documents, "No workspace documents were selected.")

	output.WriteString("## Diagnostics\n\n")
	writeWorkspaceContextDiagnostics(&output, assembly.Diagnostics)

	output.WriteString("## Unresolved Case References\n\n")
	writeWorkspaceContextUnresolvedReferences(&output, assembly.UnresolvedCaseReferences)

	output.WriteString("## Sources\n\n")
	output.WriteString(RenderSourceInventory(assembly.Sources))

	return output.String()
}

func loadWorkspaceContextProject(
	request WorkspaceContextRequest,
	assembly *WorkspaceContextAssembly,
) error {
	if request.Graph.Project == nil {
		return nil
	}

	document, source, err := loadWorkspaceContextNode(
		request.ProjectDir,
		"workspace:project",
		*request.Graph.Project,
		true,
	)
	if err != nil {
		return err
	}

	assembly.Project = document
	assembly.Sources = append(assembly.Sources, source)

	return nil
}

func loadWorkspaceContextCase(
	request WorkspaceContextRequest,
	assembly *WorkspaceContextAssembly,
) (bool, error) {
	if request.Scope != WorkspaceContextScopeCase {
		return true, nil
	}

	caseNode, ok := workspaceContextCaseNode(request.Graph, request.CaseSlug)
	if !ok {
		assembly.Sources = append(assembly.Sources, missingWorkspaceContextSource(
			"workspace:case:"+request.CaseSlug,
			"",
			"Selected case document.",
			"case was not found in workspace graph",
			true,
		))

		return false, nil
	}

	document, source, err := loadWorkspaceContextNode(
		request.ProjectDir,
		"workspace:case:"+caseNode.Slug,
		caseNode,
		true,
	)
	if err != nil {
		return false, err
	}

	assembly.Case = document
	assembly.Sources = append(assembly.Sources, source)

	return true, nil
}

func validateWorkspaceContextRequest(request WorkspaceContextRequest) error {
	switch request.Scope {
	case WorkspaceContextScopeCase:
		if strings.TrimSpace(request.CaseSlug) == "" {
			return fmt.Errorf("workspace context case slug is required")
		}
	case WorkspaceContextScopeProject:
	default:
		return fmt.Errorf("workspace context scope is invalid: %s", request.Scope)
	}

	return nil
}

func workspaceContextDocumentNodes(request WorkspaceContextRequest) []WorkspaceNode {
	switch request.Scope {
	case WorkspaceContextScopeCase:
		return workspaceContextSupportedNodes(request.Graph.CaseDocuments[request.CaseSlug])
	case WorkspaceContextScopeProject:
		return workspaceContextSupportedNodes(request.Graph.ProjectWide)
	default:
		return nil
	}
}

func workspaceContextSupportedNodes(nodes []WorkspaceNode) []WorkspaceNode {
	selected := make([]WorkspaceNode, 0, len(nodes))
	for _, node := range nodes {
		if IsRawDumpNode(node) || IsRawQuizNode(node) {
			continue
		}

		if workspaceContextSupportsNodeKind(node.Kind) {
			selected = append(selected, node)
		}
	}

	return selected
}

func workspaceContextSupportsNodeKind(kind WorkspaceNodeKind) bool {
	switch kind {
	case WorkspaceNodeDump, WorkspaceNodeQuiz, WorkspaceNodeSpec, WorkspaceNodePlan:
		return true
	default:
		return false
	}
}

func workspaceContextCaseNode(graph WorkspaceGraph, slug string) (WorkspaceNode, bool) {
	for _, node := range graph.Cases {
		if node.Slug == slug {
			return node, true
		}
	}

	return WorkspaceNode{}, false
}

func loadWorkspaceContextNode(
	projectDir string,
	id string,
	node WorkspaceNode,
	required bool,
) (ContextDocument, ContextSourceInventoryEntry, error) {
	path := node.Path
	if path == "" && node.RelativePath != "" {
		path = filepath.Join(projectDir, filepath.FromSlash(node.RelativePath))
	}

	if path == "" {
		return ContextDocument{}, missingWorkspaceContextSource(
			id,
			node.RelativePath,
			workspaceContextSourceDescription(node),
			"node path is empty",
			required,
		), nil
	}

	if !filepath.IsAbs(path) {
		path = filepath.Join(projectDir, filepath.FromSlash(path))
	}

	path = filepath.Clean(path)
	if !pathIsInsideProject(projectDir, path) {
		return ContextDocument{}, ContextSourceInventoryEntry{}, fmt.Errorf("workspace context document is outside project: %s", path)
	}

	document, err := readContextDocument(projectDir, id, path)
	if err != nil {
		if os.IsNotExist(err) {
			return ContextDocument{}, missingWorkspaceContextSource(
				id,
				workspaceContextNodeRelativePath(projectDir, node, path),
				workspaceContextSourceDescription(node),
				"file was not found",
				required,
			), nil
		}

		return ContextDocument{}, ContextSourceInventoryEntry{}, err
	}

	return document, ContextSourceInventoryEntry{
		ID:          id,
		Kind:        ContextSourceKindDocument,
		Class:       ContextSourceClassFresh,
		Path:        document.RelativePath,
		Description: workspaceContextSourceDescription(node),
		Required:    required,
		Included:    true,
	}, nil
}

func missingWorkspaceContextSource(
	id string,
	path string,
	description string,
	reason string,
	required bool,
) ContextSourceInventoryEntry {
	return ContextSourceInventoryEntry{
		ID:          id,
		Kind:        ContextSourceKindDocument,
		Class:       ContextSourceClassFresh,
		Path:        filepath.ToSlash(path),
		Description: description,
		Required:    required,
		Included:    false,
		Reason:      reason,
	}
}

func workspaceContextDocumentID(node WorkspaceNode) string {
	if node.RelativePath != "" {
		return "workspace:" + node.RelativePath
	}

	return "workspace:" + string(node.Kind) + ":" + node.Slug
}

func workspaceContextSourceDescription(node WorkspaceNode) string {
	switch node.Kind {
	case WorkspaceNodeProject:
		return "Project metadata document."
	case WorkspaceNodeCase:
		return "Selected case document."
	default:
		return "Workspace " + string(node.Kind) + " document."
	}
}

func workspaceContextNodeRelativePath(projectDir string, node WorkspaceNode, path string) string {
	if node.RelativePath != "" {
		return node.RelativePath
	}

	relPath, err := filepath.Rel(projectDir, path)
	if err != nil {
		return filepath.ToSlash(path)
	}

	return filepath.ToSlash(relPath)
}

func workspaceContextUnresolvedReferences(request WorkspaceContextRequest) []WorkspaceCaseReference {
	if request.Scope == WorkspaceContextScopeProject {
		return append([]WorkspaceCaseReference(nil), request.Graph.UnresolvedCaseReferences...)
	}

	for _, reference := range request.Graph.UnresolvedCaseReferences {
		if reference.CaseSlug == request.CaseSlug {
			return []WorkspaceCaseReference{reference}
		}
	}

	return nil
}

func writeWorkspaceContextDiagnostics(output *strings.Builder, diagnostics []WorkspaceGraphDiagnostic) {
	if len(diagnostics) == 0 {
		output.WriteString("No workspace graph diagnostics.\n\n")

		return
	}

	for _, diagnostic := range diagnostics {
		output.WriteString("- ")
		output.WriteString(diagnostic.Path)
		output.WriteString(" [")
		output.WriteString(string(diagnostic.Kind))
		output.WriteString("]: ")
		output.WriteString(diagnostic.Message)
		output.WriteString("\n")
	}

	output.WriteString("\n")
}

func writeWorkspaceContextUnresolvedReferences(output *strings.Builder, references []WorkspaceCaseReference) {
	if len(references) == 0 {
		output.WriteString("No unresolved case references.\n\n")

		return
	}

	for _, reference := range references {
		output.WriteString("- ")
		output.WriteString(reference.CaseSlug)
		output.WriteString("\n")

		for _, node := range reference.References {
			output.WriteString("  - ")
			output.WriteString(node.RelativePath)
			output.WriteString("\n")
		}
	}

	output.WriteString("\n")
}
