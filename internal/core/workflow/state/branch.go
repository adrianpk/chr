package state

import (
	"strings"
)

// ValidBranch checks the retained Git branch spelling without granting publication.
func ValidBranch(branch string) bool {
	if !validBranchText(branch) {
		return false
	}

	for _, invalid := range []string{"..", "@{", "//"} {
		if strings.Contains(branch, invalid) {
			return false
		}
	}

	return validBranchParts(branch)
}

func validBranchParts(branch string) bool {
	for _, part := range strings.Split(branch, "/") {
		if part == "" || strings.HasPrefix(part, ".") || strings.HasSuffix(part, ".") || strings.HasSuffix(part, ".lock") {
			return false
		}
	}

	return true
}

func validBranchText(branch string) bool {
	return branch != "" && branch != "@" && len(branch) <= 256 && !strings.HasPrefix(branch, "-") && !strings.HasPrefix(branch, "refs/") && !strings.ContainsAny(branch, " ~^:?*[\\\x7f") && !strings.ContainsFunc(branch, func(r rune) bool { return r < 32 })
}
