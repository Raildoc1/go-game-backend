-- name: AddEvent :exec
INSERT INTO outbox (topic, payload, created_at) VALUES ($1, $2, NOW());

-- name: FetchEvents :many
SELECT id, topic, payload FROM outbox WHERE processed_at IS NULL ORDER BY id LIMIT $1;

-- name: MarkProcessed :exec
UPDATE outbox SET processed_at = NOW() WHERE id = $1;
