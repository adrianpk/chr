package chc

import (
	"fmt"
	"path/filepath"
	"strings"
)

type ContextSourceClass string

const (
	ContextSourceClassInvariant ContextSourceClass = "invariant"
	ContextSourceClassFresh     ContextSourceClass = "fresh"
)

func (c ContextSourceClass) Valid() bool {
	switch c {
	case ContextSourceClassInvariant, ContextSourceClassFresh:
		return true
	default:
		return false
	}
}

type ContextSourceKind string

const (
	ContextSourceKindAgentIdentity   ContextSourceKind = "agent_identity"
	ContextSourceKindProject         ContextSourceKind = "project"
	ContextSourceKindProtocol        ContextSourceKind = "protocol"
	ContextSourceKindDocument        ContextSourceKind = "document"
	ContextSourceKindFilesystemTree  ContextSourceKind = "filesystem_tree"
	ContextSourceKindTextSearch      ContextSourceKind = "text_search"
	ContextSourceKindGoGraph         ContextSourceKind = "gograph"
	ContextSourceKindValidation      ContextSourceKind = "validation"
	ContextSourceKindPreviousAttempt ContextSourceKind = "previous_attempt"
	ContextSourceKindOutputContract  ContextSourceKind = "output_contract"
)

func (k ContextSourceKind) Valid() bool {
	switch k {
	case ContextSourceKindAgentIdentity,
		ContextSourceKindProject,
		ContextSourceKindProtocol,
		ContextSourceKindDocument,
		ContextSourceKindFilesystemTree,
		ContextSourceKindTextSearch,
		ContextSourceKindGoGraph,
		ContextSourceKindValidation,
		ContextSourceKindPreviousAttempt,
		ContextSourceKindOutputContract:
		return true
	default:
		return false
	}
}

type ContextComposerOptions struct {
	TicketKey   string
	TicketPath  string
	ExecutionID string
	Plan        RuntimePlan
	Sources     []ContextSourceInventoryEntry
}

type ContextCompositionPlan struct {
	ProjectDir  string
	TicketKey   string
	TicketPath  string
	ExecutionID string
	Paths       ContextArtifactPaths
	Sources     []ContextSourceInventoryEntry
	Readiness   ContextReadiness
}

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

type ContextSourceInventoryEntry struct {
	ID          string
	Kind        ContextSourceKind
	Class       ContextSourceClass
	Path        string
	Description string
	Required    bool
	Included    bool
	Reason      string
}

type ContextReadinessStatus string

const (
	ContextReadinessReady   ContextReadinessStatus = "ready"
	ContextReadinessBlocked ContextReadinessStatus = "blocked"
)

type ContextReadiness struct {
	Status                 ContextReadinessStatus
	Warnings               []string
	MissingRequiredSources []ContextSourceInventoryEntry
}

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

func normalizeContextSources(sources []ContextSourceInventoryEntry) ([]ContextSourceInventoryEntry, error) {
	normalized := make([]ContextSourceInventoryEntry, 0, len(sources))

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

func contextReadiness(sources []ContextSourceInventoryEntry) ContextReadiness {
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
