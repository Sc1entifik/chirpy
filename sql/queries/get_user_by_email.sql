-- name: GetUserByEmail :one
SELECT * FROM USERS
WHERE email = $1; 
