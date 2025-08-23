package api

import (
	"fmt"
	"net/http"
	"runtime/debug"
	"strings"
	"time"

	"github.com/goinginblind/l0-task/internal/pkg/logger"
	"github.com/goinginblind/l0-task/internal/pkg/metrics"
)

type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (r *statusRecorder) WriteHeader(code int) {
	r.status = code
	r.ResponseWriter.WriteHeader(code)
}

// normalizePath converts numeric ids into :id;
// It's crude, but it works for now (and im tired)
func normalizePath(path string) string {
	if strings.HasPrefix(path, "/orders/") {
		return "/orders/:id"
	}
	return path
}

func metricsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		{
			start := time.Now()

			rw := &statusRecorder{ResponseWriter: w, status: http.StatusOK}
			next.ServeHTTP(rw, r)

			duration := time.Since(start).Seconds()
			path := normalizePath(r.URL.Path)

			metrics.HTTPRequestCount.WithLabelValues(r.Method, path, fmt.Sprint(rw.status)).Inc()
			metrics.HTTPRequestDuration.WithLabelValues(r.Method, path).Observe(duration)
		}
	})
}

func recoveryMiddleware(next http.Handler, logger logger.Logger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rec := recover(); rec != nil {
				logger.Errorw("server encountered panic",
					"panic", rec,
					"stack", string(debug.Stack()),
					"method", r.Method,
					"url", r.URL.String(),
					"remote", r.RemoteAddr,
				)
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			}
		}()

		next.ServeHTTP(w, r)
	})
}
