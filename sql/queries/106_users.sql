-- name: StorePassword :exec
UPDATE users
SET hashed_password = $1
WHERE email = $2;