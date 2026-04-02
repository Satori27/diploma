-- name: InsertEventsBatch :exec
INSERT INTO e2e_events(order_id, sent_at)
SELECT unnest($1::uuid[]), now();

-- name: UpdateEventsBatch :exec
UPDATE e2e_events
SET received_at = now()
WHERE order_id = ANY($1::uuid[]);


-- name: DeleteEvents :exec
DELETE FROM e2e_events
WHERE sent_at < now() - interval '1 hour';


SELECT COUNT(*) FROM e2e_events
WHERE received_at NOT IS NULL;