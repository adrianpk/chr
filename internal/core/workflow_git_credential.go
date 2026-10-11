package core

import (
	"context"
	"encoding/base64"
	"fmt"
	"strings"
)

func workflowPublicationCredential(ctx context.Context, directory, endpoint string) (string, error) {
	command := workflowPublicationCommand(ctx, "git", "--no-replace-objects", "-C", directory, "-c", "credential.interactive=false", "credential", "fill")
	output := &workflowPublicationBuffer{limit: workflowPublicationTextLimit}
	command.Stdout = output

	command.Stdin = strings.NewReader("url=" + endpoint + "\n\n")
	if err := command.Run(); err != nil {
		return "", fmt.Errorf("host Git credential lookup failed")
	}

	fields := make(map[string]string)

	for _, line := range strings.Split(output.String(), "\n") {
		key, value, found := strings.Cut(line, "=")
		if found {
			fields[key] = value
		}
	}

	return workflowPublicationAuthorization(fields)
}

func workflowPublicationAuthorization(fields map[string]string) (string, error) {
	if fields["authtype"] != "" && fields["credential"] != "" {
		kind, credential := fields["authtype"], fields["credential"]
		if strings.ContainsAny(kind, " \t\r\n\x00") || strings.ContainsAny(credential, "\r\n\x00") {
			return "", fmt.Errorf("invalid host authorization fields")
		}

		return kind + " " + credential, nil
	}

	if fields["username"] == "" || fields["password"] == "" {
		return "", fmt.Errorf("host Git credential lookup returned no supported authorization")
	}

	return "Basic " + base64.StdEncoding.EncodeToString([]byte(fields["username"]+":"+fields["password"])), nil
}
