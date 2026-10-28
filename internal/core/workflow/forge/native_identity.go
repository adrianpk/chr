package forge

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"forge.adrianpk.com/chr/internal/core/contenthash"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

// NativeIdentity hashes the observed publisher API identity without retaining credentials.
func (client Client) NativeIdentity(ctx context.Context) (string, error) {
	data, err := client.Request(ctx, Request{Method: "GET", Path: "/user"})
	if err != nil || len(data) > workflowstate.FileLimit {
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

	return contenthash.Sum(data), nil
}
