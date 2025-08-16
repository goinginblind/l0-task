package api

import (
	"bytes"
	"context"
	"fmt"
	"html/template"
	"io/fs"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/goinginblind/l0-task/internal/api/ui"
	"github.com/goinginblind/l0-task/internal/config"
	"github.com/goinginblind/l0-task/internal/pkg/logger"
	"github.com/goinginblind/l0-task/internal/service"
)

// Server is the HTTP server.
type Server struct {
	service       service.OrderService
	logger        logger.Logger
	httpServer    *http.Server
	templateCache map[string]*template.Template
}

// NewServer creates a new Server.
func NewServer(service service.OrderService, logger logger.Logger, cfg config.HTTPServerConfig) (*Server, error) {
	templateCache, err := newTemplateCache()
	if err != nil {
		return nil, err
	}

	srv := &Server{
		service:       service,
		logger:        logger,
		templateCache: templateCache,
	}

	mux := http.NewServeMux()
	mux.Handle("/static/", http.FileServer(http.FS(ui.Files)))

	mux.HandleFunc("/", srv.home)
	mux.HandleFunc("/home", srv.home)
	mux.HandleFunc("/orders/", srv.orderView)

	srv.httpServer = &http.Server{
		Handler:      mux,
		ReadTimeout:  cfg.ReadTimeout,
		WriteTimeout: cfg.WriteTimeout,
		IdleTimeout:  cfg.IdleTimeout,
	}

	return srv, nil
}

func (s *Server) Start(addr string) error {
	s.httpServer.Addr = addr
	s.logger.Infow("Server listening", "addr", addr)
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

func (s *Server) home(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" && r.URL.Path != "/home" {
		http.NotFound(w, r)
		return
	}
	s.render(w, r, http.StatusOK, "home.tmpl", nil)
}

func (s *Server) orderView(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		return
	}

	uid := strings.TrimPrefix(r.URL.Path, "/orders/")
	if uid == "" {
		http.Redirect(w, r, "/home", http.StatusFound)
		return
	}

	order, err := s.service.GetOrder(r.Context(), uid)
	if err != nil {
		s.serverError(w, r, err)
		return
	}

	data := map[string]interface{}{
		"OrderUID":   uid,
		"OrderFound": order != nil,
		"Order":      order,
	}

	status := http.StatusOK
	if order == nil {
		status = http.StatusNotFound
	}

	s.render(w, r, status, "order.tmpl", data)
}

func (s *Server) render(w http.ResponseWriter, r *http.Request, status int, page string, data interface{}) {
	ts, ok := s.templateCache[page]
	if !ok {
		err := fmt.Errorf("the template %s does not exist", page)
		s.serverError(w, r, err)
		return
	}

	buf := new(bytes.Buffer)
	err := ts.ExecuteTemplate(buf, "base", data)
	if err != nil {
		s.serverError(w, r, err)
		return
	}

	w.WriteHeader(status)
	buf.WriteTo(w)
}

func (s *Server) serverError(w http.ResponseWriter, r *http.Request, err error) {
	s.logger.Errorw("server error", "error", err, "request_method", r.Method, "request_uri", r.URL.RequestURI())
	http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
}

func newTemplateCache() (map[string]*template.Template, error) {
	cache := map[string]*template.Template{}

	pages, err := fs.Glob(ui.Files, "html/pages/*.tmpl")
	if err != nil {
		return nil, err
	}

	for _, page := range pages {
		name := filepath.Base(page)

		patterns := []string{
			"html/layouts/base.tmpl",
			page,
		}

		ts, err := template.New(name).ParseFS(ui.Files, patterns...)
		if err != nil {
			return nil, err
		}

		cache[name] = ts
	}

	return cache, nil
}