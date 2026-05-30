package blog

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestServerPostEndpoints(t *testing.T) {
	server := NewServer(NewPostStore(SeedPosts())).Routes()

	tests := []struct {
		name       string
		path       string
		wantStatus int
	}{
		{
			name:       "list posts",
			path:       "/posts",
			wantStatus: http.StatusOK,
		},
		{
			name:       "get published post",
			path:       "/posts/hello-world",
			wantStatus: http.StatusOK,
		},
		{
			name:       "unknown post",
			path:       "/posts/missing",
			wantStatus: http.StatusNotFound,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			response := httptest.NewRecorder()
			request := httptest.NewRequest(http.MethodGet, test.path, nil)

			server.ServeHTTP(response, request)

			if response.Code != test.wantStatus {
				t.Fatalf("status = %d, want %d; body = %s", response.Code, test.wantStatus, response.Body.String())
			}

			assertJSONResponse(t, response)
		})
	}
}

func assertJSONResponse(t *testing.T, response *httptest.ResponseRecorder) {
	t.Helper()

	if got := response.Header().Get("Content-Type"); got != "application/json" {
		t.Fatalf("Content-Type = %q, want application/json", got)
	}

	var payload map[string]any
	if err := json.Unmarshal(response.Body.Bytes(), &payload); err != nil {
		t.Fatalf("response body is not JSON: %v", err)
	}
}
