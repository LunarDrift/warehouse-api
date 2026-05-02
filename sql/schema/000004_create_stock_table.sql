-- +goose Up
CREATE TABLE stock (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    item_id UUID NOT NULL,
    location_id UUID NOT NULL,
    quantity INT NOT NULL DEFAULT 0,
    UNIQUE (item_id, location_id),
    CONSTRAINT fk_items
    FOREIGN KEY (item_id) REFERENCES items (id),
    CONSTRAINT fk_locations
    FOREIGN KEY (location_id) REFERENCES locations (id)
);

-- +goose Down
DROP TABLE stock;
