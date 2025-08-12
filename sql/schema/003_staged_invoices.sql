-- +goose Up

CREATE TABLE staged_invoices(
    id TEXT PRIMARY KEY,
    user_id TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,

    gmail_message_id TEXT NOT NULL,
    gmail_thread_id TEXT NOT NULL,

    status TEXT NOT NULL DEFAULT 'pending_review',

    sender TEXT NOT NULL,
    subject TEXT NOT NULL,
    snippet TEXT,

    has_attachment BOOLEAN NOT NULL DEFAULT FALSE,

    received_at INTEGER NOT NULL,
    created_at INTEGER NOT NULL,
    updated_at INTEGER NOT NULL
);

CREATE INDEX idx_staged_invoices_user_status ON staged_invoices (user_id, status);

-- +goose Down 
DROP TABLE staged_invoices; 