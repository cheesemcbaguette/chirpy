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