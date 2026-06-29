package render

import (
	"fmt"
	"html/template"
	"io/fs"
	"net/http"
	"strings"

	"forge.adrianpk.com/chc/internal/webui/htmx"
	"forge.adrianpk.com/chc/internal/webui/ui"
)

type Renderer struct {
	templates *template.Template
}

func New(files fs.FS, patterns ...string) (Renderer, error) {
	tmpl, err := template.New("").Funcs(FuncMap()).ParseFS(files, patterns...)
	if err != nil {
		return Renderer{}, fmt.Errorf("parse web templates: %w", err)
	}

	return Renderer{templates: tmpl}, nil
}

func (r Renderer) Page(w http.ResponseWriter, name string, data any) error {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	return r.templates.ExecuteTemplate(w, name, data)
}

func (r Renderer) Fragment(w http.ResponseWriter, name string, data any) error {
	return r.templates.ExecuteTemplate(w, name, data)
}

func FuncMap() template.FuncMap {
	return MergeFuncMaps(template.FuncMap{
		"upper": strings.ToUpper,
		"lower": strings.ToLower,
	}, htmx.FuncMap(), ui.FuncMap())
}

func MergeFuncMaps(maps ...template.FuncMap) template.FuncMap {
	result := make(template.FuncMap)
	for _, funcs := range maps {
		for name, fn := range funcs {
			result[name] = fn
		}
	}

	return result
}
