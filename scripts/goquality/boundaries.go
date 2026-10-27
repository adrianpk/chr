package main

import (
	"fmt"
	"path/filepath"
	"sort"
	"strings"
)

const (
	moduleImportPath = "forge.adrianpk.com/chr"
	coreImportPath   = moduleImportPath + "/internal/core/"
)

// Each extracted owner lists its actual module dependencies. Its descendants
// inherit the same bounds; a new directory cannot bypass the owner contract.
var leafImports = map[string]map[string]bool{
	"internal/core/authoring/generation": {coreImportPath + "quiz": true, coreImportPath + "workflow/state": true},
	"internal/core/quiz":                 {coreImportPath + "ops/namespace": true, coreImportPath + "workflow/state": true},
	"internal/core/gate":                 {coreImportPath + "configvalue": true},
	"internal/core/project":              {coreImportPath + "database": true},
	"internal/core/notify": {
		coreImportPath + "database":         true,
		coreImportPath + "contenthash":      true,
		coreImportPath + "ops/namespace":    true,
		coreImportPath + "project":          true,
		coreImportPath + "telegram/message": true,
		coreImportPath + "workflow/state":   true,
	},
	"internal/core/telegram": {
		coreImportPath + "authoring/state":  true,
		coreImportPath + "contenthash":      true,
		coreImportPath + "database":         true,
		coreImportPath + "model":            true,
		coreImportPath + "notify":           true,
		coreImportPath + "ops/namespace":    true,
		coreImportPath + "project":          true,
		coreImportPath + "telegram/message": true,
		coreImportPath + "workflow/state":   true,
		coreImportPath + "workspace":        true,
	},
	"internal/core/telegram/message": {coreImportPath + "contenthash": true, coreImportPath + "ops/namespace": true},
	"internal/core/database":         {},
	"internal/core/model":            {},
	"internal/core/authoring/state":  {},
	"internal/core/workspace":        {},
	"internal/core/configvalue":      {},
	"internal/core/contenthash":      {},
	"internal/core/ops/namespace":    {},
	"internal/core/workflow/state":   {coreImportPath + "contenthash": true},
}

func checkLeafDependencies(importsByFile map[string][]string) []string {
	var failures []string

	for source, imports := range importsByFile {
		owner := dependencyOwner(filepath.ToSlash(filepath.Dir(source)))
		if owner == "" {
			continue
		}

		for _, dependency := range imports {
			local := dependency == moduleImportPath || strings.HasPrefix(dependency, moduleImportPath+"/")
			if !local || leafImports[owner][dependency] {
				continue
			}

			failures = append(failures, fmt.Sprintf("%s: %s cannot import %s", source, owner, dependency))
		}
	}

	sort.Strings(failures)

	return failures
}

func dependencyOwner(directory string) string {
	owner := ""
	for candidate := range leafImports {
		if len(candidate) > len(owner) && (directory == candidate || strings.HasPrefix(directory, candidate+"/")) {
			owner = candidate
		}
	}

	return owner
}
