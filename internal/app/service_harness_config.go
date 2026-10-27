package app

import (
	"context"
	"errors"
	"sort"

	"forge.adrianpk.com/chr/internal/core"
	harnesscore "forge.adrianpk.com/chr/internal/core/harness"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
	workspacecore "forge.adrianpk.com/chr/internal/core/workspace"
)

// HarnessProfileStatus offers only named choices and local readiness evidence.
type HarnessProfileStatus struct {
	Name          string                `json:"name"`
	Harness       string                `json:"harness"`
	ModelProvider string                `json:"model_provider,omitempty"`
	Model         string                `json:"model"`
	Thinking      string                `json:"thinking,omitempty"`
	Revision      string                `json:"revision"`
	Worker        harnesscore.Readiness `json:"worker"`
	Generation    harnesscore.Readiness `json:"generation"`
}

// HarnessSelectionStatus presents defaults for new operations, not retained runs.
type HarnessSelectionStatus struct {
	Snapshot   core.HarnessSelectionSnapshot `json:"snapshot"`
	Worker     harnesscore.Selection         `json:"worker"`
	Generation harnesscore.Selection         `json:"generation"`
	Profiles   []HarnessProfileStatus        `json:"profiles"`
	Problems   []string                      `json:"problems,omitempty"`
}

// InspectHarnessSelections reads both roles from one persisted configuration.
// Generation's action is explicit because omitted selection preserves per-action dispatch.
func (s service) InspectHarnessSelections(ctx context.Context, action string) (HarnessSelectionStatus, error) {
	var result HarnessSelectionStatus
	if err := ctx.Err(); err != nil {
		return result, err
	}

	config, err := core.LoadWorkspaceConfig(s.projectContext.ProjectDir)
	if err != nil {
		return result, err
	}

	result.Snapshot, err = config.HarnessSelections(s.projectContext.ProjectID)
	if err != nil {
		return result, err
	}

	if s.env == nil {
		return result, errors.New("private user configuration is required to inspect harness choices")
	}

	profiles, err := harnesscore.LoadProfiles(s.env.ConfigDir())
	if err != nil {
		return result, err
	}

	for _, field := range []struct {
		role         harnesscore.Role
		action, name string
		target       *harnesscore.Selection
	}{
		{harnesscore.Worker, "", config.Harnesses.Worker, &result.Worker}, {harnesscore.Generation, action, config.Harnesses.Generation, &result.Generation},
	} {
		resolved, err := resolveConfiguredHarnessSelection(ctx, harnessSelectionInput{Config: config, Name: field.name, Role: field.role, Action: field.action, ConfigDir: s.env.ConfigDir()}, profiles)
		*field.target = resolved.View

		if err != nil {
			result.Problems = append(result.Problems, err.Error())
		}
	}

	result.Profiles = harnessProfileChoices(ctx, config, profiles, s.env.ConfigDir())

	return result, nil
}

func harnessProfileChoices(ctx context.Context, config core.WorkspaceConfig, profiles map[string]harnesscore.Profile, configDir string) []HarnessProfileStatus {
	choices := make([]HarnessProfileStatus, 0, len(profiles))
	for _, profile := range profiles {
		worker := prepareHarnessSelection(ctx, harnessSelectionInput{Config: config, Role: harnesscore.Worker, ConfigDir: configDir}, profile)
		generation := prepareHarnessSelection(ctx, harnessSelectionInput{Config: config, Role: harnesscore.Generation, ConfigDir: configDir}, profile)
		choices = append(choices, HarnessProfileStatus{Name: profile.Name, Harness: profile.Harness, ModelProvider: profile.ModelProvider, Model: worker.View.Model, Thinking: profile.Thinking, Revision: profile.Revision, Worker: worker.View.Readiness, Generation: generation.View.Readiness})
	}

	sort.Slice(choices, func(i, j int) bool { return choices[i].Name < choices[j].Name })

	return choices
}

// ChangeHarnessSelections is the shared authorized configuration mutation. The
// trusted adapter checks its current actor/project authority even on receipt replay.
func (s service) ChangeHarnessSelections(ctx context.Context, request core.HarnessSelectionChangeRequest, authorize func(context.Context) error) (core.HarnessSelectionSnapshot, error) {
	if request.ProjectRef != s.projectContext.ProjectID {
		return core.HarnessSelectionSnapshot{}, workflowstate.ErrConflict
	}

	return core.ChangeHarnessSelections(ctx, s.projectContext.ProjectDir, core.HarnessSelectionChangeOptions{
		Request: request, Authorize: authorize,
		Validate: func(ctx context.Context, config core.WorkspaceConfig) error {
			return s.validateHarnessSelectionChange(ctx, config, request)
		},
	})
}

func (s service) validateHarnessSelectionChange(ctx context.Context, config core.WorkspaceConfig, request core.HarnessSelectionChangeRequest) error {
	for _, field := range []struct {
		role   harnesscore.Role
		change *string
	}{{harnesscore.Worker, request.Worker}, {harnesscore.Generation, request.Generation}} {
		if field.change == nil || *field.change == "" {
			continue
		}

		resolved, err := s.resolveHarnessSelection(ctx, config, field.role, "spec_generation")
		if err != nil {
			return err
		}

		if !resolved.View.Readiness.Ready() {
			return errors.New("selected harness is not locally ready")
		}
	}

	return nil
}

func localHarnessSelectionAuthority(project, actor string) func(context.Context) error {
	return func(ctx context.Context) error {
		if err := ctx.Err(); err != nil {
			return err
		}

		userspace, err := workspacecore.ResolveUserspace(workspacecore.ResolveUserspaceOptions{ProjectDir: project})
		if err != nil {
			return err
		}

		if userspace.OperatorID != actor {
			return errors.New("harness selection actor is not the active local operator")
		}

		return nil
	}
}
