# task-api-go

A small, dependency-free **REST API in Go** (standard library only): a task manager with validation, filtering,
proper HTTP status codes, a concurrency-safe in-memory store, and table-driven tests.

## Endpoints
| Method | Path | Description |
|---|---|---|
| GET | `/health` | Liveness check |
| GET | `/tasks?status=todo\|doing\|done` | List tasks (optional filter) |
| POST | `/tasks` | Create `{"title": "...", "status": "todo"}` (status optional) |
| GET | `/tasks/{id}` | Get one task |
| PUT | `/tasks/{id}` | Replace title and status |
| DELETE | `/tasks/{id}` | Delete (204) |

Errors are JSON: `{"error": "title is required"}` with 400 (invalid input), 404 (not found) or 405 (wrong method).

## Run
```bash
go run ./cmd/server            # listens on :8080 (set PORT to change)
curl -X POST localhost:8080/tasks -d '{"title":"learn go"}'
curl localhost:8080/tasks
```

## Test
```bash
go test ./... -race
```

## Structure
- `internal/store`: domain model, validation, thread-safe storage (`sync.RWMutex`)
- `internal/api`: HTTP handlers and routing (Go 1.22 `GET /tasks/{id}` patterns)
- `cmd/server`: entry point

## Ideas to extend
Persist to PostgreSQL, add pagination, JWT authentication, Dockerfile, OpenAPI spec.

> Built with AI assistance (Claude). I reviewed and run the code myself.
