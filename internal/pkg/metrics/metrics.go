package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	HTTPRequestCount = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "http_total_requests",
		Help: "Total requests to the HTTP",
	},
		[]string{"method", "path", "code"},
	)

	HTTPRequestDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name: "http_request_duration",
		Help: "Duration of HTTP requests",
	},
		[]string{"method", "path"},
	)

	// for internal/consumer/workers
	MessagesProcessedTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "consumer_processed_total",
		Help: "Total number of processed messages",
	},
		[]string{"status"},
	)

	CacheHits = promauto.NewCounter(prometheus.CounterOpts{
		Name: "cache_hits_total",
		Help: "Total number of cache hits",
	})
	CacheMisses = promauto.NewCounter(prometheus.CounterOpts{
		Name: "cache_misses_total",
		Help: "Total number of cace misses",
	})
	CacheResponseTime = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name: "cache_response_time",
		Help: "Duartion of cache response",
	},
		[]string{"operation"},
	)

	DBResponseTime = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name: "db_response_time",
		Help: "Duartion of DB response",
	},
		[]string{"operation"},
	)
	DBUptime = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "db_up",
		Help: "1 if database is reachable, 0 if not",
	})
	DbTransientErrors = promauto.NewCounter(prometheus.CounterOpts{
		Name: "db_transient_err_total",
		Help: "Total number of recoverable DB hiccups",
	})
)
