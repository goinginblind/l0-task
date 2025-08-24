package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	/* HTTP metrics */
	HTTPRequestCount = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "http_requests_total",
		Help: "Total requests to the HTTP",
	},
		[]string{"method", "path", "code"},
	)
	HTTPRequestDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name: "http_request_duration_seconds",
		Help: "Duration of HTTP requests",
	},
		[]string{"method", "path"},
	)

	/* Consumer metrics */
	MessagesProcessedTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "consumer_processed_total",
		Help: "Total number of processed messages",
	},
		[]string{"status"},
	)
	MessageProcessingLatency = promauto.NewHistogram(prometheus.HistogramOpts{
		Name:    "consumer_processing_latency_seconds",
		Help:    "Time from message creation to processing completion.",
		Buckets: prometheus.DefBuckets,
	})
	DLQMessagesTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "consumer_dlq_messages_total",
		Help: "Total number of messages sent to the Dead Letter Queue.",
	},
		[]string{"reason"},
	)
	ConsumerLag = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "consumer_lag",
		Help: "Estimated number of messages lagging behind the latest offset.",
	},
		[]string{"topic", "partition"},
	)

	/* Cache metrics */
	CacheHits = promauto.NewCounter(prometheus.CounterOpts{
		Name: "cache_hits_total",
		Help: "Total number of cache hits",
	})
	CacheMisses = promauto.NewCounter(prometheus.CounterOpts{
		Name: "cache_misses_total",
		Help: "Total number of cache misses",
	})
	CacheResponseTime = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "cache_response_time_seconds",
		Help:    "Duartion of cache response",
		Buckets: []float64{.0001, .00025, .0005, .001, .0025, .005, .01, 0.025, 0.05, 0.1, 0.25},
	},
		[]string{"operation"},
	)

	/* DB metrics */
	DBResponseTime = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "db_response_time_seconds",
		Help:    "Duartion of DB response",
		Buckets: []float64{.005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5},
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
