package main

import (
	"context"
	"log"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"github.com/Satori27/diploma/checkout-service/internal/handlers/order"
	"github.com/Satori27/diploma/checkout-service/internal/migrations"
	"github.com/Satori27/diploma/checkout-service/internal/storage"
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

	migrations.RunMigrations(storage)

	
	r.Handle("/metrics", promhttp.Handler())

	r.Post("/order", order.NewCreateOrder(ctx, storage))

	go func() {
		ticker := time.NewTicker(10 * time.Second)
		defer ticker.Stop()

		for range ticker.C {
			if ctx.Err() != nil {
				return
			}
			storage.DeleteOrder(ctx) 
		}
	}()

	err := http.ListenAndServe(":3000", r)

	if err != nil {
		log.Fatal("Can't start server ", err.Error())
	}
}
