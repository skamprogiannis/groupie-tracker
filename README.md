# Groupie Tracker

Groupie Tracker is a Zone01 Go web project that will fetch the public Groupie Tracker API and display artist, member, album, location, date, and relation data in a user-friendly website.

Current `main` contains the M0 project skeleton only. The remaining milestones are tracked in Gitea issues for the team to implement.

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

- `/` - skeleton home response
- `/healthz` - health check

## Documentation

- `docs/PRD.md` - product requirements and audit scope
- `docs/architecture.md` - package, routing, and async design
- `docs/instructions.txt` - original subject instructions
- `docs/audit-questions.txt` - audit checklist
- `docs/llm-log.txt` - LLM usage log
