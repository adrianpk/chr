package render

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"testing/fstest"
)

func TestRendererPage(t *testing.T) {
	files := fstest.MapFS{
		"templates/page.html": {
			Data: []byte(`{{define "page.html"}}<h1>{{upper .Title}}</h1>{{end}}`),
		},
	}
	renderer, err := New(files, "templates/*.html")
	if err != nil {
		t.Fatalf("New returned error: %v", err)
	}

	recorder := httptest.NewRecorder()
	err = renderer.Page(recorder, "page.html", map[string]string{"Title": "project"})
	if err != nil {
		t.Fatalf("Page returned error: %v", err)
	}

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusOK)
	}

	if got := recorder.Body.String(); got != "<h1>PROJECT</h1>" {
		t.Fatalf("body = %q, want rendered template", got)
	}
}
