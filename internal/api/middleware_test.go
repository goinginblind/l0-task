package api

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/goinginblind/l0-task/internal/pkg/metrics"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/testutil"
	"github.com/stretchr/testify/assert"
)

func Test_normalizePath(t *testing.T) {
	tests := []struct {
		name string
		path string
		want string
	}{
		{
			name: "order path with ID",
			path: "/orders/12345",
			want: "/orders/:id",
		},
		{
			name: "order path without ID",
			path: "/orders/",
			want: "/orders/:id",
		},
		{
			name: "home path",
			path: "/home",
			want: "/home",
		},
		{
			name: "root path",
			path: "/",
			want: "/",
		},
		{
			name: "static path",
			path: "/static/css/main.css",
			want: "/static/css/main.css",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, normalizePath(tt.path))
		})
	}
}

func Test_metricsMiddleware(t *testing.T) {
	metrics.HTTPRequestCount.Reset()
	metrics.HTTPRequestDuration.Reset()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	middleware := metricsMiddleware(handler)

	req := httptest.NewRequest("GET", "/orders/123", nil)
	rr := httptest.NewRecorder()

	middleware.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	expectedCount := `
		# HELP http_total_requests Total requests to the HTTP
		# TYPE http_total_requests counter
		http_total_requests{code="200",method="GET",path="/orders/:id"} 1
	`
	err := testutil.CollectAndCompare(metrics.HTTPRequestCount, bytes.NewBufferString(expectedCount), "http_total_requests")
	assert.NoError(t, err)

	metricFamilies, err := prometheus.DefaultGatherer.Gather()
	assert.NoError(t, err)

	foundDurationMetric := false
	for _, mf := range metricFamilies {
		if mf.GetName() == "http_request_duration" {
			assert.Greater(t, len(mf.GetMetric()), 0)
			assert.Equal(t, "GET", mf.GetMetric()[0].GetLabel()[0].GetValue())
			assert.Equal(t, "/orders/:id", mf.GetMetric()[0].GetLabel()[1].GetValue())
			foundDurationMetric = true
			break
		}
	}
	assert.True(t, foundDurationMetric, "http_request_duration metric not found")
}

func Test_statusRecorder(t *testing.T) {
	rr := httptest.NewRecorder()
	recorder := &statusRecorder{ResponseWriter: rr, status: http.StatusOK}

	recorder.WriteHeader(http.StatusTeapot)
	assert.Equal(t, http.StatusTeapot, recorder.status)
	assert.Equal(t, http.StatusTeapot, rr.Code)
}
