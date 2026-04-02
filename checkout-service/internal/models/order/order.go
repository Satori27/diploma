package order

type CreateOrderRequest struct {
	ID   string `json:"id"`
	Data []byte `json:"data"`
}
