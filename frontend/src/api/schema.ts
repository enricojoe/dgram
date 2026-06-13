import { api } from './client'
import type { Dialect, ParseResponse, Schema } from '@/types/schema'

/**
 * parseDDL sends DDL text to the Go backend and returns the normalized schema
 * plus any non-fatal parse warnings. This is the `api/` layer's job: it knows
 * about HTTP/endpoints so the rest of the app never does.
 */
export async function parseDDL(
  dialect: Dialect,
  ddl: string,
): Promise<ParseResponse> {
  const { data } = await api.post<ParseResponse>('/parse', { dialect, ddl })
  return data
}

/**
 * generateDDL sends a schema to the backend and returns dialect-correct DDL.
 * This powers the diagram → DDL direction of bidirectional editing.
 */
export async function generateDDL(
  dialect: Dialect,
  schema: Schema,
): Promise<string> {
  const { data } = await api.post<{ ddl: string }>('/generate', {
    dialect,
    schema,
  })
  return data.ddl
}
