-- +goose Up
ALTER TABLE users
ADD updated_at TIMESTAMP NOT NULL DEFAULT now();

-- +goose Down
ALTER TABLE users
DROP COLUMN updated_at;
