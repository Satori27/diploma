# Stage 1: Build
FROM golang:1.25-alpine AS builder
WORKDIR /app
COPY go.mod ./
COPY go.sum ./
RUN go mod download
COPY . .
RUN go build -o checkout-service cmd/main.go

# Stage 2: Minimal image
FROM alpine:3.18
WORKDIR /app
COPY --from=builder /app/checkout-service .
COPY migrations/ ./migrations/
CMD ["./checkout-service"]