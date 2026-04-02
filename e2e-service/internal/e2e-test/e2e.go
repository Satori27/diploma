package e2etest

import (
	"context"
	"log/slog"

	"github.com/Satori27/diploma/e2e-service/internal/clients"
	"github.com/Satori27/diploma/e2e-service/internal/storage"
	"github.com/google/uuid"
)

type E2ETest struct {
	storage         *storage.Storage
	checkoutService *clients.CheckoutService
}

func NewE2ETest(storage *storage.Storage, checkoutService *clients.CheckoutService) *E2ETest {
	return &E2ETest{
		storage:         storage,
		checkoutService: checkoutService,
	}
}

const testCount = 30

func (e *E2ETest) RunTests(ctx context.Context) {
	ids := make(uuid.UUIDs, 0, testCount)
	for i := 0; i < testCount; i++ {
		id := uuid.New()
		err := e.checkoutService.SendCheckout(ctx, &clients.CreateOrderRequest{
			ID:   id.String(),
			Data: []byte(id.String()),
		})
		if err != nil {
			slog.Error("Failed send checkout request: ", "error", err.Error())
			continue
		}
		ids = append(ids, id)
	}

	if len(ids) == 0 {
		return
	}

	err := e.storage.InsertOrders(ctx, ids)
	if err != nil {
		slog.Error("Failed to insert orders: ", "error", err.Error())
	}

	slog.Info("Sent msgs", "count", len(ids))
}


func (e *E2ETest) DeleteResults(ctx context.Context) {
	err := e.storage.DeleteOrders(ctx)
	if err != nil {
		slog.Error("Failed to delete orders: ", "error", err.Error())
	}
	slog.Info("Deleted orders")
}
