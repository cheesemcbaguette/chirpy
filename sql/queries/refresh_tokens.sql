-- name: GetRefreshToken :one
SELECT *
FROM refresh_tokens
WHERE  token = $1;

-- name: CreateRefreshToken :exec
INSERT INTO refresh_tokens (token, created_at, updated_at, user_id, expires_at, revoked_at)
VALUES ($1, now(), now(), $2, $3, $4);

-- name: RevokeRefreshToken :exec
UPDATE refresh_tokens
SET revoked_at = NOW(),
    updated_at = NOW()
WHERE token = $1 AND revoked_at IS NULL;