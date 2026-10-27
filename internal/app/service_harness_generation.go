package app

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"

	"forge.adrianpk.com/chr/internal/core"
	harnesscore "forge.adrianpk.com/chr/internal/core/harness"
	harnesshost "forge.adrianpk.com/chr/internal/core/harness/host"
	quizcore "forge.adrianpk.com/chr/internal/core/quiz"
	quizmodel "forge.adrianpk.com/chr/internal/core/quiz/model"
	openaiquiz "forge.adrianpk.com/chr/internal/core/quiz/openai"
	workspaceconfig "forge.adrianpk.com/chr/internal/core/workspace/config"
	"forge.adrianpk.com/chr/internal/harness/codex"
	"forge.adrianpk.com/chr/internal/harness/pi"
)

type generationOperationBinder func(service, context.Context, string, string) (*harnesshost.GenerationOperation, error)

// boundGeneration holds selected adapters and the admitted operation value.
// Callers derive their call context from their own parent; no operation context is retained.
type boundGeneration struct {
	service
	operation *harnesshost.GenerationOperation
}

func (b boundGeneration) context(parent context.Context) context.Context {
	if b.operation == nil {
		return parent
	}

	return harnesshost.WithGeneration(parent, *b.operation)
}

var errLegacyGenerationUnavailable = errors.New("legacy generation executable is unavailable")

func (s service) bindGeneration(ctx context.Context, id, purpose string) (boundGeneration, error) {
	if s.generationBinder == nil {
		return boundGeneration{service: s}, errors.New("generation operation binder is required")
	}

	operation, err := s.generationBinder(s, ctx, id, purpose)
	if err != nil {
		return boundGeneration{service: s}, err
	}

	bound, err := s.generationAdapters(operation)

	return boundGeneration{service: bound, operation: operation}, err
}

func (s service) bindHarnessGeneration(ctx context.Context, id, purpose string) (*harnesshost.GenerationOperation, error) {
	if id == "" {
		id = rand.Text()
	}

	root := s.generationArtifactsDir
	if root == "" {
		root = s.projectContext.ProjectDir
	}

	location := harnesshost.GenerationLocation{ConfigDir: s.projectContext.ProjectDir, ExecutionDir: s.generationExecutionDir, ProjectRef: s.projectContext.ProjectID}
	if location.ExecutionDir == "" {
		location.ExecutionDir = s.projectContext.ProjectDir
	}

	operation, err := core.FreezeHarnessGeneration(ctx, harnesshost.GenerationFreezeInput{Project: root, ID: id, Purpose: purpose, Locations: []harnesshost.GenerationLocation{location}}, func(ctx context.Context, config workspaceconfig.Config) (harnesshost.GenerationPreparation, error) {
		return s.prepareGenerationBinding(ctx, config, purpose)
	})
	if err != nil {
		return nil, err
	}

	return &operation, nil
}

func (s service) prepareGenerationBinding(ctx context.Context, config workspaceconfig.Config, purpose string) (harnesshost.GenerationPreparation, error) {
	if model, ok := s.legacyConversationModel(config, purpose); ok {
		return prepareLegacyAPIGeneration(config, purpose, model)
	}

	resolved, err := s.resolveHarnessSelection(ctx, config, harnesscore.Generation, purpose)
	if err != nil {
		return harnesshost.GenerationPreparation{}, err
	}

	selection := resolved.View
	if selection.Backend == "openai" {
		return prepareLegacyAPIGeneration(config, purpose, (openaiquiz.Adapter{Host: harnesshost.QuizGeneration{}}).ResolvedModel())
	}

	if !selection.Readiness.Ready() {
		if selection.Source == "legacy" {
			return harnesshost.GenerationPreparation{}, errLegacyGenerationUnavailable
		}

		if selection.Readiness.Problem != "" {
			return harnesshost.GenerationPreparation{}, fmt.Errorf("selected generation profile is not locally ready: %s", selection.Readiness.Problem)
		}

		return harnesshost.GenerationPreparation{}, errors.New("selected generation profile is not locally ready")
	}

	if selection.Backend == pi.Identity {
		return preparePiGenerationBinding(selection, resolved, s.env.ConfigDir())
	}

	if selection.Backend != codex.Identity {
		return harnesshost.GenerationPreparation{}, fmt.Errorf("unsupported generation harness %q", selection.Backend)
	}

	return prepareCodexGenerationBinding(selection, resolved.Executable)
}

func (s service) legacyConversationModel(config workspaceconfig.Config, purpose string) (string, bool) {
	if config.Harnesses.Generation != "" || purpose != "conversation" {
		return "", false
	}

	provider, ok := s.conversationModel.(*core.ConfiguredStructuredModel)
	if !ok || !provider.OpenAI {
		return "", false
	}

	return provider.API.ResolvedModel(), true
}

func prepareLegacyAPIGeneration(config workspaceconfig.Config, purpose, model string) (harnesshost.GenerationPreparation, error) {
	selection := harnesscore.Selection{
		Role: harnesscore.Generation, Action: purpose, Backend: "openai", Source: "legacy", ModelProvider: "openai", Model: model,
		SelectionRevision: config.Revision(), ProfileRevision: harnesscore.ProfileRevision(harnesscore.Profile{ModelProvider: "openai", Model: model}),
		Readiness: harnesscore.Readiness{Supported: true, Executable: "not_required", Credentials: "api_managed", Runtime: "not_required"},
	}
	binding, err := harnesscore.SealGenerationBinding(harnesscore.GenerationBinding{Selection: selection, ProtocolRevision: "openai-responses/v1"})

	return harnesshost.GenerationPreparation{Binding: binding}, err
}

