-- name: GetUser :one
SELECT * FROM users WHERE id = ? LIMIT 1;

-- name: GetUserByEmail :one
SELECT * FROM users WHERE email = ? LIMIT 1;

-- name: GetUserByUsername :one
SELECT * FROM users WHERE username = ? LIMIT 1;

-- name: ListUsers :many
SELECT * FROM users ORDER BY created_at DESC;

-- name: CreateUser :one
INSERT INTO users (username, email, password_hash)
VALUES (?, ?, ?)
RETURNING *;

-- name: UpdateUser :one
UPDATE users 
SET username = ?, email = ?, updated_at = CURRENT_TIMESTAMP
WHERE id = ?
RETURNING *;

-- name: DeleteUser :exec
DELETE FROM users WHERE id = ?;

-- name: GetPost :one
SELECT * FROM posts WHERE id = ? LIMIT 1;

-- name: ListPosts :many
SELECT * FROM posts ORDER BY created_at DESC;

-- name: ListPostsByUser :many
SELECT * FROM posts WHERE user_id = ? ORDER BY created_at DESC;

-- name: ListPublishedPosts :many
SELECT * FROM posts WHERE published = TRUE ORDER BY created_at DESC;

-- name: CreatePost :one
INSERT INTO posts (user_id, title, content, published)
VALUES (?, ?, ?, ?)
RETURNING *;

-- name: UpdatePost :one
UPDATE posts 
SET title = ?, content = ?, published = ?, updated_at = CURRENT_TIMESTAMP
WHERE id = ?
RETURNING *;

-- name: DeletePost :exec
DELETE FROM posts WHERE id = ?;

-- name: GetComment :one
SELECT * FROM comments WHERE id = ? LIMIT 1;

-- name: ListCommentsByPost :many
SELECT * FROM comments WHERE post_id = ? ORDER BY created_at ASC;

-- name: CreateComment :one
INSERT INTO comments (post_id, user_id, content)
VALUES (?, ?, ?)
RETURNING *;

-- name: UpdateComment :one
UPDATE comments 
SET content = ?
WHERE id = ?
RETURNING *;

-- name: DeleteComment :exec
DELETE FROM comments WHERE id = ?;

-- name: GetTag :one
SELECT * FROM tags WHERE id = ? LIMIT 1;

-- name: GetTagByName :one
SELECT * FROM tags WHERE name = ? LIMIT 1;

-- name: ListTags :many
SELECT * FROM tags ORDER BY name ASC;

-- name: CreateTag :one
INSERT INTO tags (name)
VALUES (?)
RETURNING *;

-- name: DeleteTag :exec
DELETE FROM tags WHERE id = ?;

-- name: AddPostTag :exec
INSERT INTO post_tags (post_id, tag_id) VALUES (?, ?);

-- name: RemovePostTag :exec
DELETE FROM post_tags WHERE post_id = ? AND tag_id = ?;

-- name: GetPostTags :many
SELECT t.* FROM tags t
JOIN post_tags pt ON t.id = pt.tag_id
WHERE pt.post_id = ?
ORDER BY t.name ASC;

-- name: GetPostsByTag :many
SELECT p.* FROM posts p
JOIN post_tags pt ON p.id = pt.post_id
WHERE pt.tag_id = ?
ORDER BY p.created_at DESC;