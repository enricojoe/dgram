// TypeScript mirror of the Go `model.Schema` contract
// (backend/internal/model/schema.go). Keep these in sync — they are the shared
// shape the /api/parse endpoint speaks.

export type Dialect = 'postgres' | 'mysql'

export interface Column {
  name: string
  type: string
  pk?: boolean
  notNull?: boolean
  unique?: boolean
  autoInc?: boolean
  default?: string
  note?: string
}

export interface Index {
  name?: string
  columns: string[]
  unique?: boolean
}

export interface Table {
  schema?: string
  name: string
  columns: Column[]
  indexes?: Index[]
  note?: string
}

export interface Ref {
  fromTable: string
  fromColumn: string
  toTable: string
  toColumn: string
  onDelete?: string
  onUpdate?: string
}

export interface Enum {
  name: string
  values: string[]
}

export interface Schema {
  tables: Table[]
  refs: Ref[]
  enums: Enum[]
}

export interface ParseResponse {
  schema: Schema
  warnings: string[]
}

export const emptySchema: Schema = { tables: [], refs: [], enums: [] }
