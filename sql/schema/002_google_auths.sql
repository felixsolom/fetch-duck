-- +goose Up

CREATE TABLE google_auths(
    user_id TEXT PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
    created_at INTEGER NOT NULL,
    updated_at INTEGER NOT NULL,
    token_expiry INTEGER NOT NULL,
    access_token TEXT NOT NULL,
    refresh_token TEXT NOT NULL
);

-- +goose Down 
DROP TABLE google_auths; 
