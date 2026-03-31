# Blog Platform - Complete Example

A complete blog platform demonstrating relationships, custom queries, and complex data models.

## Features Demonstrated

| Feature | Status | Description |
|---------|--------|-------------|
| Multiple Tables | ✅ | Users, Posts, Comments with relationships |
| Foreign Keys | ✅ | Posts → Users, Comments → Posts/Users |
| Custom JOINs | ✅ | Custom queries with multi-table joins |
| Aggregation | ✅ | COUNT, SUM queries |
| Hierarchical Data | ✅ | Comments with parent_id for nesting |

## Schema Overview

```
users (1) ──────< posts (N)
  │                   │
  │                   │
  └────< comments (N)─┘
            │
            └── parent_id (self-reference for threading)
```

## Quick Start

```bash
cd examples/blog
go mod tidy
go run main.go --dev
```

## API Endpoints

### Users
- `GET /v1/users` - List all users
- `POST /v1/users` - Create user
- `GET /v1/users/{id}` - Get user by ID
- `PUT /v1/users/{id}` - Update user
- `DELETE /v1/users/{id}` - Delete user

### Posts
- `GET /v1/posts` - List published posts
- `POST /v1/posts` - Create post
- `GET /v1/posts/{id}` - Get post by ID
- `GET /v1/posts/slug/{slug}` - Get post by slug
- `PUT /v1/posts/{id}` - Update post
- `DELETE /v1/posts/{id}` - Delete post

### Comments
- `GET /v1/comments?post_id=1` - List comments for a post
- `POST /v1/comments` - Create comment
- `GET /v1/comments/{id}` - Get comment by ID
- `PUT /v1/comments/{id}` - Update comment
- `DELETE /v1/comments/{id}` - Delete comment

## Custom Queries

The `queries/blog.sql` file contains custom queries demonstrating:

### Get Posts with Author Info
```sql
SELECT p.*, u.username, u.display_name, u.avatar_url
FROM posts p
JOIN users u ON p.author_id = u.id
WHERE p.published = TRUE
```

### Get Comment Count
```sql
SELECT 
    COUNT(*) as total,
    SUM(CASE WHEN parent_id IS NULL THEN 1 ELSE 0 END) as top_level
FROM comments
WHERE post_id = ?
```

### Get User with Post Count
```sql
SELECT u.*, COUNT(p.id) as post_count
FROM users u
LEFT JOIN posts p ON u.id = p.author_id
WHERE u.id = ?
GROUP BY u.id
```

## Example API Calls

### Create User
```bash
curl -X POST http://localhost:8081/v1/users \
  -H "Content-Type: application/json" \
  -d '{"username":"john","email":"john@example.com","display_name":"John Doe"}'
```

### Create Post
```bash
curl -X POST http://localhost:8081/v1/posts \
  -H "Content-Type: application/json" \
  -d '{
    "author_id": 1,
    "title": "My First Post",
    "slug": "my-first-post",
    "content": "Hello World!",
    "summary": "A brief introduction",
    "published": true
  }'
```

### Add Comment
```bash
curl -X POST http://localhost:8081/v1/comments \
  -H "Content-Type: application/json" \
  -d '{"post_id":1,"author_id":1,"content":"Great post!"}'
```

### Reply to Comment
```bash
curl -X POST http://localhost:8081/v1/comments \
  -H "Content-Type: application/json" \
  -d '{"post_id":1,"author_id":2,"parent_id":1,"content":"Thanks!"}'
```

## Learning Points

1. **Relationships**: See how foreign keys create natural joins
2. **Custom Queries**: Add complex queries in `queries/blog.sql`
3. **Aggregation**: Use custom queries for counts, sums, averages
4. **Hierarchical Data**: Comments demonstrate self-referencing tables

## Next Steps

- Add tags/categories tables
- Add pagination to custom queries
- Implement search functionality
- Add post likes/favorites
