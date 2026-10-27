package project

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// ErrActiveNotSelected identifies a package boundary error.
var ErrActiveNotSelected = errors.New("active project is not selected")

// Active describes a package boundary value.
type Active struct {
	ProjectID string
}

// ActiveStore defines a package boundary interface.
type ActiveStore interface {
	LoadActiveProject() (Active, error)
	SetActiveProject(activeProject Active) error
	ClearActiveProject() error
}

type activeStore struct {
	path string
}

// NewActiveStore constructs the corresponding package service or value.
func NewActiveStore(env ConfigDirectory) ActiveStore {
	return &activeStore{
		path: filepath.Join(env.ConfigDir(), "active-project"),
	}
}

// NewActive constructs the corresponding package service or value.
func NewActive(projectID string) (Active, error) {
	if projectID == "" {
		return Active{}, fmt.Errorf("active project id is required")
	}

	return Active{ProjectID: projectID}, nil
}

// LoadActiveProject loads the requested data.
func (s *activeStore) LoadActiveProject() (Active, error) {
	content, err := os.ReadFile(s.path)
	if errors.Is(err, os.ErrNotExist) {
		return Active{}, ErrActiveNotSelected
	}

	if err != nil {
		return Active{}, fmt.Errorf("load active project: %w", err)
	}

	activeProject, err := NewActive(strings.TrimSpace(string(content)))
	if err != nil {
		return Active{}, fmt.Errorf("load active project: %w", err)
	}

	return activeProject, nil
}

// SetActiveProject executes the corresponding package operation.
func (s *activeStore) SetActiveProject(activeProject Active) error {
	_, err := NewActive(activeProject.ProjectID)
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

// ClearActiveProject executes the corresponding package operation.
func (s *activeStore) ClearActiveProject() error {
	err := os.Remove(s.path)
	if errors.Is(err, os.ErrNotExist) {
		return nil
	} else if err != nil {
		return fmt.Errorf("clear active project: %w", err)
	}

	return nil
}

// ConfigDirectory supplies the active-selection configuration directory.
type ConfigDirectory interface{ ConfigDir() string }
