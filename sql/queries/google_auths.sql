-- name: GetGoogleAuthByUserID :one
SELECT user_id, access_token, refresh_token, token_expiry, created_at, updated_at 
FROM google_auths
WHERE user_id = ?;
--

-- name: UpsertGoogleAuth :exec
INSERT INTO google_auths(
    user_id,
    access_token,
    refresh_token,
    token_expiry,
    created_at,
    updated_at
) VALUES (
    ?, ?, ?, ?, ?, ?
) 
ON CONFLICT(user_id) DO UPDATE SET 
    access_token = excluded.access_token,
    refresh_token = excluded.refresh_token,
    token_expiry = excluded.token_expiry,
    updated_at = excluded.updated_at;
--