package core

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"time"

	quizcore "forge.adrianpk.com/chr/internal/core/quiz"
)

// WriteGeneratedSpec applies the existing writer under shared authoring exclusion.
func WriteGeneratedSpec(ctx context.Context, projectDir string, spec GeneratedSpec) (result GeneratedSpecPaths, err error) {
	err = WithAuthoringMutation(ctx, projectDir, func(ctx context.Context) error { result, err = writeGeneratedSpec(ctx, projectDir, spec); return err })

	return result, err
}

// WriteGeneratedPlan applies the existing writer under shared authoring exclusion.
func WriteGeneratedPlan(ctx context.Context, projectDir string, plan GeneratedPlan) (result GeneratedPlanPaths, err error) {
	err = WithAuthoringMutation(ctx, projectDir, func(ctx context.Context) error { result, err = writeGeneratedPlan(ctx, projectDir, plan); return err })

	return result, err
}

// ArchiveGeneratedPlan serializes the existing archive operation with generation.
func ArchiveGeneratedPlan(ctx context.Context, projectDir string, scope quizcore.Scope, at time.Time) (result ArchivedPlanPaths, found bool, err error) {
	err = WithAuthoringMutation(ctx, projectDir, func(ctx context.Context) error {
		result, found, err = archiveGeneratedPlan(ctx, projectDir, scope, at)

		return err
	})

	return result, found, err
}

// SynthesizeAssistedArtifact shares the session/artifact mutation owner.
func SynthesizeAssistedArtifact(ctx context.Context, opts AssistedArtifactSynthesisOptions) (result AssistedArtifactSynthesisResult, err error) {
	if err := validateAssistedArtifactSynthesisOptions(opts); err != nil {
		return result, err
	}

	if opts.TargetKind == AuthoringArtifactKindDump {
		var cancel context.CancelFunc

		ctx, cancel = context.WithTimeout(ctx, 90*time.Second)
		defer cancel()
	}

	err = WithAuthoringMutation(ctx, opts.ProjectDir, func(ctx context.Context) error { result, err = synthesizeAssistedArtifact(ctx, opts); return err })

	return result, err
}

// CurateAssistedAnswers shares the session/artifact mutation owner.
func CurateAssistedAnswers(ctx context.Context, opts AssistedAnswerCurationOptions) (result AssistedAnswerCurationResult, err error) {
	if err := validateAssistedAnswerCurationOptions(opts); err != nil {
		return result, err
	}

	err = WithAuthoringMutation(ctx, filepath.Dir(filepath.Dir(filepath.Dir(filepath.Dir(opts.SessionPaths.SessionDir)))), func(ctx context.Context) error { result, err = curateAssistedAnswers(ctx, opts); return err })

	return result, err
}

// WriteContextualQuiz preserves the existing standalone writer with a bounded lock wait.
func WriteContextualQuiz(plan ContextualQuizGenerationPlan) error {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	return WriteContextualQuizContext(ctx, plan)
}

// WriteContextualQuizContext serializes a contextual quiz with shared authoring.
func WriteContextualQuizContext(ctx context.Context, plan ContextualQuizGenerationPlan) error {
	root := strings.TrimSuffix(plan.Layout.FullPath, string(os.PathSeparator)+filepath.FromSlash(plan.Layout.Ref.Path))

	return WithAuthoringMutation(ctx, root, func(context.Context) error { return writeContextualQuiz(plan) })
}

// WriteQuizToDraftPromotion preserves the standalone writer with a bounded lock wait.
func WriteQuizToDraftPromotion(plan QuizToDraftPromotionPlan) error {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	return WriteQuizToDraftPromotionContext(ctx, plan)
}

// WriteQuizToDraftPromotionContext serializes promotion with shared authoring.
func WriteQuizToDraftPromotionContext(ctx context.Context, plan QuizToDraftPromotionPlan) error {
	root := strings.TrimSuffix(plan.Layout.FullPath, string(os.PathSeparator)+filepath.FromSlash(plan.Layout.Ref.Path))

	return WithAuthoringMutation(ctx, root, func(context.Context) error { return writeQuizToDraftPromotion(plan) })
}

// WriteQuizDraftSpec publishes quiz-generated specs under the shared owner.
func WriteQuizDraftSpec(ctx context.Context, projectDir string, spec QuizDraftSpec) (result QuizDraftSpecPaths, err error) {
	err = WithAuthoringMutation(ctx, projectDir, func(ctx context.Context) error { result, err = writeQuizDraftSpec(ctx, projectDir, spec); return err })

	return result, err
}
