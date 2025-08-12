package api

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"

	"github.com/goinginblind/l0-task/internal/pkg/logger"
	"github.com/goinginblind/l0-task/internal/service"
)

// Server is the HTTP server.
type Server struct {
	service service.OrderService
	logger  logger.Logger
}

// NewServer creates a new Server.
func NewServer(service service.OrderService, logger logger.Logger) *Server {
	return &Server{
		service: service,
		logger:  logger,
	}
}

// Start starts the HTTP server.
func (s *Server) Start(port string) {
	http.HandleFunc("/orders/", s.orderHandler)
	s.logger.Infow("Server listening", "port", port)
	if err := http.ListenAndServe(port, nil); err != nil {
		// TODO: log.Fatal(err) are the places specifically marked
		// to be refactored with retry/backoffs/gracefull restarts or shutdowns,
		// so this log stays for now
		log.Fatal(err)
	}
}

func (s *Server) orderHandler(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/orders/")
	if path == "" || strings.Contains(path, "/") {
		http.NotFound(w, r)
		return
	}

	order, err := s.service.GetOrder(r.Context(), path)
	if err != nil {
		http.Error(w, "Failed to get order", http.StatusInternalServerError)
		return
	}

	if order == nil {
		http.NotFound(w, r)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(order); err != nil {
		http.Error(w, "Failed to encode order", http.StatusInternalServerError)
	}
}
