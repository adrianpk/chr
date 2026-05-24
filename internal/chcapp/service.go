package chcapp

import (
	"context"
	"io/fs"

	"forge.adrianpk.com/chc/internal/chc"
)

type service struct {
	ctx            context.Context
	env            chc.UserEnvironment
	assetsFS       fs.FS
	projectContext commandProjectContext
	dumpEditor     DumpEditorLauncher
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
	}
}
