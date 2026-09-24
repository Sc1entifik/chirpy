-- name: RevokeRefreshToken :exec

UPDATE refresh_tokens
SET REVOKED_AT = NOW(), UPDATED_AT = NOW()
WHERE TOKEN = $1;
