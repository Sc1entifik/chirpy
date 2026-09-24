-- name: UpdateUserEmailAndPassword :exec
UPDATE users
SET EMAIL = $2, HASHED_PASSWORD = $3
WHERE ID = $1;
