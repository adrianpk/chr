package chcapp

import (
	"context"
	"io/fs"
	"time"

	"forge.adrianpk.com/chc/internal/chc"
)

type service struct {
	ctx            context.Context
	env            chc.UserEnvironment
	assetsFS       fs.FS
	projectContext commandProjectContext
	dumpEditor     DumpEditorLauncher
	quizModel      chc.QuizModelAdapter
	now            func() time.Time
}

func newService(
	ctx context.Context,
	env chc.UserEnvironment,
	projectContext commandProjectContext,
) service {
	return service{
		ctx:            ctx,
		env:            env,
		assetsFS:       assetsFS,
		projectContext: projectContext,
		dumpEditor:     defaultDumpEditorLauncher{},
		now:            time.Now,
	}
}
