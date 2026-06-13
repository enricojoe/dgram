# Phase 1 — Parser Engine (backend)

Goal: `POST /api/parse` turns PG & MySQL DDL into the normalized `model.Schema`.

## Tasks
- [x] Freeze `model.Schema` contract (`internal/model/schema.go`)
- [x] Resolve pure-Go parser libs (auxten/postgresql-parser, pingcap/tidb/pkg/parser)
- [x] `internal/parser/parser.go` — dialect dispatcher + warnings
- [x] `internal/parser/dialect/postgres.go` — PG AST → Schema
- [x] `internal/parser/dialect/mysql.go` — MySQL AST → Schema
- [x] `internal/service/schema_service.go` — thin service (room for Generate)
- [x] `internal/controller/parse_controller.go` — `POST /api/parse`
- [x] Register controller in `cmd/server/main.go`
- [x] Table-driven tests for both dialects (FK + ON DELETE, PK, unique, enum)
- [x] Verify: build, vet, `go test`, live curl both dialects

## Known limitations (revisit in a later phase)
- **PG enums**: `CREATE TYPE ... AS ENUM` is extracted via regexp (parser lib
  can't handle it). Enum *values* are captured, but a column typed as an enum
  is rewritten to `text` — the column→enum linkage is lost. MySQL inline
  `ENUM(...)` keeps the link (hoisted to `<table>_<column>`).
- **PG serial/bigserial** both normalize to `serial` + `AutoInc` (CRDB renders
  both as INT8, indistinguishable post-parse).
- `substitutePostgresEnumUsages` word-replaces enum names globally — possible
  false positive if an enum name equals a table/column name.

## Result
Phase 1 complete (2026-06-13), implemented via subagent and independently
verified by orchestrator: `go build`/`go vet`/`go test` all pass (5 parser
tests), live `/api/parse` returns correct schema + FK refs + enums for both
dialects; bad input → 400, fatal parse → 422. Code is idiomatic, tolerant
parsing with per-statement fallback. `model.Schema` contract unchanged.
