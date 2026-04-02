package clients

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"time"
)

type CheckoutService struct {
	dns string
}

type CreateOrderRequest struct {
	ID   string `json:"id"`
	Data []byte `json:"data"`
}

func NewCheckout() *CheckoutService {
	return &CheckoutService{
		dns: os.Getenv("CHECKOUT_SERVICE_DNS"),
	}
}

func (s *CheckoutService) SendCheckout(ctx context.Context, req *CreateOrderRequest) error {
    data, err := json.Marshal(req)
    if err != nil {
        return err
    }

    urlStr, err := url.JoinPath(s.dns, "/order")
    if err != nil {
        return err
    }

    client := &http.Client{
        Timeout: 1 * time.Second,
    }

    var resp *http.Response
    resp, err = client.Post(urlStr, "application/json", bytes.NewBuffer(data))

    if err != nil {
        return fmt.Errorf("failed to send checkout request: %w", err)
    }
    defer resp.Body.Close()

    if resp.StatusCode >= 300 {
        return fmt.Errorf("checkout service returned status %d", resp.StatusCode)
    }

    return nil
}
