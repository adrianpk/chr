package app

import "testing"

// Independent control journeys own separate roots, credentials, stores and HTTP
// servers. Bound their overlap so race instrumentation can finish the package
// deadline without competing with the serial shared-capacity certification.
var telegramControlFixtureSlots = make(chan struct{}, 4)

func telegramControlParallel(t *testing.T) {
	t.Helper()
	t.Parallel()

	telegramControlFixtureSlots <- struct{}{}

	t.Cleanup(func() { <-telegramControlFixtureSlots })
}
