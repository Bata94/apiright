package core

import "strings"

var Version = "dev"

var sqlToGoTypeMap = map[string]string{
	"INTEGER":   "int64",
	"BIGINT":    "int64",
	"INT":       "int64",
	"SMALLINT":  "int32",
	"TINYINT":   "int32",
	"TEXT":      "string",
	"VARCHAR":   "string",
	"CHAR":      "string",
	"BOOLEAN":   "bool",
	"BOOL":      "bool",
	"REAL":      "float64",
	"DOUBLE":    "float64",
	"FLOAT":     "float64",
	"DECIMAL":   "float64",
	"DATETIME":  "time.Time",
	"TIMESTAMP": "time.Time",
	"DATE":      "time.Time",
	"TIME":      "time.Time",
	"BLOB":      "[]byte",
	"JSON":      "string",
}

var sqlToProtoTypeMap = map[string]string{
	"INTEGER":   "int64",
	"BIGINT":    "int64",
	"INT":       "int64",
	"SMALLINT":  "int32",
	"TINYINT":   "int32",
	"BOOLEAN":   "bool",
	"BOOL":      "bool",
	"REAL":      "double",
	"DOUBLE":    "double",
	"FLOAT":     "float",
	"DECIMAL":   "double",
	"DATETIME":  "google.protobuf.Timestamp",
	"TIMESTAMP": "google.protobuf.Timestamp",
	"DATE":      "google.protobuf.Timestamp",
	"TIME":      "google.protobuf.Timestamp",
	"BLOB":      "bytes",
	"JSON":      "string",
}

func SQLToGoType(sqlType string) string {
	if goType, exists := sqlToGoTypeMap[sqlType]; exists {
		return goType
	}
	return "string"
}

func SQLToProtoType(sqlType string) string {
	if protoType, exists := sqlToProtoTypeMap[sqlType]; exists {
		return protoType
	}
	return "string"
}

func SQLTypeToOpenAPI(sqlType string) string {
	sqlType = strings.ToLower(sqlType)

	switch {
	case strings.Contains(sqlType, "int"), strings.Contains(sqlType, "serial"):
		return "integer"
	case strings.Contains(sqlType, "float"), strings.Contains(sqlType, "decimal"), strings.Contains(sqlType, "numeric"), strings.Contains(sqlType, "real"), strings.Contains(sqlType, "double"):
		return "number"
	case strings.Contains(sqlType, "bool"):
		return "boolean"
	case strings.Contains(sqlType, "text"), strings.Contains(sqlType, "char"), strings.Contains(sqlType, "varchar"), strings.Contains(sqlType, "string"):
		return "string"
	case strings.Contains(sqlType, "date"), strings.Contains(sqlType, "time"), strings.Contains(sqlType, "timestamp"):
		return "string"
	case strings.Contains(sqlType, "json"):
		return "object"
	case strings.Contains(sqlType, "blob"), strings.Contains(sqlType, "binary"), strings.Contains(sqlType, "bytes"):
		return "string"
	default:
		return "string"
	}
}

func GetExampleValue(sqlType string) any {
	sqlType = strings.ToLower(sqlType)

	switch {
	case strings.Contains(sqlType, "int"):
		return 1
	case strings.Contains(sqlType, "float"), strings.Contains(sqlType, "decimal"):
		return 1.5
	case strings.Contains(sqlType, "bool"):
		return true
	case strings.Contains(sqlType, "text"), strings.Contains(sqlType, "char"), strings.Contains(sqlType, "varchar"):
		return "example"
	case strings.Contains(sqlType, "date"):
		return "2024-01-01"
	case strings.Contains(sqlType, "time"):
		return "12:00:00"
	case strings.Contains(sqlType, "timestamp"):
		return "2024-01-01T12:00:00Z"
	default:
		return ""
	}
}

func ToPascalCase(s string) string {
	parts := strings.Split(s, "_")
	for i, part := range parts {
		if len(part) > 0 {
			parts[i] = strings.ToUpper(part[:1]) + strings.ToLower(part[1:])
		}
	}
	return strings.Join(parts, "")
}
