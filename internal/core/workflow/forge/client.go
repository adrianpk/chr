package forge

import (
	"context"
	"strings"
)

// Commands is the bounded command IO consumed by provider protocol operations.
// The host owns deadlines, output limits and child-process supervision.
type Commands interface {
	Output(context.Context, string, ...string) ([]byte, error)
	Input(context.Context, string, []byte, ...string) ([]byte, error)
}

// Endpoint binds an enabled provider command to one resolved repository or PR.
// The host resolves and validates configuration before constructing this value.
type Endpoint struct {
	Repository string
	Provider   string
	Host       string
	Owner      string
	Repo       string
	Number     int
	Command    string
}

// Request describes one provider API operation. A nil body selects output-only IO.
type Request struct {
	Method string
	Path   string
	Body   []byte
}

// Client owns provider protocol and bounded PR observation and creation.
type Client struct {
	Endpoint Endpoint
	Commands Commands
}

// Request executes one resolved API operation without retrying uncertain IO.
func (client Client) Request(ctx context.Context, request Request) ([]byte, error) {
	args := client.Endpoint.Arguments(request)
	if request.Body != nil {
		return client.Commands.Input(ctx, client.Endpoint.Command, request.Body, args...)
	}

	return client.Commands.Output(ctx, client.Endpoint.Command, args...)
}

// RepositoryPath returns the provider API path for this exact owner and repository.
func (endpoint Endpoint) RepositoryPath() string {
	return "/repos/" + endpoint.Owner + "/" + endpoint.Repo
}

// Arguments preserves the distinct Forgejo and GitHub CLI request protocols.
func (endpoint Endpoint) Arguments(request Request) []string {
	if endpoint.Provider == "github" {
		args := []string{"api", "--hostname", endpoint.Host, "-X", request.Method, strings.TrimPrefix(request.Path, "/")}
		if request.Body != nil {
			args = append(args, "--input", "-")
		}

		return args
	}

	root := strings.TrimSuffix(endpoint.Repository, "/"+endpoint.Owner+"/"+endpoint.Repo)

	args := []string{"api", "-X", request.Method, root + "/api/v1" + request.Path}
	if request.Body != nil {
		args = append(args, "-d", "@-")
	}

	return args
}
