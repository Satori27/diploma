# Stage 1: Build
FROM golang:1.25-alpine AS builder
WORKDIR /app
COPY go.mod ./
COPY go.sum ./
RUN go mod download
COPY . .
RUN go build -o e2e-service cmd/main.go

# Stage 2: Minimal image
FROM alpine:3.18
WORKDIR /app
COPY --from=builder /app/e2e-service .
COPY migrations/ ./migrations/
CMD ["./e2e-service"]