package capture

import (
	"maps"
	"crypto/rand"
	"fmt"
	"hooklet/internal/model"
	"hooklet/internal/store"
	"io"
	"net/http"
	"time"
)



const maxBodyBytes = 10 * 1024 * 1024 


type Broadcaster interface {
	Publish(req *model.WebhookRequest)
}

type Handler struct {
	store store.Store
	broker Broadcaster
}

func NewHandler(s store.Store, b Broadcaster) *Handler {
	return &Handler{store: s, broker: b}
}

func(h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request){
	r.Body = http.MaxBytesReader(w, r.Body, maxBodyBytes)
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Payload too large or unreadable", http.StatusBadRequest)
		return
	}

	headers := make(model.HeaderMap)
	
	maps.Copy(headers, r.Header)

    req := &model.WebhookRequest{
		ID:            generateID(),
		Method:        r.Method,
		Path:          r.URL.Path,
		QueryString:   r.URL.RawQuery,
		Headers:       headers,
		Body:          body,
		ContentType:   r.Header.Get("Content-Type"),
		ContentLength: int64(len(body)),
		RemoteAddr:    r.RemoteAddr,
		CreatedAt:     time.Now().UTC(),
	}

	if err := h.store.SaveRequest(r.Context(), req);err != nil {
		http.Error(w, "Failed to persist webhook", http.StatusInternalServerError)
		return
	}
	
	if h.broker != nil {
		h.broker.Publish(req)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = fmt.Fprintf(w, `{"status":"captured","id":%q}`, req.ID)
}

func generateID() string {
	var b [16]byte
	_, _ = rand.Read(b[:])
	return fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:])
}