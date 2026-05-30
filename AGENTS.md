# AI Agent Instructions - groupie-tracker

## Project Context

- Build the Zone01 Groupie Tracker web project in Go.
- Read `docs/PRD.md` for requirements and `docs/architecture.md` for design before implementation.
- Keep `docs/instructions.txt` and `docs/audit-questions.txt` as the source subject and audit references.

## Technical Rules

- Use only Go standard library packages.
- Keep HTTP handlers thin and put business logic in testable internal packages.
- The server must not crash on bad input, unknown routes, upstream errors, or template failures.
- The app must run with `go run .` and respect `PORT` when set.

## Workflow

- Run `go test ./...` before considering work complete.
- Run `gofmt` on changed Go files.
- Update `docs/llm-log.txt` for every LLM-assisted session, including the prompt/task, model or agent when known, copied versus edited content, and affected files.
- Do not add AI co-author trailers to commits.
