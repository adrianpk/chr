package dialer

import (
	"encoding/json"
	"errors"
	"net/http"
)

type Server struct {
	store *Store
}

func NewServer(store *Store) Server {
	return Server{store: store}
}

func (s Server) Routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /leads", s.listLeads)
	mux.HandleFunc("GET /leads/{id}", s.getLead)
	mux.HandleFunc("POST /leads/{id}/calls", s.recordCall)
	mux.HandleFunc("GET /agents/{id}/queue", s.agentQueue)

	return mux
}

func (s Server) listLeads(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{"leads": s.store.Leads()})
}

func (s Server) getLead(w http.ResponseWriter, r *http.Request) {
	lead, err := s.store.Lead(r.PathValue("id"))
	if err != nil {
		if errors.Is(err, ErrLeadNotFound) {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "lead not found"})

			return
		}

		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal error"})

		return
	}

	writeJSON(w, http.StatusOK, lead)
}

func (s Server) recordCall(w http.ResponseWriter, r *http.Request) {
	var request struct {
		AgentID string `json:"agent_id"`
		Outcome string `json:"outcome"`
	}
	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid json"})

		return
	}

	call, err := s.store.RecordCall(r.PathValue("id"), request.AgentID, request.Outcome)
	if err != nil {
		if errors.Is(err, ErrLeadNotFound) {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "lead not found"})

			return
		}

		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal error"})

		return
	}

	writeJSON(w, http.StatusCreated, call)
}

func (s Server) agentQueue(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{"leads": s.store.Queue(r.PathValue("id"))})
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}
