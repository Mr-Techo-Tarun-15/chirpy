-- name: GetChirpFromID :one
SELECT *
FROM chirps
WHERE id = $1;