package app

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"forge.adrianpk.com/chr/internal/core"
)

func (s service) appendPlanGenerationLog(event string, fields map[string]string) {
	entry := map[string]string{
		"time":        s.utcNow().Format("2006-01-02T15:04:05.000000000Z07:00"),
		"event":       event,
		"project_dir": s.projectContext.ProjectDir,
		"project_id":  s.projectContext.ProjectID,
	}

	for key, value := range fields {
		entry[key] = value
	}

	content, err := json.Marshal(entry)
	if err != nil {
		return
	}

	path := core.UserPath(s.projectContext.ProjectDir, "runtime", "plan-generation.log")

	err = os.MkdirAll(filepath.Dir(path), 0o755)
	if err != nil {
		return
	}

	file, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return
	}

	defer func() {
		closeErr := file.Close()
		if closeErr != nil {
			return
		}
	}()

	_, err = fmt.Fprintln(file, string(content))
	if err != nil {
		return
	}
}
