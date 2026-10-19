package app

import (
	"context"

	"github.com/urfave/cli/v3"

	"forge.adrianpk.com/chr/internal/core"
)

func contextualQuizGenerationProvider(ctx context.Context, cmd *cli.Command, project commandProjectContext, sessionID string) (core.AuthoringQuizQuestionProvider, context.Context, error) {
	provider := contextualQuizProviderFromCommand()
	if _, native := provider.(core.HarnessAuthoringQuizQuestionProvider); !native {
		return provider, ctx, nil
	}

	env, err := newUserEnvironmentFromCommand(cmd)
	if err != nil {
		return provider, ctx, err
	}

	operations := newService(ctx, env, project, nil)
	_, callCtx, err := operations.bindGeneration(ctx, "contextual-"+sessionID, "contextual_quiz")

	return provider, callCtx, err
}
