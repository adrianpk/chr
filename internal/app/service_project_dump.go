package app

import "forge.adrianpk.com/chr/internal/core"

// ReadProjectDump observes the project-level destination across ops namespaces.
func (s service) ReadProjectDump() (core.ProjectDumpObservation, error) {
	return core.ObserveProjectDump(s.ctx, s.projectContext.ProjectDir)
}

// AdoptProjectDump binds an unambiguous existing dump without altering its components.
func (s service) AdoptProjectDump(expectedRevision string) (core.ProjectMetadata, error) {
	return core.NewProjectMetadataService(s.projectContext.ProjectDir).AdoptProjectDump(s.ctx, expectedRevision)
}
