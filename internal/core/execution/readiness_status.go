package execution

import (
	"strings"
)

func workReadinessStateForStatus(status []string) WorktreeState {
	if len(status) > 0 {
		return WorktreeStateDirty
	}

	return WorktreeStateClean
}

type worktreeStatusChanges struct {
	runtimeArtifacts []string
	userChanges      []string
}

func splitWorktreeStatus(lines []string) worktreeStatusChanges {
	runtimeArtifacts := make([]string, 0)
	userChanges := make([]string, 0)

	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			continue
		}

		if worktreeStatusLineIsRuntimeArtifact(line) {
			runtimeArtifacts = append(runtimeArtifacts, line)

			continue
		}

		userChanges = append(userChanges, line)
	}

	return worktreeStatusChanges{runtimeArtifacts: runtimeArtifacts, userChanges: userChanges}
}

func worktreeStatusLineIsRuntimeArtifact(line string) bool {
	return strings.Contains(line, ".chr/default/runtime/") || strings.Contains(line, "ops/default/run/")
}

func runtimeArtifactIgnorePatterns() []string {
	return []string{".chr/default/runtime/", "ops/*/run/"}
}
