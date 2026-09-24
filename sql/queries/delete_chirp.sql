-- name: DeleteChirp :exec
DELETE FROM chirps 
WHERE ID = $1 
AND USER_ID = $2;
