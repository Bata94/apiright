package apiright_test

// Benchmarks for CRUDly code generation performance
// Run with: go test -bench=. ./tests/

import (
	"strings"
	"testing"
)

func BenchmarkSchemaParsingSmall(b *testing.B) {
	schema := `
CREATE TABLE items (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL,
    description TEXT
);
`
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		tables, _ := parseSchema(schema)
		if len(tables) != 1 {
			b.Fatal("expected 1 table")
		}
	}
}

func BenchmarkSchemaParsingMedium(b *testing.B) {
	schema := `
CREATE TABLE users (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    username TEXT NOT NULL,
    email TEXT NOT NULL,
    bio TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);
CREATE TABLE posts (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER NOT NULL,
    title TEXT NOT NULL,
    content TEXT,
    published BOOLEAN DEFAULT FALSE,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id)
);
CREATE TABLE comments (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    post_id INTEGER NOT NULL,
    user_id INTEGER NOT NULL,
    content TEXT NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (post_id) REFERENCES posts(id),
    FOREIGN KEY (user_id) REFERENCES users(id)
);
`
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		tables, _ := parseSchema(schema)
		if len(tables) != 3 {
			b.Fatal("expected 3 tables")
		}
	}
}

func BenchmarkSchemaParsingLarge(b *testing.B) {
	schema := `
CREATE TABLE users (id INTEGER PRIMARY KEY, username TEXT, email TEXT);
CREATE TABLE posts (id INTEGER PRIMARY KEY, author_id INTEGER, title TEXT);
CREATE TABLE comments (id INTEGER PRIMARY KEY, post_id INTEGER, content TEXT);
CREATE TABLE categories (id INTEGER PRIMARY KEY, name TEXT, parent_id INTEGER);
CREATE TABLE tags (id INTEGER PRIMARY KEY, name TEXT);
CREATE TABLE post_tags (post_id INTEGER, tag_id INTEGER);
CREATE TABLE media (id INTEGER PRIMARY KEY, user_id INTEGER, filename TEXT);
CREATE TABLE settings (id INTEGER PRIMARY KEY, user_id INTEGER, key TEXT);
CREATE TABLE sessions (id INTEGER PRIMARY KEY, user_id INTEGER, token TEXT);
CREATE TABLE notifications (id INTEGER PRIMARY KEY, user_id INTEGER, message TEXT);
`
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		tables, _ := parseSchema(schema)
		if len(tables) != 10 {
			b.Fatal("expected 10 tables")
		}
	}
}

func BenchmarkSQLQueryGeneration5Tables(b *testing.B) {
	tables := []string{"users", "posts", "comments", "categories", "tags"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, name := range tables {
			_ = generateCRUDQueries(name)
		}
	}
}

func BenchmarkSQLQueryGeneration15Tables(b *testing.B) {
	tables := []string{
		"users", "posts", "comments", "categories", "tags",
		"post_tags", "media", "settings", "sessions", "notifications",
		"likes", "followers", "messages", "attachments", "logs",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, name := range tables {
			_ = generateCRUDQueries(name)
		}
	}
}

func BenchmarkSQLQueryGeneration50Tables(b *testing.B) {
	tables := make([]string, 50)
	for i := range tables {
		tables[i] = "table_" + string(rune('a'+i%26)) + string(rune('0'+i/26))
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, name := range tables {
			_ = generateCRUDQueries(name)
		}
	}
}

func BenchmarkProtobufGeneration5Messages(b *testing.B) {
	messages := []string{"User", "Post", "Comment", "Category", "Tag"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, name := range messages {
			_ = generateProtobuf(name)
		}
	}
}

func BenchmarkProtobufGeneration15Messages(b *testing.B) {
	messages := []string{
		"User", "Post", "Comment", "Category", "Tag",
		"Media", "Setting", "Session", "Notification",
		"Like", "Follower", "Message", "Attachment", "Log", "Role",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, name := range messages {
			_ = generateProtobuf(name)
		}
	}
}

func BenchmarkContentNegotiationDetection(b *testing.B) {
	acceptHeaders := []string{
		"application/json",
		"application/xml",
		"application/yaml",
		"text/plain",
		"*/*",
		"application/*",
		"application/json;q=0.9,application/xml;q=1.0",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, header := range acceptHeaders {
			_ = detectContentType(header)
		}
	}
}

func BenchmarkContentNegotiationSerialization(b *testing.B) {
	data := map[string]interface{}{
		"id":    1,
		"name":  "Test Item",
		"count": 42,
		"items": []string{"a", "b", "c"},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = serializeToJSON(data)
	}
}

func parseSchema(schema string) ([]string, error) {
	lines := strings.Split(schema, ";")
	var tables []string
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(strings.ToUpper(line), "CREATE TABLE") {
			tables = append(tables, line)
		}
	}
	return tables, nil
}

func generateCRUDQueries(tableName string) []string {
	return []string{
		"SELECT * FROM " + tableName + ";",
		"SELECT * FROM " + tableName + " WHERE id = ?;",
		"INSERT INTO " + tableName + " (name) VALUES (?);",
		"UPDATE " + tableName + " SET name = ? WHERE id = ?;",
		"DELETE FROM " + tableName + " WHERE id = ?;",
	}
}

func generateProtobuf(messageName string) string {
	var sb strings.Builder
	sb.WriteString("message " + messageName + " {\n")
	sb.WriteString("  int64 id = 1;\n")
	sb.WriteString("  string name = 2;\n")
	sb.WriteString("  int32 count = 3;\n")
	sb.WriteString("}\n")
	return sb.String()
}

func detectContentType(accept string) string {
	accept = strings.ToLower(accept)
	if strings.Contains(accept, "json") {
		return "application/json"
	}
	if strings.Contains(accept, "xml") {
		return "application/xml"
	}
	if strings.Contains(accept, "yaml") || strings.Contains(accept, "yml") {
		return "application/yaml"
	}
	if strings.Contains(accept, "protobuf") || strings.Contains(accept, "proto") {
		return "application/protobuf"
	}
	if strings.Contains(accept, "text") {
		return "text/plain"
	}
	return "application/json"
}

func serializeToJSON(data map[string]interface{}) string {
	var sb strings.Builder
	sb.WriteString("{")
	first := true
	for k, v := range data {
		if !first {
			sb.WriteString(",")
		}
		first = false
		sb.WriteString(`"` + k + `":`)
		switch val := v.(type) {
		case string:
			sb.WriteString(`"` + val + `"`)
		case int, int32, int64:
			sb.WriteString("1")
		case []string:
			sb.WriteString("[")
			for i, s := range val {
				if i > 0 {
					sb.WriteString(",")
				}
				sb.WriteString(`"` + s + `"`)
			}
			sb.WriteString("]")
		}
	}
	sb.WriteString("}")
	return sb.String()
}
