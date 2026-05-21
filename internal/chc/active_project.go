package chc

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

var ErrActiveProjectNotSelected = errors.New("active project is not selected")

type ActiveProject struct {
	ProjectID string
}

type ActiveProjectStore interface {
	LoadActiveProject() (ActiveProject, error)
	SetActiveProject(activeProject ActiveProject) error
	ClearActiveProject() error
}

type activeProjectStore struct {
	path string
}

func NewActiveProjectStore(env UserEnvironment) ActiveProjectStore {
	return &activeProjectStore{
		path: filepath.Join(env.ConfigDir(), "active-project"),
	}
}

func NewActiveProject(projectID string) (ActiveProject, error) {
	if projectID == "" {
		return ActiveProject{}, fmt.Errorf("active project id is required")
	}

	return ActiveProject{ProjectID: projectID}, nil
}

func (s *activeProjectStore) LoadActiveProject() (ActiveProject, error) {
	content, err := os.ReadFile(s.path)
	if errors.Is(err, os.ErrNotExist) {
		return ActiveProject{}, ErrActiveProjectNotSelected
	}

	if err != nil {
		return ActiveProject{}, fmt.Errorf("load active project: %w", err)
	}

	activeProject, err := NewActiveProject(strings.TrimSpace(string(content)))
	if err != nil {
		return ActiveProject{}, fmt.Errorf("load active project: %w", err)
	}

	return activeProject, nil
}

func (s *activeProjectStore) SetActiveProject(activeProject ActiveProject) error {
	_, err := NewActiveProject(activeProject.ProjectID)
	if err != nil {
		return err
	}

	err = os.MkdirAll(filepath.Dir(s.path), 0o755)
	if err != nil {
		return fmt.Errorf("create active project config directory: %w", err)
	}

	tmpPath := s.path + ".tmp"
	content := []byte(activeProject.ProjectID + "\n")

	err = os.WriteFile(tmpPath, content, 0o644)
	if err != nil {
		return fmt.Errorf("write active project: %w", err)
	}

	err = os.Rename(tmpPath, s.path)
	if err != nil {
		return fmt.Errorf("replace active project: %w", err)
	}

	return nil
}

func (s *activeProjectStore) ClearActiveProject() error {
	err := os.Remove(s.path)
	if errors.Is(err, os.ErrNotExist) {
		return nil
	} else if err != nil {
		return fmt.Errorf("clear active project: %w", err)
	}

	return nil
}
