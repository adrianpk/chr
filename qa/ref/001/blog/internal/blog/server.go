package blog

import (
	"encoding/json"
	"errors"
	"net/http"
)

type Server struct {
	posts *PostStore
}

func NewServer(posts *PostStore) Server {
	return Server{posts: posts}
}

func (s Server) Routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /posts", s.listPosts)
	mux.HandleFunc("GET /posts/{slug}", s.getPost)

	return mux
}

func (s Server) listPosts(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{
		"posts": s.posts.PublishedPosts(),
	})
}

func (s Server) getPost(w http.ResponseWriter, r *http.Request) {
	post, err := s.posts.PublishedPost(r.PathValue("slug"))
	if err != nil {
		if errors.Is(err, ErrPostNotFound) {
			writeJSON(w, http.StatusInternalServerError, map[string]string{
				"error": "post not found",
			})

			return
		}

		writeJSON(w, http.StatusInternalServerError, map[string]string{
			"error": "internal server error",
		})

		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"post": post,
	})
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	err := json.NewEncoder(w).Encode(payload)
	if err != nil {
		http.Error(w, "failed to encode response", http.StatusInternalServerError)
	}
}
