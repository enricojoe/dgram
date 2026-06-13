# DGram — Frontend

React + TypeScript + Vite client for [DGram](../), a dbdiagram.io-style app:
paste SQL DDL → interactive ER diagram, edit either side, save & share.

Built with React Flow + dagre (diagram), CodeMirror (editor), React Query
(server state), and Zustand (auth/UI state).

## Prerequisites

- Node.js 22+
- A running DGram backend (see [`../backend`](../backend)) — or use the
  full Docker stack ([`../docker-compose.yml`](../docker-compose.yml)).

## Getting started

```bash
npm install
npm run dev
```

The dev server starts on http://localhost:5400 (set by `FRONTEND_PORT` in
`.env`; override as shown below). API calls to `/api` are proxied to the backend.

## Scripts

| Command           | Description                                  |
| ----------------- | -------------------------------------------- |
| `npm run dev`     | Start the Vite dev server with HMR.          |
| `npm run build`   | Type-check (`tsc -b`) and build to `dist/`.  |
| `npm run preview` | Serve the production build locally.          |
| `npm run lint`    | Run ESLint.                                  |

## Configuration

Environment variables are loaded by Vite from `.env` (committed defaults) and
`.env.local` (gitignored, per-developer overrides), or from the shell.

| Variable                | `.env` default          | Purpose                                                              |
| ----------------------- | ----------------------- | ------------------------------------------------------------------- |
| `FRONTEND_PORT`         | `5400`                  | Port the dev server (and the Docker nginx container) listens on.    |
| `VITE_API_BASE_URL`     | `/api`                  | Base URL the client uses for API requests (baked in at build time). |
| `VITE_API_PROXY_TARGET` | `http://localhost:8090` | Backend the dev server proxies `/api` to.                           |

Only `VITE_`-prefixed variables are exposed to client code; `FRONTEND_PORT` is
build/dev-server config only.

```bash
# One-off override
FRONTEND_PORT=3000 npm run dev

# Or persist it in .env.local
echo "FRONTEND_PORT=3000" >> .env.local
```

## Docker

The frontend ships as an nginx image that serves the built SPA and reverse-proxies
`/api` to the backend. nginx's listen port is templated from `FRONTEND_PORT` at
container start. Run it as part of the full stack from the repo root:

```bash
cd .. && docker compose up
```

Then visit http://localhost:5173. Ports for every service are env-driven via the
root `.env` (copy from [`../.env.example`](../.env.example)) — `FRONTEND_PORT`,
`BACKEND_PORT`, and `POSTGRES_PORT`.

## Project structure

Feature-based layout under `src/`:

```
src/
  features/
    auth/       # Login/register, JWT token store
    diagrams/   # Saved diagrams, sharing
    editor/     # DDL editor + ER diagram canvas
```

See the [project root](../) for backend, architecture, and the overall stack.
