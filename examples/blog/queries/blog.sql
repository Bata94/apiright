-- name: GetPostsWithAuthor :many
SELECT 
    p.id, p.author_id, p.title, p.slug, p.content, p.summary,
    p.featured_image, p.published, p.published_at, p.created_at, p.updated_at,
    u.username, u.display_name, u.avatar_url
FROM posts p
JOIN users u ON p.author_id = u.id
WHERE p.published = TRUE
ORDER BY p.published_at DESC;

-- name: GetPostWithAuthor :one
SELECT 
    p.id, p.author_id, p.title, p.slug, p.content, p.summary,
    p.featured_image, p.published, p.published_at, p.created_at, p.updated_at,
    u.username, u.display_name, u.avatar_url
FROM posts p
JOIN users u ON p.author_id = u.id
WHERE p.id = ?;

-- name: GetPostBySlug :one
SELECT 
    p.id, p.author_id, p.title, p.slug, p.content, p.summary,
    p.featured_image, p.published, p.published_at, p.created_at, p.updated_at,
    u.username, u.display_name, u.avatar_url
FROM posts p
JOIN users u ON p.author_id = u.id
WHERE p.slug = ?;

-- name: GetPostsByAuthor :many
SELECT 
    p.id, p.author_id, p.title, p.slug, p.content, p.summary,
    p.featured_image, p.published, p.published_at, p.created_at, p.updated_at,
    u.username, u.display_name, u.avatar_url
FROM posts p
JOIN users u ON p.author_id = u.id
WHERE p.author_id = ?
ORDER BY p.created_at DESC;

-- name: GetCommentsWithAuthor :many
SELECT 
    c.id, c.post_id, c.author_id, c.parent_id, c.content, c.created_at, c.updated_at,
    u.username, u.display_name, u.avatar_url
FROM comments c
JOIN users u ON c.author_id = u.id
WHERE c.post_id = ?
ORDER BY c.created_at ASC;

-- name: GetCommentCount :one
SELECT 
    COUNT(*) as total,
    SUM(CASE WHEN parent_id IS NULL THEN 1 ELSE 0 END) as top_level
FROM comments
WHERE post_id = ?;

-- name: GetPostsStats :one
SELECT 
    COUNT(*) as total_posts,
    SUM(CASE WHEN published THEN 1 ELSE 0 END) as published_posts,
    SUM(CASE WHEN NOT published THEN 1 ELSE 0 END) as draft_posts
FROM posts;

-- name: GetUserWithPostCount :one
SELECT 
    u.id, u.username, u.email, u.display_name, u.bio, u.avatar_url,
    u.created_at, u.updated_at,
    COUNT(p.id) as post_count
FROM users u
LEFT JOIN posts p ON u.id = p.author_id
WHERE u.id = ?
GROUP BY u.id;
