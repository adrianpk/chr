package workflow

import (
	"crypto/sha256"
	"encoding/hex"
	"io"
)

func workflowMergerFileHash(reader io.Reader) (string, error) {
	hash := sha256.New()
	if _, err := io.Copy(hash, reader); err != nil {
		return "", err
	}

	return hex.EncodeToString(hash.Sum(nil)), nil
}
