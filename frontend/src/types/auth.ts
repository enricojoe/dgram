// Auth + diagram DTOs — mirror the Go API contract (camelCase JSON).

import type { Dialect } from './schema'

export interface User {
  id: number
  email: string
  createdAt: string
}

export interface AuthResponse {
  user: User
  accessToken: string
  refreshToken: string
}

/** Per-table node position map, stored as jsonb on the backend. */
export type Layout = Record<string, { x: number; y: number }>

/** List view of a saved diagram (no ddl/layout). */
export interface DiagramSummary {
  id: number
  name: string
  dialect: Dialect
  createdAt: string
  updatedAt: string
}

/** Full saved diagram. */
export interface Diagram extends DiagramSummary {
  ddl: string
  layout: Layout
  shareId?: string | null
  isPublic: boolean
}

/** Public, read-only view returned by the share endpoint. */
export interface PublicDiagram {
  name: string
  dialect: Dialect
  ddl: string
  layout: Layout
}

export interface DiagramInput {
  name: string
  dialect: Dialect
  ddl: string
  layout: Layout
}
