-- +goose Up
ALTER TABLE locations
ADD is_active BOOLEAN NOT NULL DEFAULT TRUE;

-- +goose Down
ALTER TABLE locations
DROP COLUMN is_active;
