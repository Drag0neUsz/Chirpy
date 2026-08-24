-- name: CreateChirp :one
INSERT INTO chirps (body, user_id)
VALUES ($1, $2)
RETURNING *;


-- name: GetChirps :many
SELECT * FROM chirps
ORDER BY created_at ASC;

-- name: GetChirpsByAuthorID :many
SELECT * FROM chirps
WHERE user_id = $1
ORDER BY created_at ASC;

-- name: GetChirp :one
SELECT * FROM chirps
WHERE id = $1 LIMIT 1;

-- name: DeleteChirp :exec
DELETE FROM chirps
WHERE id = $1;