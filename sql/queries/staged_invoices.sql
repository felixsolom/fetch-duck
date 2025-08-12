-- name: CreateStagedInvoice :one
INSERT INTO staged_invoices (
    id,
    user_id,
    gmail_message_id,
    gmail_thread_id,
    sender,
    subject,
    snippet,
    has_attachment,
    received_at,
    created_at,
    updated_at
) VALUES (
    ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?
)
RETURNING *; 
--

-- name: ListStagedInvoicesByUser :many
SELECT * FROM staged_invoices
WHERE 
    user_id = ? 
    AND status = 'pending_review'
    AND received_at >= ?
    AND received_at <= ?
    ORDER BY received_at DESC
LIMIT ?
OFFSET ?; 
--

-- name: GetStagedInvoice :one 
SELECT * FROM staged_invoices
WHERE id = ? AND user_id = ?;
--

-- name: UpdateStagedInvoiceStatus :exec
UPDATE staged_invoices
SET status = ?, updated_at = ?
WHERE id = ? AND user_id = ?; 
--

-- name: GetStagedInvoicesByMessageId :many
SELECT * FROM staged_invoices
WHERE user_id = ? AND gmail_message_id = ?;
--