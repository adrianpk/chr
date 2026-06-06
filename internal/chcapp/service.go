package chcapp

import (
	"context"
	"io/fs"
	"os"
	"strings"
	"time"

	"forge.adrianpk.com/chc/internal/chc"
)

type service struct {
	ctx            context.Context
	env            chc.UserEnvironment
	assetsFS       fs.FS
	projectContext commandProjectContext
	dumpEditor     DumpEditorLauncher
	dumpCurator    DumpCurationProvider
	quizModel      chc.QuizModelAdapter
	now            func() time.Time
}

func newService(
	ctx context.Context,
	env chc.UserEnvironment,
	projectContext commandProjectContext,
	assetsFS fs.FS,
) service {
	return service{
		ctx:            ctx,
		env:            env,
		assetsFS:       assetsFS,
		projectContext: projectContext,
		dumpEditor:     defaultDumpEditorLauncher{},
		dumpCurator:    newDumpCurationProvider(projectContext.ProjectDir),
		quizModel:      newQuizModelAdapter(projectContext.ProjectDir),
		now:            time.Now,
	}
}

func newDumpCurationProvider(projectDir string) DumpCurationProvider {
	if openAIGenerationBackend("CHC_DUMP_GENERATION_BACKEND") {
		return openAIDumpCurationProvider{ProjectDir: projectDir}
	}

	return codexDumpCurationProvider{ProjectDir: projectDir}
}

func newQuizModelAdapter(projectDir string) chc.QuizModelAdapter {
	defaultAdapter := chc.OpenAIQuizModelAdapter{ProjectDir: projectDir}
	actionAdapters := map[chc.QuizModelAction]chc.QuizModelAdapter{}

	if codexGenerationBackend("CHC_SPEC_GENERATION_BACKEND") {
		actionAdapters[chc.QuizModelActionGenerateSpec] = chc.CodexQuizModelAdapter{ProjectDir: projectDir}
	}

	if codexGenerationBackend("CHC_ACCEPTANCE_FEATURE_GENERATION_BACKEND") {
		actionAdapters[chc.QuizModelActionGenerateAcceptanceFeature] = chc.CodexQuizModelAdapter{ProjectDir: projectDir}
	}

	if codexGenerationBackend("CHC_PLAN_GENERATION_BACKEND") {
		actionAdapters[chc.QuizModelActionGeneratePlan] = chc.CodexQuizModelAdapter{ProjectDir: projectDir}
	}

	if codexGenerationBackend("CHC_PLAN_TRACKING_GENERATION_BACKEND") {
		actionAdapters[chc.QuizModelActionMaterializePlanTracking] = chc.CodexQuizModelAdapter{ProjectDir: projectDir}
	}

	if len(actionAdapters) == 0 {
		return defaultAdapter
	}

	return chc.ActionRoutingQuizModelAdapter{
		ActionAdapters: actionAdapters,
		DefaultAdapter: defaultAdapter,
	}
}

func codexGenerationBackend(name string) bool {
	return !openAIGenerationBackend(name)
}

func openAIGenerationBackend(name string) bool {
	return strings.EqualFold(strings.TrimSpace(os.Getenv(name)), "openai")
}
