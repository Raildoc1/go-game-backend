-- name: UpsertPlayer :exec
INSERT INTO players (user_id, nickname) VALUES ($1, $2)
ON CONFLICT (user_id) DO UPDATE SET nickname=COALESCE(players.nickname, EXCLUDED.nickname);

-- name: GetPlayer :one
SELECT user_id, nickname FROM players WHERE user_id = $1;
