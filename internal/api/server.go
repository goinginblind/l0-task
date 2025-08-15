package api

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/goinginblind/l0-task/internal/config"
	"github.com/goinginblind/l0-task/internal/pkg/logger"
	"github.com/goinginblind/l0-task/internal/service"
)

// Server is the HTTP server.
type Server struct {
	service    service.OrderService
	logger     logger.Logger
	httpServer *http.Server
}

// NewServer creates a new Server.
func NewServer(service service.OrderService, logger logger.Logger, cfg config.HTTPServerConfig) *Server {
	srv := &Server{
		service: service,
		logger:  logger,
	}

	// New server and handlers
	mux := http.NewServeMux()
	mux.HandleFunc("/orders/", srv.orderHandler)
	srv.httpServer = &http.Server{
		Handler:      mux,
		ReadTimeout:  cfg.ReadTimeout,
		WriteTimeout: cfg.WriteTimeout,
		IdleTimeout:  cfg.IdleTimeout,
	}

	return srv
}

func (s *Server) Start(addr string) error {
	s.httpServer.Addr = addr
	s.logger.Infow("Server listening", "addr", addr)
	// http.ErrServerClosed is a "good" error, returned after Shutdown
	if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return err
	}
	return nil
}

// Shutdown gracefully shuts down the server.
func (s *Server) Shutdown(ctx context.Context) error {
	s.logger.Infow("Shutting down server...")
	return s.httpServer.Shutdown(ctx)
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
		// TODO: Differentiate between not found and other errors
		http.Error(w, "Failed to encode order", http.StatusInternalServerError)
		return
	}
}
