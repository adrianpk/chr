package app

import (
	"context"
	"fmt"
	"strings"
	"time"

	"forge.adrianpk.com/chr/internal/core"
	telegramcore "forge.adrianpk.com/chr/internal/core/telegram"
)

func (r telegramReceiver) respondHarness(ctx context.Context, user telegramcore.AllowedUser, config telegramcore.ReceiverConfig, update telegramcore.ReceivedUpdate) (bool, error) {
	parsedHarness, err := parseTelegramHarness(update)

	choices, recognized := parsedHarness.Choices, parsedHarness.Recognized
	if !recognized {
		return false, nil
	}

	if err != nil {
		return true, r.recordDiagnostic(ctx, update, config, telegramHarnessUsage+"\nUse '-' to clear a role. Select the project with /project first.")
	}

	operations, ok := r.operations.(telegramHarnessService)
	if !ok {
		return true, r.recordDiagnostic(ctx, update, config, "Harness selection is unavailable.")
	}

	projects, err := r.operations.ListRegisteredProjects(user.ProjectIDs)
	if err != nil {
		return true, err
	}

	input := telegramcore.ConversationInput{Update: update, Config: config, Projects: projects}

	target, err := r.store.ResolveHarnessTarget(ctx, input, time.Now())
	if err != nil {
		return true, r.recordDiagnostic(ctx, update, config, "Select an authorized current project with /project, then send /harness as a new command.")
	}

	if choices.Worker != nil || choices.Generation != nil {
		return true, r.prepareHarnessChange(ctx, operations, user, input, target, choices)
	}

	status, err := operations.InspectTelegramHarnessSelections(ctx, target)
	if err != nil {
		return true, r.recordDiagnostic(ctx, update, config, "Harness choices could not be loaded. Inspect private profile setup locally.")
	}

	return true, r.store.QueuePresentationResponse(ctx, update, *telegramOutcomeMessage("38", telegramHarnessText(status)), config.Digest(), time.Now())
}

func telegramHarnessText(status telegramHarnessStatus) string {
	s := status.Selections

	var text strings.Builder
	fmt.Fprintf(&text, "Harness defaults for %s\n\nNew operations only; running, paused and resumed work keeps its binding.\n", s.Snapshot.ProjectRef)
	text.WriteString(telegramHarnessRole("Worker", s.Worker, false))
	text.WriteString(telegramHarnessRole("Generation", s.Generation, true))
	text.WriteString("Named profiles (local readiness; remote authentication not tested):\n")

	for _, p := range s.Profiles[:min(8, len(s.Profiles))] {
		fmt.Fprintf(&text, "- %s · worker %s; generation %s\n", telegramHarnessLabel(p.Name, p.Harness, p.ModelProvider, p.Model), telegramHarnessReady(p.Worker), telegramHarnessReady(p.Generation))
	}

	if len(s.Profiles) > 8 {
		text.WriteString("Showing the first eight profiles; inspect the remaining configured names locally.\n")
	}

	text.WriteString("Retained worker bindings:\n")

	for _, run := range status.Retained {
		label := "Legacy worker (original settings)"
		if b := run.HarnessBinding; b != nil {
			label = telegramHarnessLabel(b.Profile, b.Harness, b.ModelProvider, b.Model)
		}

		fmt.Fprintf(&text, "- %s · %s · %s\n", run.ID, run.Status, label)
	}

	if status.More {
		text.WriteString("Showing the first four invocations; inspect workflow status for more.\n")
	}

	text.WriteString(status.RetainedProblem)
	text.WriteString("\n" + telegramHarnessUsage + "\n'-' restores legacy dispatch; omitted roles stay unchanged. Private credentials and runtime setup remain local.")

	return telegramcore.BoundReceiverText(text.String())
}

func telegramHarnessRole(role string, selection core.HarnessSelection, generation bool) string {
	label := telegramHarnessLabel(selection.Profile, selection.Harness, selection.ModelProvider, selection.Model)
	if selection.Profile == "" {
		label = "Legacy dispatch · Codex CLI default"
		if generation {
			label = "Legacy dispatch · varies by operation"
		}
	}

	return fmt.Sprintf("%s: %s · local %s\n", role, label, telegramHarnessReady(selection.Readiness))
}

func telegramHarnessReady(readiness core.HarnessReadiness) string {
	if readiness.Ready() {
		return "ready"
	}

	return "not ready"
}

func telegramHarnessLabel(profile, identity, provider, model string) string {
	identity = map[string]string{"codex": "Codex", "pi": "Pi"}[identity]

	if model == "cli_default" {
		model = "CLI default"
	}

	if provider != "" && provider != "cli_default" {
		model = provider + "/" + model
	}

	return strings.Trim(strings.Join([]string{profile, identity, model}, " · "), " ·")
}
