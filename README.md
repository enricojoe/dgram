# DGram

Parse SQL DDL into an interactive ER diagram (like dbdiagram.io). Supports
PostgreSQL and MySQL DDL, with bidirectional editing (DDL ↔ diagram).

Monorepo:

- **`/backend`** — Go + Gin + sqlx. The parsing engine and API.
- **`/frontend`** — React + TypeScript + Vite. The diagram editor.

Architecture & phased plan: see `tasks/todo.md` and `tasks/plans/`.

## Quick start (Docker)

Run the whole stack — Postgres, backend, and frontend — from the repo root:

```bash
cp .env.example .env         # adjust ports/secrets as needed
docker compose up
```

Then open http://localhost:5173. All published ports are env-driven via `.env`
(`FRONTEND_PORT`, `BACKEND_PORT`, `POSTGRES_PORT`); see `.env.example` for the
full list of variables and defaults.

## Local development

Run each piece directly for the fastest feedback loop.

### 1. Database
```bash
cd backend
docker compose up -d         # Postgres on host :5434 (→ container 5432)
```

> Postgres is mapped to host port **5434** (5432/5433 are used by other local
> DBs). The backend's default `DATABASE_URL` already points at 5434.

### 2. Backend
```bash
cd backend
cp .env.example .env
go run ./cmd/server          # http://localhost:8080
```

### 3. Frontend
```bash
cd frontend
npm install
npm run dev                  # http://localhost:5173 (proxies /api → backend)
```

Ports and other settings are configurable via environment variables — see each
package's `.env.example` / README.
