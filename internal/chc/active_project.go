package chc

import (
	"errors"
	"fmt"
)

var ErrActiveProjectNotSelected = errors.New("active project is not selected")

type ActiveProject struct {
	ProjectID string
}

func NewActiveProject(projectID string) (ActiveProject, error) {
	if projectID == "" {
		return ActiveProject{}, fmt.Errorf("active project id is required")
	}

	return ActiveProject{ProjectID: projectID}, nil
}
