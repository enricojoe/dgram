# Phase 3 — Bidirectional Editing

Goal: edit the diagram → regenerate DDL; preserve drag positions across edits.

## Backend (subagent)
- [x] `internal/parser/generator/postgres.go` — Schema → PG DDL
- [x] `internal/parser/generator/mysql.go` — Schema → MySQL DDL
- [x] `internal/parser/generator/generator.go` — dialect dispatcher
- [x] `SchemaService.Generate()` + `POST /api/generate`
- [x] Round-trip tests: parse → generate → parse ≈ equal

## Frontend (main thread)
- [x] `api/schema.ts` — add `generateDDL(dialect, schema)`
- [x] Store: `origin` guard ('ddl' | 'diagram'), `layout` map (table→position),
      `applySchemaEdit()` mutator, `setNodePosition()`
- [x] `useDdlSync` hook — on diagram edits, debounce → /generate → set DDL
      (guarded so it never loops with useSchemaSync)
- [x] Drag persistence (`lib/schemaFlow.ts` layoutWithStored + onNodeDragStop)
- [x] `DiagramCanvas` onConnect → add FK ref to schema; "+ Add table" panel
- [x] Editable `TableNode` (`lib/schemaEdits.ts` pure transforms): rename table,
      add/delete column, edit column name/type, toggle PK, delete table

## Sync-loop design
`origin` field marks the last edit source. `useSchemaSync` parses only when
`origin==='ddl'`; `useDdlSync` generates only when `origin==='diagram'`.
Generate sets DDL without flipping origin to 'ddl', so no ping-pong.

## Result
Phase 3 complete (2026-06-13). Backend generator built via subagent and
independently verified: `go build`/`vet`/`test` pass (generator round-trip tests
for both dialects). Frontend bidirectional editing built in main thread,
type-checks + builds. End-to-end through the Vite proxy:
`POST /api/generate` returns correct DDL, and a full **parse→generate→parse
round-trip yields identical tables + refs for BOTH postgres and mysql** (proves
the sync loop is stable). Diagram editing (rename/add/delete/PK/draw-FK) flows
through the store → regenerates DDL; drag positions persist across re-layout.

### Known limitation (flagged for UI)
A standalone enum not attached to any column survives in Postgres (CREATE TYPE)
but cannot be represented in MySQL. Enum→column type linkage is still lost on
the PG parse side (carried over from Phase 1).
