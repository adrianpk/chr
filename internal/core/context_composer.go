package core

import (
	"fmt"
	"path/filepath"
	"strings"

	authoringcore "forge.adrianpk.com/chr/internal/core/authoring"
)

// ContextComposerOptions describes a package boundary value.
type ContextComposerOptions struct {
	TicketKey   string
	TicketPath  string
	ExecutionID string
	Plan        RuntimePlan
	Sources     []authoringcore.ContextSourceInventoryEntry
}

// ContextCompositionPlan describes a package boundary value.
type ContextCompositionPlan struct {
	ProjectDir  string
	TicketKey   string
	TicketPath  string
	ExecutionID string
	Paths       ContextArtifactPaths
	Sources     []authoringcore.ContextSourceInventoryEntry
	Readiness   ContextReadiness
}

// ContextArtifactPaths describes a package boundary value.
type ContextArtifactPaths struct {
	RunArtifactDir      string
	MainContextPath     string
	ContextDir          string
	AgentPath           string
	ProjectPath         string
	ProtocolsPath       string
	DocumentsPath       string
	CodeContextPath     string
	PreviousAttemptPath string
	SourceInventoryPath string
}

// ContextReadinessStatus defines a package boundary type.
type ContextReadinessStatus string

const (
	// ContextReadinessReady defines a package boundary constant.
	ContextReadinessReady ContextReadinessStatus = "ready"
	// ContextReadinessBlocked defines a package boundary constant.
	ContextReadinessBlocked ContextReadinessStatus = "blocked"
)

// ContextReadiness describes a package boundary value.
type ContextReadiness struct {
	Status                 ContextReadinessStatus
	Warnings               []string
	MissingRequiredSources []authoringcore.ContextSourceInventoryEntry
}

// PlanContextComposition executes the corresponding package operation.
func PlanContextComposition(opts ContextComposerOptions) (ContextCompositionPlan, error) {
	err := validateContextComposerOptions(opts)
	if err != nil {
		return ContextCompositionPlan{}, err
	}

	sources, err := normalizeContextSources(opts.Sources)
	if err != nil {
		return ContextCompositionPlan{}, err
	}

	return ContextCompositionPlan{
		ProjectDir:  opts.Plan.ProjectDir,
		TicketKey:   opts.TicketKey,
		TicketPath:  opts.TicketPath,
		ExecutionID: opts.ExecutionID,
		Paths:       contextArtifactPaths(opts.Plan),
		Sources:     sources,
		Readiness:   contextReadiness(sources),
	}, nil
}

func contextArtifactPaths(plan RuntimePlan) ContextArtifactPaths {
	contextDir := filepath.Join(plan.RunArtifactDir, "context")

	return ContextArtifactPaths{
		RunArtifactDir:      plan.RunArtifactDir,
		MainContextPath:     plan.WorkerContextPath,
		ContextDir:          contextDir,
		AgentPath:           filepath.Join(contextDir, "agent.md"),
		ProjectPath:         filepath.Join(contextDir, "project.md"),
		ProtocolsPath:       filepath.Join(contextDir, "protocols.md"),
		DocumentsPath:       filepath.Join(contextDir, "documents.md"),
		CodeContextPath:     filepath.Join(contextDir, "code-context.md"),
		PreviousAttemptPath: filepath.Join(contextDir, "previous-attempt.md"),
		SourceInventoryPath: filepath.Join(contextDir, "source-inventory.md"),
	}
}

func normalizeContextSources(sources []authoringcore.ContextSourceInventoryEntry) ([]authoringcore.ContextSourceInventoryEntry, error) {
	normalized := make([]authoringcore.ContextSourceInventoryEntry, 0, len(sources))

	for _, source := range sources {
		source.ID = strings.TrimSpace(source.ID)
		source.Path = strings.TrimSpace(source.Path)
		source.Description = strings.TrimSpace(source.Description)
		source.Reason = strings.TrimSpace(source.Reason)

		if source.ID == "" {
			return nil, fmt.Errorf("context source id is required")
		}

		if !source.Kind.Valid() {
			return nil, fmt.Errorf("context source kind is invalid: %s", source.Kind)
		}

		if !source.Class.Valid() {
			return nil, fmt.Errorf("context source class is invalid: %s", source.Class)
		}

		normalized = append(normalized, source)
	}

	return normalized, nil
}

func contextReadiness(sources []authoringcore.ContextSourceInventoryEntry) ContextReadiness {
	readiness := ContextReadiness{Status: ContextReadinessReady}

	for _, source := range sources {
		if source.Required && !source.Included {
			readiness.MissingRequiredSources = append(readiness.MissingRequiredSources, source)
		}
	}

	if len(readiness.MissingRequiredSources) > 0 {
		readiness.Status = ContextReadinessBlocked
		readiness.Warnings = append(readiness.Warnings, "required context sources are missing")
	}

	return readiness
}

func validateContextComposerOptions(opts ContextComposerOptions) error {
	required := []struct {
		name  string
		value string
	}{
		{name: "project dir", value: opts.Plan.ProjectDir},
		{name: "ticket key", value: opts.TicketKey},
		{name: "ticket path", value: opts.TicketPath},
		{name: "execution id", value: opts.ExecutionID},
		{name: "run artifact dir", value: opts.Plan.RunArtifactDir},
		{name: "worker context path", value: opts.Plan.WorkerContextPath},
	}

	for _, field := range required {
		if strings.TrimSpace(field.value) == "" {
			return fmt.Errorf("context composer %s is required", field.name)
		}
	}

	return nil
}
