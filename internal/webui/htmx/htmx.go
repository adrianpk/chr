package htmx

import (
	"fmt"
	"html/template"
	"net/http"
	"sort"
	"strings"
)

const (
	HeaderRequest  = "HX-Request"
	HeaderRedirect = "HX-Redirect"
	HeaderRefresh  = "HX-Refresh"
	HeaderRetarget = "HX-Retarget"
	HeaderReswap   = "HX-Reswap"
	HeaderTrigger  = "HX-Trigger"
)

type Action struct {
	method string
	url    string
}

func Get(url string) Action {
	return Action{method: "get", url: url}
}

func Post(url string) Action {
	return Action{method: "post", url: url}
}

func (a Action) Attr() string {
	if a.method == "" {
		return ""
	}

	return "hx-" + a.method
}

type Swap string

const (
	SwapInner Swap = "innerHTML"
	SwapOuter Swap = "outerHTML"
	SwapNone  Swap = "none"
)

type Attrs struct {
	action  *Action
	target  string
	swap    Swap
	confirm string
	boost   *bool
	extra   map[string]string
}

func NewAttrs() *Attrs {
	return &Attrs{}
}

func HX() *Attrs {
	return NewAttrs()
}

func (a *Attrs) Get(url string) *Attrs {
	action := Get(url)
	a.action = &action

	return a
}

func (a *Attrs) Post(url string) *Attrs {
	action := Post(url)
	a.action = &action

	return a
}

func (a *Attrs) TargetID(id string) *Attrs {
	a.target = "#" + id

	return a
}

func (a *Attrs) TargetThis() *Attrs {
	a.target = "this"

	return a
}

func (a *Attrs) SwapOuter() *Attrs {
	a.swap = SwapOuter

	return a
}

func (a *Attrs) SwapInner() *Attrs {
	a.swap = SwapInner

	return a
}

func (a *Attrs) SwapNone() *Attrs {
	a.swap = SwapNone

	return a
}

func (a *Attrs) Confirm(message string) *Attrs {
	a.confirm = message

	return a
}

func (a *Attrs) Boost() *Attrs {
	value := true
	a.boost = &value

	return a
}

func (a *Attrs) Set(key, value string) *Attrs {
	if a.extra == nil {
		a.extra = make(map[string]string)
	}

	a.extra[key] = value

	return a
}

func (a *Attrs) Map() map[string]string {
	result := make(map[string]string)
	if a.action != nil {
		result[a.action.Attr()] = a.action.url
	}

	if a.target != "" {
		result["hx-target"] = a.target
	}

	if a.swap != "" {
		result["hx-swap"] = string(a.swap)
	}

	if a.confirm != "" {
		result["hx-confirm"] = a.confirm
	}

	if a.boost != nil {
		result["hx-boost"] = fmt.Sprintf("%t", *a.boost)
	}

	for key, value := range a.extra {
		result[key] = value
	}

	return result
}

func (a *Attrs) String() string {
	values := a.Map()
	if len(values) == 0 {
		return ""
	}

	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	parts := make([]string, 0, len(keys))
	for _, key := range keys {
		parts = append(parts, fmt.Sprintf(`%s="%s"`, key, template.HTMLEscapeString(values[key])))
	}

	return strings.Join(parts, " ")
}

func (a *Attrs) HTML() template.HTMLAttr {
	return template.HTMLAttr(a.String())
}

func IsRequest(r *http.Request) bool {
	return r.Header.Get(HeaderRequest) == "true"
}

func Redirect(w http.ResponseWriter, url string) {
	w.Header().Set(HeaderRedirect, url)
}

func Refresh(w http.ResponseWriter) {
	w.Header().Set(HeaderRefresh, "true")
}

func Retarget(w http.ResponseWriter, selector string) {
	w.Header().Set(HeaderRetarget, selector)
}

func Reswap(w http.ResponseWriter, swap Swap) {
	w.Header().Set(HeaderReswap, string(swap))
}

func Trigger(w http.ResponseWriter, event string) {
	w.Header().Set(HeaderTrigger, event)
}

type OOB string

const (
	OOBSwap      OOB = "true"
	OOBOuter     OOB = "outerHTML"
	OOBInner     OOB = "innerHTML"
	OOBBeforeEnd OOB = "beforeend"
)

func OOBAttr(oob OOB) template.HTMLAttr {
	return template.HTMLAttr(fmt.Sprintf(`hx-swap-oob="%s"`, template.HTMLEscapeString(string(oob))))
}

func FuncMap() template.FuncMap {
	return template.FuncMap{
		"hx":      HX,
		"oobAttr": OOBAttr,
	}
}
