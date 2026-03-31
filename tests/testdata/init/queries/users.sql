-- name: GetUserByEmail :one
SELECT id, name, email, created_at, updated_at
FROM users
WHERE email = ? LIMIT 1;

-- name: GetUsersCreatedAfter :many
SELECT id, name, email, created_at, updated_at
FROM users
WHERE created_at > ? 
ORDER BY created_at DESC;
