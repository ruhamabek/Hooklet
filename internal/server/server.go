package server

import (
	"encoding/json"
	"errors"
	"fmt"
	"hooklet/internal/capture"
	"hooklet/internal/event"
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
	// 1. Webhook Ingestion: capture any incoming webhook hitting /wh/*
	s.mux.Handle("/wh/", capture.NewHandler(s.store, s.broker))

	// 2. REST APIs
	s.mux.HandleFunc("GET /api/requests", s.handleListRequests)
	s.mux.HandleFunc("GET /api/requests/{id}", s.handleGetRequest)
	s.mux.HandleFunc("POST /api/requests/{id}/replay", s.handleReplay)

	// 3. Real-time SSE Stream
	s.mux.HandleFunc("GET /api/events", s.handleSSE)

	// 4. Exact root dashboard (remove the redundant GET /)
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
	attempt, err := s.dispatcher.Replay(r.Context(), req, targetURL)
	if attempt == nil && err != nil {
		http.Error(w, err.Error(), http.StatusBadGateway)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(attempt)
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
	defer s.broker.Unsusbcribe(ch)
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