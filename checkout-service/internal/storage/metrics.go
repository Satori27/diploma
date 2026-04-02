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
)
