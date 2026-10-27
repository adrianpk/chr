package app

import (
	"crypto/sha256"
	"io/fs"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"forge.adrianpk.com/chr/internal/core"
	opscore "forge.adrianpk.com/chr/internal/core/ops"
	"forge.adrianpk.com/chr/internal/webui"
)

// HTTP browsing selects every concrete identity, keeps opening previews bounded,
// escapes document text, and leaves both files and model boundaries untouched.
func TestWebPlanCatalog(t *testing.T) {
	provider, model := webPlanFixture(t)
	handler := webQuizHandler(t, provider)
	refs := seedWebPlanCatalog(t, provider)
	before := webPlanFiles(t, provider)

	catalog, err := provider.CatalogView(t.Context(), "plans")
	if err != nil || len(catalog.Items) != len(refs) {
		t.Fatal(catalog, err)
	}

	response := webPlanGet(handler, "/plans")
	if response.Code != http.StatusOK || strings.Count(response.Body.String(), `class="document-card"`) != len(refs) {
		t.Fatal(response.Body.String())
	}

	checkWebPlanAbsent(t, response.Body.String(), []string{"Artifacts", `class="kv-list"`, "case_slug:", `class="artifact-card"`, `class="summary-grid"`})

	for _, card := range catalog.Items {
		checkWebPlanPreview(t, card)
		checkWebPlanDocument(t, provider, handler, card)
	}

	if !reflect.DeepEqual(before, webPlanFiles(t, provider)) || len(model.requests) != 0 {
		t.Fatal("browsing mutated files or generated content")
	}
}

func checkWebPlanDocument(t *testing.T, provider webViewProvider, handler http.Handler, card webui.ArtifactCardView) {
	t.Helper()

	document, err := core.ReadPlanDocument(t.Context(), provider.projectContext.ProjectDir, card.ID)
	if err != nil {
		t.Fatal(err)
	}

	detail, err := provider.ArtifactDetailView(t.Context(), webui.ArtifactDetailRequest{Family: "plans", ID: card.ID})
	if err != nil || detail.Content != document.Body || detail.ID != card.ID {
		t.Fatal(detail, err)
	}

	for _, suffix := range []string{"", "/fragment"} {
		response := webPlanGet(handler, "/plans/detail"+suffix+"?id="+url.QueryEscape(card.ID))

		body := response.Body.String()
		if response.Code != http.StatusOK || !strings.Contains(body, "END "+card.Title) || strings.Contains(body, "<script>unsafe") {
			t.Fatal(body)
		}

		checkWebPlanAbsent(t, body, []string{"<textarea", `name="content"`, ">Save<", ">Append<", ">Edit<", "case_slug:"})
	}
}

func seedWebPlanCatalog(t *testing.T, provider webViewProvider) []string {
	t.Helper()

	layout, err := opscore.ResolveLayout(opscore.LayoutOptions{ProjectDir: provider.projectContext.ProjectDir, Namespace: "default"})
	if err != nil {
		t.Fatal(err)
	}

	var refs []string

	for _, item := range []struct{ title, body string }{{"Short plan", "Opening short body."}, {"Long plan", strings.Repeat("Opening long body.\n", 150)}, {strings.Repeat("UnbrokenTitle", 12), strings.Repeat("UnbrokenBody", 200)}} {
		artifact, err := opscore.CreateArtifact(opscore.ArtifactOptions{Layout: layout, Type: "plan", Title: item.title, Producer: "test"})
		if err != nil {
			t.Fatal(err)
		}

		body := "---\ntitle: " + item.title + "\ncase_slug: named-work\n---\n\n# " + item.title + "\n\n" + item.body + "\n<script>unsafe</script>\nEND " + item.title
		if err := os.WriteFile(filepath.Join(artifact.Dir, "artifact.md"), []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}

		relative, err := filepath.Rel(provider.projectContext.ProjectDir, filepath.Join(artifact.Dir, "artifact.md"))
		if err != nil {
			t.Fatal(err)
		}

		refs = append(refs, filepath.ToSlash(relative))
	}

	return refs
}

func webPlanGet(handler http.Handler, path string) *httptest.ResponseRecorder {
	request := httptest.NewRequest(http.MethodGet, path, nil)
	request.Header.Set("HX-Request", "true")

	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)

	return response
}

func webPlanFiles(t *testing.T, provider webViewProvider) map[string][32]byte {
	t.Helper()

	files := map[string][32]byte{}

	err := filepath.WalkDir(provider.projectContext.ProjectDir, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if entry.IsDir() {
			return nil
		}

		content, err := os.ReadFile(path)
		if err != nil {
			return err
		}

		files[path] = sha256.Sum256(content)

		return nil
	})
	if err != nil {
		t.Fatal(err)
	}

	return files
}

func checkWebPlanAbsent(t *testing.T, body string, unwanted []string) {
	t.Helper()

	for _, value := range unwanted {
		if strings.Contains(body, value) {
			t.Fatal(value, body)
		}
	}
}

func checkWebPlanPreview(t *testing.T, card webui.ArtifactCardView) {
	t.Helper()

	if len([]rune(card.Preview)) > 1024 || strings.Contains(card.Preview, "title:") || strings.HasPrefix(card.Preview, "# "+card.Title) {
		t.Fatal(card)
	}
}
