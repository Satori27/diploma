package storage

import (
	"context"
	"log"
	"log/slog"
	"os"
	"time"

	queries "github.com/Satori27/diploma/checkout-service/internal/db/queries"
	"github.com/amirsalarsafaei/sqlc-pgx-monitoring/dbtracer"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/prometheus/client_golang/prometheus"
)

func NewStorage(ctx context.Context) *Storage {
	dbHost := os.Getenv("DB_HOST")
	dbPort := os.Getenv("DB_PORT")
	dbUser := os.Getenv("DB_USER")
	dbPassword := os.Getenv("DB_PASSWORD")

	config, err := pgxpool.ParseConfig(
		"postgres://" + dbUser + ":" + dbPassword + "@" + dbHost + ":" + dbPort + "/checkout",
	)
	if err != nil {
		log.Fatal("Error parse config: ", err.Error())
	}

	config.MaxConnLifetime = 10 * time.Minute   // убить соединение через 10 минут
	config.MaxConnIdleTime = 5 * time.Minute

	tracer, err := dbtracer.NewDBTracer(
		"orders-db",
		dbtracer.WithLogger(slog.Default()),
		dbtracer.WithLatencyHistogramConfig("sql_query_duration_ms", "ms", "Duration of SQL queries", 1, 5, 10, 50, 100, 500, 1000, 5000),
	)

	if err != nil {
		log.Fatal("Error init tracer: ", err.Error())
	}

	config.ConnConfig.Tracer = tracer

	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		log.Fatal("Error connect to postgres: ", err.Error())
	}

	s := &Storage{
		db:      pool,
		queries: queries.New(pool),
	}

	s.initMetrics()

	return s
}

type Storage struct {
	db      *pgxpool.Pool
	queries *queries.Queries
}

func (s *Storage) GetPool() *pgxpool.Pool {
	return s.db
}

func (s *Storage) Close() {
	s.db.Close()
}

func (s *Storage) initMetrics() {
	prometheus.MustRegister(poolTotalConns, poolIdleConns, poolAcquiredConns)
	go func() {
		ticker := time.NewTicker(5 * time.Second)
		defer ticker.Stop()

		for range ticker.C {
			stats := s.db.Stat()
			poolTotalConns.Set(float64(stats.TotalConns()))
			poolIdleConns.Set(float64(stats.IdleConns()))
			poolAcquiredConns.Set(float64(stats.AcquiredConns()))
		}
	}()
}
