import { useEffect } from 'react'
import { useParams } from 'react-router-dom'
import { useQuery } from '@tanstack/react-query'
import { getShared } from '@/api/share'
import { useEditorStore } from '@/features/editor/store/editorStore'
import { useSchemaSync } from '@/features/editor/hooks/useSchemaSync'
import DiagramCanvas from '@/features/editor/components/DiagramCanvas'

/**
 * SharePage renders a publicly shared diagram read-only. It loads the diagram
 * into the editor store and reuses the (read-only) canvas; useSchemaSync parses
 * the DDL into the schema the canvas draws.
 */
export default function SharePage() {
  useSchemaSync()
  const { shareId } = useParams()
  const loadDiagram = useEditorStore((s) => s.loadDiagram)

  const { data, isLoading, error } = useQuery({
    queryKey: ['share', shareId],
    queryFn: () => getShared(shareId as string),
    enabled: !!shareId,
  })

  useEffect(() => {
    if (data) {
      loadDiagram({
        ddl: data.ddl,
        dialect: data.dialect,
        layout: data.layout ?? {},
      })
    }
  }, [data, loadDiagram])

  if (isLoading) return <p className="p-6 text-sm text-gray-500">Loading…</p>
  if (error || !data)
    return (
      <p className="p-6 text-sm text-red-600">
        This diagram is not available.
      </p>
    )

  return (
    <div className="flex h-full flex-col">
      <div className="border-b border-gray-200 px-3 py-2 text-sm font-medium text-gray-700">
        {data.name} <span className="text-gray-400">· read-only</span>
      </div>
      <div className="min-h-0 flex-1">
        <DiagramCanvas readOnly />
      </div>
    </div>
  )
}
