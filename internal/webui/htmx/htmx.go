// Package htmx provides local htmx helpers for Web UI templates and handlers.
package htmx

import (
	"fmt"
	"html/template"
	"net/http"
	"sort"
	"strings"
)

const (
	// HeaderRequest identifies htmx requests.
	HeaderRequest = "HX-Request"
	// HeaderRedirect identifies htmx redirect responses.
	HeaderRedirect = "HX-Redirect"
	// HeaderRefresh identifies htmx refresh responses.
	HeaderRefresh = "HX-Refresh"
	// HeaderRetarget identifies htmx retarget responses.
	HeaderRetarget = "HX-Retarget"
	// HeaderReswap identifies htmx reswap responses.
	HeaderReswap = "HX-Reswap"
	// HeaderTrigger identifies htmx trigger responses.
	HeaderTrigger = "HX-Trigger"
)

// Action describes one htmx HTTP action attribute.
type Action struct {
	method string
	url    string
}

// Get creates an htmx GET action.
func Get(url string) Action {
	return Action{method: "get", url: url}
}

// Post creates an htmx POST action.
func Post(url string) Action {
	return Action{method: "post", url: url}
}

// Attr returns the htmx attribute name for the action.
func (a Action) Attr() string {
	if a.method == "" {
		return ""
	}

	return "hx-" + a.method
}

// Swap identifies an htmx swap strategy.
type Swap string

const (
	// SwapInner swaps the target inner HTML.
	SwapInner Swap = "innerHTML"
	// SwapOuter swaps the target outer HTML.
	SwapOuter Swap = "outerHTML"
	// SwapNone disables target swapping.
	SwapNone Swap = "none"
)

// Attrs builds escaped htmx HTML attributes.
type Attrs struct {
	action  *Action
	target  string
	swap    Swap
	confirm string
	boost   *bool
	extra   map[string]string
}

// NewAttrs creates an empty htmx attribute builder.
func NewAttrs() *Attrs {
	return &Attrs{}
}

// HX creates an empty htmx attribute builder.
func HX() *Attrs {
	return NewAttrs()
}

// Get sets an htmx GET action.
func (a *Attrs) Get(url string) *Attrs {
	action := Get(url)
	a.action = &action

	return a
}

// Post sets an htmx POST action.
func (a *Attrs) Post(url string) *Attrs {
	action := Post(url)
	a.action = &action

	return a
}

// TargetID sets the htmx target to an element ID.
func (a *Attrs) TargetID(id string) *Attrs {
	a.target = "#" + id

	return a
}

// TargetThis sets the htmx target to the element itself.
func (a *Attrs) TargetThis() *Attrs {
	a.target = "this"

	return a
}

// SwapOuter sets the htmx swap strategy to outerHTML.
func (a *Attrs) SwapOuter() *Attrs {
	a.swap = SwapOuter

	return a
}

// SwapInner sets the htmx swap strategy to innerHTML.
func (a *Attrs) SwapInner() *Attrs {
	a.swap = SwapInner

	return a
}

// SwapNone disables htmx content swapping.
func (a *Attrs) SwapNone() *Attrs {
	a.swap = SwapNone

	return a
}

// Confirm sets an htmx confirmation prompt.
func (a *Attrs) Confirm(message string) *Attrs {
	a.confirm = message

	return a
}

// Boost enables htmx boosted navigation.
func (a *Attrs) Boost() *Attrs {
	value := true
	a.boost = &value

	return a
}

// Set adds a custom htmx or HTML attribute.
func (a *Attrs) Set(key, value string) *Attrs {
	if a.extra == nil {
		a.extra = make(map[string]string)
	}

	a.extra[key] = value

	return a
}

// Map returns the htmx attributes as a map.
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

// HTML returns escaped attributes for templates.
func (a *Attrs) HTML() template.HTMLAttr {
	return template.HTMLAttr(a.String())
}

// IsRequest reports whether a request came from htmx.
func IsRequest(r *http.Request) bool {
	return r.Header.Get(HeaderRequest) == "true"
}

// Redirect sets an htmx client-side redirect response header.
func Redirect(w http.ResponseWriter, url string) {
	w.Header().Set(HeaderRedirect, url)
}

// Refresh asks htmx to refresh the page.
func Refresh(w http.ResponseWriter) {
	w.Header().Set(HeaderRefresh, "true")
}

// Retarget changes the htmx response target.
func Retarget(w http.ResponseWriter, selector string) {
	w.Header().Set(HeaderRetarget, selector)
}

// Reswap changes the htmx response swap strategy.
func Reswap(w http.ResponseWriter, swap Swap) {
	w.Header().Set(HeaderReswap, string(swap))
}

// Trigger emits an htmx client-side event.
func Trigger(w http.ResponseWriter, event string) {
	w.Header().Set(HeaderTrigger, event)
}

// OOB identifies an htmx out-of-band swap mode.
type OOB string

const (
	// OOBSwap swaps an element out of band.
	OOBSwap OOB = "true"
	// OOBOuter swaps an element outer HTML out of band.
	OOBOuter OOB = "outerHTML"
	// OOBInner swaps an element inner HTML out of band.
	OOBInner OOB = "innerHTML"
	// OOBBeforeEnd appends content before the target end out of band.
	OOBBeforeEnd OOB = "beforeend"
)

// OOBAttr renders an htmx out-of-band swap attribute.
func OOBAttr(oob OOB) template.HTMLAttr {
	return template.HTMLAttr(fmt.Sprintf(`hx-swap-oob="%s"`, template.HTMLEscapeString(string(oob))))
}

// FuncMap returns htmx helpers for templates.
func FuncMap() template.FuncMap {
	return template.FuncMap{
		"hx":      HX,
		"oobAttr": OOBAttr,
	}
}
