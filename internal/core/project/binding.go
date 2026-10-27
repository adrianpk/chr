package project

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"path/filepath"
)

// ErrRootUnavailable preserves the safe diagnostic for an unavailable registered root.
var ErrRootUnavailable = errors.New("registered project root is unavailable")

// RootBindingDigest binds identity, registered absolute path and resolved root.
// Both notification and conversation targets use these exact canonical bytes.
func RootBindingDigest(registered Project) (string, error) {
	root, err := filepath.Abs(registered.Path)
	if err != nil {
		return "", err
	}

	registeredRoot := root

	root, err = filepath.EvalSymlinks(root)
	if err != nil {
		return "", ErrRootUnavailable
	}

	sum := sha256.Sum256([]byte(registered.ID + "\x00" + registeredRoot + "\x00" + root))

	return hex.EncodeToString(sum[:]), nil
}
