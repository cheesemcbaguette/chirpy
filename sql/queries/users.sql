-- name: CreateUser :one
INSERT INTO users (id, created_at, updated_at, hashed_password, email)
VALUES (
           gen_random_uuid (), now(), now(), $1, $2
       )
RETURNING *;

-- name: DeleteAllUsers :exec
DELETE FROM users;

-- name: GetUserByEmail :one
SELECT id, email, created_at, updated_at, hashed_password
FROM users
WHERE email = $1;

-- name: UpdateUser :exec
UPDATE users
SET email = $1, hashed_password = $2, updated_at = now()
WHERE id = $3;