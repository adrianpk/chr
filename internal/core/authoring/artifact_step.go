package authoring

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func writeAuthoringArtifactStep(layout AuthoringArtifactLayout, content, description string) error {
	projectDir := strings.TrimSuffix(layout.FullPath, string(os.PathSeparator)+filepath.FromSlash(layout.Ref.Path))
	if err := EnsureAuthoringArtifactMetadata(projectDir, layout.Ref); err != nil {
		return err
	}

	if err := os.MkdirAll(layout.Dir, 0o755); err != nil {
		return fmt.Errorf("create %s directory: %w", description, err)
	}

	if err := os.WriteFile(layout.FullPath, []byte(content), 0o644); err != nil {
		return fmt.Errorf("write %s artifact: %w", description, err)
	}

	return nil
}
