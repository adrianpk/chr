package chc

import (
	"fmt"
	"strings"
)

type ContextReadinessHandoffOptions struct {
	Composition   ContextCompositionPlan
	Plan          RuntimePlan
	RequiredKinds []ContextSourceKind
}

type ContextReadinessHandoff struct {
	Status       ContextReadinessStatus
	Completeness ContextCompleteness
	Issues       []ContextReadinessIssue
	Warnings     []string
	Artifacts    []ContextArtifactLink
	RuntimeNotes []string
}

type ContextCompleteness struct {
	Passed int
	Total  int
}

type ContextReadinessIssue struct {
	Field   string
	Message string
}

type ContextArtifactLink struct {
	Name          string
	HostPath      string
	ContainerPath string
}

func EvaluateContextReadinessHandoff(opts ContextReadinessHandoffOptions) ContextReadinessHandoff {
	checks := contextReadinessChecks(opts)
	issues := contextReadinessIssues(checks)

	status := ContextReadinessReady
	if len(issues) > 0 {
		status = ContextReadinessBlocked
	}

	return ContextReadinessHandoff{
		Status:       status,
		Completeness: contextCompleteness(checks),
		Issues:       issues,
		Warnings:     contextReadinessWarnings(opts.Composition),
		Artifacts:    contextArtifactLinks(opts.Plan, opts.Composition.Paths),
		RuntimeNotes: contextRuntimeNotes(),
	}
}

func RenderContextReadinessHandoff(handoff ContextReadinessHandoff) string {
	var output strings.Builder

	output.WriteString("# Context Readiness Handoff\n\n")
	writeMarkdownField(&output, "Status", string(handoff.Status))
	writeMarkdownField(&output, "Completeness", handoff.Completeness.String())
	writeReadinessIssues(&output, handoff.Issues)
	writeMarkdownList(&output, "Warnings", handoff.Warnings)
	writeArtifactLinks(&output, handoff.Artifacts)
	writeMarkdownList(&output, "Runtime Workspace Notes", handoff.RuntimeNotes)

	return output.String()
}

func (c ContextCompleteness) String() string {
	if c.Total == 0 {
		return "0/0"
	}

	return fmt.Sprintf("%d/%d", c.Passed, c.Total)
}

type contextReadinessCheck struct {
	field  string
	ok     bool
	reason string
}

func contextReadinessChecks(opts ContextReadinessHandoffOptions) []contextReadinessCheck {
	checks := []contextReadinessCheck{}
	checks = append(checks, identityContextChecks(opts.Composition)...)
	checks = append(checks, runtimeContextChecks(opts)...)
	checks = append(checks, sourceKindChecks(opts.Composition.Sources, requiredContextKinds(opts.RequiredKinds))...)
	checks = append(checks, requiredSourceChecks(opts.Composition.Sources)...)

	return checks
}

func identityContextChecks(composition ContextCompositionPlan) []contextReadinessCheck {
	return []contextReadinessCheck{
		{field: "execution_id", ok: composition.ExecutionID != "", reason: "execution id is required"},
		{field: "ticket_key", ok: composition.TicketKey != "", reason: "ticket key is required"},
		{field: "ticket_path", ok: composition.TicketPath != "", reason: "ticket path is required"},
	}
}

func runtimeContextChecks(opts ContextReadinessHandoffOptions) []contextReadinessCheck {
	plan := opts.Plan
	paths := opts.Composition.Paths

	return []contextReadinessCheck{
		{field: "project_dir", ok: plan.ProjectDir != "", reason: "runtime project dir is required"},
		{field: "worker_context_path", ok: paths.MainContextPath != "", reason: "worker context path is required"},
		{field: "result_path", ok: plan.ResultPath != "", reason: "result artifact path is required"},
		{field: "log_path", ok: plan.LogPath != "", reason: "runtime log path is required"},
		{field: "validation_path", ok: plan.ValidationPath != "", reason: "validation artifact path is required"},
		{field: "validation_command", ok: strings.TrimSpace(plan.ValidationCommand) != "", reason: "validation command is required"},
	}
}

func sourceKindChecks(sources []ContextSourceInventoryEntry, requiredKinds []ContextSourceKind) []contextReadinessCheck {
	checks := make([]contextReadinessCheck, 0, len(requiredKinds))

	for _, kind := range requiredKinds {
		checks = append(checks, contextReadinessCheck{
			field:  "source:" + string(kind),
			ok:     hasIncludedSourceKind(sources, kind),
			reason: "required context source kind is missing: " + string(kind),
		})
	}

	return checks
}

