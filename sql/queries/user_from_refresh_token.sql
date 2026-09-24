-- name: UserFromRefreshToken :one
SELECT USER_ID FROM refresh_tokens
WHERE TOKEN = $1 
AND EXPIRES_AT > NOW()
AND REVOKED_AT IS NULL;
