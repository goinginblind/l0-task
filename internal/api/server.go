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

	"errors"

	"github.com/goinginblind/l0-task/internal/api/ui"
	"github.com/goinginblind/l0-task/internal/config"
	"github.com/goinginblind/l0-task/internal/pkg/logger"
	"github.com/goinginblind/l0-task/internal/service"
	"github.com/goinginblind/l0-task/internal/store"
	"github.com/prometheus/client_golang/prometheus/promhttp"
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

	mainMux := http.NewServeMux()
	mainMux.Handle("/metrics", promhttp.Handler())
	mainMux.Handle("/", metricsMiddleware(mux))

	srv.httpServer = &http.Server{
		Handler:      mainMux,
		ReadTimeout:  cfg.ReadTimeout,
		WriteTimeout: cfg.WriteTimeout,
		IdleTimeout:  cfg.IdleTimeout,
	}

	return srv, nil
}

// Start the server
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

// Home page handler
func (s *Server) home(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == "/" {
		http.Redirect(w, r, "/home", http.StatusFound)
		return
	}
	if r.URL.Path != "/home" {
		http.NotFound(w, r)
		return
	}

	uid := r.URL.Query().Get("uid")
	errMsg := r.URL.Query().Get("error")

	data := map[string]any{
		"OrderUID":   uid,
		"OrderFound": false,
		"Error":      "",
	}

	if errMsg == "not_found" {
		data["Error"] = fmt.Sprintf("Order with UID '%s' not found", uid)
	}

	s.render(w, r, http.StatusOK, "home.tmpl", data)
}

// order page handler
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
		if errors.Is(err, store.ErrNotFound) {
			redirectURL := fmt.Sprintf("/home?error=not_found&uid=%s", uid)
			http.Redirect(w, r, redirectURL, http.StatusFound)
			return
		}
		s.serverError(w, r, err)
		return
	}

	data := map[string]any{
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

// render the template: its usueless actually and couldve been doen w/o templating
// but it is what it is...
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
