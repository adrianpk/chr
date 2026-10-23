package app

import (
	"path/filepath"

	"forge.adrianpk.com/chr/internal/webui"
)

func webServeProjectContext(config resolvedWebServeConfig) commandProjectContext {
	project := config.ProjectContext
	if project.ProjectID == "" {
		project.ProjectID = projectID(config.ProjectDir)
	}

	if project.ProjectName == "" {
		project.ProjectName = filepath.Base(config.ProjectDir)
	}

	return project
}

func webServeProvider(config resolvedWebServeConfig) webui.DataProvider {
	provider := config.Provider
	if provider == nil && config.Env != nil && config.AssetsFS != nil {
		provider = webViewProvider{env: config.Env, assetsFS: config.AssetsFS, projectContext: webServeProjectContext(config), config: config}
	}

	if view, ok := provider.(webViewProvider); ok && config.guidedHost != nil {
		view.config.guidedHost = config.guidedHost
		provider = view
	}

	return provider
}
