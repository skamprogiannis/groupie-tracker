# Groupie Tracker

Groupie Tracker is a Zone01 Go web project that will fetch the public Groupie Tracker API and display artist, member, album, location, date, and relation data in a user-friendly website.

The current implementation includes the data and server foundation: API fetching, normalized in-memory catalog data, basic route rendering, search JSON, static assets, health checks, and shared error handling.

## Run

```bash
go run .
```

The server listens on `http://localhost:8080` by default. Set `PORT` to use another port:

```bash
PORT=3000 go run .
```

## Test

```bash
go test ./...
gofmt -w .
```

The project uses only Go standard library packages.

## Routes

- `/` - catalog home page
- `/artist/{id}` - artist detail page
- `/api/search?q=...` - JSON search endpoint
- `/healthz` - health check
- `/static/*` - embedded static assets

## Documentation

- `docs/PRD.md` - product requirements and audit scope
- `docs/architecture.md` - package, routing, and async design
- `docs/instructions.txt` - original subject instructions
- `docs/audit-questions.txt` - audit checklist
- `docs/llm-log.txt` - LLM usage log
