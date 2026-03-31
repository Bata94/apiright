-- name: GetTodosByStatus :many
SELECT id, title, description, completed, priority, due_date, created_at, updated_at
FROM todos
WHERE completed = ?
ORDER BY priority DESC, created_at DESC;

-- name: GetTodosByPriority :many
SELECT id, title, description, completed, priority, due_date, created_at, updated_at
FROM todos
WHERE priority >= ?
ORDER BY priority DESC, created_at DESC;

-- name: GetOverdueTodos :many
SELECT id, title, description, completed, priority, due_date, created_at, updated_at
FROM todos
WHERE due_date < datetime('now') AND completed = FALSE
ORDER BY due_date ASC;

-- name: GetTodosCount :one
SELECT COUNT(*) as total, 
       SUM(CASE WHEN completed THEN 1 ELSE 0 END) as completed,
       SUM(CASE WHEN NOT completed AND due_date < datetime('now') THEN 1 ELSE 0 END) as overdue
FROM todos;
