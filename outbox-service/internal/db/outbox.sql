-- name: GetOutbox :many
SELECT id, aggregate_type, aggregate_id, event_type, payload
FROM outbox
WHERE processed_at IS NULL
ORDER BY created_at
LIMIT $1
FOR UPDATE SKIP LOCKED;

-- name: UpdateOutbox :exec
UPDATE outbox
SET processed_at = now()
WHERE id = ANY($1::bigint[]);

-- name: CountPending :one
SELECT count(*) FROM outbox WHERE processed_at IS NULL;
