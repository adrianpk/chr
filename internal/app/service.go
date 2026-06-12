package app

import (
	"context"
	"io/fs"
	"os"
	"strings"
	"time"

	"forge.adrianpk.com/chc/internal/core"
)

type service struct {
	ctx            context.Context
	env            core.UserEnvironment
	assetsFS       fs.FS
	projectContext commandProjectContext
	dumpEditor     DumpEditorLauncher
	dumpCurator    DumpCurationProvider
	quizModel      core.QuizModelAdapter
	now            func() time.Time
}

func newService(
	ctx context.Context,
	env core.UserEnvironment,
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

func newQuizModelAdapter(projectDir string) core.QuizModelAdapter {
	defaultAdapter := core.OpenAIQuizModelAdapter{ProjectDir: projectDir}
	actionAdapters := map[core.QuizModelAction]core.QuizModelAdapter{}

	if codexGenerationBackend("CHC_SPEC_GENERATION_BACKEND") {
		actionAdapters[core.QuizModelActionGenerateSpec] = core.CodexQuizModelAdapter{ProjectDir: projectDir}
	}

	if codexGenerationBackend("CHC_ACCEPTANCE_FEATURE_GENERATION_BACKEND") {
		actionAdapters[core.QuizModelActionGenerateAcceptanceFeature] = core.CodexQuizModelAdapter{ProjectDir: projectDir}
	}

	if codexGenerationBackend("CHC_PLAN_GENERATION_BACKEND") {
		actionAdapters[core.QuizModelActionGeneratePlan] = core.CodexQuizModelAdapter{ProjectDir: projectDir}
	}

	if codexGenerationBackend("CHC_PLAN_TRACKING_GENERATION_BACKEND") {
		actionAdapters[core.QuizModelActionMaterializePlanTracking] = core.CodexQuizModelAdapter{ProjectDir: projectDir}
	}

	if len(actionAdapters) == 0 {
		return defaultAdapter
	}

	return core.ActionRoutingQuizModelAdapter{
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
