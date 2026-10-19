package pi

import (
	"encoding/json"
	"errors"
	"net"
	"net/url"
	"strings"
)

func workerModelNetwork(data []byte) (bool, error) {
	var definition any
	if json.Unmarshal(data, &definition) != nil {
		return false, ErrCapability
	}

	return workerEndpointFields(definition)
}

func workerEndpointFields(value any) (bool, error) {
	gateway := false

	switch fields := value.(type) {
	case map[string]any:
		return workerEndpointMap(fields)
	case []any:
		for _, child := range fields {
			found, err := workerEndpointFields(child)
			if err != nil {
				return false, err
			}

			gateway = gateway || found
		}
	}

	return gateway, nil
}

func workerEndpoint(value any) (bool, error) {
	text, ok := value.(string)
	if !ok {
		return false, ErrCapability
	}

	parsed, err := url.Parse(text)
	if err != nil {
		return false, ErrCapability
	}

	host := strings.ToLower(parsed.Hostname())

	ip := net.ParseIP(host)
	if host == "localhost" || strings.HasSuffix(host, ".localhost") || (ip != nil && (ip.IsLoopback() || ip.IsUnspecified())) {
		return false, errors.New("pi worker endpoint uses container loopback; configure a container-reachable endpoint such as host.docker.internal")
	}

	return host == "host.docker.internal", nil
}

func workerEndpointMap(fields map[string]any) (bool, error) {
	gateway := false

	for key, child := range fields {
		var (
			found bool
			err   error
		)
		if key == "baseUrl" {
			found, err = workerEndpoint(child)
		} else {
			found, err = workerEndpointFields(child)
		}

		if err != nil {
			return false, err
		}

		gateway = gateway || found
	}

	return gateway, nil
}
