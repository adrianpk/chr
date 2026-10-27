package app

import (
	"context"
	"path/filepath"
	"strings"

	"forge.adrianpk.com/chr/internal/core"
	authoringcore "forge.adrianpk.com/chr/internal/core/authoring"
	opscore "forge.adrianpk.com/chr/internal/core/ops"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

// ReadGuidedUseCaseOriginal returns a complete committed image without reading
// a replacement working copy or recovering any stage.
func (s service) ReadGuidedUseCaseOriginal(ctx context.Context, id, namespace, kind string) (GuidedUseCaseArtifact, error) {
	scope, err := s.guidedUseCaseScope(id, namespace)
	if err != nil {
		return GuidedUseCaseArtifact{}, err
	}

	session, err := core.ReadGuidedUseCase(ctx, s.projectContext.ProjectDir, scope)
	if err != nil {
		return GuidedUseCaseArtifact{}, err
	}

	for _, checkpoint := range session.Guided.Checkpoints {
		for _, artifact := range checkpoint.Artifacts {
			if !guidedOriginalMatches(session, checkpoint.Stage, artifact.Path, kind) {
				continue
			}

			content, err := core.ReadGuidedCommittedArtifact(ctx, s.projectContext.ProjectDir, checkpoint, artifact.Path)

			return GuidedUseCaseArtifact{Stage: checkpoint.Stage, Reference: artifact.Path, Commit: checkpoint.Commit, Content: string(content)}, err
		}
	}

	return GuidedUseCaseArtifact{}, workflowstate.ErrConflict
}

func guidedOriginalMatches(session authoringcore.Session, stage, path, kind string) bool {
	if kind == "tracker" {
		if session.Guided.Work == nil || session.Guided.Work.Materialization == nil {
			return false
		}

		ref := session.Guided.Work.Materialization.PlanRevision.CurrentTrackerRef

		return stage == "preparation" && path == filepath.ToSlash(filepath.Join(opscore.DirectoryName, session.Namespace, ref))
	}

	stages := map[string]string{"dump": "dump", "quiz": "quiz", "specification": "specification", "plan": "preparation"}
	if stages[kind] != stage || !strings.HasSuffix(path, "/artifact.md") {
		return false
	}

	if kind == "plan" {
		return session.Guided.Work != nil && session.Guided.Work.Publication != nil && path == session.Guided.Work.Publication.Reference
	}

	return true
}
