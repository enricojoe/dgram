# Phase 0 — Scaffolding

Goal: both apps exist, build, and run; Postgres available locally.

## Backend
- [x] Directory layout (`cmd/`, `internal/{config,controller,service,repository,model,middleware,util,parser}`)
- [x] `go.mod` + deps: gin, sqlx, lib/pq, godotenv
- [x] `internal/config/config.go` — env loading
- [x] `internal/util/response.go` — JSON response helpers
- [x] `internal/controller/health_controller.go` — `/health`
- [x] `internal/middleware/cors.go` — CORS for frontend
- [x] `cmd/server/main.go` — wire config + router, start server
- [x] `docker-compose.yml` — Postgres 16 (host port **5434**)
- [x] `.env.example`
- [x] `go build ./...` passes; server runs and `/api/health` returns 200

## Frontend
- [x] Vite + React + TypeScript scaffold in `/frontend`
- [x] Install: @xyflow/react, @dagrejs/dagre, @uiw/react-codemirror,
      @codemirror/lang-sql, @tanstack/react-query, zustand, react-router-dom,
      axios, tailwindcss v4 (@tailwindcss/vite)
- [x] Base folder layout (`api/ components/ features/ pages/ lib/ types/ routes/`)
- [x] `@` path alias, Vite `/api` → `:8080` proxy, React Query + Router wired
- [x] `npm run build` passes

## Result
Phase 0 complete (2026-06-13). Backend builds + `/api/health` → 200. Frontend
type-checks and bundles. Postgres 16 healthy on host port **5434** (5432/5433
were occupied by the user's other DBs). Ready for Phase 1.
