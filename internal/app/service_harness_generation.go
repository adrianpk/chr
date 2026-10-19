package app

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"

	"forge.adrianpk.com/chr/internal/core"
	"forge.adrianpk.com/chr/internal/harness/codex"
)

type generationOperationBinder func(service, context.Context, string, string) (context.Context, error)

var errLegacyGenerationUnavailable = errors.New("legacy generation executable is unavailable")

func (s service) bindGeneration(ctx context.Context, id, purpose string) (service, context.Context, error) {
	if s.generationBinder == nil {
		return s, ctx, errors.New("generation operation binder is required")
	}

	ctx, err := s.generationBinder(s, ctx, id, purpose)
	if err != nil {
		return s, ctx, err
	}

	s.ctx = ctx

	return s.generationAdapters(ctx), ctx, nil
}

func (s service) bindHarnessGeneration(ctx context.Context, id, purpose string) (context.Context, error) {
	if id == "" {
		id = rand.Text()
	}

	root := s.generationArtifactsDir
	if root == "" {
		root = s.projectContext.ProjectDir
	}

	location := core.HarnessGenerationLocation{ConfigDir: s.projectContext.ProjectDir, ExecutionDir: s.generationExecutionDir, ProjectRef: s.projectContext.ProjectID}
	if location.ExecutionDir == "" {
		location.ExecutionDir = s.projectContext.ProjectDir
	}

	operation, err := core.FreezeHarnessGeneration(ctx, root, id, purpose, func(ctx context.Context, config core.WorkspaceConfig) (core.HarnessGenerationBinding, string, error) {
		return s.prepareGenerationBinding(ctx, config, purpose)
	}, location)
	if err != nil {
		return ctx, err
	}

	return core.WithHarnessGeneration(ctx, operation), nil
}

func (s service) prepareGenerationBinding(ctx context.Context, config core.WorkspaceConfig, purpose string) (core.HarnessGenerationBinding, string, error) {
	if config.Harnesses.Generation == "" && purpose == "conversation" {
		if provider, ok := s.conversationModel.(*core.ConfiguredStructuredModel); ok && provider.OpenAI {
			return prepareLegacyAPIGeneration(config, purpose, provider.API.ResolvedModel())
		}
	}

	resolved, err := s.resolveHarnessSelection(ctx, config, core.HarnessGeneration, purpose)
	if err != nil {
		return core.HarnessGenerationBinding{}, "", err
	}

	selection := resolved.View
	if selection.Backend == "openai" {
		return prepareLegacyAPIGeneration(config, purpose, (core.OpenAIQuizModelAdapter{}).ResolvedModel())
	}

	if !selection.Readiness.Ready() {
		if selection.Source == "legacy" {
			return core.HarnessGenerationBinding{}, "", errLegacyGenerationUnavailable
		}

		return core.HarnessGenerationBinding{}, "", errors.New("selected generation profile is not locally ready")
	}

	if selection.Backend != codex.Identity {
		return core.HarnessGenerationBinding{}, "", fmt.Errorf("unsupported generation harness %q", selection.Backend)
	}

	return prepareCodexGenerationBinding(selection, resolved.Executable)
}

func prepareLegacyAPIGeneration(config core.WorkspaceConfig, purpose, model string) (core.HarnessGenerationBinding, string, error) {
	selection := core.HarnessSelection{
		Role: core.HarnessGeneration, Action: purpose, Backend: "openai", Source: "legacy", ModelProvider: "openai", Model: model,
		SelectionRevision: config.Revision(), ProfileRevision: core.HarnessProfileRevision(core.HarnessProfile{ModelProvider: "openai", Model: model}),
		Readiness: core.HarnessReadiness{Supported: true, Executable: "not_required", Credentials: "api_managed", Runtime: "not_required"},
	}
	binding, err := core.SealHarnessGenerationBinding(core.HarnessGenerationBinding{Selection: selection, ProtocolRevision: "openai-responses/v1"})

	return binding, "", err
}

