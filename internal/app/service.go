package app

import (
	"context"
	"io/fs"
	"os"
	"strings"
	"time"

	"forge.adrianpk.com/chr/internal/core"
	"forge.adrianpk.com/chr/internal/harness"
	"forge.adrianpk.com/chr/internal/harness/codex"
)

type service struct {
	afterGuidedPreparation func(string) error
	afterGuidedSpec        func(string) error
	afterGuidedQuiz        func() error
	guidedContext          *guidedGenerationContext
	ctx                    context.Context
	env                    core.UserEnvironment
	assetsFS               fs.FS
	projectContext         commandProjectContext
	dumpEditor             DumpEditorLauncher
	dumpCurator            DumpCurationProvider
	quizModel              core.QuizModelAdapter
	conversationModel      core.StructuredModelProvider
	generationBinder       generationOperationBinder
	generationArtifactsDir string
	generationExecutionDir string
	generationRunner       harness.AgentRunner
	workflowWorker         []string
	workflowNamespace      string
	workflowPullRequests   core.WorkflowPullRequestProvider
	workflowIntegrationGit core.WorkflowIntegrationGit
	workflowPreparationGit core.WorkflowPreparationGit
	now                    func() time.Time
}

func newService(
	ctx context.Context,
	env core.UserEnvironment,
	projectContext commandProjectContext,
	assetsFS fs.FS,
) service {
	operations := service{
		ctx:               ctx,
		env:               env,
		assetsFS:          assetsFS,
		projectContext:    projectContext,
		dumpEditor:        defaultDumpEditorLauncher{},
		dumpCurator:       newDumpCurationProvider(projectContext.ProjectDir),
		quizModel:         newQuizModelAdapter(projectContext.ProjectDir),
		conversationModel: newConversationModel(),
		now:               time.Now,
	}
	operations.generationBinder = service.bindHarnessGeneration

	return operations
}

func newDumpCurationProvider(projectDir string) DumpCurationProvider {
	if openAIGenerationBackend("CHR_DUMP_GENERATION_BACKEND") {
		return openAIDumpCurationProvider{ProjectDir: projectDir}
	}

	return harnessDumpCurationProvider{ProjectDir: projectDir, Runner: codex.Runner{}}
}

func newQuizModelAdapter(projectDir string) core.QuizModelAdapter {
	defaultAdapter := core.OpenAIQuizModelAdapter{ProjectDir: projectDir}
	actionAdapters := map[core.QuizModelAction]core.QuizModelAdapter{}

	if codexGenerationBackend("CHR_SPEC_GENERATION_BACKEND") {
		actionAdapters[core.QuizModelActionGenerateSpec] = core.HarnessQuizModelAdapter{ProjectDir: projectDir, Runner: codex.Runner{}}
	}

	if codexGenerationBackend("CHR_ACCEPTANCE_FEATURE_GENERATION_BACKEND") {
		actionAdapters[core.QuizModelActionGenerateAcceptanceFeature] = core.HarnessQuizModelAdapter{ProjectDir: projectDir, Runner: codex.Runner{}}
	}

	if codexGenerationBackend("CHR_PLAN_GENERATION_BACKEND") {
		actionAdapters[core.QuizModelActionGeneratePlan] = core.HarnessQuizModelAdapter{ProjectDir: projectDir, Runner: codex.Runner{}}
	}

	if codexGenerationBackend("CHR_PLAN_TRACKING_GENERATION_BACKEND") {
		actionAdapters[core.QuizModelActionMaterializePlanTracking] = core.HarnessQuizModelAdapter{ProjectDir: projectDir, Runner: codex.Runner{}}
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

func newConversationModel() *core.ConfiguredStructuredModel {
	model := core.NewConfiguredStructuredModel(openAIGenerationBackend("CHR_CONVERSATION_GENERATION_BACKEND"))
	model.Runner = codex.Runner{}

	return model
}
