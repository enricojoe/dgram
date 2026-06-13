# DGram

Parse SQL DDL into an interactive ER diagram (like dbdiagram.io). Supports
PostgreSQL and MySQL DDL, with bidirectional editing (DDL ↔ diagram).

Monorepo:

- **`/backend`** — Go + Gin + sqlx. The parsing engine and API.
- **`/frontend`** — React + TypeScript + Vite. The diagram editor.

Architecture & phased plan: see `tasks/todo.md` and `tasks/plans/`.

## Local development

### 1. Database
```bash
cd backend
docker compose up -d        # Postgres on :5432
```

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
npm run dev                  # http://localhost:5173 (proxies /api → :8080)
```
