-- name: ChirpyRedUpgrade :exec

UPDATE users
SET IS_CHIRPY_RED = TRUE
WHERE ID = $1;
