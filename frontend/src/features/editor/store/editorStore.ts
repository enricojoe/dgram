import { create } from 'zustand'
import type { Dialect, Schema } from '@/types/schema'
import { emptySchema } from '@/types/schema'

const SAMPLE_DDL = `CREATE TABLE users (
  id serial PRIMARY KEY,
  email varchar(255) UNIQUE NOT NULL,
  name varchar(100)
);

CREATE TABLE posts (
  id serial PRIMARY KEY,
  user_id int NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  title varchar(200) NOT NULL,
  body text
);

CREATE TABLE comments (
  id serial PRIMARY KEY,
  post_id int NOT NULL REFERENCES posts(id) ON DELETE CASCADE,
  author_id int REFERENCES users(id),
  body text NOT NULL
);`

/** Which side last drove a change — used to break the DDL↔diagram sync loop. */
export type EditOrigin = 'ddl' | 'diagram'

export interface XY {
  x: number
  y: number
}

/**
 * The editor's single source of truth.
 *
 * Bidirectional sync is coordinated by `origin`:
 *  - User types in the DDL pane → origin 'ddl' → useSchemaSync parses → schema.
 *  - User edits the diagram     → origin 'diagram' → useDdlSync generates → ddl.
 * Each sync hook only fires for its own origin, so neither overwrites the other.
 *
 * `layout` remembers per-table node positions so user drags survive re-layout.
 */
interface EditorState {
  dialect: Dialect
  ddl: string
  schema: Schema
  warnings: string[]
  error: string | null
  parsing: boolean
  generating: boolean
  origin: EditOrigin
  layout: Record<string, XY>

  setDialect: (d: Dialect) => void
  setDdl: (ddl: string) => void
  setParseResult: (schema: Schema, warnings: string[]) => void
  setError: (error: string | null) => void
  setParsing: (parsing: boolean) => void
  setGenerating: (generating: boolean) => void
  /** Set DDL produced by /generate; keeps origin 'diagram' so it isn't re-parsed. */
  setGeneratedDdl: (ddl: string) => void
  /** Apply a diagram-driven schema edit; flips origin to 'diagram' to trigger regen. */
  applySchemaEdit: (updater: (schema: Schema) => Schema) => void
  /** Persist a dragged node position (does not touch schema or DDL). */
  setNodePosition: (table: string, pos: XY) => void
  /** Load a saved diagram into the workspace (origin 'ddl' → re-parses). */
  loadDiagram: (input: {
    ddl: string
    dialect: Dialect
    layout: Record<string, XY>
  }) => void
}

export const useEditorStore = create<EditorState>((set) => ({
  dialect: 'postgres',
  ddl: SAMPLE_DDL,
  schema: emptySchema,
  warnings: [],
  error: null,
  parsing: false,
  generating: false,
  origin: 'ddl',
  layout: {},

  // Switching dialect re-parses the current DDL.
  setDialect: (dialect) => set({ dialect, origin: 'ddl' }),
  setDdl: (ddl) => set({ ddl, origin: 'ddl' }),
  setParseResult: (schema, warnings) => set({ schema, warnings, error: null }),
  setError: (error) => set({ error }),
  setParsing: (parsing) => set({ parsing }),
  setGenerating: (generating) => set({ generating }),
  setGeneratedDdl: (ddl) => set({ ddl }),
  applySchemaEdit: (updater) =>
    set((state) => ({ schema: updater(state.schema), origin: 'diagram' })),
  setNodePosition: (table, pos) =>
    set((state) => ({ layout: { ...state.layout, [table]: pos } })),
  loadDiagram: ({ ddl, dialect, layout }) =>
    set({ ddl, dialect, layout: layout ?? {}, origin: 'ddl', error: null, warnings: [] }),
}))
