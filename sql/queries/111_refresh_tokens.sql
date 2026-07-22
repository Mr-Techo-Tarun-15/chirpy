-- name: UpdateRefreshTokenUpdate :exec
UPDATE refresh_tokens
SET updated_at = $1
WHERE token = $2;