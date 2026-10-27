package core

import workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"

type projectDumpRead struct {
	Artifact ProjectDumpArtifact
	Bytes    int
}
type (
	projectDumpFileImages struct{ Body, Metadata []byte }
	projectDumpChanges    struct {
		Changes []workflowstate.FileChange
		Result  ProjectDumpMutationResult
	}
)

type projectDumpReceiptLookup struct {
	Result ProjectDumpMutationResult
	Found  bool
}
type (
	metadataMarkdown struct{ Frontmatter, Body string }
	documentField    struct {
		Name, Value string
		Found       bool
	}
)
