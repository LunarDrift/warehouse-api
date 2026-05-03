-- name: GetAllStockWithThreshold :many
SELECT s.*, i.name, i.sku, i.low_stock_threshold
FROM stock s
JOIN items i ON s.item_id = i.id
ORDER BY s.location_id;

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

-- name: GetLowStockItems :many
SELECT s.*, i.name, i.sku, i.low_stock_threshold
FROM stock s
JOIN items i ON s.item_id = i.id
WHERE s.quantity < i.low_stock_threshold
AND i.low_stock_threshold > 0;