func prepareCodexGenerationBinding(selection core.HarnessSelection, program string) (core.HarnessGenerationBinding, string, error) {
	executable, command, err := core.FreezeHarnessExecutable([]string{program})
	if err != nil {
		return core.HarnessGenerationBinding{}, "", err
	}

	if selection.ModelProvider == "" {
		selection.ModelProvider = "cli_default"
	}

	binding, err := core.SealHarnessGenerationBinding(core.HarnessGenerationBinding{Selection: selection, ExecutableRef: executable, ProtocolRevision: "codex-exec-generation/v1"})

	return binding, command[0], err
}

func (s service) generationAdapters(ctx context.Context) service {
	_, binding, bound := core.HarnessGenerationSelection(ctx)
	if !bound {
		return s
	}

	selection := binding.Selection

	s.quizModel = s.boundQuizAdapter(selection)
	if provider, ok := s.conversationModel.(*core.ConfiguredStructuredModel); ok {
		s.conversationModel = provider.WithGenerationSelection(selection.Backend == "openai", selection.Model, codex.Runner{})
	}

	s.dumpCurator = s.boundDumpCurator(selection)

	return s
}

func (s service) boundQuizAdapter(selection core.HarnessSelection) core.QuizModelAdapter {
	switch adapter := s.quizModel.(type) {
	case core.OpenAIQuizModelAdapter:
		return s.selectedQuizAdapter(selection, adapter)
	case core.ActionRoutingQuizModelAdapter:
		api, ok := adapter.DefaultAdapter.(core.OpenAIQuizModelAdapter)
		if !ok {
			api = core.OpenAIQuizModelAdapter{ProjectDir: s.projectContext.ProjectDir}
		}

		return s.selectedQuizAdapter(selection, api)
	case core.HarnessQuizModelAdapter:
		return s.selectedQuizAdapter(selection, core.OpenAIQuizModelAdapter{ProjectDir: s.projectContext.ProjectDir})
	default:
		// An injected provider retains its transport contract and receives the
		// same owning context; production assembly uses the concrete adapters above.
		return s.quizModel
	}
}

func (s service) selectedQuizAdapter(selection core.HarnessSelection, api core.OpenAIQuizModelAdapter) core.QuizModelAdapter {
	if selection.Backend == "openai" {
		api.Model = selection.Model
		api.ProjectDir = s.projectContext.ProjectDir

		return api
	}

	return core.HarnessQuizModelAdapter{ProjectDir: s.projectContext.ProjectDir, Runner: codex.Runner{}}
}

func (s service) boundDumpCurator(selection core.HarnessSelection) DumpCurationProvider {
	switch provider := s.dumpCurator.(type) {
	case openAIDumpCurationProvider:
		if selection.Backend == "openai" {
			provider.Model = selection.Model

			return provider
		}
	case harnessDumpCurationProvider:
	default:
		return s.dumpCurator
	}

	if selection.Backend == "openai" {
		return openAIDumpCurationProvider{ProjectDir: s.projectContext.ProjectDir, Model: selection.Model}
	}

	return harnessDumpCurationProvider{ProjectDir: s.projectContext.ProjectDir, Runner: codex.Runner{}}
}

func (s service) singlePassPlan(ctx context.Context) bool {
	if _, binding, bound := core.HarnessGenerationSelection(ctx); bound {
		return binding.Selection.Backend != "openai"
	}

	return codexGenerationBackend("CHR_PLAN_GENERATION_BACKEND")
}

func (s service) materializeBoundPlanTracking(ctx context.Context, id string, scope core.QuizScope, plan PlanView, path string) (PlanTrackingResult, error) {
	bound, ctx, err := s.bindGeneration(ctx, "tracking-"+dumpRequestID(id), string(core.QuizModelActionMaterializePlanTracking))
	if err != nil {
		return PlanTrackingResult{}, err
	}

	return bound.materializePlanTracking(ctx, scope, plan, path)
}
