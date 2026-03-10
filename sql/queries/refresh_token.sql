-- name: CreateRefreshToken :one
INSERT INTO refresh_tokens (token, created_at, updated_at, user_id, expires_at, revoked_at)
Values (
    $1,
    Now(),
    NOW(),
    $2,
    $3,
    NULL
)
RETURNING *;




-- name: GetUserFromRefreshToken :one
SELECT * FROM refresh_tokens
WHERE token = $1;


-- name: RevokeRefreshToken :exec
UPDATE refresh_tokens
SET updated_at = Now(),
    revoked_at = Now()
WHERE token = $1
    AND user_id = $2;