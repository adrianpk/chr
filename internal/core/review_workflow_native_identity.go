package core

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
)

func (adapter RemoteReviewAdapter) readWorkflowNativeIdentity(ctx context.Context, pr WorkflowPullRequestSnapshot) (string, error) {
	reference, config, err := adapter.resolve(pr.URL)
	if err != nil {
		return "", err
	}

	data, err := adapter.workflowProviderRunner().Output(ctx, config.CLI, workflowRemoteAPIArgs(pr.Repository, reference, "GET", "/user", false)...)
	if err != nil || len(data) > WorkflowFileLimit {
		return "", fmt.Errorf("current publisher API identity could not be observed")
	}

	var identity struct {
		ID    int64  `json:"id"`
		Login string `json:"login"`
	}
	if err := json.Unmarshal(data, &identity); err != nil || identity.ID < 1 || identity.Login == "" || len(identity.Login) > 256 || strings.ContainsAny(identity.Login, "\x00\r\n") {
		return "", fmt.Errorf("current publisher API identity is incomplete")
	}

	data, err = json.Marshal(identity)
	if err != nil {
		return "", err
	}

	return workflowHash(data), nil
}
