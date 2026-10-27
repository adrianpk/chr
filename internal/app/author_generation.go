package app

import (
	"context"

	"github.com/urfave/cli/v3"

	"forge.adrianpk.com/chr/internal/core"
	authoringcore "forge.adrianpk.com/chr/internal/core/authoring"
)

func withContextualQuizGeneration(ctx context.Context, cmd *cli.Command, project commandProjectContext, sessionID string, use func(context.Context, authoringcore.AuthoringQuizQuestionProvider) error) error {
	provider := contextualQuizProviderFromCommand()
	if _, native := provider.(core.HarnessAuthoringQuizQuestionProvider); !native {
		return use(ctx, provider)
	}

	env, err := newUserEnvironmentFromCommand(cmd)
	if err != nil {
		return err
	}

	operations := newService(ctx, env, project, nil)

	bound, callCtx, err := operations.bindGeneration(ctx, "contextual-"+sessionID, "contextual_quiz")
	if native, ok := provider.(core.HarnessAuthoringQuizQuestionProvider); ok && bound.generationRunner != nil {
		native.Runner = bound.generationRunner
		provider = native
	}

	if err != nil {
		return err
	}

	return use(callCtx, provider)
}
