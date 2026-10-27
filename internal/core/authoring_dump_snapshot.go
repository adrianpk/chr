package core

import (
	"path/filepath"
	"strings"

	authoringcore "forge.adrianpk.com/chr/internal/core/authoring"
	"forge.adrianpk.com/chr/internal/core/contenthash"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
	workspacecore "forge.adrianpk.com/chr/internal/core/workspace"
)

// Dump observations include every alias image and the resolved canonical body.
// This keeps an old session reference usable without dropping its edit guard.
func (r *fileAuthoringSessionRepository) readAuthoringDumpObservations(snapshot *AuthoringSnapshot, ref authoringcore.ArtifactRef, images map[string][]byte) (ProjectDumpArtifact, error) {
	if err := ValidateProjectDumpReference(ref.Path); err != nil || ref.Path == "" || strings.Split(ref.Path, "/")[3] != ref.ID {
		return ProjectDumpArtifact{}, workflowstate.ErrConflict
	}

	resolved, err := resolveProjectDumpAlias(ref.Path, func(reference string) (ProjectDumpArtifact, error) {
		return r.readAuthoringDumpImage(snapshot, ref, reference, images)
	})
	if err != nil {
		return resolved, err
	}

	if resolved.Raw || resolved.CaseSlug != "" {
		if !validAuthoringArtifactRef(snapshot.Session.Namespace, ref) {
			return resolved, workflowstate.ErrConflict
		}

		return resolved, nil
	}

	project, err := r.authoringDumpProject(images)
	if err != nil {
		return resolved, err
	}

	if err := guardAuthoringDumpReference(snapshot.Session.Namespace, ref, resolved, project); err != nil {
		return resolved, err
	}

	snapshot.projectMetadataHash = projectDumpMetadataImage(project)

	return resolved, nil
}

func (r *fileAuthoringSessionRepository) readAuthoringDumpImage(snapshot *AuthoringSnapshot, original authoringcore.ArtifactRef, reference string, images map[string][]byte) (ProjectDumpArtifact, error) {
	body, err := r.readAuthoringDumpBytes(reference, images)
	if err != nil {
		return ProjectDumpArtifact{}, err
	}

	encoded, err := r.readAuthoringDumpBytes(filepath.Join(filepath.Dir(reference), "metadata.json"), images)
	if err != nil {
		return ProjectDumpArtifact{}, err
	}

	metadata, err := decodeProjectDumpMetadata(reference, encoded)
	if err != nil {
		return ProjectDumpArtifact{}, err
	}

	fields, err := projectDumpBodyFields(body)
	if err != nil {
		return ProjectDumpArtifact{}, err
	}

	artifact, err := projectDumpArtifact(reference, projectDumpFileImages{Body: body, Metadata: encoded}, fields, metadata)
	if err != nil {
		return artifact, err
	}

	if err := appendAuthoringDumpObservation(snapshot, original, artifact, body, encoded); err != nil {
		return artifact, err
	}

	return artifact, nil
}

func (r *fileAuthoringSessionRepository) authoringDumpProject(images map[string][]byte) (ProjectMetadata, error) {
	reference := filepath.Join(workspacecore.DirectoryName, UserDirectoryName, "project.md")
	if content, found := images[reference]; found {
		project, err := ParseProjectMetadata(string(content))
		project.Exists, project.SourceHash = true, contenthash.Sum(content)

		return project, err
	}

	return NewProjectMetadataService(r.projectDir).Load()
}

func (r *fileAuthoringSessionRepository) readAuthoringDumpBytes(reference string, images map[string][]byte) ([]byte, error) {
	if content, found := images[reference]; found {
		return content, nil
	}

	return r.readAuthoringDocument(reference)
}

func guardAuthoringDumpReference(namespace string, ref authoringcore.ArtifactRef, resolved ProjectDumpArtifact, project ProjectMetadata) error {
	if project.ProjectDumpRef != "" && project.ProjectDumpRef != resolved.Reference {
		return workflowstate.ErrConflict
	}

	if !validAuthoringArtifactRef(namespace, ref) && project.ProjectDumpRef != resolved.Reference {
		return workflowstate.ErrConflict
	}

	return nil
}

func appendAuthoringDumpObservation(snapshot *AuthoringSnapshot, original authoringcore.ArtifactRef, artifact ProjectDumpArtifact, body, encoded []byte) error {
	if authoringArtifactObservation(*snapshot, artifact.Reference) != nil {
		return nil
	}

	for _, alias := range snapshot.dumpAliases {
		if alias.Ref.Path == artifact.Reference {
			return nil
		}
	}

	ref := authoringDumpRef(artifact.Reference)
	ref.Role = original.Role
	observed := AuthoringArtifactObservation{Ref: ref, SourceHash: artifact.SourceHash, MetadataHash: artifact.MetadataHash}

	snapshot.artifactBytes += len(body) + len(encoded)
	if artifact.CanonicalReference != "" {
		snapshot.dumpAliases = append(snapshot.dumpAliases, observed)
	} else {
		observed.Body = string(body)
		snapshot.Artifacts = append(snapshot.Artifacts, observed)
	}

	return validateAuthoringArtifactContext(*snapshot)
}
