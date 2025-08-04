-- +goose Up
CREATE TABLE users(
    id TEXT PRIMARY KEY,
    created_at INTEGER NOT NULL,
    updated_at INTEGER NOT NULL,
    email TEXT NOT NULL UNIQUE

);

-- +goose Down
DROP TABLE users;