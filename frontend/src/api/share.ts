import { api } from './client'
import type { PublicDiagram } from '@/types/auth'

/** Make a diagram public; returns its share token. */
export async function enableShare(id: number): Promise<{ shareId: string }> {
  const { data } = await api.post<{ shareId: string; isPublic: boolean }>(
    `/diagrams/${id}/share`,
  )
  return { shareId: data.shareId }
}

/** Make a diagram private again. */
export async function disableShare(id: number): Promise<void> {
  await api.delete(`/diagrams/${id}/share`)
}

/** Fetch a publicly shared diagram (no auth). */
export async function getShared(shareId: string): Promise<PublicDiagram> {
  const { data } = await api.get<PublicDiagram>(`/share/${shareId}`)
  return data
}