func requiredSourceChecks(sources []ContextSourceInventoryEntry) []contextReadinessCheck {
	checks := []contextReadinessCheck{}

	for _, source := range sources {
		if !source.Required {
			continue
		}

		checks = append(checks, contextReadinessCheck{
			field:  "source:" + source.ID,
			ok:     source.Included,
			reason: "required context source is missing: " + source.ID,
		})
	}

	return checks
}

func requiredContextKinds(kinds []ContextSourceKind) []ContextSourceKind {
	if len(kinds) > 0 {
		return append([]ContextSourceKind(nil), kinds...)
	}

	return []ContextSourceKind{
		ContextSourceKindAgentIdentity,
		ContextSourceKindProtocol,
		ContextSourceKindProject,
		ContextSourceKindDocument,
	}
}

func hasIncludedSourceKind(sources []ContextSourceInventoryEntry, kind ContextSourceKind) bool {
	for _, source := range sources {
		if source.Kind == kind && source.Included {
			return true
		}
	}

	return false
}

func contextReadinessIssues(checks []contextReadinessCheck) []ContextReadinessIssue {
	issues := []ContextReadinessIssue{}

	for _, check := range checks {
		if check.ok {
			continue
		}

		issues = append(issues, ContextReadinessIssue{
			Field:   check.field,
			Message: check.reason,
		})
	}

	return issues
}

func contextCompleteness(checks []contextReadinessCheck) ContextCompleteness {
	completeness := ContextCompleteness{Total: len(checks)}

	for _, check := range checks {
		if check.ok {
			completeness.Passed++
		}
	}

	return completeness
}

func contextReadinessWarnings(composition ContextCompositionPlan) []string {
	warnings := append([]string(nil), composition.Readiness.Warnings...)
	if composition.Readiness.Status == ContextReadinessBlocked {
		warnings = append(warnings, "context composition readiness is blocked")
	}

	return warnings
}

func contextArtifactLinks(plan RuntimePlan, paths ContextArtifactPaths) []ContextArtifactLink {
	return []ContextArtifactLink{
		contextArtifactLink(plan, "context", paths.MainContextPath),
		contextArtifactLink(plan, "agent", paths.AgentPath),
		contextArtifactLink(plan, "project", paths.ProjectPath),
		contextArtifactLink(plan, "protocols", paths.ProtocolsPath),
		contextArtifactLink(plan, "documents", paths.DocumentsPath),
		contextArtifactLink(plan, "code_context", paths.CodeContextPath),
		contextArtifactLink(plan, "previous_attempt", paths.PreviousAttemptPath),
		contextArtifactLink(plan, "source_inventory", paths.SourceInventoryPath),
		contextArtifactLink(plan, "result", plan.ResultPath),
		contextArtifactLink(plan, "runtime_log", plan.LogPath),
		contextArtifactLink(plan, "validation", plan.ValidationPath),
	}
}

func contextArtifactLink(plan RuntimePlan, name, hostPath string) ContextArtifactLink {
	return ContextArtifactLink{
		Name:          name,
		HostPath:      hostPath,
		ContainerPath: containerProjectPath(plan, hostPath),
	}
}

func contextRuntimeNotes() []string {
	return []string{
		"Initial execution uses a mounted host workspace; users should avoid concurrent edits while the agent is running.",
		"Ephemeral worktree or clone execution remains a future runtime-workspace concern.",
		"PAT-backed clone, commit, push, and PR behavior from inside the container remains out of scope for this slicing.",
	}
}

func writeReadinessIssues(output *strings.Builder, issues []ContextReadinessIssue) {
	output.WriteString("## Issues\n\n")

	if len(issues) == 0 {
		output.WriteString("No readiness issues were found.\n\n")

		return
	}

	for _, issue := range issues {
		output.WriteString("- `")
		output.WriteString(issue.Field)
		output.WriteString("`: ")
		output.WriteString(issue.Message)
		output.WriteString("\n")
	}

	output.WriteString("\n")
}

func writeArtifactLinks(output *strings.Builder, links []ContextArtifactLink) {
	output.WriteString("## Artifact Links\n\n")

	for _, link := range links {
		output.WriteString("- ")
		output.WriteString(link.Name)
		output.WriteString(": `")
		output.WriteString(link.ContainerPath)
		output.WriteString("`")
		output.WriteString(" (host: `")
		output.WriteString(link.HostPath)
		output.WriteString("`)\n")
	}

	output.WriteString("\n")
}
