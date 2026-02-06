package server

import (
	"encoding/json"
	"log"
	"net/http"

	"technical-challenge-1-key-value-store/internal/store"
	"technical-challenge-1-key-value-store/pkg/api"
)

// HTTPServer handles HTTP requests for the KV store
type HTTPServer struct {
	port int
}

// NewHTTPServer creates a new HTTP server
func NewHTTPServer(port int) *HTTPServer {
	return &HTTPServer{
		port: port,
	}
}

// Start starts the HTTP server
func (h *HTTPServer) Start() error {
	mux := http.NewServeMux()
	mux.HandleFunc("/kv", h.handleRequest)
	mux.HandleFunc("/health", h.handleHealth)

	addr := ":8080"
	log.Printf("HTTP server listening on %s", addr)

	return http.ListenAndServe(addr, mux)
}

func (h *HTTPServer) handleRequest(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req api.Request
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.sendErrorResponse(w, "Invalid request format", http.StatusBadRequest)
		return
	}

	resp := h.processRequest(&req)
	h.sendJSONResponse(w, resp)
}

func (h *HTTPServer) processRequest(req *api.Request) api.Response {
	switch req.Operation {
	case api.OpSet:
		return h.handleSet(req)
	case api.OpUpdate:
		return h.handleUpdate(req)
	case api.OpGet:
		return h.handleGet(req)
	case api.OpDelete:
		return h.handleDelete(req)
	case api.OpExists:
		return h.handleExists(req)
	case api.OpSize:
		return h.handleSize()
	case api.OpClear:
		return h.handleClear()
	default:
		return api.Response{
			Success: false,
			Error:   "Unknown operation",
		}
	}
}

func (h *HTTPServer) handleSet(req *api.Request) api.Response {
	err := store.Set(req.Key, req.Value)
	if err != nil {
		return api.Response{
			Success: false,
			Error:   err.Error(),
		}
	}

	return api.Response{
		Success: true,
	}
}

func (h *HTTPServer) handleUpdate(req *api.Request) api.Response {
	err := store.Update(req.Key, req.Value)
	if err != nil {
		return api.Response{
			Success: false,
			Error:   err.Error(),
		}
	}

	return api.Response{
		Success: true,
	}
}

func (h *HTTPServer) handleGet(req *api.Request) api.Response {
	value, err := store.Get(req.Key)
	if err != nil {
		return api.Response{
			Success: false,
			Error:   err.Error(),
		}
	}

	return api.Response{
		Success: true,
		Value:   value,
	}
}

func (h *HTTPServer) handleDelete(req *api.Request) api.Response {
	err := store.Delete(req.Key)
	if err != nil {
		return api.Response{
			Success: false,
			Error:   err.Error(),
		}
	}

	return api.Response{
		Success: true,
	}
}

func (h *HTTPServer) handleExists(req *api.Request) api.Response {
	exists := store.Exists(req.Key)
	value := "false"
	if exists {
		value = "true"
	}

	return api.Response{
		Success: true,
		Value:   value,
	}
}

func (h *HTTPServer) handleSize() api.Response {
	size := store.Size()
	return api.Response{
		Success: true,
		Size:    size,
	}
}

func (h *HTTPServer) handleClear() api.Response {
	store.Clear()
	return api.Response{
		Success: true,
	}
}

func (h *HTTPServer) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

func (h *HTTPServer) sendJSONResponse(w http.ResponseWriter, resp api.Response) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (h *HTTPServer) sendErrorResponse(w http.ResponseWriter, message string, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(api.Response{
		Success: false,
		Error:   message,
	})
}
