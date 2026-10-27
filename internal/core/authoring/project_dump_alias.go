package authoring

import (
	"context"
	"fmt"
	"time"
)

const projectDumpAliasLimit = 16

// ResolveProjectDumpReference follows retained general aliases without writing.
// Scoped and raw artifacts cannot act as an alias destination.
func ResolveProjectDumpReference(ctx context.Context, projectDir, reference string) (ProjectDumpArtifact, error) {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	r := &fileAuthoringSessionRepository{projectDir: projectDir}
	if err := r.checkProjectDumpInspection(ctx); err != nil {
		return ProjectDumpArtifact{}, err
	}

	resolved, err := resolveProjectDumpAlias(reference, func(ref string) (ProjectDumpArtifact, error) {
		if err := ctx.Err(); err != nil {
			return ProjectDumpArtifact{}, err
		}

		dumpRead, err := r.readProjectDumpArtifact(ref)
		artifact := dumpRead.Artifact

		return artifact, err
	})
	if err != nil {
		return resolved, err
	}

	return resolved, r.checkProjectDumpInspection(ctx)
}

func resolveProjectDumpAlias(reference string, read func(string) (ProjectDumpArtifact, error)) (ProjectDumpArtifact, error) {
	seen := map[string]bool{}
	for depth := 0; depth <= projectDumpAliasLimit; depth++ {
		if reference == "" || seen[reference] {
			return ProjectDumpArtifact{}, fmt.Errorf("%w: missing or cyclic dump alias", ErrProjectDumpReconciliation)
		}

		seen[reference] = true

		artifact, err := read(reference)
		if err != nil {
			return artifact, fmt.Errorf("%w: unresolved dump alias: %w", ErrProjectDumpReconciliation, err)
		}

		if depth > 0 && (artifact.CaseSlug != "" || artifact.Raw) {
			return artifact, fmt.Errorf("%w: a dump alias must resolve to a general destination", ErrProjectDumpReconciliation)
		}

		if artifact.CanonicalReference == "" {
			return artifact, nil
		}

		reference = artifact.CanonicalReference
	}

	return ProjectDumpArtifact{}, fmt.Errorf("%w: dump alias exceeds %d links", ErrProjectDumpReconciliation, projectDumpAliasLimit)
}

func validateProjectDumpAliases(artifacts []ProjectDumpArtifact) error {
	byReference := make(map[string]ProjectDumpArtifact, len(artifacts))
	for _, artifact := range artifacts {
		byReference[artifact.Reference] = artifact
	}

	for _, artifact := range artifacts {
		if artifact.CanonicalReference == "" {
			continue
		}

		_, err := resolveProjectDumpAlias(artifact.Reference, func(reference string) (ProjectDumpArtifact, error) {
			found, exists := byReference[reference]
			if !exists {
				return found, fmt.Errorf("missing canonical destination %s", reference)
			}

			return found, nil
		})
		if err != nil {
			return err
		}
	}

	return nil
}
