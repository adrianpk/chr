// Package render provides server-side template rendering helpers for the Web UI.
package render

import (
	"fmt"
	"html/template"
	"io/fs"
	"net/http"
	"strings"

	"forge.adrianpk.com/chr/internal/webui/htmx"
	"forge.adrianpk.com/chr/internal/webui/ui"
)

// Renderer executes parsed Web UI templates.
type Renderer struct {
	templates *template.Template
}

// New parses templates from an FS.
func New(files fs.FS, patterns ...string) (Renderer, error) {
	tmpl, err := template.New("").Funcs(FuncMap()).ParseFS(files, patterns...)
	if err != nil {
		return Renderer{}, fmt.Errorf("parse web templates: %w", err)
	}

	return Renderer{templates: tmpl}, nil
}

// Page renders a full HTML page.
func (r Renderer) Page(w http.ResponseWriter, name string, data any) error {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	return r.templates.ExecuteTemplate(w, name, data)
}

// Fragment renders an htmx fragment.
func (r Renderer) Fragment(w http.ResponseWriter, name string, data any) error {
	return r.templates.ExecuteTemplate(w, name, data)
}

// FuncMap returns template helpers shared by Web UI templates.
func FuncMap() template.FuncMap {
	return MergeFuncMaps(template.FuncMap{
		"upper": strings.ToUpper,
		"lower": strings.ToLower,
	}, htmx.FuncMap(), ui.FuncMap())
}

// MergeFuncMaps merges template function maps.
func MergeFuncMaps(maps ...template.FuncMap) template.FuncMap {
	result := make(template.FuncMap)

	for _, funcs := range maps {
		for name, fn := range funcs {
			result[name] = fn
		}
	}

	return result
}