func prepareCodexGenerationBinding(selection harnesscore.Selection, program string) (harnesshost.GenerationPreparation, error) {
	executable, err := harnesshost.FreezeExecutable([]string{program})
	if err != nil {
		return harnesshost.GenerationPreparation{}, err
	}

	if selection.ModelProvider == "" {
		selection.ModelProvider = "cli_default"
	}

	binding, err := harnesscore.SealGenerationBinding(harnesscore.GenerationBinding{Selection: selection, ExecutableRef: executable.Reference, ProtocolRevision: "codex-exec-generation/v1"})

	return harnesshost.GenerationPreparation{Binding: binding, Launch: harnesshost.GenerationLaunch{Program: executable.Command[0]}}, err
}

func (s service) generationAdapters(generation *harnesshost.GenerationOperation) (service, error) {
	if generation == nil {
		return s, nil
	}

	selection := generation.Binding.Selection
	switch selection.Backend {
	case pi.Identity:
		if generation.Binding.ProtocolRevision != pi.ProtocolRevision {
			return s, errors.New("unsupported retained Pi generation protocol")
		}

		s.generationRunner = pi.Runner{}
	case codex.Identity:
		s.generationRunner = codex.Runner{}
	case "openai":
	default:
		return s, errors.New("unsupported retained generation harness")
	}

	s.quizModel = s.boundQuizAdapter(selection)
	if provider, ok := s.conversationModel.(*core.ConfiguredStructuredModel); ok {
		s.conversationModel = provider.WithGenerationSelection(selection.Backend == "openai", selection.Model, s.generationRunner)
	}

	s.dumpCurator = s.boundDumpCurator(selection)

	return s, nil
}

func (s service) boundQuizAdapter(selection harnesscore.Selection) quizmodel.Adapter {
	switch adapter := s.quizModel.(type) {
	case openaiquiz.Adapter:
		return s.selectedQuizAdapter(selection, adapter)
	case core.ActionRoutingQuizModelAdapter:
		api, ok := adapter.DefaultAdapter.(openaiquiz.Adapter)
		if !ok {
			api = openaiquiz.Adapter{Host: harnesshost.QuizGeneration{}, ProjectDir: s.projectContext.ProjectDir}
		}

		return s.selectedQuizAdapter(selection, api)
	case core.HarnessQuizModelAdapter:
		return s.selectedQuizAdapter(selection, openaiquiz.Adapter{Host: harnesshost.QuizGeneration{}, ProjectDir: s.projectContext.ProjectDir})
	default:
		// An injected provider retains its transport contract and receives the
		// same owning context; production assembly uses the concrete adapters above.
		return s.quizModel
	}
}

func (s service) selectedQuizAdapter(selection harnesscore.Selection, api openaiquiz.Adapter) quizmodel.Adapter {
	if selection.Backend == "openai" {
		api.Model = selection.Model
		api.ProjectDir = s.projectContext.ProjectDir

		return api
	}

	return core.HarnessQuizModelAdapter{ProjectDir: s.projectContext.ProjectDir, Runner: s.generationRunner}
}

func (s service) boundDumpCurator(selection harnesscore.Selection) DumpCurationProvider {
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

	return harnessDumpCurationProvider{ProjectDir: s.projectContext.ProjectDir, Runner: s.generationRunner}
}

func preparePiGenerationBinding(selection harnesscore.Selection, resolved resolvedHarnessSelection, configDir string) (harnesshost.GenerationPreparation, error) {
	resources, err := pi.ResolveResources(configDir, resolved.Profile.CredentialRef, resolved.Profile.ModelConfigRef, selection.ModelProvider)
	if err != nil {
		return harnesshost.GenerationPreparation{}, err
	}

	executable, err := harnesshost.FreezeExecutable([]string{resolved.Executable})
	if err != nil {
		return harnesshost.GenerationPreparation{}, err
	}

	binding, err := harnesscore.SealGenerationBinding(harnesscore.GenerationBinding{Selection: selection, ExecutableRef: executable.Reference, ProtocolRevision: pi.ProtocolRevision})

	return harnesshost.GenerationPreparation{Binding: binding, Launch: harnesshost.GenerationLaunch{Program: executable.Command[0], Resources: resources}}, err
}

func (s service) singlePassPlan(ctx context.Context) bool {
	if generation, bound := harnesshost.GenerationSelection(ctx); bound {
		return generation.Binding.Selection.Backend != "openai"
	}

	return codexGenerationBackend("CHR_PLAN_GENERATION_BACKEND")
}

func (s service) materializeBoundPlanTracking(ctx context.Context, id string, scope quizcore.Scope, plan PlanView, path string) (PlanTrackingResult, error) {
	bound, err := s.bindGeneration(ctx, "tracking-"+dumpRequestID(id), string(quizmodel.ActionMaterializePlanTracking))
	ctx = bound.context(ctx)

	if err != nil {
		return PlanTrackingResult{}, err
	}

	return bound.materializePlanTracking(ctx, scope, plan, path)
}
