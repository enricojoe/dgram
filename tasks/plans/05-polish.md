# Phase 5 — Polish

Goal: export, public share links, code-splitting, read-only mode.

## Share API contract (camelCase)
- `POST /api/diagrams/:id/share` (auth, owner) → 200 `{shareId, isPublic:true}`
  (idempotent: creates shareId if absent, sets is_public=true)
- `DELETE /api/diagrams/:id/share` (auth, owner) → 200 `{isPublic:false}`
- `GET /api/share/:shareId` (PUBLIC, no auth) → 200 `{name,dialect,ddl,layout}`
  or 404 if not public/unknown
- Full diagram GET/POST/PUT responses now also include `shareId` (nullable) and
  `isPublic`.

## Backend (subagent)
- [x] Migration `000002_share`: `share_id TEXT UNIQUE`, `is_public BOOL`
- [x] model/repo/service/controller for share enable/disable + public get
- [x] Route wiring (public `GET /api/share/:shareId`)
- [x] Verified: enable→public get; disable→404; non-owner→404; cross-user→404

## Frontend (main thread)
- [x] `features/editor/editorMode.tsx`: `readOnly` context
- [x] DiagramCanvas honors readOnly + Export PNG (html-to-image) control
- [x] TableNode hides edit controls when readOnly (EditableText disabled)
- [x] Export SQL (download .sql) in editor toolbar
- [x] Share button (saved diagrams): enable → copies `/s/:shareId` link
- [x] `pages/SharePage` at `/s/:shareId` (public, read-only canvas)
- [x] `api/share.ts`: enableShare/disableShare/getShared
- [x] Code-split routes with React.lazy + Suspense

## Result
Phase 5 complete (2026-06-13). Backend share via subagent, verified. Frontend
builds + type-checks; code-splitting confirmed (editor chunk 447kB split out;
initial load no longer ships it; >500kB warning gone). Share E2E through proxy
ALL PASS (share → public get no-auth → diagram includes shareId/isPublic →
unshare → 404 → cross-user 404). Export SQL/PNG wired; read-only share view
reuses the canvas via EditorMode context. Final regression: backend
build/vet/test green; both servers healthy (200).
