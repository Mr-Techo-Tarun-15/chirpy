-- name: DeleteChripWithID :exec
DELETE FROM chirps
WHERE ID = $1;