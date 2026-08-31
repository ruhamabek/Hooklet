package server

import (
	"encoding/json"
	"errors"
	"fmt"
	"hooklet/internal/capture"
	"hooklet/internal/event"
	"hooklet/internal/model"
	"hooklet/internal/replay"
	"hooklet/internal/store"
	"hooklet/web"
	"net/http"
)

type Server struct {
	store         store.Store
	broker        *event.Broker
	dispatcher    *replay.Dispatcher
	defaultTarget string
	mux           *http.ServeMux
}

func New(s store.Store, b *event.Broker, d *replay.Dispatcher, defaultTarget string) *Server {
	srv := &Server{
		store:         s,
		broker:        b,
		dispatcher:    d,
		defaultTarget: defaultTarget,
		mux:           http.NewServeMux(),
	}
	srv.routes()
	return srv
}

func (s *Server) Handler() http.Handler {
	return s.mux
}

func (s *Server) routes() {
 	s.mux.Handle("/wh/", capture.NewHandler(s.store, s.broker))

 	s.mux.HandleFunc("GET /api/requests", s.handleListRequests)
	s.mux.HandleFunc("DELETE /api/requests", s.handleClearRequests)
	s.mux.HandleFunc("DELETE /api/requests/{id}", s.handleDeleteRequest)
	s.mux.HandleFunc("GET /api/requests/{id}", s.handleGetRequest)
	s.mux.HandleFunc("POST /api/requests/{id}/replay", s.handleReplay)

 	s.mux.HandleFunc("GET /api/events", s.handleSSE)

 	s.mux.Handle("GET /static/", http.StripPrefix("/static/", web.StaticHandler()))
	s.mux.HandleFunc("GET /{$}", s.handleDashboard)
}

func (s *Server) handleListRequests(w http.ResponseWriter, r *http.Request) {
	requests, err := s.store.ListRequests(r.Context(), 50, 0)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(requests)
}

func (s *Server) handleGetRequest(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	req, err := s.store.GetRequest(r.Context(), id)
	if errors.Is(err, store.ErrNotFound) || req == nil {
		http.Error(w, "Webhook not found", http.StatusNotFound)
		return
	}
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(req)
}


type customReplayPayload struct {
	Target  string          `json:"target"`
	Method  string          `json:"method"`
	Headers model.HeaderMap `json:"headers"`
	Body    string          `json:"body"`
}

func (s *Server) handleReplay(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	req, err := s.store.GetRequest(r.Context(), id)
	if errors.Is(err, store.ErrNotFound) || req == nil {
		http.Error(w, "Webhook not found", http.StatusNotFound)
		return
	}
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	targetURL := r.URL.Query().Get("target")
	if targetURL == "" {
		targetURL = s.defaultTarget
	}

 	if r.Header.Get("Content-Type") == "application/json" {
		var payload customReplayPayload
		if err := json.NewDecoder(r.Body).Decode(&payload); err == nil {
			if payload.Target != "" {
				targetURL = payload.Target
			}
			clone := *req
			if payload.Method != "" {
				clone.Method = payload.Method
			}
			if len(payload.Headers) > 0 {
				clone.Headers = payload.Headers
			}
			if payload.Body != "" {
				clone.Body = []byte(payload.Body)
			}
			req = &clone
		}
	}

	attempt, err := s.dispatcher.Replay(r.Context(), req, targetURL)
	if attempt == nil && err != nil {
		http.Error(w, err.Error(), http.StatusBadGateway)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(attempt)
}

func (s *Server) handleClearRequests(w http.ResponseWriter, r *http.Request) {
	if err := s.store.ClearRequests(r.Context()); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_, _ = w.Write([]byte(`{"status":"cleared"}`))
}

func (s *Server) handleDeleteRequest(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if err := s.store.DeleteRequest(r.Context(), id); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_, _ = w.Write([]byte(`{"status":"deleted"}`))
}

func (s *Server) handleSSE(w http.ResponseWriter, r *http.Request) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming unsupported", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	ch := s.broker.Subscribe()
	defer s.broker.Unsubscribe(ch)
	for {
		select {
		case <-r.Context().Done():
			return
		case req, ok := <-ch:
			if !ok {
				return
			}
			data, _ := json.Marshal(req)
			fmt.Fprintf(w, "data: %s\n\n", data)
			flusher.Flush()
		}
	}
}

func (s *Server) handleDashboard(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = w.Write(web.IndexHTML)
}