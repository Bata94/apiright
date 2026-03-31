# Proto Extensions Guide

APIRight supports extending generated protobuf definitions through custom proto files. This allows you to add custom messages, services, and RPC methods that work alongside the auto-generated CRUD operations.

## How It Works

1. Create custom proto files in your project's `proto/` directory
2. Import the generated `db_ar_gen.proto` to reference your database entities
3. Run `apiright gen` - your proto extensions are automatically processed
4. Extensions are copied to `gen/proto/ext_*.proto` for compilation

## Basic Example

```protobuf
// proto/todo_extensions.proto
syntax = "proto3";
package extensions;

import "gen/proto/db_ar_gen.proto";

message TodoFilter {
  bool completed = 1;
  string priority = 2;
}

message BulkCreateRequest {
  repeated db.Todo todos = 1;
}

service TodoExtensions {
  rpc GetTodosByFilter(TodoFilter) returns (db.ListTodoResponse);
  rpc BulkCreateTodos(BulkCreateRequest) returns (db.CreateTodoResponse);
}
```

## Import Pattern

All proto extensions must import the generated database proto:

```protobuf
import "gen/proto/db_ar_gen.proto";
```

This gives you access to:
- `db.<TableName>` - Entity messages (Todo, User, Post, etc.)
- `db.<TableName>Request` - CRUD request types
- `db.<TableName>Response` - CRUD response types

## Custom Messages

Add custom message types for request/response bodies:

```protobuf
message Pagination {
  int32 page = 1;
  int32 page_size = 2;
}

message PaginatedResponse {
  repeated db.Todo items = 1;
  int32 total_count = 2;
  int32 page = 3;
  int32 page_size = 4;
}
```

## Custom Services

Define new service methods that use generated types:

```protobuf
service ReportingService {
  rpc GetTodoStats(StatsRequest) returns (StatsResponse);
  rpc ExportTodos(ExportRequest) returns (stream db.Todo);
}
```

## Plugin-Based Extensions

You can also provide proto extensions through plugins by implementing the `ProtoExtension` interface:

```go
type MyProtoExtension struct{}

func (m *MyProtoExtension) Name() string {
    return "my-proto-extension"
}

func (m *MyProtoExtension) ProtoFiles() []string {
    return []string{"./proto/my_extension.proto"}
}

func (m *MyProtoExtension) Imports() []string {
    return []string{"gen/proto/db_ar_gen.proto"}
}
```

## Generated Output

After running `apiright gen`:

```
gen/
└── proto/
    ├── db_ar_gen.proto      # Auto-generated
    ├── api_ar_gen.proto     # Auto-generated
    └── ext_*.proto          # Your extensions (copied here)
```

## Compilation

Proto extensions are copied to `gen/proto/` for compilation. You'll need to:

1. Run `apiright gen` to generate and copy extensions
2. Compile all protos with protoc:

```bash
protoc --go_out=. --go-grpc_out=. \
  -I . \
  gen/proto/*.proto
```

## Best Practices

1. **Use a separate package** - Keep extensions in their own `extensions` or `custom` package
2. **Import generated types** - Don't redefine types that already exist in `db_ar_gen.proto`
3. **Follow proto conventions** - Use consistent naming and numbering
4. **Document your extensions** - Add comments explaining custom services

## Limitations

- Proto extensions are processed but not automatically compiled
- You need protoc installed to compile the final protos
- Field numbers must not conflict with generated messages
