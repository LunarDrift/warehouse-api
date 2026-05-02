-- +goose Up
CREATE TABLE stock_movements (
    id UUID PRIMARY KEY NOT NULL DEFAULT gen_random_uuid(),
    item_id UUID NOT NULL,
    from_location_id UUID,
    to_location_id UUID,
    quantity INTEGER NOT NULL,
    moved_by UUID NOT NULL,
    moved_at TIMESTAMP NOT NULL DEFAULT now(),
    CONSTRAINT fk_items
      FOREIGN KEY (item_id) REFERENCES items (id),
    CONSTRAINT fk_from_location
      FOREIGN KEY (from_location_id) REFERENCES locations (id),
    CONSTRAINT fk_to_location
      FOREIGN KEY (to_location_id) REFERENCES locations (id),
    CONSTRAINT fk_moved_by
      FOREIGN KEY (moved_by) REFERENCES users (id)
);

-- +goose Down
DROP TABLE stock_movements;
