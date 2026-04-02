package order

import (
	"context"
	"encoding/json"
	"log"
	"net/http"

	"github.com/Satori27/diploma/checkout-service/internal/models/order"
)

type CreateOrderStorage interface {
	CreateOrder(ctx context.Context, data []byte) error
}

func NewCreateOrder(ctx context.Context, storage CreateOrderStorage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req order.CreateOrderRequest
		err := json.NewDecoder(r.Body).Decode(&req)
		defer r.Body.Close()
		if err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			log.Printf("Error decoding request body: %v", err)
			return
		}
		log.Printf("Received order creation request: %s", req.ID)
		log.Printf("Request data: %s", string(req.Data))
		err = storage.CreateOrder(ctx, []byte(req.Data))
		if err != nil {
			http.Error(w, "Failed to create order", http.StatusInternalServerError)
			log.Printf("Error creating order: %v", err)
			return
		}
		w.WriteHeader(http.StatusCreated)
	}

}
