package htmx

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestAttrsRendersEscapedHTMXAttributes(t *testing.T) {
	attrs := HX().
		Get("/project?q=<unsafe>").
		TargetID("workspace").
		SwapOuter().
		Confirm(`Use "now"?`)

	got := attrs.String()
	for _, want := range []string{
		`hx-confirm="Use &#34;now&#34;?"`,
		`hx-get="/project?q=&lt;unsafe&gt;"`,
		`hx-swap="outerHTML"`,
		`hx-target="#workspace"`,
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("Attrs.String() = %q, want %q", got, want)
		}
	}
}

func TestResponseHeaders(t *testing.T) {
	recorder := httptest.NewRecorder()

	Redirect(recorder, "/next")
	Refresh(recorder)
	Retarget(recorder, "#workspace")
	Reswap(recorder, SwapOuter)
	Trigger(recorder, "workspace:refresh")

	headers := recorder.Result().Header
	for name, want := range map[string]string{
		HeaderRedirect: "/next",
		HeaderRefresh:  "true",
		HeaderRetarget: "#workspace",
		HeaderReswap:   "outerHTML",
		HeaderTrigger:  "workspace:refresh",
	} {
		if got := headers.Get(name); got != want {
			t.Fatalf("%s = %q, want %q", name, got, want)
		}
	}
}

func TestIsRequest(t *testing.T) {
	request := httptest.NewRequest(http.MethodGet, "/", nil)
	if IsRequest(request) {
		t.Fatal("IsRequest() = true, want false without HX-Request")
	}

	request.Header.Set(HeaderRequest, "true")

	if !IsRequest(request) {
		t.Fatal("IsRequest() = false, want true with HX-Request")
	}
}
