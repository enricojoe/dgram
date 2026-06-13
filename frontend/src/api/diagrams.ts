import {
  useMutation,
  useQuery,
  useQueryClient,
} from '@tanstack/react-query'
import { api } from './client'
import type { Diagram, DiagramInput, DiagramSummary } from '@/types/auth'

// --- raw endpoint functions ---

async function listDiagrams(): Promise<DiagramSummary[]> {
  const { data } = await api.get<DiagramSummary[]>('/diagrams')
  return data
}

async function getDiagram(id: number): Promise<Diagram> {
  const { data } = await api.get<Diagram>(`/diagrams/${id}`)
  return data
}

async function createDiagram(input: DiagramInput): Promise<Diagram> {
  const { data } = await api.post<Diagram>('/diagrams', input)
  return data
}

async function updateDiagram(
  id: number,
  input: Partial<DiagramInput>,
): Promise<Diagram> {
  const { data } = await api.put<Diagram>(`/diagrams/${id}`, input)
  return data
}

async function deleteDiagram(id: number): Promise<void> {
  await api.delete(`/diagrams/${id}`)
}

// --- React Query hooks (server state lives here, not in components) ---

const keys = {
  all: ['diagrams'] as const,
  detail: (id: number) => ['diagrams', id] as const,
}

export function useDiagrams() {
  return useQuery({ queryKey: keys.all, queryFn: listDiagrams })
}

export function useDiagram(id: number | undefined) {
  return useQuery({
    queryKey: keys.detail(id ?? -1),
    queryFn: () => getDiagram(id as number),
    enabled: id != null,
  })
}

export function useCreateDiagram() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: createDiagram,
    onSuccess: () => qc.invalidateQueries({ queryKey: keys.all }),
  })
}

export function useUpdateDiagram() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: ({ id, input }: { id: number; input: Partial<DiagramInput> }) =>
      updateDiagram(id, input),
    onSuccess: (d) => {
      qc.invalidateQueries({ queryKey: keys.all })
      qc.invalidateQueries({ queryKey: keys.detail(d.id) })
    },
  })
}

export function useDeleteDiagram() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: deleteDiagram,
    onSuccess: () => qc.invalidateQueries({ queryKey: keys.all }),
  })
}
