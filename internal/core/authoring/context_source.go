package authoring

import (
	"strings"

	opsmetadata "forge.adrianpk.com/chr/internal/core/ops/metadata"
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

// ContextDocumentLoad preserves loaded documents and independent source observations.
type ContextDocumentLoad struct {
	Documents []ContextDocument
	Sources   []ContextSourceInventoryEntry
}

// RenderSourceInventory renders the corresponding package data.
func RenderSourceInventory(sources []ContextSourceInventoryEntry) string {
	var output strings.Builder

	output.WriteString("# Context Source Inventory\n\n")
	output.WriteString("| ID | Kind | Class | Path | Included | Required | Reason |\n")
	output.WriteString("| --- | --- | --- | --- | --- | --- | --- |\n")

	for _, source := range sources {
		output.WriteString("| ")
		output.WriteString(opsmetadata.ContextTableCell(source.ID))
		output.WriteString(" | ")
		output.WriteString(opsmetadata.ContextTableCell(string(source.Kind)))
		output.WriteString(" | ")
		output.WriteString(opsmetadata.ContextTableCell(string(source.Class)))
		output.WriteString(" | ")
		output.WriteString(opsmetadata.ContextTableCell(source.Path))
		output.WriteString(" | ")
		output.WriteString(opsmetadata.YesNo(source.Included))
		output.WriteString(" | ")
		output.WriteString(opsmetadata.YesNo(source.Required))
		output.WriteString(" | ")
		output.WriteString(opsmetadata.ContextTableCell(source.Reason))
		output.WriteString(" |\n")
	}

	return output.String()
}

// ContextDocument describes a package boundary value.
type ContextDocument struct {
	ID           string
	Path         string
	RelativePath string
	Content      string
}
