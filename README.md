# Groupie Tracker

Groupie Tracker is a Zone01 Go web project that fetches the public Groupie Tracker API and displays artist, member, album, location, date, and relation data in a user-friendly website.

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

- `/` - artist catalog and client-server search event
- `/artist/{id}` - artist detail page
- `/api/search?q=queen` - JSON search endpoint
- `/healthz` - deployment health check

## Documentation

- `docs/PRD.md` - product requirements and audit scope
- `docs/architecture.md` - package, routing, and async design
- `docs/instructions.txt` - original subject instructions
- `docs/audit-questions.txt` - audit checklist
- `docs/llm-log.txt` - LLM usage log
- `docs/issues.md` - tracker-ready milestones and issue bodies
