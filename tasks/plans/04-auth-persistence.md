# Phase 4 — Auth + Persistence

Goal: user accounts (JWT) + save/load/list diagrams. Anonymous parse/generate
still works; saving requires login.

## API contract (camelCase JSON)
- `POST /api/auth/register` `{email,password}` → 201 `{user,accessToken,refreshToken}`
- `POST /api/auth/login` `{email,password}` → 200 `{user,accessToken,refreshToken}`
- `POST /api/auth/refresh` `{refreshToken}` → 200 `{accessToken,refreshToken}`
- `GET /api/me` (auth) → `{id,email,createdAt}`
- `GET /api/diagrams` (auth) → `[{id,name,dialect,createdAt,updatedAt}]`
- `POST /api/diagrams` (auth) `{name,dialect,ddl,layout}` → 201 full diagram
- `GET /api/diagrams/:id` (auth, owner) → full diagram `{id,name,dialect,ddl,layout,createdAt,updatedAt}`
- `PUT /api/diagrams/:id` (auth, owner) `{name?,dialect?,ddl?,layout?}` → full diagram
- `DELETE /api/diagrams/:id` (auth, owner) → 204

`user` = `{id,email,createdAt}`. `layout` = JSON object `{tableName:{x,y}}` (jsonb).
Tokens are stateless JWTs (access ~24h, refresh ~30d, `typ` claim distinguishes).

## Backend (subagent)
- [x] Migrations (golang-migrate, embedded iofs): `users`, `diagrams` (+ index)
- [x] `model/user.go`, `model/diagram.go`; `internal/db/db.go` (Connect+Migrate)
- [x] `util/password.go` (bcrypt), `util/jwt.go` (access 24h/refresh 30d, typ+sub)
- [x] `repository/user_repository.go`, `diagram_repository.go` (sqlx, owner-scoped)
- [x] `service/auth_service.go`, `diagram_service.go`
- [x] `controller/auth_controller.go`, `diagram_controller.go`
- [x] `middleware/auth.go` (RequireAuth + UserID helper)
- [x] `main.go`: connect DB (5434) fail-fast, run migrations, protected group

## Frontend (main thread)
- [x] `api/auth.ts`, `api/diagrams.ts` + React Query hooks
- [x] `features/auth/store/authStore.ts` (persist), `client.ts` 401→refresh→retry
- [x] `features/auth` AuthForm; `pages/LoginPage`, `RegisterPage`
- [x] `routes`: ProtectedRoute; `/` `/login` `/register` `/dashboard` `/d/:id`
- [x] `pages/DashboardPage` list (open/delete/new)
- [x] `components/AppShell` header; EditorPage toolbar Save (load/save incl layout)

## Result
Phase 4 complete (2026-06-13). Backend via subagent, independently verified
(build/vet/test pass; tables users/diagrams/schema_migrations created). Frontend
builds + type-checks. Full E2E through the Vite proxy — register, login, /me
auth (401 without token), create-with-layout, get, list, update, **cross-user
404 (no leak)**, refresh, delete — ALL PASS. Persistence real (Postgres :5434).
Contract met exactly; password hash never serialized; stateless JWT tokens.
