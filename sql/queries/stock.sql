-- name: GetAllStock :many
SELECT * FROM stock
ORDER BY location_id;

-- name: GetStockForItemAllLocations :many
SELECT * FROM stock
WHERE item_id =
$1;

-- name: GetStockForItemSpecificLocation :many
SELECT * FROM stock
WHERE location_id = $1;

-- name: MoveStock :one
UPDATE stock
  SET quantity = quantity + $1
  WHERE item_id = $2 AND location_id = $3
RETURNING *;

-- name: ReceiveStock :one
INSERT INTO stock (
  item_id,
  location_id,
  quantity
) VALUES ($1, $2, $3)
ON CONFLICT (item_id, location_id)
DO UPDATE SET quantity = stock.quantity + EXCLUDED.quantity
RETURNING *;
