-- name: CreateUser :one
INSERT INTO users (id, created_at, updated_at, hashed_password, email)
VALUES (
           gen_random_uuid (), now(), now(), $1, $2
       )
RETURNING *;

-- name: DeleteAllUsers :exec
DELETE FROM users;

-- name: GetUserByEmail :one
SELECT *
FROM users
WHERE email = $1;

-- name: GetUserByID :one
SELECT *
FROM users
WHERE id = $1;

-- name: UpdateUser :exec
UPDATE users
SET email = $1, hashed_password = $2, updated_at = now()
WHERE id = $3;

-- name: UpgradeUserToChirpRed :exec
UPDATE users
SET is_chirpy_red = TRUE, updated_at = now()
WHERE id = $1;