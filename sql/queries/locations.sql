-- name: CreateLocation :one
INSERT INTO locations (name, description)
VALUES ($1, $2)
RETURNING *;

-- name: GetAllLocations :many
SELECT * FROM locations
WHERE is_active = true
ORDER BY created_at;

-- name: GetLocationFromID :one
SELECT * FROM locations
WHERE id = $1;

-- name: UpdateLocation :one
UPDATE locations
SET name = $2, description = $3
WHERE id = $1
RETURNING *;

-- name: DeleteLocation :exec
UPDATE locations
SET is_active = false
WHERE id = $1;
