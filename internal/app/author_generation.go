package app

import (
	"context"

	"github.com/urfave/cli/v3"

	authoringcore "forge.adrianpk.com/chr/internal/core/authoring"
	harnesshost "forge.adrianpk.com/chr/internal/core/harness/host"
)

func withContextualQuizGeneration(ctx context.Context, cmd *cli.Command, project commandProjectContext, sessionID string, use func(context.Context, authoringcore.QuizQuestionProvider) error) error {
	provider := contextualQuizProviderFromCommand()
	if _, native := provider.(harnesshost.AuthoringQuizQuestionProvider); !native {
		return use(ctx, provider)
	}

	env, err := newUserEnvironmentFromCommand(cmd)
	if err != nil {
		return err
	}

	operations := newService(ctx, env, project, nil)

	bound, err := operations.bindGeneration(ctx, "contextual-"+sessionID, "contextual_quiz")

	callCtx := bound.context(ctx)
	if native, ok := provider.(harnesshost.AuthoringQuizQuestionProvider); ok && bound.generationRunner != nil {
		native.Runner = bound.generationRunner
		provider = native
	}

	if err != nil {
		return err
	}

	return use(callCtx, provider)
}
