-- name: AddUser :one
INSERT INTO player_credentials (login_token) VALUES ($1) RETURNING id;

-- name: FindUserByLoginToken :one
SELECT id FROM player_credentials WHERE login_token = $1;
