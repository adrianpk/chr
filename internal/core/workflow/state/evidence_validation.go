package state

import (
	"fmt"
	"path/filepath"
	"strings"
)

// CheckEvidenceRef checks a retained reference against its invocation namespace and hash spelling.
func CheckEvidenceRef(id string, ref EvidenceRef) error {
	if !ValidRequestKey(id) || filepath.Base(id) != id || strings.ContainsAny(ref.Path, "\\\x00") || !filepath.IsLocal(ref.Path) || filepath.ToSlash(filepath.Clean(ref.Path)) != ref.Path || !ValidFingerprint(ref.Hash) {
		return ErrConflict
	}

	root := "run/" + id + "/"
	if !strings.HasPrefix(ref.Path, root+"evidence/") && !strings.HasPrefix(ref.Path, root+"extension/chr/") {
		return fmt.Errorf("%w: evidence belongs to another invocation", ErrConflict)
	}

	return nil
}
