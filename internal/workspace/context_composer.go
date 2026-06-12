package workspace

import (
	"fmt"
	"path/filepath"
	"strings"
)

// ContextSourceClass defines a package boundary type.
type ContextSourceClass string

const (
	// ContextSourceClassInvariant defines a package boundary constant.
	ContextSourceClassInvariant ContextSourceClass = "invariant"
	// ContextSourceClassFresh defines a package boundary constant.
	ContextSourceClassFresh ContextSourceClass = "fresh"
)

// Valid reports whether the value is one of the supported constants.
func (c ContextSourceClass) Valid() bool {
	switch c {
	case ContextSourceClassInvariant, ContextSourceClassFresh:
		return true
	default:
		return false
	}
}

// ContextSourceKind defines a package boundary type.
type ContextSourceKind string

const (
	// ContextSourceKindAgentIdentity defines a package boundary constant.
	ContextSourceKindAgentIdentity ContextSourceKind = "agent_identity"
	// ContextSourceKindProject defines a package boundary constant.
	ContextSourceKindProject ContextSourceKind = "project"
	// ContextSourceKindProtocol defines a package boundary constant.
	ContextSourceKindProtocol ContextSourceKind = "protocol"
	// ContextSourceKindDocument defines a package boundary constant.
	ContextSourceKindDocument ContextSourceKind = "document"
	// ContextSourceKindFilesystemTree defines a package boundary constant.
	ContextSourceKindFilesystemTree ContextSourceKind = "filesystem_tree"
	// ContextSourceKindTextSearch defines a package boundary constant.
	ContextSourceKindTextSearch ContextSourceKind = "text_search"
	// ContextSourceKindGoGraph defines a package boundary constant.
	ContextSourceKindGoGraph ContextSourceKind = "gograph"
	// ContextSourceKindValidation defines a package boundary constant.
	ContextSourceKindValidation ContextSourceKind = "validation"
	// ContextSourceKindPreviousAttempt defines a package boundary constant.
	ContextSourceKindPreviousAttempt ContextSourceKind = "previous_attempt"
	// ContextSourceKindOutputContract defines a package boundary constant.
	ContextSourceKindOutputContract ContextSourceKind = "output_contract"
)

// Valid reports whether the value is one of the supported constants.
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

// ContextComposerOptions describes a package boundary value.
type ContextComposerOptions struct {
	TicketKey   string
	TicketPath  string
	ExecutionID string
	Plan        RuntimePlan
	Sources     []ContextSourceInventoryEntry
}

// ContextCompositionPlan describes a package boundary value.
type ContextCompositionPlan struct {
	ProjectDir  string
	TicketKey   string
	TicketPath  string
	ExecutionID string
	Paths       ContextArtifactPaths
	Sources     []ContextSourceInventoryEntry
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

// ContextSourceInventoryEntry describes a package boundary value.
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
	MissingRequiredSources []ContextSourceInventoryEntry
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
