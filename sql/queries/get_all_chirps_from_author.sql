-- name: GetAllChirpsFromAuthor :many

SELECT *
FROM Chirps
WHERE user_id = $1
ORDER BY created_at ASC;

