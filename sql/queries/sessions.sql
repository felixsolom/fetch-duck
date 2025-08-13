-- name: CreateSession :one
INSERT INTO sessions (
    token, user_id, expiry, created_at
) VALUES (
    ?, ?, ?, ?
) RETURNING *;
--

-- name: GetUserBySessionToken :one
SELECT users.* FROM users
JOIN sessions ON users.id = sessions.user_id
WHERE sessions.token = ? AND sessions.expiry > ?;
--