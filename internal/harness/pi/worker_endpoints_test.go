package pi

import (
	"fmt"
	"testing"
)

// Container loopback cannot name a host provider. Only an explicit gateway
// name requests Docker's host-gateway mapping; ordinary endpoints keep routing.
func TestPiWorkerEndpoint(t *testing.T) {
	for _, test := range []struct {
		name, endpoint string
		gateway, valid bool
	}{
		{"localhost", "http://localhost:8080/v1", false, false},
		{"localhost subdomain", "http://model.localhost/v1", false, false},
		{"IPv4 loopback", "http://127.0.0.2/v1", false, false},
		{"IPv6 loopback", "http://[::1]/v1", false, false},
		{"IPv4 mapped loopback", "http://[::ffff:127.0.0.1]/v1", false, false},
		{"unspecified", "http://0.0.0.0/v1", false, false},
		{"gateway", "http://host.docker.internal:8080/v1", true, true},
		{"container address", "http://172.18.0.2:8080/v1", false, true},
		{"remote provider", "https://model.example/v1", false, true},
	} {
		t.Run(test.name, func(t *testing.T) {
			data := []byte(fmt.Sprintf(`{"providers":{"fixture":{"models":[{"baseUrl":%q}]}}}`, test.endpoint))

			gateway, err := workerModelNetwork(data)
			if gateway != test.gateway || (err == nil) != test.valid {
				t.Fatal("wrong explicit worker endpoint boundary", gateway, err)
			}
		})
	}
}
