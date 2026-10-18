package app

import (
	"path/filepath"
	"testing"

	"forge.adrianpk.com/chr/internal/core"
)

// Shared application operations preserve the destination through metadata edits
// without rewriting the existing dump or enrolling case-specific material.
func TestServiceProjectDump(t *testing.T) {
	operations, reference := newProjectDumpTestService(t)

	observed, err := operations.ReadProjectDump()
	if err != nil || observed.Destination != filepath.ToSlash(reference) {
		t.Fatal("service missed existing dump", observed, err)
	}

	saved, err := operations.AdoptProjectDump(observed.Revision)
	if err != nil || saved.ProjectDumpRef != observed.Destination {
		t.Fatal("service adoption lost identity", saved, err)
	}

	name := "Updated project name"
	if _, err := operations.PatchProjectMetadata(ProjectMetadataPatchRequest{Name: &name}); err != nil {
		t.Fatal(err)
	}

	reloaded, err := operations.ReadProjectDump()
	if err != nil || reloaded.Destination != observed.Destination || reloaded.Project.Name != name {
		t.Fatal("metadata edit lost canonical destination", reloaded, err)
	}
}

func newProjectDumpTestService(t *testing.T) (service, string) {
	t.Helper()

	project := newTestWorkspace(t)

	layout, err := core.ResolveOpsLayout(core.ResolveOpsLayoutOptions{ProjectDir: project, Namespace: "team-a"})
	if err != nil {
		t.Fatal(err)
	}

	artifact, err := core.CreateOpsArtifact(core.CreateOpsArtifactOptions{Layout: layout, Type: "dump", Title: "Existing project context", Producer: "chr"})
	if err != nil {
		t.Fatal(err)
	}

	reference, err := filepath.Rel(project, filepath.Join(artifact.Dir, core.OpsArtifactBodyFile()))
	if err != nil {
		t.Fatal(err)
	}

	return newTestService(t, project), reference
}
