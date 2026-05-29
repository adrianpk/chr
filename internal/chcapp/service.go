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
		dumpCurator:    openAIDumpCurationProvider{ProjectDir: projectContext.ProjectDir},
		quizModel:      newQuizModelAdapter(projectContext.ProjectDir),
		now:            time.Now,
	}
}

func newQuizModelAdapter(projectDir string) chc.QuizModelAdapter {
	defaultAdapter := chc.OpenAIQuizModelAdapter{ProjectDir: projectDir}
	actionAdapters := map[chc.QuizModelAction]chc.QuizModelAdapter{}

	if codexGenerationBackend("CHC_SPEC_GENERATION_BACKEND") {
		actionAdapters[chc.QuizModelActionGenerateSpec] = chc.CodexQuizModelAdapter{ProjectDir: projectDir}
	}

	if codexGenerationBackend("CHC_PLAN_GENERATION_BACKEND") {
		actionAdapters[chc.QuizModelActionGeneratePlan] = chc.CodexQuizModelAdapter{ProjectDir: projectDir}
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
	return strings.EqualFold(strings.TrimSpace(os.Getenv(name)), "codex")
}
