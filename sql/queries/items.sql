-- name: CreateItem :one
INSERT INTO items (sku, name, description, low_stock_threshold)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: GetAllItems :many
SELECT * FROM items
ORDER BY created_at;

-- name: GetItemFromID :one
SELECT * FROM items
WHERE id = $1;

-- name: UpdateItem :one
UPDATE items
SET sku = $2, name = $3, description = $4
WHERE id = $1
RETURNING *;

-- name: DeleteItem :exec
DELETE FROM items
WHERE id = $1;
