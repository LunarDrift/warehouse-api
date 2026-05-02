-- name: CreateMovement :one
INSERT INTO stock_movements (item_id, from_location_id, to_location_id, quantity, moved_by)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: GetAllMovements :many
SELECT * FROM stock_movements
ORDER BY moved_at DESC;

-- name: GetMovementsForItem :many
SELECT * FROM stock_movements
WHERE item_id = $1
ORDER BY moved_at DESC;
