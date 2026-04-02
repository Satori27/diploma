-- name: CreateOrder :one
-- name: CreateOrder :one
INSERT INTO orders (data)
VALUES ($1)
RETURNING *;

-- name: InsertOutboxEvent :exec
INSERT INTO outbox (
    aggregate_type,
    aggregate_id,
    event_type,
    payload
)
VALUES ($1, $2, $3, $4);

-- name: DeleteProcessedOutboxBatch :exec
DELETE FROM outbox
WHERE id IN (
    SELECT id
    FROM outbox
    WHERE processed_at IS NOT NULL
    AND processed_at < now() - $1::interval
    LIMIT $2
);

-- name: DeleteOldOrdersBatch :exec
DELETE FROM orders
WHERE id IN (
    SELECT id
    FROM orders
    WHERE created_at < now() - $1::interval
    LIMIT $2
);
