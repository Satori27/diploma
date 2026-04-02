-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS e2e_events  (
    order_id UUID PRIMARY KEY,
    sent_at TIMESTAMP,
    received_at TIMESTAMP
);

CREATE INDEX idx_e2e_events_lost
ON e2e_events(sent_at)
WHERE received_at IS NULL;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE e2e_events;
DROP INDEX idx_e2e_events_lost;
-- +goose StatementEnd
