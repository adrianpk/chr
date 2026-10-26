// Package webui owns the local Web UI handler and embedded assets.
//
//nolint:wsl_v5 // Handler methods keep request decoding and guarded rendering steps together.
package webui

import (
	"fmt"
	"net/http"
	"time"
)

func (h handler) withRequestLogging(next http.Handler) http.Handler {
	if h.logRequests == nil {
		return next
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		recorder := &statusRecorder{ResponseWriter: w, status: http.StatusOK}
		started := time.Now()

		next.ServeHTTP(recorder, r)

		_, err := fmt.Fprintf(
			h.logRequests,
			"web request: method=%s path=%s status=%d duration=%s\n",
			r.Method,
			r.URL.EscapedPath(),
			recorder.status,
			time.Since(started).Round(time.Millisecond),
		)
		if err != nil {
			return
		}
	})
}

type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (r *statusRecorder) WriteHeader(status int) {
	r.status = status
	r.ResponseWriter.WriteHeader(status)
}
