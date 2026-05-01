-- name: CreateUser :one
INSERT INTO users (username, hashed_password)
VALUES ($1, $2)
RETURNING *;

-- name: UpdateUserRole :one
UPDATE users
SET role = $2
WHERE id = $1
RETURNING *;

-- name: GetUserFromID :one
SELECT * FROM users
WHERE id = $1;

-- name: GetUserFromName :one
SELECT * FROM users
WHERE username = $1;
