package storage

import (
	"context"
	"log"
	"log/slog"
	"os"
	"strconv"
	"time"

	db "github.com/Satori27/diploma/outbox-service/internal/db/queries"
	queries "github.com/Satori27/diploma/outbox-service/internal/db/queries"
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
	config.MaxConnLifetime = 2 * time.Minute 
	config.MaxConnIdleTime = 1 * time.Minute
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
	prometheus.MustRegister(
		poolTotalConns, 
		poolIdleConns, 
		poolAcquiredConns,
		outboxProcessed,
		outboxLag,
		outboxBatchDuration,
		outboxPending,
		outboxPublishErrors,
		outboxBatchSize,
	)
	go func() {
		ticker := time.NewTicker(5 * time.Second)
		defer ticker.Stop()

		for range ticker.C {
			stats := s.db.Stat()
			poolTotalConns.Set(float64(stats.TotalConns()))
			poolIdleConns.Set(float64(stats.IdleConns()))
			poolAcquiredConns.Set(float64(stats.AcquiredConns()))
			s.trackPendingEvents()
		}
	}()
}

type Publisher interface {
	Publish(message [][]byte) error
}

func (s *Storage) trackPendingEvents() {

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)

	count, err := s.queries.CountPending(ctx)
	if err == nil {
		outboxPending.Set(float64(count))
	}

	cancel()
}


func (s *Storage) DoOutbox(ctx context.Context, publisher Publisher) ([]db.GetOutboxRow, error) {
	logger := slog.Default()
	timeout := getDBTimeout()
	start := time.Now()
	defer func() {
		outboxBatchDuration.Observe(float64(time.Since(start).Seconds()))
	}()	
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	queries := s.queries.WithTx(tx)
	limit, err := getBatchSize()
	if err != nil {
		return nil, err
	}

	events, err := queries.GetOutbox(ctx, int32(limit))
	if err != nil {
		return nil, err
	}
	if len(events) == 0 {
		return nil, nil
	}
	outboxBatchSize.Observe(float64(len(events)))
	logger.Info("Fetched outbox events", "count", len(events))

	if len(events) == 0 {
		return events, nil
	}

	payloads := make([][]byte, 0, len(events))
	for _, event := range events {
		payloads = append(payloads, event.Payload)
	}

	timeKafkaLag := time.Now()
	err = publisher.Publish(payloads)
	outboxLag.Observe(float64(time.Since(timeKafkaLag).Seconds()))
	if err != nil {
		outboxPublishErrors.Inc()
		outboxProcessed.WithLabelValues("error").Add(float64(len(events)))
		return nil, err
	}

	ids := make([]int64, 0, len(events))
	for _, event := range events {
		ids = append(ids, event.ID)
	}

	err = queries.UpdateOutbox(ctx, ids)
	if err != nil {
		outboxProcessed.WithLabelValues("error").Add(float64(len(events)))
		return nil, err
	}

	err = tx.Commit(ctx)
	if err != nil {
		outboxProcessed.WithLabelValues("error").Add(float64(len(events)))
		return nil, err
	}

	logger.Info("Processed outbox events", "count", len(events))
	outboxProcessed.WithLabelValues("success").Add(float64(len(events)))

	return events, err
}

func getDBTimeout() time.Duration {
	timeoutStr := os.Getenv("DB_TIMEOUT_MS")
	if timeoutStr == "" {
		timeoutStr = "5000"
	}
	timeoutMs, err := strconv.Atoi(timeoutStr)
	if err != nil {
		log.Printf("Invalid DB_TIMEOUT_MS value: %v, using default 5000ms", err)
		timeoutMs = 5000
	}
	return time.Duration(timeoutMs) * time.Millisecond
}

func getBatchSize() (int, error) {
	batchSize := os.Getenv("BATCH_SIZE")
	if batchSize == "" {
		batchSize = "1000"
	}
	limit, err := strconv.Atoi(batchSize)
	if err != nil {
		return 0, err
	}
	return limit, nil
}
