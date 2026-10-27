package app

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"

	"forge.adrianpk.com/chr/internal/core"
	quizcore "forge.adrianpk.com/chr/internal/core/quiz"
)

func (s service) observePlanContext(ctx context.Context, scope quizcore.QuizScope) (core.PlanSourceObservation, error) {
	sources, err := core.ObservePlanSources(ctx, s.projectContext.ProjectDir, scope)
	if err != nil || scope.CaseSlug != "" {
		return sources, err
	}

	// Preserve the existing project-scoped API's repository baseline. Observe it
	// once per snapshot; all model stages receive the captured values.
	sections, err := s.projectWidePlanningSections()
	if err != nil {
		return sources, err
	}

	for _, section := range sections {
		hash := sha256.Sum256([]byte(section.Content))
		sources.Inputs = append(sources.Inputs, core.PlanSourceInput{Kind: section.Name, Reference: section.Authority, Revision: hex.EncodeToString(hash[:]), Content: section.Content})
	}

	encoded, err := json.Marshal(sources.Inputs)
	if err != nil {
		return sources, err
	}

	if len(encoded) > 96*1024 {
		return sources, errors.New("plan source context exceeds 96 KiB")
	}

	hash := sha256.Sum256(append([]byte(sources.Hash), encoded...))
	sources.Hash = hex.EncodeToString(hash[:])

	return sources, ctx.Err()
}
