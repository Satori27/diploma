package storage

import (
	"context"
	"fmt"

	log "log/slog"

	queries "github.com/Satori27/diploma/checkout-service/internal/db/queries"
	"github.com/jackc/pgx/v5/pgtype"
)

func (s *Storage) CreateOrder(ctx context.Context, data []byte) error {
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	qtx := s.queries.WithTx(tx)

	order, err := qtx.CreateOrder(ctx, data)
	if err != nil {
		return err
	}

	err = qtx.InsertOutboxEvent(ctx, queries.InsertOutboxEventParams{
		AggregateType: "order",
		AggregateID:   int64(order.ID),
		EventType:     "order_created",
		Payload:       data,
	})
	if err != nil {
		return err
	}

	return tx.Commit(ctx)
}

const hourInMicroseconds = 3600 * 1e6

func (s *Storage) DeleteOrder(ctx context.Context) {
	tx, err := s.db.Begin(ctx)
	if err != nil {
		log.Error(fmt.Sprintf("Failed to begin transaction DeleteOrder: %s", err.Error()))
		return
	}
	defer tx.Rollback(ctx)

	qtx := s.queries.WithTx(tx)

	err = qtx.DeleteOldOrdersBatch(ctx, queries.DeleteOldOrdersBatchParams{
		Column1: pgtype.Interval{Microseconds: hourInMicroseconds},
		Limit: 1000,
	})
	if err != nil {
		log.Error(fmt.Sprintf("Failed to delete old orders: %s", err))
		return
	}

	err = qtx.DeleteProcessedOutboxBatch(ctx, queries.DeleteProcessedOutboxBatchParams{
		Column1: pgtype.Interval{Microseconds: hourInMicroseconds},
		Limit: 100,
	})

	if err != nil {
		log.Error(fmt.Sprintf("Failed to delete old outbox events: %s", err.Error()))
	}

	if err := tx.Commit(ctx); err != nil {
		log.Error(fmt.Sprintf("Failed to commit transaction DeleteOrder: %s", err.Error()))
	}
}