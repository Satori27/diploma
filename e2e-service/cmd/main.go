package main

import (
	"context"
	"log"
	"log/slog"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"github.com/Satori27/diploma/e2e-service/internal/broker"
	"github.com/Satori27/diploma/e2e-service/internal/clients"
	e2etest "github.com/Satori27/diploma/e2e-service/internal/e2e-test"
	k8sclient "github.com/Satori27/diploma/e2e-service/internal/k8s-client"
	"github.com/Satori27/diploma/e2e-service/internal/migrations"
	"github.com/Satori27/diploma/e2e-service/internal/storage"
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

	migrations.RunMigrations(storage)

	broker := broker.NewBroker("checkout", storage)
	defer broker.Close()

	checkout := clients.NewCheckout()

	test := e2etest.NewE2ETest(storage, checkout)

	kuberClient, err := k8sclient.NewKubeClient()
	if err != nil {
		log.Fatal("Failed to create kube client: ", err.Error())
	}

	tickTest := time.NewTicker(5 * time.Second)
	tickDelete := time.NewTicker(7 * time.Second)
	tickPodDelete := time.NewTicker(7*time.Minute)
	go func() {
		for range tickTest.C {
			if ctx.Err() != nil {
				slog.Error("Stopping test runner", "error", ctx.Err())
				return
			}
			test.RunTests(ctx)
		}
	}()

	go func() {
		for range tickPodDelete.C {
			if ctx.Err() != nil {
				slog.Info("Stopping pod killer", "error", ctx.Err())
				return
			}
			kuberClient.KillCheckoutPod()

		}
	}()

	go func() {
		for range tickDelete.C {
			if ctx.Err() != nil {
				slog.Info("Stopping results deleter", "error", ctx.Err())
				return
			}
			test.DeleteResults(ctx)
		}
	}()

	err = http.ListenAndServe(":3002", r)

	if err != nil {
		log.Fatal("Can't start server ", err.Error())
	}
}
