# 06 — Frontend modernization + profile feature

Full spec: `~/.claude/plans/so-here-s-what-i-mighty-rocket.md`

## Scope
1. Branded favicon + page title.
2. Dark mode (class-based, persisted, OS default).
3. Header avatar dropdown (My Diagrams / Profile / dark toggle / Log out).
4. Hamburger (☰) collapse toggle for the DDL editor pane.
5. Profile feature — **required backend work** (User was read-only).
6. Dashboard redesigned as a card grid.

## Backend (profile capability)
- [x] Migration `000003_add_user_display_name` — `display_name TEXT NOT NULL DEFAULT ''`.
- [x] `model.User`/`UserView` + `View()` carry `DisplayName`.
- [x] Repo: `display_name` added to Create/GetByEmail/GetByID; new `UpdateDisplayName`, `UpdatePasswordHash`.
- [x] Service: `UpdateDisplayName`, `UpdatePassword` (verifies current password, reuses `ErrInvalidCredentials`).
- [x] Controller: `PATCH /api/me`, `POST /api/me/password`.

## Frontend
- [x] `public/favicon.svg` (ERD glyph) + `index.html` title/description.
- [x] `features/theme/`: `themeStore` (zustand persist, OS default), `useApplyTheme`, `ThemeToggle`.
- [x] `index.css`: `@custom-variant dark` + body surface defaults; existing button-pointer base rule kept.
- [x] `components/UserMenu.tsx` + `AppShell` rewrite (avatar dropdown, click-out/Esc, dark variants).
- [x] `components/icons.tsx` (`BurgerIcon`); `EditorPage` + `DdlEditor` use it; collapse hides pane, toolbar ☰ restores; CodeMirror `theme` follows store.
- [x] `pages/ProfilePage.tsx` + `/profile` protected route; `api/auth` `updateProfile`/`changePassword`; `useProfile` hook; `authStore.updateUser`; `identity.ts` helpers.
- [x] `pages/DashboardPage.tsx` card grid + dark; `AuthForm` dark variants.

## Results / verification
- `go build ./...` clean; `go vet` clean.
- Migration applied: `users.display_name` present, `schema_migrations` version=3, dirty=f (verified via psql).
- HTTP smoke (built binary on a free port): register→displayName ""; PATCH /me→"Jane Doe"; change-password wrong-old→401, correct→200; login old→401, new→200 with persisted displayName.
- Frontend `tsc -b` + `vite build` succeed (ProfilePage chunk emitted).
- Manual UI checks (dark toggle, dropdown, ☰ collapse, dashboard cards) left for the user in `npm run dev`.

## Notes
- Dev machine has port clutter (8090 occupied by a running instance; sandbox also blocks `listen`). Smoke test ran the compiled binary on a verified-free port with the sandbox lifted.
- React Flow canvas internals were not dark-themed (out of scope); surrounding chrome flips correctly.
