package pi

import (
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"testing"
)

// Private preparation rejects unsafe inputs without launching a model or exposing paths.
func TestPiGenerationResources(t *testing.T) {
	for _, scenario := range []struct{ name, auth, model, kind string }{
		{"missing", "", "", "missing"},
		{"wrong provider", `{}`, "", ""},
		{"oauth", `{"fixture":{"type":"oauth","access":"FIXTURE_PRIVATE_KEY","refresh":"fixture-refresh","expires":1}}`, "", ""},
		{"credential command", `{"fixture":{"type":"api_key","key":"!touch forbidden"}}`, "", ""},
		{"credential environment", `{"fixture":{"type":"api_key","key":"$FIXTURE_KEY"}}`, "", ""},
		{"credential options", `{"fixture":{"type":"api_key","key":"FIXTURE_PRIVATE_KEY","extra":true}}`, "", ""},
		{"public", validPiAuth, "", "public"},
		{"link", validPiAuth, "", "link"},
		{"pipe", validPiAuth, "", "pipe"},
		{"directory", validPiAuth, "", "directory"},
		{"oversized", strings.Repeat(" ", resourceLimit) + validPiAuth, "", ""},
		{"null model", validPiAuth, `{"providers":{"fixture":null}}`, ""},
		{"missing model", validPiAuth, `{"providers":{}}`, ""},
		{"credential model", validPiAuth, `{"providers":{"fixture":{"apiKey":"FIXTURE_PRIVATE_KEY"}}}`, ""},
		{"credential headers", validPiAuth, `{"providers":{"fixture":{"models":[{"id":"model","headers":{"Authorization":"FIXTURE_PRIVATE_KEY"}}]}}}`, ""},
		{"dynamic model", validPiAuth, `{"providers":{"fixture":{"baseUrl":"$ENDPOINT"}}}`, ""},
		{"endpoint credentials", validPiAuth, `{"providers":{"fixture":{"baseUrl":"https://user:FIXTURE_PRIVATE_KEY@example.invalid/v1"}}}`, ""},
		{"endpoint query", validPiAuth, `{"providers":{"fixture":{"baseUrl":"https://example.invalid/v1?key=FIXTURE_PRIVATE_KEY"}}}`, ""},
		{"unknown provider field", validPiAuth, `{"providers":{"fixture":{"futureOption":true}}}`, ""},
	} {
		t.Run(scenario.name, func(t *testing.T) {
			dir := t.TempDir()
			auth := preparePiResourceFile(t, dir, scenario.auth, scenario.kind)

			var model string
			if scenario.model != "" {
				model = filepath.Join(dir, "models.json")
				writePiFixtureFile(t, model, scenario.model)
			}

			_, err := ResolveResources(dir, auth, model, "fixture")
			if err == nil || strings.Contains(err.Error(), dir) || strings.Contains(err.Error(), "FIXTURE_PRIVATE_KEY") {
				t.Fatal("unsafe resource accepted or exposed private evidence")
			}
		})
	}
}

const validPiAuth = `{"fixture":{"type":"api_key","key":"FIXTURE_PRIVATE_KEY"}}`

func preparePiResourceFile(t *testing.T, dir, auth, kind string) string {
	t.Helper()

	path := filepath.Join(dir, "auth.json")

	var err error

	switch kind {
	case "missing":
		return ""
	case "directory":
		err = os.Mkdir(path, 0o700)
	case "pipe":
		err = syscall.Mkfifo(path, 0o600)
	case "link":
		writePiFixtureFile(t, filepath.Join(dir, "source"), auth)

		err = os.Symlink(filepath.Join(dir, "source"), path)
	default:
		writePiFixtureFile(t, path, auth)

		if kind == "public" {
			err = os.Chmod(path, 0o644)
		}
	}

	if err != nil {
		t.Fatal(err)
	}

	return path
}

// Only the selected credential-free model definition contributes to its frozen identity.
func TestPiGenerationResourceIdentity(t *testing.T) {
	dir := t.TempDir()
	auth, models := filepath.Join(dir, "auth.json"), filepath.Join(dir, "models.json")
	writePiFixtureFile(t, auth, validPiAuth)
	writePiFixtureFile(t, models, `{"providers":{"fixture":{"baseUrl":"http://127.0.0.1:1/v1","api":"openai-completions"},"unrelated":{"apiKey":"unrelated-fixture"}}}`)

	first, err := ResolveResources(dir, "auth.json", "models.json", "fixture")
	if err != nil {
		t.Fatal(err)
	}

	writePiFixtureFile(t, auth, `{"fixture":{"type":"api_key","key":"ROTATED_FIXTURE_KEY"}}`)
	writePiFixtureFile(t, models, `{"providers":{"fixture": {"api":"openai-completions", "baseUrl":"http://127.0.0.1:1/v1"}}}`)

	second, err := ResolveResources(dir, "auth.json", "models.json", "fixture")
	if err != nil || first != second || len(first.ModelConfigHash) != 64 {
		t.Fatal("credential rotation or unrelated profiles changed frozen model semantics", err)
	}
}
