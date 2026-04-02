package storage

import (
	"github.com/prometheus/client_golang/prometheus"
)

var (
	poolTotalConns = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "db_pool_total_connections",
			Help: "Total connections in pgx pool",
		},
	)

	poolIdleConns = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "db_pool_idle_connections",
			Help: "Idle connections in pgx pool",
		},
	)

	poolAcquiredConns = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "db_pool_acquired_connections",
			Help: "Currently acquired connections",
		},
	)

	outboxProcessed = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "outbox_events_processed_total",
			Help: "Total number of processed outbox events",
		},
		[]string{"status"},
	)

	outboxLag = prometheus.NewHistogram(
		prometheus.HistogramOpts{
			Name:    "outbox_event_publish_seconds",
			Help:    "Time for publishing kafka events from outbox",
			Buckets: prometheus.DefBuckets,
		},
	)

	outboxBatchDuration = prometheus.NewHistogram(
		prometheus.HistogramOpts{
			Name:    "outbox_batch_duration_seconds",
			Help:    "Duration of outbox batch processing",
			Buckets: prometheus.DefBuckets,
		},
	)

		outboxBatchSize = prometheus.NewHistogram(
		prometheus.HistogramOpts{
			Name:    "outbox_batch_size",
			Help:    "Size of outbox batch processing",
			Buckets: []float64{1, 5, 10, 20, 50, 100, 200, 500, 1000, 10000},
		},
	)

	outboxPending = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "outbox_pending_events",
			Help: "Number of unprocessed outbox events",
		},
	)

	outboxPublishErrors = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "outbox_publish_errors_total",
			Help: "Total number of publish errors",
		},
	)
)
