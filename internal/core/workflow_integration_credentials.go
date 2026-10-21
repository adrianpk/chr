package core

import (
	"context"
	"errors"
	"fmt"
	"os/exec"
	"strconv"
	"strings"
)

// Temporary bare stores do not inherit the source repository's local Git
// configuration. Carry only credential settings in the child environment;
// helper results and secrets are never written to the temporary Git config.
func (reader workflowIntegrationReader) withSourceCredentials(ctx context.Context, source string) (workflowIntegrationReader, error) {
	config := workflowIntegrationReader{dir: source, environment: reader.environment}

	data, err := config.rawOutput(ctx, "config", "--null", "--local", "--get-regexp", `^credential(\..*)?\.(helper|username|usehttppath)$`)
	if err != nil {
		var exit *exec.ExitError
		if errors.As(err, &exit) && exit.ExitCode() == 1 {
			return reader, nil
		}

		return reader, fmt.Errorf("read integration credential context: %w", err)
	}

	if len(data) > 65536 {
		return reader, fmt.Errorf("integration credential configuration exceeds bounds")
	}

	environment, err := workflowCredentialEnvironment(reader.process(ctx, "version").Env, data)
	if err != nil {
		return reader, err
	}

	reader.environment = environment

	return reader, nil
}

func workflowCredentialEnvironment(environment []string, data string) ([]string, error) {
	entries := strings.Split(strings.TrimSuffix(data, "\x00"), "\x00")

	count, err := workflowCredentialEnvironmentCount(environment)
	if err != nil || len(entries) > 64 || count+len(entries) > 256 {
		return nil, fmt.Errorf("integration credential configuration exceeds bounds")
	}

	result := append([]string{}, environment...)

	for _, entry := range entries {
		if entry == "" {
			continue
		}

		key, value, ok := strings.Cut(entry, "\n")
		if !ok {
			return nil, fmt.Errorf("invalid integration credential configuration")
		}

		result = append(result, fmt.Sprintf("GIT_CONFIG_KEY_%d=%s", count, key), fmt.Sprintf("GIT_CONFIG_VALUE_%d=%s", count, value))
		count++
	}

	return append(result, "GIT_CONFIG_COUNT="+strconv.Itoa(count)), nil
}

func workflowCredentialEnvironmentCount(environment []string) (int, error) {
	count := 0

	for _, value := range environment {
		raw, found := strings.CutPrefix(value, "GIT_CONFIG_COUNT=")
		if !found {
			continue
		}

		var err error

		count, err = strconv.Atoi(raw)
		if err != nil || count < 0 || count > 256 {
			return 0, fmt.Errorf("invalid inherited Git configuration count")
		}
	}

	return count, nil
}
