# Phase 2 — Frontend Render

Goal: paste/edit DDL → live interactive ER diagram.

## Tasks
- [x] `types/schema.ts` — TS mirror of Go `model.Schema`
- [x] `api/schema.ts` — typed `parseDDL()` call
- [x] `features/editor/store/editorStore.ts` — Zustand single source of truth
- [x] `lib/schemaFlow.ts` — schema→nodes/edges + dagre layout (pure)
- [x] `features/editor/hooks/useSchemaSync.ts` — debounced parse, stale-guard
- [x] `features/editor/components/TableNode.tsx` — custom node, per-column handles
- [x] `features/editor/components/DdlEditor.tsx` — CodeMirror + dialect select
- [x] `features/editor/components/DiagramCanvas.tsx` — React Flow + minimap
- [x] `pages/EditorPage.tsx` — split layout, activates sync
- [x] React Flow CSS imported in `main.tsx`
- [x] Verify: `npm run build`, both servers up, proxied `/api/parse` returns refs

## Notes / deferred
- Re-layout runs on every parse; user drags are NOT yet persisted across edits
  (→ Phase 3) and not saved to backend (→ Phase 4).
- Used a debounced effect for parse (not React Query) — parse is derived from
  live input, not cached server state. React Query reserved for diagram CRUD.
- Bundle >500kB warning (parser/codemirror/reactflow) — code-split in Phase 5.

## Result
Phase 2 complete (2026-06-13). Frontend builds + type-checks. Backend (:8080)
and Vite (:5173) run together; SPA serves 200; proxied POST /api/parse returns
correct schema with FK refs. Editing DDL re-parses (500ms debounce) and the
diagram reflects tables (with PK/type/NN/U markers) and FK edges with dagre
auto-layout. Visual confirmation pending user view at http://localhost:5173.
