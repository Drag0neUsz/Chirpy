-- name: CreateRefreshToken :one
INSERT INTO refresh_tokens (token, user_id, expires_at)
VALUES ($1, $2, NOW() + INTERVAL '60 days')
RETURNING *;

-- name: GetUserIDFromRefreshToken :one
SELECT user_id, revoked_at, expires_at FROM refresh_tokens WHERE token = $1;

-- name: RevokeRefreshToken :exec
UPDATE refresh_tokens SET revoked_at = NOW(), updated_at = NOW() WHERE token = $1;