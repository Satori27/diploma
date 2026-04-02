package main

import (
	"context"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/Satori27/diploma/outbox-service/internal/broker"
	"github.com/Satori27/diploma/outbox-service/internal/storage"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
	r := chi.NewRouter()
	r.Use(
		middleware.Logger,
		middleware.Recoverer,
	)
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGTERM)
	defer stop()

	storage := storage.NewStorage(ctx)
	defer storage.Close()
	r.Handle("/metrics", promhttp.Handler())

	broker := broker.NewBroker("checkout")
	defer broker.Close()
	logger := slog.Default()
	tickerDuration := getTickerDuration()
	ticker := time.NewTicker(tickerDuration)

	go func() {
		for range ticker.C {
			if ctx.Err() != nil {
				log.Println("Shutting down outbox processing...")
				return
			}
			_, err := storage.DoOutbox(ctx, broker)
			if err != nil {
				logger.Error("Error processing outbox", "error", err.Error())
			}
		}
	}()

	err := http.ListenAndServe(":3001", r)

	if err != nil {
		log.Fatal("Can't start server ", err.Error())
	}
}

func getTickerDuration() time.Duration {
	tickerDurationStr := os.Getenv("TICKER_DURATION_MS")
	if tickerDurationStr == "" {
		return 500 * time.Millisecond
	}
	tickerDurationMs, err := strconv.Atoi(tickerDurationStr)
	if err != nil {
		log.Printf("Invalid TICKER_DURATION_MS value: %s, using default 500ms", tickerDurationStr)
		return 500 * time.Millisecond
	}
	return time.Duration(tickerDurationMs) * time.Millisecond
}
