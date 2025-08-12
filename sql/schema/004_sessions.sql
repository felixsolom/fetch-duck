-- +goose Up

CREATE TABLE sessions(
    token TEXT PRIMARY KEY,
    user_id TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    expiry INTEGER NOT NULL,
    created_at INTEGER NOT NULL
);

-- +goose Down
DROP TABLE sessions;