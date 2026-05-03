-- +goose Up
ALTER TABLE items
ADD is_active BOOLEAN NOT NULL DEFAULT TRUE;

-- +goose Down
ALTER TABLE items
DROP COLUMN is_active;
