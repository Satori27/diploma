package migrations

import (
	"log"

	"github.com/Satori27/diploma/checkout-service/internal/storage"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
)

func RunMigrations(st *storage.Storage) {
	if err := goose.SetDialect("pgx"); err != nil {
		log.Fatalf("Failed to set dialect: %v", err)
	}
	db := stdlib.OpenDBFromPool(st.GetPool())
	err := goose.Up(db, "./migrations")
	if err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}
}
