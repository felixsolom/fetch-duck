-- name: CreateUser :exec
INSERT INTO users (id, created_at, updated_at, email)
VALUES (
    ?,
    ?,
    ?,
    ?
);
--

-- name: GetUser :one
SELECT id, created_at, updated_at, email
FROM users 
WHERE email = ?;
--

-- name: GetUsers :many
SELECT * FROM users;
-- 