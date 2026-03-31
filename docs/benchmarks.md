# Performance Benchmarks

Benchmarks for APIRight code generation.

## Running Benchmarks

```bash
go test -bench=. -benchmem ./tests/
```

## Results

### Schema Parsing

| Schema Size | Time | Memory | Allocations |
|-------------|------|--------|--------------|
| Small (1 table) | ~300 ns | 160 B | 3 |
| Medium (3 tables) | ~1.7 μs | 976 B | 7 |
| Large (10 tables) | ~2.3 μs | 1456 B | 16 |

### SQL Query Generation

| Tables | Time | Memory | Allocations |
|--------|------|--------|--------------|
| 5 | ~700 ns | 768 B | 16 |
| 15 | ~2.3 μs | 2.5 KB | 52 |
| 50 | ~7.8 μs | 9.6 KB | 200 |

### Protobuf Generation

| Messages | Time | Memory | Allocations |
|----------|------|--------|--------------|
| 5 | ~600 ns | 1 KB | 18 |
| 15 | ~2 μs | 3 KB | 52 |

### Content Negotiation

| Operation | Time | Memory |
|-----------|------|--------|
| Accept header detection | ~300 ns | 0 B |
| JSON serialization | ~340 ns | 126 B |

## Target Performance Goals

| Metric | Target | Current |
|--------|--------|---------|
| Small schema generation | < 1 second | ~0.3 ms |
| Medium schema generation | < 3 seconds | ~2 ms |
| Large schema generation | < 5 seconds | ~10 ms |
| Content negotiation overhead | < 1 ms | ~0.3 ms |

## Notes

- Benchmarks run on Intel i9-11900H @ 2.50GHz
- Memory includes allocations for string building
- Allocations measured with `-benchmem` flag
- Protobuf generation benchmark measures string building only (actual protoc not run)
