-- +goose Up
ALTER TABLE orders ADD COLUMN created_at TIMESTAMPTZ DEFAULT NOW();
ALTER TABLE orders ADD COLUMN updated_at TIMESTAMPTZ DEFAULT NOW();


-- +goose Down
ALTER TABLE orders DROP COLUMN created_at;
ALTER TABLE orders DROP COLUMN updated_at;
